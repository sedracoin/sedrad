package difficultymanager

import (
	"math"
	"testing"
)

// TestBlockInDifficultyResetWindow verifies the v2.0.0 fair-restart gating: the difficulty
// is reset only for blocks whose DAA score is in [activation, activation+window), and never
// when dormant (window == 0).
func TestBlockInDifficultyResetWindow(t *testing.T) {
	const activation = 53_700_000 // e.g. a disclosed post-relaunch DAA score
	const window = 2641           // ~one DAA window of bootstrap blocks

	tests := []struct {
		name     string
		daaScore uint64
		activ    uint64
		window   uint64
		want     bool
	}{
		{"dormant window=0 inside range", activation, activation, 0, false},
		{"dormant window=0 at zero", 0, 0, 0, false},
		{"one block before activation", activation - 1, activation, window, false},
		{"exactly at activation (first in-window)", activation, activation, window, true},
		{"mid window", activation + window/2, activation, window, true},
		{"last in-window block", activation + window - 1, activation, window, true},
		{"first block after window", activation + window, activation, window, false},
		{"far after window", activation + window + 1_000_000, activation, window, false},
		{"far before window", 1, activation, window, false},
		{"overflow guard: activation near max", math.MaxUint64 - 5, math.MaxUint64 - 10, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blockInDifficultyResetWindow(tt.daaScore, tt.activ, tt.window)
			if got != tt.want {
				t.Errorf("blockInDifficultyResetWindow(daa=%d, activation=%d, window=%d) = %v, want %v",
					tt.daaScore, tt.activ, tt.window, got, tt.want)
			}
		})
	}
}
