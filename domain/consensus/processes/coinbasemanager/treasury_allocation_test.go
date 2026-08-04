package coinbasemanager

import (
	"math"
	"testing"
)

// TestBlockInTreasuryAllocationWindow verifies the v2.0.0 treasury-allocation
// gating: the extra coinbase output is emitted only for blocks whose DAA score
// is in the half-open window [activation, activation+window), and never when the
// allocation is dormant (window == 0).
func TestBlockInTreasuryAllocationWindow(t *testing.T) {
	const activation = 80_000_000
	const window = 86_400 // ~1 day of blocks @ 1 blk/s

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
			got := blockInTreasuryAllocationWindow(tt.daaScore, tt.activ, tt.window)
			if got != tt.want {
				t.Errorf("blockInTreasuryAllocationWindow(daa=%d, activation=%d, window=%d) = %v, want %v",
					tt.daaScore, tt.activ, tt.window, got, tt.want)
			}
		})
	}
}

// TestTreasuryAllocationWindowTotal documents that per-block amount * window equals
// the intended 2B allocation (the disclosed schedule), independent of I/O.
func TestTreasuryAllocationWindowTotal(t *testing.T) {
	const seepPerSedra = 100_000_000
	const targetSDR = 2_000_000_000 // 2B SDR
	const window = 86_400

	// perBlock is floor(target / window); remainder is folded into disclosure.
	targetSeep := uint64(targetSDR) * seepPerSedra
	perBlock := targetSeep / window
	realized := perBlock * window
	remainder := targetSeep - realized

	if perBlock == 0 {
		t.Fatal("per-block allocation computed as zero")
	}
	// Realized total must be within one window-block of the target (rounding only).
	if targetSeep-realized >= perBlock {
		t.Errorf("rounding gap too large: target=%d realized=%d gap=%d perBlock=%d",
			targetSeep, realized, remainder, perBlock)
	}
	t.Logf("window=%d blocks, perBlock=%d seep (%.4f SDR), realized=%d seep (%.6f B SDR), remainder=%d seep",
		window, perBlock, float64(perBlock)/seepPerSedra, realized,
		float64(realized)/seepPerSedra/1e9, remainder)
}
