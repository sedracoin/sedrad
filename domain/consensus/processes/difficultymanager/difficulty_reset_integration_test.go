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

// TestDifficultyResetRelaunchTransition dress-rehearses the ACTUAL relaunch shape that
// TestDifficultyResetIntegration does not: a chain that resumes from an old tip after a
// long stall (a large forward jump in block timestamps at the fork), with the reset
// window sized to EQUAL the difficulty window (as mainnet is: both 2641). It asserts:
//   1. the large timestamp jump at resume is accepted (median-time-past + future-tolerance
//      both hold - a block dated "now" on a months-old parent is valid);
//   2. every in-window block carries genesis difficulty AND validates (build==validate);
//   3. blocks AFTER the reset window still validate (build==validate) - i.e. the first
//      real-difficulty block computes over a gap-free window (the ~months-long stall gap is
//      no longer inside the window), so there is no fork and no stall at the transition.
func TestDifficultyResetRelaunchTransition(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		if consensusConfig.DisableDifficultyAdjustment {
			return
		}
		if consensusConfig.GenesisBlock.Header.TimeInMilliseconds() > mstime.ToMSTime(time.Now().Add(-time.Hour)).UnixMilliseconds() {
			t.Fatalf("requires the GenesisBlock to be at least 1 hour old")
		}

		const activation = 6
		const window = 12
		consensusConfig.K = 1
		// Mainnet-like: the reset window EQUALS the difficulty window, so the first
		// post-reset block's window is exactly the post-reset blocks (gap-free).
		consensusConfig.DifficultyAdjustmentWindowSize = window
		consensusConfig.DifficultyResetActivationDaaScore = activation
		consensusConfig.DifficultyResetWindowBlocks = window
		genesisBits := consensusConfig.GenesisBlock.Header.Bits()

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestDifficultyResetRelaunchTransition")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Pre-fork blocks keep old (genesis-based) timestamps. At the fork we jump the
		// timestamp to ~now, simulating a chain that stalled for a long time and resumed.
		tip := consensusConfig.GenesisHash
		oldTime := consensusConfig.GenesisBlock.Header.TimeInMilliseconds()
		resumeTime := mstime.ToMSTime(time.Now().Add(-time.Minute)).UnixMilliseconds()
		sawInWindow, sawPostReset, jumped := false, false, false

		for i := 0; i < activation+window+8; i++ {
			block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tip}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents (%d): %+v", i, err)
			}

			var blockTime int64
			daaPreview := block.Header.DAAScore()
			if daaPreview >= activation {
				if !jumped {
					blockTime = resumeTime // the stall gap: months-old parent -> "now"
					jumped = true
				} else {
					resumeTime += consensusConfig.TargetTimePerBlock.Milliseconds()
					blockTime = resumeTime
				}
			} else {
				oldTime += consensusConfig.TargetTimePerBlock.Milliseconds()
				blockTime = oldTime
			}
			mutableHeader := block.Header.ToMutable()
			mutableHeader.SetTimeInMilliseconds(blockTime)
			block.Header = mutableHeader.ToImmutable()

			if err := tc.ValidateAndInsertBlock(block, true); err != nil {
				t.Fatalf("ValidateAndInsertBlock (%d, daa=%d, time=%d): %+v", i, block.Header.DAAScore(), blockTime, err)
			}

			daa := block.Header.DAAScore()
			switch {
			case daa >= activation && daa < activation+window:
				sawInWindow = true
				if block.Header.Bits() != genesisBits {
					t.Errorf("in-window block (DAA=%d): bits=%08x, want genesis %08x", daa, block.Header.Bits(), genesisBits)
				}
			case daa >= activation+window:
				// First real-difficulty block(s) after the reset. The only property that
				// matters for safety is that build==validate accepted it (done above) -
				// i.e. no fork and no stall at the gap-free transition.
				sawPostReset = true
			}
			tip = consensushashing.BlockHash(block)
		}
		if !sawInWindow {
			t.Fatalf("test never produced an in-window block; nothing was verified")
		}
		if !sawPostReset {
			t.Fatalf("test never produced a post-reset block; the transition was not exercised")
		}
	})
}

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
