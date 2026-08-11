package consensus_test

import (
	"bytes"
	"testing"

	"github.com/sedracoin/sedrad/domain/consensus"
	"github.com/sedracoin/sedrad/domain/consensus/model"
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/consensushashing"
	"github.com/sedracoin/sedrad/domain/consensus/utils/constants"
	"github.com/sedracoin/sedrad/domain/consensus/utils/transactionhelper"
	"github.com/sedracoin/sedrad/domain/consensus/utils/testutils"
)

// TestTreasuryAllocationFork is the end-to-end proof of the v2.0.0 treasury
// allocation (Model B). On a fork-enabled node it mines through the activation
// window and asserts:
//   - the extra treasury output appears in a block's coinbase IFF the block's DAA
//     score is inside [activation, activation+window);
//   - the miner reward (all non-treasury coinbase outputs) is UNCHANGED by the
//     allocation (the output is purely additive);
//   - a node WITHOUT the rule REJECTS a window block (proving a real consensus
//     fork) while accepting every out-of-window block (staying in sync pre-fork).
func TestTreasuryAllocationFork(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		const activation = 5
		const window = 6
		const perBlockSeep = 700 * constants.SeepPerSedra // distinct from the 500 SDR subsidy
		treasuryScript := &externalapi.ScriptPublicKey{Script: []byte{0x51}, Version: 0}

		// Fork-enabled node.
		consensusConfig.TreasuryAllocationActivationDaaScore = []uint64{activation}
		consensusConfig.TreasuryAllocationWindowBlocks = []uint64{window}
		consensusConfig.TreasuryAllocationPerBlockSeep = []uint64{perBlockSeep}
		consensusConfig.TreasuryScriptPublicKey = treasuryScript

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestTreasuryAllocationFork")
		if err != nil {
			t.Fatalf("Error setting up fork consensus: %+v", err)
		}
		defer teardown(false)

		// "Old" node: identical config but the allocation is dormant (pre-fork client).
		oldConfig := *consensusConfig
		oldConfig.TreasuryAllocationActivationDaaScore = nil
		oldConfig.TreasuryAllocationWindowBlocks = nil
		oldConfig.TreasuryAllocationPerBlockSeep = nil
		oldConfig.TreasuryScriptPublicKey = nil
		oldTc, oldTeardown, err := factory.NewTestConsensus(&oldConfig, "TestTreasuryAllocationFork_old")
		if err != nil {
			t.Fatalf("Error setting up old consensus: %+v", err)
		}
		defer oldTeardown(false)

		stagingArea := model.NewStagingArea()
		tipHash := consensusConfig.GenesisHash

		isTreasuryOutput := func(o *externalapi.DomainTransactionOutput) bool {
			return o.Value == perBlockSeep &&
				o.ScriptPublicKey.Version == treasuryScript.Version &&
				bytes.Equal(o.ScriptPublicKey.Script, treasuryScript.Script)
		}

		var sawWindowOutput, oldNodeRejected bool
		var minerRewardInWindow, minerRewardOutOfWindow uint64
		oldNodeInSync := true // stop feeding the old node once the chains diverge

		for i := 0; i < activation+window+4; i++ {
			block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("block #%d: fork node failed to build: %+v", i, err)
			}
			// The fork node must accept its own block (builder == validator).
			if err := tc.ValidateAndInsertBlock(block, true); err != nil {
				t.Fatalf("block #%d: fork node rejected its own block: %+v", i, err)
			}
			blockHash := consensushashing.BlockHash(block)

			daaScore, err := tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), stagingArea, blockHash)
			if err != nil {
				t.Fatalf("block #%d: DAAScore: %+v", i, err)
			}

			coinbase := block.Transactions[transactionhelper.CoinbaseTransactionIndex]
			treasuryOutputs, minerReward := 0, uint64(0)
			for _, o := range coinbase.Outputs {
				if isTreasuryOutput(o) {
					treasuryOutputs++
				} else {
					minerReward += o.Value
				}
			}

			inWindow := daaScore >= activation && daaScore < activation+window
			if inWindow {
				if treasuryOutputs != 1 {
					t.Errorf("in-window block (daa=%d) has %d treasury outputs, want exactly 1", daaScore, treasuryOutputs)
				}
				minerRewardInWindow = minerReward
			} else {
				if treasuryOutputs != 0 {
					t.Errorf("out-of-window block (daa=%d) has %d treasury outputs, want 0", daaScore, treasuryOutputs)
				}
				if daaScore >= 2 { // skip genesis-merge edge blocks
					minerRewardOutOfWindow = minerReward
				}
			}

			// Cross-validate against the old (no-rule) node. A coinbase/UTXO-invalid
			// block is not returned as an insert error; it is marked
			// StatusDisqualifiedFromChain, so we inspect the resulting status.
			if oldNodeInSync {
				errOld := oldTc.ValidateAndInsertBlock(block, true)
				status := externalapi.StatusInvalid
				if errOld == nil {
					info, e := oldTc.GetBlockInfo(blockHash)
					if e != nil {
						t.Fatalf("block #%d: old node GetBlockInfo: %+v", i, e)
					}
					status = info.BlockStatus
				}
				rejected := errOld != nil || status != externalapi.StatusUTXOValid
				if inWindow {
					if !rejected {
						t.Errorf("old node ACCEPTED a window block (daa=%d, status=%s) - not a real fork!", daaScore, status)
					} else {
						oldNodeRejected = true
					}
					oldNodeInSync = false // chains diverge from here
				} else if rejected {
					t.Fatalf("old node rejected a pre-fork block (daa=%d, err=%v, status=%s)", daaScore, errOld, status)
				}
			}

			if inWindow {
				sawWindowOutput = true
			}
			tipHash = blockHash
		}

		if !sawWindowOutput {
			t.Fatal("never observed a treasury output inside the activation window")
		}
		if !oldNodeRejected {
			t.Fatal("old (no-rule) node never rejected a fork block")
		}
		// Miner reward is identical in and out of the window -> allocation is purely additive.
		if minerRewardInWindow != minerRewardOutOfWindow {
			t.Errorf("miner reward changed by the allocation: in-window=%d out-of-window=%d",
				minerRewardInWindow, minerRewardOutOfWindow)
		}
	})
}
