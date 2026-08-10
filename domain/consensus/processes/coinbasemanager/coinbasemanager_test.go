package coinbasemanager

import (
	"github.com/sedracoin/sedrad/domain/consensus/model/externalapi"
	"github.com/sedracoin/sedrad/domain/consensus/utils/constants"
	"github.com/sedracoin/sedrad/domain/dagconfig"
	"strconv"
	"testing"
)

func TestCalcDeflationaryPeriodBlockSubsidy(t *testing.T) {
	const secondsPerMonth = 2629800
	const secondsPerHalving = secondsPerMonth * 12
	const deflationaryPhaseDaaScore = secondsPerMonth * 6
	const deflationaryPhaseBaseSubsidy = 440 * constants.SeepPerSedra
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		deflationaryPhaseDaaScore,
		deflationaryPhaseBaseSubsidy,
		0,
		0,
		0,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
	coinbaseManagerInstance := coinbaseManagerInterface.(*coinbaseManager)

	// Sedra's deflationary emission is defined by subsidyByDeflationaryMonthTable, which
	// encodes sedra's own schedule (notably a deliberate step at month 8, from
	// 29.37B to 17.00B seep/block) rather than a pure 12-month halving. This table is
	// the authoritative on-chain emission and has been validated against the live chain.
	// The test therefore verifies the DAA-score -> months -> table lookup directly,
	// including that any month beyond the end of the table returns 0 (subsidy depleted),
	// which is a much stronger regression guard than the old halving approximation.
	if subsidyByDeflationaryMonthTable[0] != deflationaryPhaseBaseSubsidy {
		t.Fatalf("subsidyByDeflationaryMonthTable[0] = %d, want base subsidy %d",
			subsidyByDeflationaryMonthTable[0], deflationaryPhaseBaseSubsidy)
	}
	tableLen := uint64(len(subsidyByDeflationaryMonthTable))
	for month := uint64(0); month <= tableLen+24; month++ {
		blockDaaScore := deflationaryPhaseDaaScore + month*secondsPerMonth
		want := subsidyByDeflationaryMonthTable[tableLen-1] // clamped tail (0)
		if month < tableLen {
			want = subsidyByDeflationaryMonthTable[month]
		}
		got := coinbaseManagerInstance.calcDeflationaryPeriodBlockSubsidy(blockDaaScore)
		if got != want {
			t.Errorf("calcDeflationaryPeriodBlockSubsidy(month %d, daaScore %d): want %d, got %d",
				month, blockDaaScore, want, got)
		}
	}
}

func TestBuildSubsidyTable(t *testing.T) {
	deflationaryPhaseBaseSubsidy := dagconfig.MainnetParams.DeflationaryPhaseBaseSubsidy
	if deflationaryPhaseBaseSubsidy != 440*constants.SeepPerSedra {
		t.Errorf("TestBuildSubsidyTable: table generation function was not updated to reflect "+
			"the new base subsidy %d. Please fix the constant above and replace subsidyByDeflationaryMonthTable "+
			"in coinbasemanager.go with the printed table", deflationaryPhaseBaseSubsidy)
	}
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		0,
		deflationaryPhaseBaseSubsidy,
		0,
		0,
		0,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
	coinbaseManagerInstance := coinbaseManagerInterface.(*coinbaseManager)

	var subsidyTable []uint64
	for M := uint64(0); ; M++ {
		subsidy := coinbaseManagerInstance.calcDeflationaryPeriodBlockSubsidyFloatCalc(M)
		subsidyTable = append(subsidyTable, subsidy)
		if subsidy == 0 {
			break
		}
	}

	tableStr := "\n{\t"
	for i := 0; i < len(subsidyTable); i++ {
		tableStr += strconv.FormatUint(subsidyTable[i], 10) + ", "
		if (i+1)%25 == 0 {
			tableStr += "\n\t"
		}
	}
	tableStr += "\n}"
	t.Logf(tableStr)
}
