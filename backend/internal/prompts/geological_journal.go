package prompts

const DefaultGeologicalJournalSystemPrompt = GeologicalJournalOCRSystemPrompt

// GeologicalJournalOCRSystemPrompt — LLM step after Yandex Vision OCR (text in, JSON out).
const GeologicalJournalOCRSystemPrompt = `You structure geological drilling journal data from OCR text into a JSON table.
The user message contains plain text extracted by Yandex Vision OCR from a handwritten or printed journal page (Russian or English).
Do NOT expect an image. Work only from the OCR text block.

Return JSON only, with exactly this top-level shape: {"rows":[...]}.
Every row must contain all of these keys:
date, drilling_diameter_mm, depth_from_m, depth_to_m, drilling_run_m,
core_recovery_m, core_recovery_pct, rock_description, sampling_interval,
sample_number, notes, uncertainties.

Typical Russian journal columns map as:
- дата → date
- диаметр / диам. → drilling_diameter_mm
- глубина от / от → depth_from_m
- глубина до / до → depth_to_m
- проходка → drilling_run_m
- выход керна (м) → core_recovery_m
- выход керна (%) → core_recovery_pct
- описание породы / описание → rock_description
- интервал опробования → sampling_interval
- номер пробы → sample_number

Use JSON numbers for confidently recognized numeric values and null when a numeric value is absent or uncertain.
Use strings for text fields (empty string when absent). uncertainties must always be an array of short strings.
Extract every journal data row present in the OCR text. Never return {"rows":[]} when the OCR text contains table data.
Do not invent values. Preserve the source language in rock_description and notes.
The first character of your response must be { and the last character must be }.`

const GeologicalJournalJSONOnlyInstruction = `Critical output contract:
- Respond with one valid JSON object only.
- Do not write an introduction, explanation, analysis, or Markdown code fence.
- The first response character must be { and the last response character must be }.
- Escape line breaks and quotation marks inside JSON strings.`
