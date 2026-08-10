package libsedrawallet

import (
	"github.com/sedracoin/go-secp256k1"
	"github.com/sedracoin/sedrad/cmd/sedrawallet/libsedrawallet/serialization"
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/consensushashing"
	"github.com/sedracoin/sedrad/domain/consensus/utils/constants"
	"github.com/sedracoin/sedrad/domain/consensus/utils/subnetworks"
	"github.com/sedracoin/sedrad/domain/consensus/utils/txscript"
	"github.com/sedracoin/sedrad/domain/consensus/utils/utxo"
	"github.com/sedracoin/sedrad/util"
	"github.com/pkg/errors"
)

// Time-locked distribution and locking (v2.0.0 light-wallet features).
//
// Everything here is built from the same primitive proven in
// txscript/treasury_steward_test.go: a P2SH output whose redeem script is
//
//	<unlockDaaScore> OP_CHECKLOCKTIMEVERIFY <schnorrPubKey> OP_CHECKSIG
//
// The recipient can spend the output only with a transaction whose LockTime is >=
// unlockDaaScore (and a non-final input sequence). Because the CHAIN enforces the
// unlock, a schedule created once keeps releasing on time with nothing running -
// it survives the wallet being closed and our servers going away.
//
// Uses (all no-consensus-change, wallet-level only):
//   - Treasury scheduled distribution (Model 1): split a total into N tranches that
//     unlock evenly over a period, all funded by ONE 2-of-3 signature at setup.
//   - User self-lock: lock your own funds until a chosen height.
//   - Aligned lock: a durable self-lock recognized OFF-CHAIN (Keeper eligibility,
//     reduced service fees, governance weight) - just a sustained lock, no new rules.

// scriptPublicKeyVersion is the script version used for wallet-built outputs.
const scriptPublicKeyVersion uint16 = 0

// MinAlignedLockBlocks is the minimum duration (in blocks ~= seconds at 1 block/s) a
// self-lock must still span to be recognized as "aligned" by off-chain services. 30 days.
const MinAlignedLockBlocks uint64 = 30 * 24 * 60 * 60

// Tranche is one time-locked slice of a distribution (or a single lock). The recipient
// encoded in RedeemScript can spend Amount seep only once a spending transaction sets
// LockTime >= UnlockDaaScore.
type Tranche struct {
	Index           int
	Amount          uint64
	UnlockDaaScore  uint64
	RedeemScript    []byte
	ScriptPublicKey *externalapi.ScriptPublicKey
}

// SchnorrPubKeyFromAddress extracts the 32-byte schnorr public key from a standard
// pay-to-pubkey sedra address. Timelocks require a pubkey (P2PK) recipient because the
// redeem script ends in <pubkey> OP_CHECKSIG; script-hash addresses are rejected.
func SchnorrPubKeyFromAddress(address util.Address) ([]byte, error) {
	pubKeyAddress, ok := address.(*util.AddressPublicKey)
	if !ok {
		return nil, errors.Errorf("timelock recipient must be a pay-to-pubkey (sedra:q...) address, got %T", address)
	}
	pubKey := pubKeyAddress.ScriptAddress()
	if len(pubKey) != 32 {
		return nil, errors.Errorf("expected a 32-byte schnorr public key, got %d bytes", len(pubKey))
	}
	return pubKey, nil
}

// CLTVRedeemScript builds <unlockDaaScore> OP_CHECKLOCKTIMEVERIFY <pubKey> OP_CHECKSIG.
// This fork's OP_CHECKLOCKTIMEVERIFY POPS its operand, so there is deliberately no
// trailing OP_DROP (a trailing OP_DROP would corrupt the stack and break the script).
func CLTVRedeemScript(unlockDaaScore uint64, serializedSchnorrPubKey []byte) ([]byte, error) {
	if len(serializedSchnorrPubKey) != 32 {
		return nil, errors.Errorf("expected a 32-byte schnorr public key, got %d bytes", len(serializedSchnorrPubKey))
	}
	// A lock value at or above LockTimeThreshold is interpreted as a unix timestamp, not
	// a DAA score. Keep DAA-score locks strictly below it (~15,000 years of blocks away).
	if unlockDaaScore >= constants.LockTimeThreshold {
		return nil, errors.Errorf("unlockDaaScore %d must be below LockTimeThreshold %d to be a DAA-score lock",
			unlockDaaScore, uint64(constants.LockTimeThreshold))
	}
	builder := txscript.NewScriptBuilder()
	builder.AddInt64(int64(unlockDaaScore))
	builder.AddOp(txscript.OpCheckLockTimeVerify)
	builder.AddData(serializedSchnorrPubKey)
	builder.AddOp(txscript.OpCheckSig)
	return builder.Script()
}

func trancheScriptPublicKey(redeemScript []byte) (*externalapi.ScriptPublicKey, error) {
	p2sh, err := txscript.PayToScriptHashScript(redeemScript)
	if err != nil {
		return nil, err
	}
	return &externalapi.ScriptPublicKey{Script: p2sh, Version: scriptPublicKeyVersion}, nil
}

// BuildTrancheSchedule splits `total` seep into `numTranches` tranches for the recipient
// (a 32-byte schnorr pubkey), unlocking evenly across `periodBlocks` starting after
// `startDaaScore`: tranche i unlocks at startDaaScore + (i+1)*periodBlocks/numTranches.
// The first numTranches-1 tranches carry total/numTranches seep; the last carries the
// remainder, so the amounts sum to exactly `total`.
func BuildTrancheSchedule(recipientSchnorrPubKey []byte, total, startDaaScore, periodBlocks uint64, numTranches int) ([]*Tranche, error) {
	if numTranches <= 0 {
		return nil, errors.Errorf("numTranches must be positive, got %d", numTranches)
	}
	if total == 0 {
		return nil, errors.Errorf("total must be positive")
	}
	if uint64(numTranches) > total {
		return nil, errors.Errorf("cannot split %d seep into %d tranches (each needs >= 1 seep)", total, numTranches)
	}
	if periodBlocks == 0 {
		return nil, errors.Errorf("periodBlocks must be positive")
	}
	step := periodBlocks / uint64(numTranches)
	if step == 0 {
		return nil, errors.Errorf("periodBlocks (%d) must be >= numTranches (%d) so tranches unlock at distinct heights", periodBlocks, numTranches)
	}
	perTranche := total / uint64(numTranches)
	tranches := make([]*Tranche, numTranches)
	var distributed uint64
	for i := 0; i < numTranches; i++ {
		amount := perTranche
		if i == numTranches-1 {
			amount = total - distributed // remainder to the final tranche
		}
		distributed += amount
		unlockDaaScore := startDaaScore + step*uint64(i+1)
		redeemScript, err := CLTVRedeemScript(unlockDaaScore, recipientSchnorrPubKey)
		if err != nil {
			return nil, err
		}
		scriptPublicKey, err := trancheScriptPublicKey(redeemScript)
		if err != nil {
			return nil, err
		}
		tranches[i] = &Tranche{
			Index:           i,
			Amount:          amount,
			UnlockDaaScore:  unlockDaaScore,
			RedeemScript:    redeemScript,
			ScriptPublicKey: scriptPublicKey,
		}
	}
	return tranches, nil
}

// SelfLock builds a single self-vesting tranche: `amount` locked to the owner's own key
// until currentDaaScore+lockBlocks. Trustless savings/commitment, no counterparty.
func SelfLock(ownerSchnorrPubKey []byte, amount, currentDaaScore, lockBlocks uint64) (*Tranche, error) {
	schedule, err := BuildTrancheSchedule(ownerSchnorrPubKey, amount, currentDaaScore, lockBlocks, 1)
	if err != nil {
		return nil, err
	}
	return schedule[0], nil
}

// AlignedLock builds a "lock to align" tranche: the owner locks `amount` to their own key
// for at least MinAlignedLockBlocks. On-chain it is an ordinary self-lock; its meaning
// ("aligned") is recognized OFF-CHAIN by services reading chain state - no consensus change.
func AlignedLock(ownerSchnorrPubKey []byte, amount, currentDaaScore, lockBlocks uint64) (*Tranche, error) {
	if lockBlocks < MinAlignedLockBlocks {
		return nil, errors.Errorf("aligned lock must span at least %d blocks (~30d), got %d", MinAlignedLockBlocks, lockBlocks)
	}
	return SelfLock(ownerSchnorrPubKey, amount, currentDaaScore, lockBlocks)
}

// IsAlignedLock reports whether `tranche` still qualifies as aligned as of
// currentDaaScore - i.e. it has at least MinAlignedLockBlocks remaining until unlock.
func IsAlignedLock(tranche *Tranche, currentDaaScore uint64) bool {
	if tranche.UnlockDaaScore <= currentDaaScore {
		return false
	}
	return tranche.UnlockDaaScore-currentDaaScore >= MinAlignedLockBlocks
}

// distributionOutputs turns tranches (+ any change outputs) into transaction outputs.
func distributionOutputs(tranches []*Tranche, changeOutputs []*externalapi.DomainTransactionOutput) []*externalapi.DomainTransactionOutput {
	outputs := make([]*externalapi.DomainTransactionOutput, 0, len(tranches)+len(changeOutputs))
	for _, tranche := range tranches {
		outputs = append(outputs, &externalapi.DomainTransactionOutput{
			Value:           tranche.Amount,
			ScriptPublicKey: tranche.ScriptPublicKey,
		})
	}
	return append(outputs, changeOutputs...)
}

// CreateUnsignedDistributionTransaction builds an unsigned (partially-signed) transaction
// that funds every tranche's time-locked output from the given (treasury) UTXOs. The
// treasury multisig signs this SINGLE transaction once via the normal Sign flow; once it
// confirms, the whole schedule exists on-chain and needs nothing running. `changeOutputs`
// may return change to the treasury; inputs - tranche outputs - changeOutputs = fee.
func CreateUnsignedDistributionTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	tranches []*Tranche,
	changeOutputs []*externalapi.DomainTransactionOutput,
	selectedUTXOs []*UTXO) ([]byte, error) {

	if len(tranches) == 0 {
		return nil, errors.Errorf("no tranches to distribute")
	}
	sortPublicKeys(extendedPublicKeys)
	outputs := distributionOutputs(tranches, changeOutputs)
	unsignedTransaction, err := createUnsignedTransactionWithOutputs(extendedPublicKeys, minimumSignatures, outputs, selectedUTXOs)
	if err != nil {
		return nil, err
	}
	return serialization.SerializePartiallySignedTransaction(unsignedTransaction)
}

// BuildTrancheClaimTransaction builds and signs a transaction claiming a single tranche
// (at fundingOutpoint) to destinationScriptPublicKey. It sets LockTime =
// tranche.UnlockDaaScore and a non-final input sequence, so the chain accepts it only
// at/after unlock. Only the recipient's own key is required.
func BuildTrancheClaimTransaction(
	tranche *Tranche,
	fundingOutpoint *externalapi.DomainOutpoint,
	fundingBlockDaaScore uint64,
	recipientKey *secp256k1.SchnorrKeyPair,
	fee uint64,
	destinationScriptPublicKey *externalapi.ScriptPublicKey) (*externalapi.DomainTransaction, error) {

	if fee >= tranche.Amount {
		return nil, errors.Errorf("fee %d must be less than tranche amount %d", fee, tranche.Amount)
	}
	tx := &externalapi.DomainTransaction{
		Version: constants.MaxTransactionVersion,
		Inputs: []*externalapi.DomainTransactionInput{{
			PreviousOutpoint: *fundingOutpoint,
			Sequence:         0, // non-final so OP_CHECKLOCKTIMEVERIFY is enforced
			UTXOEntry:        utxo.NewUTXOEntry(tranche.Amount, tranche.ScriptPublicKey, false, fundingBlockDaaScore),
		}},
		Outputs: []*externalapi.DomainTransactionOutput{{
			Value:           tranche.Amount - fee,
			ScriptPublicKey: destinationScriptPublicKey,
		}},
		LockTime:     tranche.UnlockDaaScore,
		SubnetworkID: subnetworks.SubnetworkIDNative,
	}

	signature, err := txscript.RawTxInSignature(tx, 0, consensushashing.SigHashAll, recipientKey, &consensushashing.SighashReusedValues{})
	if err != nil {
		return nil, err
	}
	signatureScript, err := txscript.NewScriptBuilder().
		AddData(signature).
		AddData(tranche.RedeemScript).
		Script()
	if err != nil {
		return nil, err
	}
	tx.Inputs[0].SignatureScript = signatureScript
	return tx, nil
}
