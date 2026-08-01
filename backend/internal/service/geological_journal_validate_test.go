package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/stretchr/testify/require"
)

func floatPtr(v float64) *float64 { return &v }

func TestValidateGeologicalJournalOutputDepthOrder(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{
				Date:            "24.07.25",
				DepthFromM:      floatPtr(10),
				DepthToM:        floatPtr(9),
				DrillingRunM:    floatPtr(1),
				RockDescription: "granite",
				Uncertainties:   []string{},
			},
		},
	}
	ValidateGeologicalJournalOutput(output)
	require.Contains(t, output.Rows[0].FieldIssues["depth_to_m"], "depth_order")
	require.NotEmpty(t, output.Rows[0].Uncertainties)
}

func TestValidateGeologicalJournalOutputRunMatchesDepth(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{
				Date:            "24.07.25",
				DepthFromM:      floatPtr(1),
				DepthToM:        floatPtr(2),
				DrillingRunM:    floatPtr(2),
				CoreRecoveryM:   floatPtr(0.5),
				CoreRecoveryPct: floatPtr(50),
				RockDescription: "granite",
				Uncertainties:   []string{},
			},
		},
	}
	ValidateGeologicalJournalOutput(output)
	require.Contains(t, output.Rows[0].FieldIssues["drilling_run_m"], "run_mismatch")
}

func TestValidateGeologicalJournalOutputCoreRecovery(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{
				Date:            "24.07.25",
				DepthFromM:      floatPtr(1),
				DepthToM:        floatPtr(2),
				DrillingRunM:    floatPtr(1),
				CoreRecoveryM:   floatPtr(1.5),
				CoreRecoveryPct: floatPtr(150),
				RockDescription: "granite",
				Uncertainties:   []string{},
			},
		},
	}
	ValidateGeologicalJournalOutput(output)
	require.Contains(t, output.Rows[0].FieldIssues["core_recovery_m"], "core_m_exceeds_run")
	require.Contains(t, output.Rows[0].FieldIssues["core_recovery_pct"], "core_pct_range")
}

func TestValidateGeologicalJournalOutputRowContinuity(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{
				Date:            "24.07.25",
				DepthFromM:      floatPtr(1),
				DepthToM:        floatPtr(5),
				DrillingRunM:    floatPtr(4),
				RockDescription: "granite",
				Uncertainties:   []string{},
			},
			{
				Date:            "24.07.26",
				DepthFromM:      floatPtr(3),
				DepthToM:        floatPtr(6),
				DrillingRunM:    floatPtr(3),
				RockDescription: "slate",
				Uncertainties:   []string{},
			},
		},
	}
	ValidateGeologicalJournalOutput(output)
	require.Contains(t, output.Rows[1].FieldIssues["depth_from_m"], "depth_continuity")
}

func TestValidateGeologicalJournalOutputValidRow(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{
				Date:               "24.07.25",
				DrillingDiameterMM: floatPtr(76),
				DepthFromM:         floatPtr(155.4),
				DepthToM:           floatPtr(156.2),
				DrillingRunM:       floatPtr(0.8),
				CoreRecoveryM:      floatPtr(0.7),
				CoreRecoveryPct:    floatPtr(87),
				RockDescription:    "Сланцы",
				Uncertainties:      []string{},
			},
		},
	}
	ValidateGeologicalJournalOutput(output)
	require.Nil(t, output.Rows[0].FieldIssues)
}
