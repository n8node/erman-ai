package service

import (
	"math"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const (
	geologicalJournalDepthToleranceM   = 0.2
	geologicalJournalRunToleranceM     = 0.25
	geologicalJournalRunToleranceRatio = 0.12
	geologicalJournalMinDiameterMM     = 20
	geologicalJournalMaxDiameterMM     = 500
)

var geologicalJournalDatePattern = regexp.MustCompile(`^(\d{1,2}[./-]\d{1,2}([./-]\d{2,4})?|\d{4}[./-]\d{1,2}[./-]\d{1,2})$`)

// ValidateGeologicalJournalOutput adds deterministic field_issues and row-level
// uncertainties without mutating recognized values. Issue values are stable codes
// for frontend i18n (see geologicalJournal.validation.*).
func ValidateGeologicalJournalOutput(output *model.GeologicalJournalOutput) {
	if output == nil {
		return
	}
	for i := range output.Rows {
		issues := validateGeologicalJournalRow(&output.Rows[i], i, output.Rows)
		output.Rows[i].FieldIssues = issues
		mergeGeologicalJournalUncertainties(&output.Rows[i], issues)
	}
}

func validateGeologicalJournalRow(
	row *model.GeologicalJournalRow,
	index int,
	all []model.GeologicalJournalRow,
) map[string][]string {
	issues := make(map[string][]string)
	addIssue := func(field, code string) {
		if strings.TrimSpace(code) == "" {
			return
		}
		issues[field] = append(issues[field], code)
	}

	hasDepthData := row.DepthFromM != nil || row.DepthToM != nil || row.DrillingRunM != nil
	if hasDepthData && strings.TrimSpace(row.Date) == "" {
		addIssue("date", "missing_date")
	} else if strings.TrimSpace(row.Date) != "" && !geologicalJournalDatePattern.MatchString(strings.TrimSpace(row.Date)) {
		addIssue("date", "invalid_date")
	}

	if row.DrillingDiameterMM != nil {
		diameter := *row.DrillingDiameterMM
		if diameter < geologicalJournalMinDiameterMM || diameter > geologicalJournalMaxDiameterMM {
			addIssue("drilling_diameter_mm", "diameter_range")
		}
	}

	if row.DepthFromM != nil && row.DepthToM != nil {
		from := *row.DepthFromM
		to := *row.DepthToM
		if to < from {
			addIssue("depth_to_m", "depth_order")
			addIssue("depth_from_m", "depth_order")
		}
		if row.DrillingRunM != nil {
			expectedRun := to - from
			if !geologicalJournalApproxEqual(*row.DrillingRunM, expectedRun, geologicalJournalRunToleranceM, geologicalJournalRunToleranceRatio) {
				addIssue("drilling_run_m", "run_mismatch")
			}
		}
	}

	if row.DrillingRunM != nil && row.CoreRecoveryM != nil {
		if *row.CoreRecoveryM > *row.DrillingRunM+geologicalJournalRunToleranceM {
			addIssue("core_recovery_m", "core_m_exceeds_run")
		}
	}

	if row.CoreRecoveryPct != nil {
		pct := *row.CoreRecoveryPct
		if pct < 0 || pct > 100 {
			addIssue("core_recovery_pct", "core_pct_range")
		}
	}

	if row.DrillingRunM != nil && row.CoreRecoveryM != nil && row.CoreRecoveryPct != nil && *row.DrillingRunM > 0 {
		expectedPct := (*row.CoreRecoveryM / *row.DrillingRunM) * 100
		if math.Abs(expectedPct-*row.CoreRecoveryPct) > 8 {
			addIssue("core_recovery_pct", "core_pct_inconsistent")
		}
	}

	if index > 0 && row.DepthFromM != nil {
		prev := all[index-1]
		if prev.DepthToM != nil && *row.DepthFromM < *prev.DepthToM-geologicalJournalDepthToleranceM {
			addIssue("depth_from_m", "depth_continuity")
		}
	}

	if len(strings.TrimSpace(row.RockDescription)) == 0 && hasDepthData {
		addIssue("rock_description", "missing_rock_description")
	}

	if len(issues) == 0 {
		return nil
	}
	return issues
}

func mergeGeologicalJournalUncertainties(row *model.GeologicalJournalRow, issues map[string][]string) {
	if len(issues) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(row.Uncertainties))
	for _, item := range row.Uncertainties {
		seen[item] = struct{}{}
	}
	for field, codes := range issues {
		for _, code := range codes {
			entry := field + ":" + code
			if _, ok := seen[entry]; ok {
				continue
			}
			seen[entry] = struct{}{}
			row.Uncertainties = append(row.Uncertainties, entry)
		}
	}
}

func geologicalJournalApproxEqual(actual, expected, absTol, ratioTol float64) bool {
	if math.Abs(actual-expected) <= absTol {
		return true
	}
	if expected == 0 {
		return math.Abs(actual) <= absTol
	}
	return math.Abs(actual-expected)/math.Abs(expected) <= ratioTol
}
