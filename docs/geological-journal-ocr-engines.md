# Локальный OCR Geological Journal

## Рекомендуемый режим

В production preprocessor по умолчанию использует RapidOCR:

```yaml
OCR_ENGINE: rapid
OCR_FALLBACK: tesseract
OCR_DPI: 300
ORIENTATION_DPI: 150
OCR_TABLE_REGIONS: true
```

PaddleOCR распознаёт текст с координатами и confidence. Для каждой страницы
сохраняются:

- `ocr.txt` — текст с восстановленными строками;
- `ocr.tsv` — слова, bbox, confidence и фактический engine;
- `ocr.json` — слова страницы и отдельные результаты по табличным областям;
- `analysis.json` — engine, средний confidence, ориентация и найденные таблицы.

## Переключение движка

`rapid` — production-вариант для CPU через ONNX Runtime. Он используется по умолчанию.

`paddle` — экспериментальный вариант для серверов, где совместимы PaddlePaddle
и CPU runtime. На текущем production CPU при запуске Paddle может выдавать ошибку
`ConvertPirAttribute2RuntimeAttribute`, поэтому включать его нужно явно.

`tesseract` — fallback/сравнение. Он запускается с `--oem 1 --psm 11`,
что лучше подходит для разреженных блоков, чем прежний `--psm 3`.

Например, для проверки RapidOCR:

```powershell
$env:OCR_ENGINE = "rapid"
$env:OCR_FALLBACK = "tesseract"
docker compose up -d --build journal-preprocessor
```

## Что делает pipeline

1. Рендерит PDF в 300 DPI.
2. Определяет ориентацию по тексту и confidence OCR.
3. Выполняет perspective correction и deskew.
4. Создаёт копию изображения без длинных линий таблиц для OCR.
5. Распознаёт страницу и до восьми крупных табличных областей отдельно.
6. Сохраняет bbox/confidence, чтобы следующие этапы собирали строки и ячейки,
   а не передавали LLM одну неструктурированную строку.

## Проверка

После запуска:

```powershell
curl http://localhost:8090/health
```

В ответе должны быть `ocr_engine`, `available_ocr_engines` и `ocr_dpi`.
Если PaddleOCR временно недоступен, результат явно содержит фактически
использованный fallback engine и `fallback_errors`.

OCR-результат всё равно требует проверки геологических значений. Особенно
важные числовые ячейки должны подтверждаться по изображению и confidence.