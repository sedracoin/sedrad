package difficultymanager_test

import (
	"testing"
	"time"

	"github.com/sedracoin/sedrad/domain/consensus"
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/consensushashing"
	"github.com/sedracoin/sedrad/domain/consensus/utils/testutils"
	"github.com/sedracoin/sedrad/util/mstime"
)

// TestDifficultyResetIntegration arms the v2.0.0 fair-restart difficulty reset and verifies:
//  1. every block whose DAA score is in [activation, activation+window) carries exactly the
//     genesis difficulty (the mineable bootstrap target), and
//  2. every such block passes ValidateAndInsertBlock — i.e. block BUILDING (RequiredDifficulty)
//     and block VALIDATION (StageDAADataAndReturnRequiredDifficulty) AGREE on the reset
//     difficulty. A disagreement here is exactly what would fork the network at relaunch.
func TestDifficultyResetIntegration(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		if consensusConfig.DisableDifficultyAdjustment {
			return
		}
		if consensusConfig.GenesisBlock.Header.TimeInMilliseconds() > mstime.ToMSTime(time.Now().Add(-time.Hour)).UnixMilliseconds() {
			t.Fatalf("requires the GenesisBlock to be at least 1 hour old")
		}

		const activation = 5
		const window = 12
		consensusConfig.K = 1
		consensusConfig.DifficultyAdjustmentWindowSize = 20
		consensusConfig.DifficultyResetActivationDaaScore = activation
		consensusConfig.DifficultyResetWindowBlocks = window
		genesisBits := consensusConfig.GenesisBlock.Header.Bits()

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestDifficultyResetIntegration")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		tip := consensusConfig.GenesisHash
		tipTime := consensusConfig.GenesisBlock.Header.TimeInMilliseconds()
		sawInWindow := false
		for i := 0; i < activation+window+6; i++ {
			tipTime += consensusConfig.TargetTimePerBlock.Milliseconds()
			block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tip}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents (%d): %+v", i, err)
			}
			mutableHeader := block.Header.ToMutable()
			mutableHeader.SetTimeInMilliseconds(tipTime)
			block.Header = mutableHeader.ToImmutable()

			if err := tc.ValidateAndInsertBlock(block, true); err != nil {
				t.Fatalf("ValidateAndInsertBlock (%d, daa=%d): %+v", i, block.Header.DAAScore(), err)
			}

			daa := block.Header.DAAScore()
			if daa >= activation && daa < activation+window {
				sawInWindow = true
				if block.Header.Bits() != genesisBits {
					t.Errorf("in-window block (DAA=%d): bits=%08x, want genesis %08x",
						daa, block.Header.Bits(), genesisBits)
				}
			}
			tip = consensushashing.BlockHash(block)
		}
		if !sawInWindow {
			t.Fatalf("test never produced a block inside the reset window; nothing was verified")
		}
	})
}
