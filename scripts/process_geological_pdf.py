#!/usr/bin/env python3
"""Local, provider-neutral PDF pipeline for geological journals.

The script keeps the source PDF unchanged and creates per-page artifacts:
orientation candidates, the selected oriented image, OCR text/TSV, content
classification, detected table geometry, and a document manifest.

OCR is intentionally optional. Install Tesseract separately for local OCR and
pass --ocr tesseract. Without it, the script still renders pages, extracts
embedded PDF text, detects table-like lines, and produces a review manifest.
"""

from __future__ import annotations

import argparse
import csv
import io
import json
import re
import shutil
import statistics
import sys
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any


try:
    import fitz  # type: ignore
except ImportError:  # pragma: no cover - dependency check is user-facing
    fitz = None

try:
    import cv2  # type: ignore
    import numpy as np  # type: ignore
except ImportError:  # pragma: no cover - dependency check is user-facing
    cv2 = None
    np = None

try:
    from PIL import Image
except ImportError:  # pragma: no cover - dependency check is user-facing
    Image = None

try:
    import pytesseract  # type: ignore
except ImportError:  # pragma: no cover - dependency check is user-facing
    pytesseract = None

try:
    from paddleocr import PaddleOCR  # type: ignore
except ImportError:  # pragma: no cover - optional engine
    PaddleOCR = None

try:
    from rapidocr_onnxruntime import RapidOCR  # type: ignore
except ImportError:  # pragma: no cover - optional engine
    RapidOCR = None


DEPTH_PATTERN = re.compile(r"(?<!\d)(\d{1,5}[.,]\d{1,3}|\d{1,5})(?:\s*[-–—]\s*)(\d{1,5}[.,]\d{1,3}|\d{1,5})(?!\d)")
DATE_PATTERN = re.compile(r"\b(?:\d{1,2}[./-]){1,2}\d{2,4}\b")
SAMPLE_PATTERN = re.compile(r"(?i)\b(?:проба|обр(?:азец)?|sample|№|no\.?)[\s№:#-]*[A-Za-zА-Яа-я0-9/-]+")
GEOLOGY_TERMS = (
    "скваж", "керн", "порода", "глуб", "интервал", "опроб", "литолог",
    "бурен", "геолог", "depth", "core", "sample", "rock", "drilling",
)


@dataclass
class OrientationResult:
    correction_degrees: int
    score: float
    method: str
    confidence: float
    reason: str


@dataclass
class PageAnalysis:
    page_number: int
    source_rotation: int
    selected_orientation: OrientationResult
    content_type: str
    content_confidence: float
    width: int
    height: int
    embedded_text_chars: int
    ocr_text_chars: int
    ocr_available: bool
    table_like: bool
    detected_table_count: int
    depth_intervals: list[str]
    dates: list[str]
    sample_references: list[str]
    warnings: list[str]


def fail(message: str) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    raise SystemExit(2)


def resolve_tesseract_executable() -> str | None:
    """Find Tesseract even when the current terminal has an old PATH."""
    candidates = [
        shutil.which("tesseract"),
        r"C:\Program Files\Tesseract-OCR\tesseract.exe",
        r"C:\Program Files (x86)\Tesseract-OCR\tesseract.exe",
        str(Path.home() / "AppData/Local/Programs/Tesseract-OCR/tesseract.exe"),
    ]
    for candidate in candidates:
        if candidate and Path(candidate).is_file():
            return candidate
    return None


def require_dependencies(ocr_mode: str) -> None:
    missing = []
    if fitz is None:
        missing.append("PyMuPDF")
    if Image is None:
        missing.append("Pillow")
    if cv2 is None or np is None:
        missing.append("opencv-python-headless and numpy")
    if ocr_mode == "tesseract" and pytesseract is None:
        missing.append("pytesseract")
    if ocr_mode == "paddle" and PaddleOCR is None:
        missing.append("paddleocr and paddlepaddle")
    if ocr_mode == "rapid" and RapidOCR is None:
        missing.append("rapidocr_onnxruntime")
    if missing:
        fail("missing Python packages: " + ", ".join(missing) + ". Install with: pip install " + " ".join({
            "PyMuPDF": "PyMuPDF",
            "Pillow": "Pillow",
            "opencv-python-headless and numpy": "opencv-python-headless numpy",
            "pytesseract": "pytesseract",
            "paddleocr and paddlepaddle": "paddleocr paddlepaddle",
            "rapidocr_onnxruntime": "rapidocr_onnxruntime",
        }[item] for item in missing))
    if ocr_mode == "tesseract" and resolve_tesseract_executable() is None:
        fail("Tesseract executable was not found in PATH. Install Tesseract OCR and Russian language data, or use --ocr none.")
    if ocr_mode == "tesseract" and pytesseract is not None:
        pytesseract.pytesseract.tesseract_cmd = resolve_tesseract_executable() or "tesseract"


def parse_pages(raw: str, total: int) -> list[int]:
    raw = raw.strip().lower()
    if raw in {"", "all", "*"}:
        return list(range(1, total + 1))
    selected: set[int] = set()
    for part in raw.split(","):
        token = part.strip()
        if not token:
            continue
        if "-" in token:
            left, right = token.split("-", 1)
            try:
                start, end = int(left), int(right)
            except ValueError:
                fail(f"invalid page range: {token!r}")
            if start > end:
                start, end = end, start
            selected.update(range(start, end + 1))
        else:
            try:
                selected.add(int(token))
            except ValueError:
                fail(f"invalid page number: {token!r}")
    invalid = sorted(page for page in selected if page < 1 or page > total)
    if invalid:
        fail(f"page numbers outside document: {invalid}; document has {total} pages")
    return sorted(selected)


def render_page(page: Any, output: Path, dpi: int) -> Image.Image:
    matrix = fitz.Matrix(dpi / 72.0, dpi / 72.0)
    pixmap = page.get_pixmap(matrix=matrix, alpha=False)
    pixmap.save(str(output))
    return Image.open(output).convert("RGB")


def rotate_image(image: Image.Image, correction_degrees: int) -> Image.Image:
    # Pillow uses counter-clockwise positive angles. The metadata and output
    # report use the same convention: the correction applied to the image.
    return image.rotate(correction_degrees, expand=True, fillcolor="white")


def run_tesseract(image: Image.Image, language: str) -> tuple[str, str, dict[str, Any]]:
    if pytesseract is None:
        return "", "", {"engine": "none", "words": [], "mean_confidence": 0.0}
    config = "--oem 1 --psm 11"
    data = pytesseract.image_to_data(
        image, lang=language, config=config, output_type=pytesseract.Output.DICT
    )
    words = []
    buffer = io.StringIO()
    writer = csv.DictWriter(buffer, fieldnames=["text", "left", "top", "width", "height", "confidence", "engine"], delimiter="\t", lineterminator="\n")
    writer.writeheader()
    row_count = len(data.get("text", []))
    for index in range(row_count):
        text = str(data["text"][index]).strip()
        if not text:
            continue
        word = {"text": text, "left": int(data["left"][index]), "top": int(data["top"][index]), "width": int(data["width"][index]), "height": int(data["height"][index]), "confidence": max(0.0, float(data["conf"][index]) / 100.0), "engine": "tesseract"}
        words.append(word)
        writer.writerow(word)
    words.sort(key=lambda item: (item["top"], item["left"]))
    return "\n".join(item["text"] for item in words), buffer.getvalue(), {"engine": "tesseract", "words": words, "mean_confidence": statistics.mean([item["confidence"] for item in words]) if words else 0.0}


def run_ocr(image: Image.Image, engine: str, language: str) -> tuple[str, str, dict[str, Any]]:
    if engine == "tesseract":
        return run_tesseract(image, language)
    if engine == "paddle":
        if PaddleOCR is None:
            raise RuntimeError("PaddleOCR is not installed")
        if not hasattr(run_ocr, "paddle"):
            run_ocr.paddle = PaddleOCR(lang="ru", use_doc_orientation_classify=False, use_doc_unwarping=False, use_textline_orientation=True)
        result = run_ocr.paddle.predict(np.asarray(image))
        words = []
        for page in result:
            payload = page.json if hasattr(page, "json") else page
            if isinstance(payload, str):
                payload = json.loads(payload)
            data = payload.get("res", payload) if isinstance(payload, dict) else {}
            for text, score, box in zip(data.get("rec_texts", []), data.get("rec_scores", []), data.get("rec_boxes", [])):
                points = np.asarray(box).reshape(-1, 2)
                x, y = points.min(axis=0); x2, y2 = points.max(axis=0)
                words.append({"text": str(text), "left": int(x), "top": int(y), "width": int(x2 - x), "height": int(y2 - y), "confidence": float(score), "engine": "paddle"})
        words.sort(key=lambda item: (item["top"], item["left"]))
        return "\n".join(item["text"] for item in words), "", {"engine": "paddle", "words": words, "mean_confidence": statistics.mean([item["confidence"] for item in words]) if words else 0.0}
    if engine == "rapid":
        if RapidOCR is None:
            raise RuntimeError("RapidOCR is not installed")
        if not hasattr(run_ocr, "rapid"):
            run_ocr.rapid = RapidOCR()
        result, _ = run_ocr.rapid(np.asarray(image))
        words = []
        for box, text, score in result or []:
            points = np.asarray(box).reshape(-1, 2)
            x, y = points.min(axis=0); x2, y2 = points.max(axis=0)
            words.append({"text": str(text), "left": int(x), "top": int(y), "width": int(x2 - x), "height": int(y2 - y), "confidence": float(score), "engine": "rapid"})
        words.sort(key=lambda item: (item["top"], item["left"]))
        return "\n".join(item["text"] for item in words), "", {"engine": "rapid", "words": words, "mean_confidence": statistics.mean([item["confidence"] for item in words]) if words else 0.0}
    raise ValueError(f"unsupported OCR engine: {engine}")


def orientation_score(text: str, image: Image.Image) -> float:
    normalized = text.lower()
    if not normalized:
        return 0.0
    words = re.findall(r"[a-zа-яё]{2,}", normalized)
    terms = sum(normalized.count(term) for term in GEOLOGY_TERMS)
    numeric = len(re.findall(r"\d", normalized))
    lines = len([line for line in text.splitlines() if line.strip()])
    width, height = image.size
    aspect_bonus = 0.2 if width >= height else 0.0
    score = min(1.0, 0.15 * min(len(words) / 40, 1) + 0.25 * min(terms / 3, 1) +
                0.25 * min(numeric / 20, 1) + 0.2 * min(lines / 20, 1) + aspect_bonus)
    return round(score, 4)


def choose_orientation(image: Image.Image, embedded_text: str, language: str, ocr_mode: str) -> tuple[OrientationResult, dict[int, str]]:
    if ocr_mode == "none":
        # The embedded PDF text is not enough to reliably infer a scan rotation,
        # so preserve the page and make the uncertainty explicit.
        return OrientationResult(0, 0.0, "metadata_only", 0.0, "OCR disabled; orientation requires manual review"), {}

    candidates: dict[int, str] = {}
    scores: list[tuple[int, float]] = []
    for correction in (0, 90, 180, 270):
        candidate = rotate_image(image, correction)
        text, _, result = run_ocr(candidate, ocr_mode, language)
        candidates[correction] = text.strip()
        score = orientation_score(text, candidate) + float(result["mean_confidence"]) * 0.25
        scores.append((correction, score))
    scores.sort(key=lambda item: item[1], reverse=True)
    best_angle, best_score = scores[0]
    second_score = scores[1][1] if len(scores) > 1 else 0.0
    confidence = round(min(1.0, max(0.0, best_score + (best_score - second_score))), 4)
    reason = "best OCR readability score"
    if embedded_text.strip():
        reason += "; embedded PDF text is also available"
    return OrientationResult(best_angle, best_score, f"{ocr_mode}_rotation_candidates", confidence, reason), candidates


def remove_table_lines(image: Image.Image) -> Image.Image:
    array = cv2.cvtColor(np.asarray(image), cv2.COLOR_RGB2GRAY)
    binary = cv2.threshold(array, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)[1]
    horizontal = cv2.morphologyEx(binary, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (max(30, image.width // 25), 1)))
    vertical = cv2.morphologyEx(binary, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(30, image.height // 25))))
    return Image.fromarray(cv2.bitwise_not(cv2.bitwise_and(binary, cv2.bitwise_not(cv2.bitwise_or(horizontal, vertical))))).convert("RGB")


def detect_tables(image: Image.Image) -> list[dict[str, Any]]:
    if cv2 is None or np is None:
        return []
    array = cv2.cvtColor(np.asarray(image), cv2.COLOR_RGB2GRAY)
    threshold = cv2.adaptiveThreshold(
        array, 255, cv2.ADAPTIVE_THRESH_MEAN_C, cv2.THRESH_BINARY_INV, 31, 15
    )
    horizontal_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (max(20, image.width // 30), 1))
    vertical_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(20, image.height // 30)))
    horizontal = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, horizontal_kernel)
    vertical = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, vertical_kernel)
    grid = cv2.add(horizontal, vertical)
    contours, _ = cv2.findContours(grid, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    tables = []
    page_area = image.width * image.height
    for contour in contours:
        x, y, width, height = cv2.boundingRect(contour)
        area = width * height
        if area < page_area * 0.015 or width < 120 or height < 80:
            continue
        line_pixels = int(cv2.countNonZero(grid[y:y + height, x:x + width]))
        if line_pixels < 100:
            continue
        tables.append({
            "bbox_px": [int(x), int(y), int(x + width), int(y + height)],
            "line_pixels": line_pixels,
            "area_ratio": round(area / page_area, 5),
        })
    tables.sort(key=lambda item: item["area_ratio"], reverse=True)
    return tables[:20]


def classify_page(text: str, tables: list[dict[str, Any]], image: Image.Image) -> tuple[str, float]:
    chars = len(text.strip())
    table_signal = bool(tables)
    long_text_signal = chars >= 500
    if table_signal and long_text_signal:
        return "mixed", 0.78
    if table_signal:
        return "table", 0.76
    if long_text_signal:
        return "free_text", 0.72
    if chars < 20 and image.width * image.height > 100_000:
        return "unknown", 0.35
    return "free_text", 0.48


def extract_facts(text: str) -> dict[str, list[str]]:
    intervals = []
    for match in DEPTH_PATTERN.finditer(text):
        value = f"{match.group(1)}–{match.group(2)}".replace(",", ".")
        if value not in intervals:
            intervals.append(value)
    dates = list(dict.fromkeys(DATE_PATTERN.findall(text)))
    samples = list(dict.fromkeys(SAMPLE_PATTERN.findall(text)))
    return {"depth_intervals": intervals, "dates": dates, "sample_references": samples}


def write_csv(path: Path, rows: list[dict[str, Any]]) -> None:
    if not rows:
        return
    with path.open("w", newline="", encoding="utf-8-sig") as handle:
        writer = csv.DictWriter(handle, fieldnames=list(rows[0].keys()))
        writer.writeheader()
        writer.writerows(rows)


def build_summary(analyses: list[PageAnalysis], output: Path) -> None:
    by_type: dict[str, int] = {}
    for item in analyses:
        by_type[item.content_type] = by_type.get(item.content_type, 0) + 1
    lines = [
        "# Geological journal PDF analysis",
        "",
        f"Pages processed: {len(analyses)}",
        "",
        "## Content types",
        "",
    ]
    lines.extend(f"- `{key}`: {value}" for key, value in sorted(by_type.items()))
    lines.extend(["", "## Pages requiring review", ""])
    review = [item for item in analyses if item.selected_orientation.confidence < 0.55 or item.warnings]
    if review:
        for item in review:
            lines.append(
                f"- Page {item.page_number}: orientation confidence "
                f"{item.selected_orientation.confidence:.2f}; "
                f"type `{item.content_type}`; warnings: {', '.join(item.warnings) or 'none'}"
            )
    else:
        lines.append("- No pages were flagged by the automatic checks.")
    lines.extend([
        "",
        "## Important",
        "",
        "This report is a local preprocessing and OCR analysis. It does not replace",
        "human verification of geological values or the final document summary.",
    ])
    (output / "summary.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def process(args: argparse.Namespace) -> None:
    require_dependencies(args.ocr)
    input_path = Path(args.input).expanduser().resolve()
    output = Path(args.output).expanduser().resolve()
    if not input_path.is_file():
        fail(f"input PDF not found: {input_path}")
    output.mkdir(parents=True, exist_ok=True)

    document = fitz.open(str(input_path))
    pages = parse_pages(args.pages, document.page_count)
    analyses: list[PageAnalysis] = []
    manifest: dict[str, Any] = {
        "input": str(input_path),
        "page_count": document.page_count,
        "selected_pages": pages,
        "dpi": args.dpi,
        "ocr": args.ocr,
        "ocr_language": args.language,
        "pages": [],
    }

    for page_number in pages:
        page = document.load_page(page_number - 1)
        page_dir = output / f"page-{page_number:04d}"
        page_dir.mkdir(parents=True, exist_ok=True)
        original_path = page_dir / "original.png"
        image = render_page(page, original_path, args.dpi)
        embedded_text = page.get_text("text").strip()
        (page_dir / "embedded-text.txt").write_text(embedded_text + "\n", encoding="utf-8")

        selected, candidate_texts = choose_orientation(image, embedded_text, args.language, args.ocr)
        oriented = rotate_image(image, selected.correction_degrees)
        oriented_path = page_dir / "oriented.png"
        oriented.save(oriented_path)
        for degrees, text in candidate_texts.items():
            (page_dir / f"ocr-candidate-{degrees:03d}.txt").write_text(text + "\n", encoding="utf-8")

        ocr_text = embedded_text
        ocr_tsv = ""
        ocr_result: dict[str, Any] = {"engine": "none", "words": [], "mean_confidence": 0.0}
        if args.ocr != "none":
            ocr_text, ocr_tsv, ocr_result = run_ocr(remove_table_lines(oriented), args.ocr, args.language)
            (page_dir / "ocr.txt").write_text(ocr_text + "\n", encoding="utf-8")
            (page_dir / "ocr.tsv").write_text(ocr_tsv, encoding="utf-8")
            (page_dir / "ocr.json").write_text(json.dumps(ocr_result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        else:
            (page_dir / "ocr.txt").write_text(embedded_text + "\n", encoding="utf-8")

        tables = detect_tables(oriented)
        (page_dir / "tables.json").write_text(json.dumps(tables, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        write_csv(page_dir / "table-regions.csv", tables)
        content_type, content_confidence = classify_page(ocr_text, tables, oriented)
        facts = extract_facts(ocr_text)
        warnings = []
        if selected.confidence < 0.55:
            warnings.append("orientation_needs_review")
        if not ocr_text.strip():
            warnings.append("no_text_detected")
        if content_type == "table" and not tables:
            warnings.append("table_classification_without_geometry")

        analysis = PageAnalysis(
            page_number=page_number,
            source_rotation=int(page.rotation),
            selected_orientation=selected,
            content_type=content_type,
            content_confidence=content_confidence,
            width=oriented.width,
            height=oriented.height,
            embedded_text_chars=len(embedded_text),
            ocr_text_chars=len(ocr_text),
            ocr_available=args.ocr != "none",
            table_like=bool(tables),
            detected_table_count=len(tables),
            depth_intervals=facts["depth_intervals"],
            dates=facts["dates"],
            sample_references=facts["sample_references"],
            warnings=warnings,
        )
        (page_dir / "analysis.json").write_text(json.dumps(asdict(analysis), ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        analyses.append(analysis)
        manifest["pages"].append(asdict(analysis))
        print(f"page {page_number}: {content_type}, rotation={selected.correction_degrees}°, confidence={selected.confidence:.2f}")

    manifest["pages_processed"] = len(analyses)
    (output / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    build_summary(analyses, output)
    document.close()
    print(f"Done. Results: {output}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Analyze and OCR geological journal PDF pages locally")
    parser.add_argument("input", help="input PDF path")
    parser.add_argument("-o", "--output", default="artifacts/geological-journal", help="output directory")
    parser.add_argument("--pages", default="all", help="pages/ranges, e.g. 1,3-5,9 or all")
    parser.add_argument("--dpi", type=int, default=300, help="render resolution, default: 300")
    parser.add_argument("--ocr", choices=("none", "paddle", "rapid", "tesseract"), default="none", help="OCR engine")
    parser.add_argument("--language", default="rus+eng", help="Tesseract languages, default: rus+eng")
    return parser


if __name__ == "__main__":
    process(build_parser().parse_args())