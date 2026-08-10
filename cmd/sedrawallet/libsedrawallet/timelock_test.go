package libsedrawallet

import (
	"testing"

	"github.com/sedracoin/go-secp256k1"
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/consensushashing"
	"github.com/sedracoin/sedrad/domain/consensus/utils/constants"
	"github.com/sedracoin/sedrad/domain/consensus/utils/txscript"
	"github.com/sedracoin/sedrad/domain/consensus/utils/utxo"
	"github.com/sedracoin/sedrad/util"
)

func newSchnorrKey(t *testing.T) (*secp256k1.SchnorrKeyPair, []byte) {
	t.Helper()
	key, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		t.Fatalf("GenerateSchnorrKeyPair: %v", err)
	}
	pubKey, err := key.SchnorrPublicKey()
	if err != nil {
		t.Fatalf("SchnorrPublicKey: %v", err)
	}
	serialized, err := pubKey.Serialize()
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	return key, serialized[:]
}

// claimResult runs a claim of `tranche` with a specific spending key and LockTime through
// the real consensus script engine and returns whether it is accepted. Using an explicit
// lockTime lets us test the before/at/after-unlock boundary directly.
func claimResult(t *testing.T, tranche *Tranche, spendKey *secp256k1.SchnorrKeyPair, lockTime, sequence uint64) bool {
	t.Helper()
	tx := &externalapi.DomainTransaction{
		Version: 0,
		Inputs: []*externalapi.DomainTransactionInput{{
			PreviousOutpoint: externalapi.DomainOutpoint{Index: 0},
			Sequence:         sequence,
			UTXOEntry:        utxo.NewUTXOEntry(tranche.Amount, tranche.ScriptPublicKey, false, 0),
		}},
		Outputs:  []*externalapi.DomainTransactionOutput{{Value: tranche.Amount - 1000, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x51}, Version: 0}}},
		LockTime: lockTime,
	}
	signature, err := txscript.RawTxInSignature(tx, 0, consensushashing.SigHashAll, spendKey, &consensushashing.SighashReusedValues{})
	if err != nil {
		t.Fatalf("RawTxInSignature: %v", err)
	}
	sigScript, err := txscript.NewScriptBuilder().AddData(signature).AddData(tranche.RedeemScript).Script()
	if err != nil {
		t.Fatalf("build sigScript: %v", err)
	}
	tx.Inputs[0].SignatureScript = sigScript

	engine, err := txscript.NewEngine(tranche.ScriptPublicKey, tx, 0, txscript.ScriptNoFlags, nil, nil, &consensushashing.SighashReusedValues{})
	if err != nil {
		return false
	}
	return engine.Execute() == nil
}

func TestBuildTrancheSchedule(t *testing.T) {
	_, pubKey := newSchnorrKey(t)
	const (
		total       = uint64(2_000_000_000_00000000) // 2B SDR in seep
		start       = uint64(53_625_513)
		periodBlock = uint64(4 * 365 * 24 * 60 * 60) // ~4 years of blocks
		numTranches = 16
	)
	tranches, err := BuildTrancheSchedule(pubKey, total, start, periodBlock, numTranches)
	if err != nil {
		t.Fatalf("BuildTrancheSchedule: %v", err)
	}
	if len(tranches) != numTranches {
		t.Fatalf("got %d tranches, want %d", len(tranches), numTranches)
	}
	var sum, prevUnlock uint64
	step := periodBlock / numTranches
	for i, tranche := range tranches {
		sum += tranche.Amount
		if tranche.UnlockDaaScore <= prevUnlock {
			t.Errorf("tranche %d unlock %d not strictly after previous %d", i, tranche.UnlockDaaScore, prevUnlock)
		}
		wantUnlock := start + step*uint64(i+1)
		if tranche.UnlockDaaScore != wantUnlock {
			t.Errorf("tranche %d unlock = %d, want %d", i, tranche.UnlockDaaScore, wantUnlock)
		}
		prevUnlock = tranche.UnlockDaaScore
		if len(tranche.ScriptPublicKey.Script) == 0 {
			t.Errorf("tranche %d has empty scriptPublicKey", i)
		}
	}
	if sum != total {
		t.Errorf("tranche amounts sum to %d, want exactly %d", sum, total)
	}
}

func TestDistributionTrancheCLTVGating(t *testing.T) {
	recipientKey, recipientPub := newSchnorrKey(t)
	strangerKey, _ := newSchnorrKey(t)

	tranches, err := BuildTrancheSchedule(recipientPub, 4_000_000, 1_000_000, 40_000, 4)
	if err != nil {
		t.Fatalf("BuildTrancheSchedule: %v", err)
	}

	for _, tranche := range tranches {
		// Before unlock: a spend with LockTime < unlock must be rejected.
		if claimResult(t, tranche, recipientKey, tranche.UnlockDaaScore-1, 0) {
			t.Errorf("tranche %d: spend BEFORE unlock (locktime=%d) was accepted, want rejected", tranche.Index, tranche.UnlockDaaScore-1)
		}
		// At unlock: recipient spend must be accepted.
		if !claimResult(t, tranche, recipientKey, tranche.UnlockDaaScore, 0) {
			t.Errorf("tranche %d: recipient spend AT unlock was rejected, want accepted", tranche.Index)
		}
		// After unlock but by the wrong key: must be rejected.
		if claimResult(t, tranche, strangerKey, tranche.UnlockDaaScore, 0) {
			t.Errorf("tranche %d: spend by WRONG key was accepted, want rejected", tranche.Index)
		}
		// A finalized input (sequence == max) bypasses CLTV and must be rejected.
		if claimResult(t, tranche, recipientKey, tranche.UnlockDaaScore, constants.MaxTxInSequenceNum) {
			t.Errorf("tranche %d: finalized-input spend was accepted, want rejected", tranche.Index)
		}
	}
}

func TestBuildTrancheClaimTransaction(t *testing.T) {
	recipientKey, recipientPub := newSchnorrKey(t)
	tranche, err := SelfLock(recipientPub, 1_000_000, 500, 10_000)
	if err != nil {
		t.Fatalf("SelfLock: %v", err)
	}
	dest := &externalapi.ScriptPublicKey{Script: []byte{0x51}, Version: 0}
	outpoint := &externalapi.DomainOutpoint{Index: 0}
	tx, err := BuildTrancheClaimTransaction(tranche, outpoint, 0, recipientKey, 1000, dest)
	if err != nil {
		t.Fatalf("BuildTrancheClaimTransaction: %v", err)
	}
	if tx.LockTime != tranche.UnlockDaaScore {
		t.Errorf("claim LockTime = %d, want %d", tx.LockTime, tranche.UnlockDaaScore)
	}
	if tx.Outputs[0].Value != tranche.Amount-1000 {
		t.Errorf("claim output = %d, want %d", tx.Outputs[0].Value, tranche.Amount-1000)
	}
	engine, err := txscript.NewEngine(tranche.ScriptPublicKey, tx, 0, txscript.ScriptNoFlags, nil, nil, &consensushashing.SighashReusedValues{})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	if err := engine.Execute(); err != nil {
		t.Errorf("built claim transaction rejected by script engine: %v", err)
	}
}

func TestAlignedLock(t *testing.T) {
	_, pub := newSchnorrKey(t)
	current := uint64(1_000_000)

	// Too-short lock is rejected.
	if _, err := AlignedLock(pub, 1_000_000, current, MinAlignedLockBlocks-1); err == nil {
		t.Errorf("AlignedLock accepted a lock shorter than the minimum, want error")
	}
	// A valid aligned lock.
	tranche, err := AlignedLock(pub, 1_000_000, current, MinAlignedLockBlocks*3)
	if err != nil {
		t.Fatalf("AlignedLock: %v", err)
	}
	if !IsAlignedLock(tranche, current) {
		t.Errorf("fresh aligned lock not recognized as aligned")
	}
	// Once too little time remains, it is no longer aligned.
	almostUnlocked := tranche.UnlockDaaScore - (MinAlignedLockBlocks - 1)
	if IsAlignedLock(tranche, almostUnlocked) {
		t.Errorf("near-expiry lock still recognized as aligned, want not aligned")
	}
	// After unlock it is not aligned.
	if IsAlignedLock(tranche, tranche.UnlockDaaScore+1) {
		t.Errorf("unlocked lock still recognized as aligned")
	}
}

func TestSchnorrPubKeyFromAddress(t *testing.T) {
	_, pub := newSchnorrKey(t)
	address, err := util.NewAddressPublicKey(pub, util.Bech32PrefixSedra)
	if err != nil {
		t.Fatalf("NewAddressPublicKey: %v", err)
	}
	got, err := SchnorrPubKeyFromAddress(address)
	if err != nil {
		t.Fatalf("SchnorrPubKeyFromAddress: %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("got %d-byte pubkey, want 32", len(got))
	}
	for i := range got {
		if got[i] != pub[i] {
			t.Fatalf("recovered pubkey differs from original at byte %d", i)
		}
	}

	// A script-hash address must be rejected (timelocks need a pubkey recipient).
	scriptAddress, err := util.NewAddressScriptHash([]byte("some-redeem-script-bytes-32------"), util.Bech32PrefixSedra)
	if err == nil {
		if _, err := SchnorrPubKeyFromAddress(scriptAddress); err == nil {
			t.Errorf("SchnorrPubKeyFromAddress accepted a script-hash address, want error")
		}
	}
}

func TestDistributionOutputsMapping(t *testing.T) {
	_, pub := newSchnorrKey(t)
	tranches, err := BuildTrancheSchedule(pub, 3_000_000, 100, 30_000, 3)
	if err != nil {
		t.Fatalf("BuildTrancheSchedule: %v", err)
	}
	change := &externalapi.DomainTransactionOutput{Value: 12345, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x51}, Version: 0}}
	outputs := distributionOutputs(tranches, []*externalapi.DomainTransactionOutput{change})
	if len(outputs) != len(tranches)+1 {
		t.Fatalf("got %d outputs, want %d", len(outputs), len(tranches)+1)
	}
	for i, tranche := range tranches {
		if outputs[i].Value != tranche.Amount {
			t.Errorf("output %d value = %d, want %d", i, outputs[i].Value, tranche.Amount)
		}
		if outputs[i].ScriptPublicKey != tranche.ScriptPublicKey {
			t.Errorf("output %d scriptPublicKey does not match tranche %d", i, i)
		}
	}
	if outputs[len(tranches)] != change {
		t.Errorf("change output not appended last")
	}
}
