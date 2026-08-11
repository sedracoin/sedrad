// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math/big"
	"time"

	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"

	"github.com/sedracoin/sedrad/app/appmessage"
	"github.com/sedracoin/sedrad/util/network"

	"github.com/pkg/errors"

	"github.com/sedracoin/sedrad/util"
)

// These variables are the DAG proof-of-work limit parameters for each default
// network.
var (
	// bigOne is 1 represented as a big.Int. It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowMax is the highest proof of work value a sedra block can
	// have for the main network. It is the value 2^255 - 1.
	mainPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// testnetPowMax is the highest proof of work value a sedra block
	// can have for the test network. It is the value 2^255 - 1.
	testnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// simnetPowMax is the highest proof of work value a sedra block
	// can have for the simulation test network. It is the value 2^255 - 1.
	simnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// devnetPowMax is the highest proof of work value a sedra block
	// can have for the development network. It is the value
	// 2^255 - 1.
	devnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)
)

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType uint8

// Params defines a sedra network by its parameters. These parameters may be
// used by sedra applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	// K defines the K parameter for GHOSTDAG consensus algorithm.
	// See ghostdag.go for further details.
	K externalapi.KType

	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net appmessage.SedraNet

	// RPCPort defines the rpc server port
	RPCPort string

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []string

	// GRPCSeeds defines a list of GRPC seeds for the network that are used
	// as one method to discover peers.
	GRPCSeeds []string

	// DefaultFallbackPeers is a list of hardcoded bootstrap peers (host:port)
	// that a node will permanently connect to at startup when the operator has
	// specified neither --addpeer nor --connect. This provides a DNS-independent
	// bootstrap path so a fresh node can join the network even if every DNS seed
	// is unreachable (the failure mode that froze the original chain). Discovery
	// stays enabled, so these are a fallback, not an allowlist.
	DefaultFallbackPeers []string

	// GenesisBlock defines the first block of the DAG.
	GenesisBlock *externalapi.DomainBlock

	// GenesisHash is the starting block hash.
	GenesisHash *externalapi.DomainHash

	// PowMax defines the highest allowed proof of work value for a block
	// as a uint256.
	PowMax *big.Int

	// BlockCoinbaseMaturity is the number of blocks required before newly mined
	// coins can be spent.
	BlockCoinbaseMaturity uint64

	// SubsidyGenesisReward SubsidyMergeSetRewardMultiplier, and
	// SubsidyPastRewardMultiplier are part of the block subsidy equation.
	// Further details: https://hashdag.medium.com/sedra-launch-plan-9a63f4d754a6
	SubsidyGenesisReward            uint64
	PreDeflationaryPhaseBaseSubsidy uint64
	DeflationaryPhaseBaseSubsidy    uint64

	// Treasury allocation (v2.0.0 hard fork). A one-time, height-gated treasury
	// emission that does NOT alter the miner subsidy curve (Model B): for every
	// block whose DAA score is in [TreasuryAllocationActivationDaaScore,
	// TreasuryAllocationActivationDaaScore+TreasuryAllocationWindowBlocks), the
	// coinbase includes one extra output of TreasuryAllocationPerBlockSeep paying
	// TreasuryScriptPublicKey (the disclosed 2-of-3, CLTV-timelocked multisig).
	// The window's total is the 2B allocation; multi-year vesting is enforced by
	// the CLTV timelock on the script, not by the emission schedule.
	// DORMANT unless TreasuryAllocationWindowBlocks>0 AND TreasuryScriptPublicKey
	// is set — so an unconfigured fork build is byte-identical to pre-fork.
	// A window (rather than a single block) is required because coinbase is built
	// and validated per-block from each block's own DAA data; no single-block
	// condition is unique in a DAG.
	TreasuryAllocationActivationDaaScore uint64
	TreasuryAllocationWindowBlocks       uint64
	TreasuryAllocationPerBlockSeep       uint64
	TreasuryScriptPublicKey              *externalapi.ScriptPublicKey

	// DifficultyResetActivationDaaScore / DifficultyResetWindowBlocks implement the
	// v2.0.0 fair-restart difficulty reset: for blocks with DAA score in
	// [DifficultyResetActivationDaaScore, +DifficultyResetWindowBlocks), the required
	// difficulty is forced to the genesis difficulty so public mining reopens at a
	// mineable level after the stall; the DAA then re-ramps as hashrate joins.
	// DORMANT unless both are > 0. Miner subsidy is unaffected.
	DifficultyResetActivationDaaScore uint64
	DifficultyResetWindowBlocks       uint64

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// FinalityDuration is the duration of the finality window.
	FinalityDuration time.Duration

	// TimestampDeviationTolerance is the maximum offset a block timestamp
	// is allowed to be in the future before it gets delayed
	TimestampDeviationTolerance int

	// DifficultyAdjustmentWindowSize is the size of window that is inspected
	// to calculate the required difficulty of each block.
	DifficultyAdjustmentWindowSize int

	// These fields are related to voting on consensus rule changes as
	// defined by BIP0009.
	//
	// RuleChangeActivationThreshold is the number of blocks in a threshold
	// state retarget window for which a positive vote for a rule change
	// must be cast in order to lock in a rule change. It should typically
	// be 95% for the main network and 75% for test networks.
	//
	// MinerConfirmationWindow is the number of blocks in each threshold
	// state retarget window.
	//
	// Deployments define the specific consensus rule changes to be voted
	// on.
	RuleChangeActivationThreshold uint64
	MinerConfirmationWindow       uint64

	// Mempool parameters
	RelayNonStdTxs bool

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable bool

	// Human-readable prefix for Bech32 encoded addresses
	Prefix util.Bech32Prefix

	// Address encoding magics
	PrivateKeyID byte // First byte of a WIF private key

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks bool

	// DisableDifficultyAdjustment determine whether to use difficulty
	DisableDifficultyAdjustment bool

	// SkipProofOfWork indicates whether proof of work should be checked.
	SkipProofOfWork bool

	// MaxCoinbasePayloadLength is the maximum length in bytes allowed for a block's coinbase's payload
	MaxCoinbasePayloadLength uint64

	// MaxBlockMass is the maximum mass a block is allowed
	MaxBlockMass uint64

	// MaxBlockParents is the maximum number of blocks a block is allowed to point to
	MaxBlockParents externalapi.KType

	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	MassPerTxByte uint64

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	MassPerScriptPubKeyByte uint64

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	MassPerSigOp uint64

	// MergeSetSizeLimit is the maximum number of blocks in a block's merge set
	MergeSetSizeLimit uint64

	// CoinbasePayloadScriptPublicKeyMaxLength is the maximum allowed script public key in the coinbase's payload
	CoinbasePayloadScriptPublicKeyMaxLength uint8

	// PruningProofM is the 'm' constant in the pruning proof. For more details see: https://github.com/sedracoin/research/issues/3
	PruningProofM uint64

	// DeflationaryPhaseDaaScore is the DAA score after which the monetary policy switches
	// to its deflationary phase
	DeflationaryPhaseDaaScore uint64

	DisallowDirectBlocksOnTopOfGenesis bool

	// MaxBlockLevel is the maximum possible block level.
	MaxBlockLevel int

	MergeDepth uint64
}

// NormalizeRPCServerAddress returns addr with the current network default
// port appended if there is not already a port specified.
func (p *Params) NormalizeRPCServerAddress(addr string) (string, error) {
	return network.NormalizeAddress(addr, p.RPCPort)
}

// FinalityDepth returns the finality duration represented in blocks
func (p *Params) FinalityDepth() uint64 {
	return uint64(p.FinalityDuration / p.TargetTimePerBlock)
}

// PruningDepth returns the pruning duration represented in blocks
func (p *Params) PruningDepth() uint64 {
	return 2*p.FinalityDepth() + 4*p.MergeSetSizeLimit*uint64(p.K) + 2*uint64(p.K) + 2
}

// MainnetParams defines the network parameters for the main sedra network.
var MainnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "sedra-mainnet",
	Net:         appmessage.Mainnet,
	RPCPort:     "22110",
	DefaultPort: "22111",
	DNSSeeds: []string{
		"sedra-seed-1.sedracoin.com",
		"sedra-seed-2.sedracoin.com",
		"sedra-seed-3.sedracoin.com",
	},

	// DefaultFallbackPeers are our operated bootstrap nodes, used only when the
	// operator specified neither --addpeer nor --connect. They give a
	// DNS-independent bootstrap path. Port omitted -> normalized to DefaultPort
	// (22111). TODO(relaunch): confirm each is live and serving 22111 before
	// tagging the v1.1.0 release; extend/replace as independent operators join.
	DefaultFallbackPeers: []string{
		"169.58.126.141", // hub node
		"169.58.111.191", // secondary node
	},

	// Foundation treasury (v2.0.0). Disclosed 2-of-3 multisig receive address:
	//   sedra:pp8a4l47q0qjwlwgqvnmp5ppssk6zpn9hat58vaw6jkxkee2aqduue9252wc3
	// P2SH ScriptPublicKey below. Allocation stays DORMANT until Window/PerBlock/
	// ActivationDaaScore are set at the v2.0.0 release (after the securities review
	// and the disclosed activation height). Intended: Window=86400 (~1d),
	// PerBlock=2B*1e8/86400 seep, streaming 2B to this address; miner subsidy untouched.
	TreasuryScriptPublicKey: &externalapi.ScriptPublicKey{Version: 0, Script: []byte{
		0xaa, 0x20, 0x4f, 0xda, 0xfe, 0xbe, 0x03, 0xc1, 0x27, 0x7d, 0xc8, 0x03, 0x27, 0xb0, 0xd0, 0x21,
		0x84, 0x2d, 0xa1, 0x06, 0x65, 0xbf, 0x57, 0x43, 0xb3, 0xae, 0xd4, 0xac, 0x6b, 0x67, 0x2a, 0xe8,
		0x1b, 0xce, 0x87,
	}},
	// ARMED 2026-08-11: the 2,000,000,000 SDR regenesis-foundation allocation. Each block
	// with DAA score in [activation, activation+window) adds an extra coinbase output of
	// TreasuryPerBlockSeep to the treasury multisig above (miner subsidy untouched).
	// activation = frozen tip DAA (53,632,352) + 20 buffer, so no already-mined block is
	// in-window. window(200) * perBlock(1e15 seep) = 2e17 seep = 2,000,000,000 SDR.
	TreasuryAllocationActivationDaaScore: 53_633_267,
	TreasuryAllocationWindowBlocks:       454,
	TreasuryAllocationPerBlockSeep:       100_000_000_000_000,

	// Fair-restart difficulty reset (v2.0.0). Blocks with DAA score in
	// [activation, activation+2641) use the genesis difficulty so public mining reopens
	// mineably after the stall; the DAA then re-ramps with real hashrate. Window = the
	// DAA window size (2641) for a clean handoff. activation = restored tip DAA
	// (53,625,512) + 1, so it applies to the FIRST new block and never re-validates the
	// existing tip. RE-CONFIRM the live tip DAA right before the final release build and
	// update if the chain advanced. (Treasury allocation above stays dormant separately.)
	DifficultyResetActivationDaaScore: 53_625_512,
	DifficultyResetWindowBlocks:       2641,

	// DAG parameters
	GenesisBlock:                    &genesisBlock,
	GenesisHash:                     genesisHash,
	PowMax:                          mainPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1916, // 95% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016, //

	// Mempool parameters
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixSedra,

	// Address encoding magics
	PrivateKeyID: 0x80, // starts with 5 (uncompressed) or K (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,
	// 	DisallowDirectBlocksOnTopOfGenesis:      true,
	DisallowDirectBlocksOnTopOfGenesis: false,

	// This is technically 255, but we clamped it at 256 - block level of mainnet genesis
	// This means that any block that has a level lower or equal to genesis will be level 0.
	MaxBlockLevel: 225,
	MergeDepth:    defaultMergeDepth,
}

// TestnetParams defines the network parameters for the test sedra network.
var TestnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "sedra-testnet-10",
	Net:         appmessage.Testnet,
	RPCPort:     "22210",
	DefaultPort: "22211",
	DNSSeeds:    []string{
		//"sedracoin.com",
	},

	// DAG parameters
	GenesisBlock:                    &testnetGenesisBlock,
	GenesisHash:                     testnetGenesisHash,
	PowMax:                          testnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,

	// Mempool parameters
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixSedraTest,

	// Address encoding magics
	PrivateKeyID: 0xef, // starts with 9 (uncompressed) or c (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// SimnetParams defines the network parameters for the simulation test sedra
// network. This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing. The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules. This is important as otherwise it would
// just turn into another public testnet.
var SimnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "sedra-simnet",
	Net:         appmessage.Simnet,
	RPCPort:     "22510",
	DefaultPort: "22511",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &simnetGenesisBlock,
	GenesisHash:                     simnetGenesisHash,
	PowMax:                          simnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
	TargetTimePerBlock:              time.Millisecond,
	FinalityDuration:                time.Minute,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 75, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       100,

	// Mempool parameters
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	PrivateKeyID: 0x64, // starts with 4 (uncompressed) or F (compressed)
	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixSedraSim,

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: true,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// DevnetParams defines the network parameters for the development sedra network.
var DevnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "sedra-devnet",
	Net:         appmessage.Devnet,
	RPCPort:     "22610",
	DefaultPort: "22611",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &devnetGenesisBlock,
	GenesisHash:                     devnetGenesisHash,
	PowMax:                          devnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,

	// Mempool parameters
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: true,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixSedraDev,

	// Address encoding magics
	PrivateKeyID: 0xef, // starts with 9 (uncompressed) or c (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// ErrDuplicateNet describes an error where the parameters for a sedra
// network could not be set due to the network already being a standard
// network or previously-registered into this package.
var ErrDuplicateNet = errors.New("duplicate sedra network")

var registeredNets = make(map[appmessage.SedraNet]struct{})

// Register registers the network parameters for a sedra network. This may
// error with ErrDuplicateNet if the network is already registered (either
// due to a previous Register call, or the network being one of the default
// networks).
//
// Network parameters should be registered into this package by a main package
// as early as possible. Then, library packages may lookup networks or network
// parameters based on inputs and work regardless of the network being standard
// or not.
func Register(params *Params) error {
	if _, ok := registeredNets[params.Net]; ok {
		return ErrDuplicateNet
	}
	registeredNets[params.Net] = struct{}{}

	return nil
}

// mustRegister performs the same function as Register except it panics if there
// is an error. This should only be called from package init functions.
func mustRegister(params *Params) {
	if err := Register(params); err != nil {
		panic("failed to register network: " + err.Error())
	}
}

func init() {
	// Register all default networks when the package is initialized.
	mustRegister(&MainnetParams)
	mustRegister(&TestnetParams)
	mustRegister(&SimnetParams)
	mustRegister(&DevnetParams)
}
