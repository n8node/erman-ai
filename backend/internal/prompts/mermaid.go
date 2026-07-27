package prompts

// MermaidDiagramRules is appended to strategy prompts so LLM output renders in Mermaid 11.
const MermaidDiagramRules = `
MERMAID DIAGRAM RULES (mandatory — invalid syntax breaks the UI):
- diagrams[].mermaid must contain ONLY raw Mermaid source: NO markdown fences, NO ` + "`" + `mermaid, NO HTML tags, NO comments
- Line 1 MUST declare diagram type: use "flowchart TD" or "flowchart LR" only (do NOT use timeline, gantt, sequenceDiagram, classDiagram)
- Node IDs: ASCII letters/digits only (A, B, C1, phase1). NO spaces, NO Cyrillic, NO punctuation in IDs
- Labels: put readable text inside brackets with the ASCII id: A[Сбор данных] --> B[Обработка заявок]
- One edge or node per line; use --> for arrows; max 8 nodes and max 40 characters per label
- Do NOT use: subgraph, classDef, style, click, linkStyle, emoji, parentheses in node IDs, semicolon chains
- In JSON strings use \n for newlines and escape double quotes inside labels as \"
- Self-check before output: every line must be valid flowchart syntax

Example architecture (diagram 1):
flowchart TD
  A[Источники данных] --> B[ETL и качество]
  B --> C[Хранилище данных]
  C --> D[AI сервисы]
  D --> E[Бизнес-пользователи]

Example roadmap (diagram 2):
flowchart LR
  Q1[Пилот] --> Q2[Масштабирование]
  Q2 --> Q3[Оптимизация]
  Q3 --> Q4[Автоматизация]`
