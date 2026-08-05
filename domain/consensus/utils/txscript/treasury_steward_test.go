package txscript

import (
	"testing"

	"github.com/sedracoin/go-secp256k1"
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/consensushashing"
	"github.com/sedracoin/sedrad/domain/consensus/utils/constants"
	"github.com/sedracoin/sedrad/domain/consensus/utils/utxo"
	"github.com/sedracoin/sedrad/domain/dagconfig"
)

// Steward treasury tests: prove, at the consensus script-engine level, the two
// properties that actually secure the 2B v2.0.0 allocation:
//   1. a 2-of-3 multisig is spendable by ANY two stewards and NOT by one;
//   2. a CLTV (OP_CHECKLOCKTIMEVERIFY) lock enforces a DAA-score timelock (vesting).

func genStewardKeys(t *testing.T, n int) ([]*secp256k1.SchnorrKeyPair, [][]byte) {
	t.Helper()
	keys := make([]*secp256k1.SchnorrKeyPair, n)
	pubs := make([][]byte, n)
	for i := 0; i < n; i++ {
		k, err := secp256k1.GenerateSchnorrKeyPair()
		if err != nil {
			t.Fatalf("keygen %d: %v", i, err)
		}
		pk, err := k.SchnorrPublicKey()
		if err != nil {
			t.Fatalf("pubkey %d: %v", i, err)
		}
		ser, err := pk.Serialize()
		if err != nil {
			t.Fatalf("serialize %d: %v", i, err)
		}
		keys[i] = k
		pubs[i] = ser[:]
	}
	return keys, pubs
}

func multisigRedeemScript(t *testing.T, m int, pubs [][]byte) []byte {
	t.Helper()
	b := NewScriptBuilder()
	b.AddInt64(int64(m))
	for _, pk := range pubs {
		b.AddData(pk)
	}
	b.AddInt64(int64(len(pubs)))
	b.AddOp(OpCheckMultiSig)
	s, err := b.Script()
	if err != nil {
		t.Fatalf("build multisig redeem: %v", err)
	}
	return s
}

// spend builds a 1-input tx spending a UTXO locked by p2shScriptPubKey, with the
// given lockTime/sequence, signs the input with each key in signers (ascending
// pubkey order), and returns whether the script engine accepts it.
func spendAccepts(t *testing.T, redeemScript []byte, p2shScriptPubKey *externalapi.ScriptPublicKey,
	signers []*secp256k1.SchnorrKeyPair, lockTime uint64, sequence uint64) bool {
	t.Helper()

	tx := &externalapi.DomainTransaction{
		Version: constants.MaxTransactionVersion,
		Inputs: []*externalapi.DomainTransactionInput{{
			PreviousOutpoint: externalapi.DomainOutpoint{Index: 0},
			Sequence:         sequence,
			UTXOEntry:        utxo.NewUTXOEntry(1_000_000, p2shScriptPubKey, false, 10),
		}},
		Outputs:  []*externalapi.DomainTransactionOutput{{Value: 900_000, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x51}, Version: 0}}},
		LockTime: lockTime,
	}

	sb := NewScriptBuilder()
	for _, k := range signers {
		sig, err := RawTxInSignature(tx, 0, consensushashing.SigHashAll, k, &consensushashing.SighashReusedValues{})
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		sb.AddData(sig)
	}
	sb.AddData(redeemScript)
	sigScript, err := sb.Script()
	if err != nil {
		t.Fatalf("build sigScript: %v", err)
	}

	return checkScripts("steward", tx, 0, sigScript, p2shScriptPubKey) == nil
}

func TestStewardMultisig2of3(t *testing.T) {
	keys, pubs := genStewardKeys(t, 3)
	redeem := multisigRedeemScript(t, 2, pubs) // 2-of-3
	p2sh, err := PayToScriptHashScript(redeem)
	if err != nil {
		t.Fatalf("p2sh: %v", err)
	}
	scriptPubKey := &externalapi.ScriptPublicKey{Script: p2sh, Version: 0}

	if _, addr, err := ExtractScriptPubKeyAddress(scriptPubKey, &dagconfig.MainnetParams); err == nil {
		t.Logf("treasury 2-of-3 P2SH address: %s", addr.EncodeAddress())
	}

	const seq = 0 // finalized-check irrelevant for multisig; keep < max
	// ANY two of three must spend.
	for _, pair := range [][]int{{0, 1}, {0, 2}, {1, 2}} {
		signers := []*secp256k1.SchnorrKeyPair{keys[pair[0]], keys[pair[1]]}
		if !spendAccepts(t, redeem, scriptPubKey, signers, 0, seq) {
			t.Errorf("2-of-3 spend with stewards %v was REJECTED, want accepted", pair)
		}
	}
	// A single steward must NOT spend.
	for _, one := range [][]int{{0}, {1}, {2}} {
		signers := []*secp256k1.SchnorrKeyPair{keys[one[0]]}
		if spendAccepts(t, redeem, scriptPubKey, signers, 0, seq) {
			t.Errorf("single-steward spend by %v was ACCEPTED, want rejected", one)
		}
	}
}

func TestStewardCLTVTimelock(t *testing.T) {
	keys, pubs := genStewardKeys(t, 1)
	const unlockDaaScore = 1_000_000

	// redeem: <unlockDaa> OP_CHECKLOCKTIMEVERIFY <pubkey> OP_CHECKSIG
	// NOTE: this kaspa-fork's OP_CHECKLOCKTIMEVERIFY POPS its operand (unlike
	// Bitcoin, which leaves it and needs a trailing OP_DROP) — so no OP_DROP here.
	b := NewScriptBuilder()
	b.AddInt64(int64(unlockDaaScore))
	b.AddOp(OpCheckLockTimeVerify)
	b.AddData(pubs[0])
	b.AddOp(OpCheckSig)
	redeem, err := b.Script()
	if err != nil {
		t.Fatalf("build cltv redeem: %v", err)
	}
	p2sh, err := PayToScriptHashScript(redeem)
	if err != nil {
		t.Fatalf("p2sh: %v", err)
	}
	scriptPubKey := &externalapi.ScriptPublicKey{Script: p2sh, Version: 0}

	// Sequence must NOT be the max value, or CLTV is bypassed ("finalized").
	const seq = 0
	// Before unlock: tx.LockTime < unlockDaaScore -> rejected.
	if spendAccepts(t, redeem, scriptPubKey, keys, unlockDaaScore-1, seq) {
		t.Errorf("CLTV spend BEFORE unlock (locktime=%d) was ACCEPTED, want rejected", unlockDaaScore-1)
	}
	// At/after unlock: tx.LockTime >= unlockDaaScore -> accepted.
	if !spendAccepts(t, redeem, scriptPubKey, keys, unlockDaaScore, seq) {
		t.Errorf("CLTV spend AT unlock (locktime=%d) was REJECTED, want accepted", unlockDaaScore)
	}
	// A finalized input (sequence == max) must fail CLTV even after unlock.
	if spendAccepts(t, redeem, scriptPubKey, keys, unlockDaaScore, constants.MaxTxInSequenceNum) {
		t.Errorf("CLTV spend with finalized input was ACCEPTED, want rejected")
	}
}
