import type { GeologicalJournalRow } from "./api-geological-journal";

export type GeologicalJournalIssueCode =
  | "missing_date"
  | "invalid_date"
  | "diameter_range"
  | "depth_order"
  | "run_mismatch"
  | "core_m_exceeds_run"
  | "core_pct_range"
  | "core_pct_inconsistent"
  | "depth_continuity"
  | "missing_rock_description";

export type GeologicalJournalFieldKey = Exclude<
  keyof GeologicalJournalRow,
  "uncertainties" | "field_issues"
>;

export type GeologicalJournalFieldIssues = Partial<
  Record<GeologicalJournalFieldKey, GeologicalJournalIssueCode[]>
>;

const DEPTH_TOLERANCE_M = 0.2;
const RUN_TOLERANCE_M = 0.25;
const RUN_TOLERANCE_RATIO = 0.12;
const MIN_DIAMETER_MM = 20;
const MAX_DIAMETER_MM = 500;

const DATE_PATTERN =
  /^(\d{1,2}[./-]\d{1,2}([./-]\d{2,4})?|\d{4}[./-]\d{1,2}[./-]\d{1,2})$/;

function approxEqual(
  actual: number,
  expected: number,
  absTol: number,
  ratioTol: number
) {
  if (Math.abs(actual - expected) <= absTol) return true;
  if (expected === 0) return Math.abs(actual) <= absTol;
  return Math.abs(actual - expected) / Math.abs(expected) <= ratioTol;
}

export function validateGeologicalJournalRow(
  row: GeologicalJournalRow,
  index: number,
  allRows: GeologicalJournalRow[]
): GeologicalJournalFieldIssues {
  const issues: GeologicalJournalFieldIssues = {};
  const addIssue = (field: GeologicalJournalFieldKey, code: GeologicalJournalIssueCode) => {
    issues[field] = [...(issues[field] ?? []), code];
  };

  const hasDepthData =
    row.depth_from_m != null || row.depth_to_m != null || row.drilling_run_m != null;

  if (hasDepthData && !row.date.trim()) {
    addIssue("date", "missing_date");
  } else if (row.date.trim() && !DATE_PATTERN.test(row.date.trim())) {
    addIssue("date", "invalid_date");
  }

  if (row.drilling_diameter_mm != null) {
    const diameter = row.drilling_diameter_mm;
    if (diameter < MIN_DIAMETER_MM || diameter > MAX_DIAMETER_MM) {
      addIssue("drilling_diameter_mm", "diameter_range");
    }
  }

  if (row.depth_from_m != null && row.depth_to_m != null) {
    const from = row.depth_from_m;
    const to = row.depth_to_m;
    if (to < from) {
      addIssue("depth_to_m", "depth_order");
      addIssue("depth_from_m", "depth_order");
    }
    if (row.drilling_run_m != null) {
      const expectedRun = to - from;
      if (
        !approxEqual(
          row.drilling_run_m,
          expectedRun,
          RUN_TOLERANCE_M,
          RUN_TOLERANCE_RATIO
        )
      ) {
        addIssue("drilling_run_m", "run_mismatch");
      }
    }
  }

  if (row.drilling_run_m != null && row.core_recovery_m != null) {
    if (row.core_recovery_m > row.drilling_run_m + RUN_TOLERANCE_M) {
      addIssue("core_recovery_m", "core_m_exceeds_run");
    }
  }

  if (row.core_recovery_pct != null) {
    if (row.core_recovery_pct < 0 || row.core_recovery_pct > 100) {
      addIssue("core_recovery_pct", "core_pct_range");
    }
  }

  if (
    row.drilling_run_m != null &&
    row.core_recovery_m != null &&
    row.core_recovery_pct != null &&
    row.drilling_run_m > 0
  ) {
    const expectedPct = (row.core_recovery_m / row.drilling_run_m) * 100;
    if (Math.abs(expectedPct - row.core_recovery_pct) > 8) {
      addIssue("core_recovery_pct", "core_pct_inconsistent");
    }
  }

  if (index > 0 && row.depth_from_m != null) {
    const prev = allRows[index - 1];
    if (prev.depth_to_m != null && row.depth_from_m < prev.depth_to_m - DEPTH_TOLERANCE_M) {
      addIssue("depth_from_m", "depth_continuity");
    }
  }

  if (!row.rock_description.trim() && hasDepthData) {
    addIssue("rock_description", "missing_rock_description");
  }

  return issues;
}

function mergeIssueCodes(
  left: GeologicalJournalFieldIssues,
  right: GeologicalJournalFieldIssues
): GeologicalJournalFieldIssues {
  const merged: GeologicalJournalFieldIssues = { ...left };
  for (const [field, codes] of Object.entries(right) as Array<
    [GeologicalJournalFieldKey, GeologicalJournalIssueCode[]]
  >) {
    merged[field] = [...new Set([...(merged[field] ?? []), ...codes])];
  }
  return merged;
}

export function rowFieldIssues(
  row: GeologicalJournalRow,
  index: number,
  allRows: GeologicalJournalRow[]
): GeologicalJournalFieldIssues {
  const serverIssues = (row.field_issues ?? {}) as GeologicalJournalFieldIssues;
  const liveIssues = validateGeologicalJournalRow(row, index, allRows);
  return mergeIssueCodes(serverIssues, liveIssues);
}

export function countIssueCells(rows: GeologicalJournalRow[]) {
  let count = 0;
  rows.forEach((row, index) => {
    count += Object.keys(rowFieldIssues(row, index, rows)).length;
  });
  return count;
}

export function countIssueRows(rows: GeologicalJournalRow[]) {
  return rows.reduce((total, row, index) => {
    return total + (Object.keys(rowFieldIssues(row, index, rows)).length > 0 ? 1 : 0);
  }, 0);
}

export function firstIssueRowIndex(rows: GeologicalJournalRow[]) {
  return rows.findIndex(
    (row, index) => Object.keys(rowFieldIssues(row, index, rows)).length > 0
  );
}

export function rowHasIssues(
  row: GeologicalJournalRow,
  index: number,
  allRows: GeologicalJournalRow[]
) {
  return Object.keys(rowFieldIssues(row, index, allRows)).length > 0;
}

export function issueLabels(
  codes: GeologicalJournalIssueCode[],
  translate: (code: GeologicalJournalIssueCode) => string
) {
  return [...new Set(codes)].map(translate);
}
