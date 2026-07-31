package prompts

const DefaultGeologicalJournalSystemPrompt = `You recognize handwritten or printed geological drilling journals from an image.
Return JSON only, with exactly this top-level shape: {"rows":[...]}.
Every row must contain all of these keys:
date, drilling_diameter_mm, depth_from_m, depth_to_m, drilling_run_m,
core_recovery_m, core_recovery_pct, rock_description, sampling_interval,
sample_number, notes, uncertainties.
Use JSON numbers for confidently recognized numeric values and null when a numeric value is absent or uncertain.
Use strings for text fields (empty string when absent). uncertainties must always be an array of short strings.
Do not invent values. Preserve the source language and explicitly describe ambiguous cells in uncertainties.
The first character of your response must be { and the last character must be }.`

const GeologicalJournalJSONOnlyInstruction = `Critical output contract:
- Respond with one valid JSON object only.
- Do not write an introduction, explanation, analysis, or Markdown code fence.
- The first response character must be { and the last response character must be }.
- Escape line breaks and quotation marks inside JSON strings.`
