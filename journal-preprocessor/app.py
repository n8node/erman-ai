import io
import json
import logging
import math
import os
from pathlib import Path
import tempfile
import gc
import csv
import statistics
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import cv2
import numpy as np
import fitz
import pytesseract
from PIL import Image, ImageOps, UnidentifiedImageError

# OCR engines are loaded lazily. Importing PaddleOCR at process startup can take
# minutes or block on native runtime initialization, which must not prevent the
# HTTP health endpoint from becoming available.
PaddleOCR = None
RapidOCR = None


MAX_BODY_BYTES = 10 * 1024 * 1024
MAX_PDF_BODY_BYTES = 250 * 1024 * 1024
MAX_IMAGE_PIXELS = 40_000_000
MAX_OUTPUT_DIMENSION = 4500
PORT = int(os.getenv("PORT", "8090"))
OCR_DPI = int(os.getenv("OCR_DPI", "300"))
ORIENTATION_DPI = int(os.getenv("ORIENTATION_DPI", "150"))
TESSERACT_LANG = os.getenv("TESSERACT_LANG", "rus+eng")
OCR_ENGINE = os.getenv("OCR_ENGINE", "paddle").strip().lower()
OCR_FALLBACK = os.getenv("OCR_FALLBACK", "rapid,tesseract").strip().lower().split(",")
OCR_TABLE_REGIONS = os.getenv("OCR_TABLE_REGIONS", "true").lower() not in {"0", "false", "no"}

Image.MAX_IMAGE_PIXELS = MAX_IMAGE_PIXELS
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO").upper(),
    format="%(asctime)s %(levelname)s %(message)s",
)
logger = logging.getLogger("journal-preprocessor")


def _order_points(points: np.ndarray) -> np.ndarray:
    ordered = np.zeros((4, 2), dtype=np.float32)
    sums = points.sum(axis=1)
    differences = np.diff(points, axis=1).reshape(-1)
    ordered[0] = points[np.argmin(sums)]
    ordered[2] = points[np.argmax(sums)]
    ordered[1] = points[np.argmin(differences)]
    ordered[3] = points[np.argmax(differences)]
    return ordered


def _correct_perspective(image: np.ndarray) -> tuple[np.ndarray, bool]:
    height, width = image.shape[:2]
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    edges = cv2.Canny(blurred, 50, 150)
    contours, _ = cv2.findContours(
        edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
    )
    image_area = float(width * height)

    for contour in sorted(contours, key=cv2.contourArea, reverse=True)[:10]:
        area = cv2.contourArea(contour)
        if area < image_area * 0.60:
            break
        perimeter = cv2.arcLength(contour, True)
        polygon = cv2.approxPolyDP(contour, 0.02 * perimeter, True)
        if len(polygon) != 4 or not cv2.isContourConvex(polygon):
            continue

        points = _order_points(polygon.reshape(4, 2).astype(np.float32))
        top_left, top_right, bottom_right, bottom_left = points
        target_width = int(
            max(
                np.linalg.norm(bottom_right - bottom_left),
                np.linalg.norm(top_right - top_left),
            )
        )
        target_height = int(
            max(
                np.linalg.norm(top_right - bottom_right),
                np.linalg.norm(top_left - bottom_left),
            )
        )
        if target_width < 320 or target_height < 320:
            continue

        destination = np.array(
            [
                [0, 0],
                [target_width - 1, 0],
                [target_width - 1, target_height - 1],
                [0, target_height - 1],
            ],
            dtype=np.float32,
        )
        matrix = cv2.getPerspectiveTransform(points, destination)
        return (
            cv2.warpPerspective(
                image,
                matrix,
                (target_width, target_height),
                borderMode=cv2.BORDER_REPLICATE,
            ),
            True,
        )
    return image, False


def _deskew(image: np.ndarray) -> tuple[np.ndarray, float]:
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    edges = cv2.Canny(gray, 50, 150, apertureSize=3)
    minimum_length = max(80, int(min(image.shape[:2]) * 0.18))
    lines = cv2.HoughLinesP(
        edges,
        1,
        np.pi / 180,
        threshold=80,
        minLineLength=minimum_length,
        maxLineGap=20,
    )
    if lines is None or len(lines) == 0:
        return image, 0.0

    angles: list[float] = []
    for x1, y1, x2, y2 in lines.reshape(-1, 4)[:200]:
        angle = math.degrees(math.atan2(y2 - y1, x2 - x1))
        if angle < -45:
            angle += 90
        elif angle > 45:
            angle -= 90
        if abs(angle) <= 10:
            angles.append(angle)
    if len(angles) < 3:
        return image, 0.0

    angle = float(np.median(angles))
    if abs(angle) < 0.25 or abs(angle) > 8:
        return image, 0.0

    height, width = image.shape[:2]
    matrix = cv2.getRotationMatrix2D((width / 2, height / 2), angle, 1.0)
    return (
        cv2.warpAffine(
            image,
            matrix,
            (width, height),
            flags=cv2.INTER_CUBIC,
            borderMode=cv2.BORDER_REPLICATE,
        ),
        angle,
    )


def preprocess_image(raw: bytes) -> tuple[bytes, dict[str, object]]:
    try:
        with Image.open(io.BytesIO(raw)) as source:
            oriented = ImageOps.exif_transpose(source)
            width, height = oriented.size
            if width < 64 or height < 64:
                raise ValueError("image dimensions are too small")
            if width * height > MAX_IMAGE_PIXELS:
                raise ValueError("decoded image is too large")
            rgb = oriented.convert("RGB")
            image = cv2.cvtColor(np.asarray(rgb), cv2.COLOR_RGB2BGR)
    except (UnidentifiedImageError, OSError, Image.DecompressionBombError) as exc:
        raise ValueError("unsupported or damaged image") from exc

    image, perspective_corrected = _correct_perspective(image)
    image, deskew_angle = _deskew(image)

    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
    contrast = clahe.apply(gray)
    denoised = cv2.fastNlMeansDenoising(
        contrast,
        None,
        h=7,
        templateWindowSize=7,
        searchWindowSize=21,
    )

    max_dimension = max(denoised.shape[:2])
    scale = min(2.0, MAX_OUTPUT_DIMENSION / max_dimension)
    if scale > 1.05:
        denoised = cv2.resize(
            denoised,
            None,
            fx=scale,
            fy=scale,
            interpolation=cv2.INTER_CUBIC,
        )
    else:
        scale = 1.0

    blurred = cv2.GaussianBlur(denoised, (0, 0), 1.0)
    sharpened = cv2.addWeighted(denoised, 1.25, blurred, -0.25, 0)
    encoded, output = cv2.imencode(
        ".png", sharpened, [cv2.IMWRITE_PNG_COMPRESSION, 3]
    )
    if not encoded:
        raise RuntimeError("failed to encode preprocessed image")

    metadata: dict[str, object] = {
        "perspective_corrected": perspective_corrected,
        "deskew_angle": round(deskew_angle, 3),
        "scale": round(scale, 3),
        "width": int(sharpened.shape[1]),
        "height": int(sharpened.shape[0]),
    }
    return output.tobytes(), metadata


def _rotate_pil(image: Image.Image, degrees: int) -> Image.Image:
    return image.rotate(degrees, expand=True, fillcolor="white")


def _remove_table_lines(image: Image.Image) -> Image.Image:
    """Return an OCR copy with long table rules removed, preserving characters."""
    array = cv2.cvtColor(np.asarray(image), cv2.COLOR_RGB2GRAY)
    binary = cv2.threshold(array, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)[1]
    horizontal_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (max(30, image.width // 25), 1))
    vertical_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(30, image.height // 25)))
    horizontal = cv2.morphologyEx(binary, cv2.MORPH_OPEN, horizontal_kernel)
    vertical = cv2.morphologyEx(binary, cv2.MORPH_OPEN, vertical_kernel)
    lines = cv2.bitwise_or(horizontal, vertical)
    cleaned = cv2.bitwise_and(binary, cv2.bitwise_not(lines))
    cleaned = cv2.bitwise_not(cleaned)
    return Image.fromarray(cleaned).convert("RGB")


def _normalise_ocr_result(result: list[dict[str, object]], engine: str) -> dict[str, object]:
    words = [item for item in result if str(item.get("text", "")).strip()]
    words.sort(key=lambda item: (float(item.get("top", 0)), float(item.get("left", 0))))
    lines: list[str] = []
    current_line_top: float | None = None
    for word in words:
        top = float(word.get("top", 0))
        if current_line_top is None or abs(top - current_line_top) > max(8, float(word.get("height", 12)) * 0.7):
            lines.append(str(word["text"]))
            current_line_top = top
        else:
            lines[-1] += " " + str(word["text"])
    confidence_values = [float(item["confidence"]) for item in words if item.get("confidence") is not None]
    return {
        "engine": engine,
        "text": "\n".join(lines).strip(),
        "words": words,
        "mean_confidence": round(statistics.mean(confidence_values), 4) if confidence_values else 0.0,
        "word_count": len(words),
    }


def _run_tesseract(image: Image.Image) -> dict[str, object]:
    data = pytesseract.image_to_data(image, lang=TESSERACT_LANG, config="--oem 1 --psm 11", output_type=pytesseract.Output.DICT)
    words = []
    for index, text in enumerate(data.get("text", [])):
        if not str(text).strip():
            continue
        confidence = float(data["conf"][index])
        words.append({
            "text": str(text), "left": int(data["left"][index]), "top": int(data["top"][index]),
            "width": int(data["width"][index]), "height": int(data["height"][index]),
            "confidence": max(0.0, confidence / 100.0),
        })
    return _normalise_ocr_result(words, "tesseract")


def _run_paddle(image: Image.Image) -> dict[str, object]:
    global PaddleOCR
    if PaddleOCR is None:
        try:
            from paddleocr import PaddleOCR as PaddleOCREngine  # type: ignore
            PaddleOCR = PaddleOCREngine
        except Exception as exc:  # pragma: no cover - optional runtime
            raise RuntimeError(f"PaddleOCR import failed: {exc}") from exc
    if PaddleOCR is None:
        raise RuntimeError("PaddleOCR is not installed")
    if not hasattr(_run_paddle, "engine"):
        _run_paddle.engine = PaddleOCR(lang="ru", use_doc_orientation_classify=False, use_doc_unwarping=False, use_textline_orientation=True)
    result = _run_paddle.engine.predict(np.asarray(image))
    words = []
    for page in result:
        payload = page.json if hasattr(page, "json") else page
        if isinstance(payload, str):
            payload = json.loads(payload)
        data = payload.get("res", payload) if isinstance(payload, dict) else {}
        texts = data.get("rec_texts", [])
        scores = data.get("rec_scores", [])
        boxes = data.get("rec_boxes", data.get("dt_polys", []))
        for text, score, box in zip(texts, scores, boxes):
            points = np.asarray(box).reshape(-1, 2)
            x, y = points.min(axis=0)
            x2, y2 = points.max(axis=0)
            words.append({"text": str(text), "left": int(x), "top": int(y), "width": int(x2 - x), "height": int(y2 - y), "confidence": float(score)})
    return _normalise_ocr_result(words, "paddle")


def _run_rapid(image: Image.Image) -> dict[str, object]:
    global RapidOCR
    if RapidOCR is None:
        try:
            from rapidocr_onnxruntime import RapidOCR as RapidOCREngine  # type: ignore
            RapidOCR = RapidOCREngine
        except Exception as exc:  # pragma: no cover - optional runtime
            raise RuntimeError(f"RapidOCR import failed: {exc}") from exc
    if RapidOCR is None:
        raise RuntimeError("RapidOCR is not installed")
    if not hasattr(_run_rapid, "engine"):
        _run_rapid.engine = RapidOCR()
    result, _ = _run_rapid.engine(np.asarray(image))
    words = []
    for box, text, score in result or []:
        points = np.asarray(box).reshape(-1, 2)
        x, y = points.min(axis=0)
        x2, y2 = points.max(axis=0)
        words.append({"text": str(text), "left": int(x), "top": int(y), "width": int(x2 - x), "height": int(y2 - y), "confidence": float(score)})
    return _normalise_ocr_result(words, "rapid")


def run_ocr(image: Image.Image) -> dict[str, object]:
    runners = {"paddle": _run_paddle, "rapid": _run_rapid, "tesseract": _run_tesseract}
    attempts = [OCR_ENGINE] + [item for item in OCR_FALLBACK if item and item != OCR_ENGINE]
    errors = []
    for engine in attempts:
        runner = runners.get(engine)
        if runner is None:
            errors.append(f"unknown engine: {engine}")
            continue
        try:
            result = runner(image)
            if result["word_count"]:
                if errors:
                    result["fallback_errors"] = errors
                return result
        except Exception as exc:  # pragma: no cover - depends on optional runtimes
            logger.warning("OCR engine %s failed: %s", engine, exc)
            errors.append(f"{engine}: {exc}")
    return {"engine": "none", "text": "", "words": [], "mean_confidence": 0.0, "word_count": 0, "fallback_errors": errors}


def _orientation_score(text: str, image: Image.Image) -> float:
    normalized = text.lower()
    if not normalized:
        return 0.0
    words = len([word for word in normalized.split() if len(word) >= 2])
    terms = sum(normalized.count(term) for term in (
        "скваж", "керн", "пород", "глуб", "руд", "мед", "геолог",
        "depth", "core", "rock", "drill", "ore", "copper",
    ))
    digits = sum(character.isdigit() for character in normalized)
    lines = len([line for line in text.splitlines() if line.strip()])
    width, height = image.size
    return round(min(1.0, 0.15 * min(words / 40, 1) +
                     0.25 * min(terms / 3, 1) +
                     0.25 * min(digits / 20, 1) +
                     0.2 * min(lines / 20, 1) +
                     (0.15 if width >= height else 0.0)), 4)


def _table_regions(image: Image.Image) -> list[dict[str, object]]:
    array = cv2.cvtColor(np.asarray(image), cv2.COLOR_RGB2GRAY)
    threshold = cv2.adaptiveThreshold(
        array, 255, cv2.ADAPTIVE_THRESH_MEAN_C, cv2.THRESH_BINARY_INV, 31, 15
    )
    horizontal_kernel = cv2.getStructuringElement(
        cv2.MORPH_RECT, (max(20, image.width // 30), 1)
    )
    vertical_kernel = cv2.getStructuringElement(
        cv2.MORPH_RECT, (1, max(20, image.height // 30))
    )
    horizontal = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, horizontal_kernel)
    vertical = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, vertical_kernel)
    grid = cv2.add(horizontal, vertical)
    contours, _ = cv2.findContours(grid, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    page_area = image.width * image.height
    regions = []
    for contour in contours:
        x, y, width, height = cv2.boundingRect(contour)
        area = width * height
        if area < page_area * 0.015 or width < 120 or height < 80:
            continue
        line_pixels = int(cv2.countNonZero(grid[y:y + height, x:x + width]))
        if line_pixels < 100:
            continue
        regions.append({
            "bbox_px": [int(x), int(y), int(x + width), int(y + height)],
            "line_pixels": line_pixels,
            "area_ratio": round(area / page_area, 5),
        })
    regions.sort(key=lambda item: item["area_ratio"], reverse=True)
    return regions[:20]


def analyze_pdf_page(pdf_path: str, page_number: int, output_dir: str) -> dict[str, object]:
    source = Path(pdf_path)
    if not source.is_file():
        raise ValueError("pdf file not found")
    target = Path(output_dir)
    target.mkdir(parents=True, exist_ok=True)
    with fitz.open(str(source)) as document:
        if page_number < 1 or page_number > document.page_count:
            raise ValueError("page number outside document")
        page = document.load_page(page_number - 1)
        pixmap = page.get_pixmap(matrix=fitz.Matrix(OCR_DPI / 72, OCR_DPI / 72), alpha=False)
        original_path = target / "original.png"
        pixmap.save(str(original_path))
    original = Image.open(original_path).convert("RGB")
    orientation_image = original.copy()
    orientation_image.thumbnail((1600, 1600), Image.Resampling.LANCZOS)
    candidates = []
    candidate_texts = {}
    for degrees in (0, 90, 180, 270):
        candidate = _rotate_pil(orientation_image, degrees)
        result = run_ocr(candidate)
        text = str(result["text"])
        candidate_texts[degrees] = text
        candidates.append((degrees, _orientation_score(text, candidate) + float(result["mean_confidence"]) * 0.25))
        candidate.close()
    orientation_image.close()
    candidates.sort(key=lambda item: item[1], reverse=True)
    selected_degrees, score = candidates[0]
    second_score = candidates[1][1] if len(candidates) > 1 else 0.0
    confidence = round(min(1.0, max(0.0, score + score - second_score)), 4)
    oriented = _rotate_pil(original, selected_degrees)
    oriented_path = target / "oriented.png"
    oriented.save(oriented_path)
    preprocessed, metadata = preprocess_image(oriented_path.read_bytes())
    original.close()
    oriented.close()
    gc.collect()
    preprocessed_path = target / "preprocessed.png"
    preprocessed_path.write_bytes(preprocessed)
    processed_image = Image.open(preprocessed_path).convert("RGB")
    ocr_image = _remove_table_lines(processed_image)
    result = run_ocr(ocr_image)
    text = str(result["text"])
    table_results = []
    if OCR_TABLE_REGIONS:
        for region in _table_regions(processed_image)[:8]:
            x1, y1, x2, y2 = region["bbox_px"]
            crop = _remove_table_lines(processed_image.crop((x1, y1, x2, y2)))
            table_result = run_ocr(crop)
            table_result["bbox_px"] = region["bbox_px"]
            table_results.append(table_result)
    tsv_buffer = io.StringIO()
    writer = csv.DictWriter(tsv_buffer, fieldnames=["text", "left", "top", "width", "height", "confidence", "engine"], delimiter="\t", lineterminator="\n")
    writer.writeheader()
    for word in result["words"]:
        writer.writerow({**word, "engine": result["engine"]})
    tsv = tsv_buffer.getvalue()
    (target / "ocr.txt").write_text(text + "\n", encoding="utf-8")
    (target / "ocr.tsv").write_text(tsv, encoding="utf-8")
    (target / "ocr.json").write_text(json.dumps({"page": result, "table_regions": table_results}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    tables = _table_regions(processed_image)
    content_type = "mixed" if tables and len(text) >= 500 else "table" if tables else "free_text" if text else "unknown"
    analysis = {
        "page_number": page_number,
        "orientation_degrees": selected_degrees,
        "orientation_confidence": confidence,
        "orientation_candidates": {str(degrees): value for degrees, value in candidates},
        "content_type": content_type,
        "text_char_count": len(text),
        "word_count": len(text.split()),
        "tables": tables,
        "preprocessing": metadata,
        "ocr": {"engine": result["engine"], "mean_confidence": result["mean_confidence"], "word_count": result["word_count"], "table_regions": len(table_results)},
    }
    (target / "analysis.json").write_text(json.dumps(analysis, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    return {
        "status": "needs_review" if confidence < 0.55 or not text else "done",
        "phase": "done",
        "orientation_degrees": selected_degrees,
        "orientation_confidence": confidence,
        "content_type": content_type,
        "text_char_count": len(text),
        "table_count": len(tables),
        "ocr_text": text,
        "ocr_tsv": tsv,
        "ocr_engine": result["engine"],
        "ocr_confidence": result["mean_confidence"],
        "ocr_json": {"page": result, "table_regions": table_results},
        "analysis": analysis,
        "original_asset_path": str(original_path),
        "oriented_asset_path": str(oriented_path),
        "preprocessed_asset_path": str(preprocessed_path),
    }


def analyze_pdf_page_bytes(raw: bytes, page_number: int, output_dir: str) -> dict[str, object]:
    with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as handle:
        handle.write(raw)
        temporary_path = handle.name
    try:
        return analyze_pdf_page(temporary_path, page_number, output_dir)
    finally:
        try:
            os.unlink(temporary_path)
        except OSError:
            pass


def preview_pdf_page(pdf_path: str, page_number: int, output_dir: str) -> dict[str, object]:
    """Render a page and classify its layout without invoking any OCR engine."""
    source = Path(pdf_path)
    target = Path(output_dir)
    target.mkdir(parents=True, exist_ok=True)
    with fitz.open(str(source)) as document:
        if page_number < 1 or page_number > document.page_count:
            raise ValueError("page number outside document")
        page = document.load_page(page_number - 1)
        dpi = min(150, max(96, ORIENTATION_DPI))
        pixmap = page.get_pixmap(matrix=fitz.Matrix(dpi / 72, dpi / 72), alpha=False)
        original_path = target / "original.png"
        pixmap.save(str(original_path))
        source_rotation = int(page.rotation or 0)
        embedded_text = page.get_text("text").strip()
    original = Image.open(original_path).convert("RGB")
    oriented = _rotate_pil(original, source_rotation)
    oriented_path = target / "oriented.png"
    oriented.save(oriented_path)
    array = cv2.cvtColor(np.asarray(oriented), cv2.COLOR_RGB2GRAY)
    threshold = cv2.adaptiveThreshold(array, 255, cv2.ADAPTIVE_THRESH_MEAN_C, cv2.THRESH_BINARY_INV, 31, 15)
    horizontal = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (max(20, oriented.width // 30), 1)))
    vertical = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(20, oriented.height // 30))))
    grid = cv2.bitwise_or(horizontal, vertical)
    line_ratio = float(cv2.countNonZero(grid)) / max(1, array.size)
    tables = _table_regions(oriented)
    dark_ratio = float(cv2.countNonZero(threshold)) / max(1, array.size)
    content_type = "table" if tables or line_ratio > 0.012 else "free_text" if embedded_text or dark_ratio > 0.01 else "unknown"
    confidence = min(1.0, 0.45 + (0.25 if source_rotation in {0, 90, 180, 270} else 0) + min(0.3, line_ratio * 10 + dark_ratio))
    analysis = {
        "mode": "preview",
        "page_number": page_number,
        "source_rotation": source_rotation,
        "orientation_degrees": source_rotation,
        "orientation_confidence": round(confidence, 4),
        "content_type": content_type,
        "text_char_count": len(embedded_text),
        "table_count": len(tables),
        "embedded_text_available": bool(embedded_text),
        "dpi": dpi,
        "warnings": ["deep_analysis_required"],
    }
    (target / "embedded-text.txt").write_text(embedded_text + "\n", encoding="utf-8")
    (target / "analysis.json").write_text(json.dumps(analysis, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    original.close()
    oriented.close()
    return {
        "status": "needs_review", "phase": "preview", "orientation_degrees": source_rotation,
        "orientation_confidence": round(confidence, 4), "content_type": content_type,
        "text_char_count": len(embedded_text), "table_count": len(tables), "ocr_text": embedded_text,
        "analysis": analysis, "original_asset_path": str(original_path), "oriented_asset_path": str(oriented_path),
        "preprocessed_asset_path": "",
    }


def preview_pdf_document(raw: bytes, output_dir: str) -> dict[str, object]:
    """Render/classify every page in one request; never invokes OCR."""
    with fitz.open(stream=raw, filetype="pdf") as document:
        pages = []
        for page_number in range(1, document.page_count + 1):
            page_dir = Path(output_dir) / f"page-{page_number:04d}"
            page_dir.mkdir(parents=True, exist_ok=True)
            page_path = page_dir / "original.png"
            page = document.load_page(page_number - 1)
            page_pixmap = page.get_pixmap(matrix=fitz.Matrix(min(150, max(96, ORIENTATION_DPI)) / 72, min(150, max(96, ORIENTATION_DPI)) / 72), alpha=False)
            page_pixmap.save(str(page_path))
            pages.append(preview_pdf_page_from_render(page, page_path, page_number, page_dir))
    return {"pages": pages}


def preview_pdf_page_from_render(page: Any, original_path: Path, page_number: int, target: Path) -> dict[str, object]:
    original = Image.open(original_path).convert("RGB")
    source_rotation = int(page.rotation or 0)
    oriented = _rotate_pil(original, source_rotation)
    oriented_path = target / "oriented.png"
    oriented.save(oriented_path)
    array = cv2.cvtColor(np.asarray(oriented), cv2.COLOR_RGB2GRAY)
    threshold = cv2.adaptiveThreshold(array, 255, cv2.ADAPTIVE_THRESH_MEAN_C, cv2.THRESH_BINARY_INV, 31, 15)
    horizontal = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (max(20, oriented.width // 30), 1)))
    vertical = cv2.morphologyEx(threshold, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(20, oriented.height // 30))))
    line_ratio = float(cv2.countNonZero(cv2.bitwise_or(horizontal, vertical))) / max(1, array.size)
    tables = _table_regions(oriented)
    embedded_text = page.get_text("text").strip()
    dark_ratio = float(cv2.countNonZero(threshold)) / max(1, array.size)
    content_type = "table" if tables or line_ratio > 0.012 else "free_text" if embedded_text or dark_ratio > 0.01 else "unknown"
    analysis = {"mode": "preview", "page_number": page_number, "source_rotation": source_rotation, "orientation_degrees": source_rotation, "orientation_confidence": 0.7, "content_type": content_type, "text_char_count": len(embedded_text), "table_count": len(tables), "embedded_text_available": bool(embedded_text), "warnings": ["deep_analysis_required"]}
    (target / "embedded-text.txt").write_text(embedded_text + "\n", encoding="utf-8")
    (target / "analysis.json").write_text(json.dumps(analysis, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    original.close(); oriented.close()
    return {"status": "needs_review", "phase": "preview", "orientation_degrees": source_rotation, "orientation_confidence": 0.7, "content_type": content_type, "text_char_count": len(embedded_text), "table_count": len(tables), "ocr_text": embedded_text, "analysis": analysis, "original_asset_path": str(original_path), "oriented_asset_path": str(oriented_path), "preprocessed_asset_path": ""}


class Handler(BaseHTTPRequestHandler):
    server_version = "ErmanJournalPreprocessor/0.1"

    def _json(self, status: HTTPStatus, payload: dict[str, object]) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:
        if self.path == "/health":
            # Paddle/Rapid are intentionally not imported for health checks.
            available = ["paddle", "rapid", "tesseract"]
            self._json(HTTPStatus.OK, {"status": "ok", "ocr_engine": OCR_ENGINE, "available_ocr_engines": available, "ocr_dpi": OCR_DPI})
            return
        if self.path != "/health":
            self._json(HTTPStatus.NOT_FOUND, {"error": "not found"})
            return

    def do_POST(self) -> None:
        if self.path == "/pdf-info":
            self._handle_pdf_info()
            return
        if self.path == "/analyze-page":
            self._handle_analyze_page()
            return
        if self.path == "/preview-page":
            self._handle_preview_page()
            return
        if self.path == "/preview-document":
            self._handle_preview_document()
            return
        if self.path != "/preprocess":
            self._json(HTTPStatus.NOT_FOUND, {"error": "not found"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            length = 0
        if length <= 0 or length > MAX_BODY_BYTES:
            self._json(
                HTTPStatus.REQUEST_ENTITY_TOO_LARGE,
                {"error": "image must be between 1 byte and 10 MB"},
            )
            return

        raw = self.rfile.read(length)
        try:
            output, metadata = preprocess_image(raw)
        except ValueError as exc:
            self._json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})
            return
        except Exception:
            logger.exception("image preprocessing failed")
            self._json(
                HTTPStatus.INTERNAL_SERVER_ERROR,
                {"error": "image preprocessing failed"},
            )
            return

        self.send_response(HTTPStatus.OK)
        self.send_header("Content-Type", "image/png")
        self.send_header("Content-Length", str(len(output)))
        self.send_header(
            "X-Preprocessing-Metadata",
            json.dumps(metadata, separators=(",", ":")),
        )
        self.end_headers()
        self.wfile.write(output)

    def _read_json(self) -> dict[str, object]:
        length = int(self.headers.get("Content-Length", "0"))
        if length <= 0 or length > 1024 * 1024:
            raise ValueError("json request is too large")
        return json.loads(self.rfile.read(length))

    def _handle_pdf_info(self) -> None:
        try:
            if self.headers.get("Content-Type", "").lower().startswith("application/pdf"):
                raw = self._read_body(MAX_PDF_BODY_BYTES)
                with fitz.open(stream=raw, filetype="pdf") as document:
                    self._json(HTTPStatus.OK, {"page_count": document.page_count})
                return
            payload = self._read_json()
            path = str(payload.get("pdf_path", ""))
            with fitz.open(path) as document:
                self._json(HTTPStatus.OK, {"page_count": document.page_count})
        except Exception as exc:
            self._json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})

    def _handle_analyze_page(self) -> None:
        try:
            if self.headers.get("Content-Type", "").lower().startswith("application/pdf"):
                raw = self._read_body(MAX_PDF_BODY_BYTES)
                page_number = int(self.headers.get("X-Page-Number", "0"))
                output_dir = self.headers.get("X-Output-Dir", "")
                result = analyze_pdf_page_bytes(raw, page_number, output_dir)
                self._json(HTTPStatus.OK, result)
                return
            payload = self._read_json()
            result = analyze_pdf_page(
                str(payload.get("pdf_path", "")),
                int(payload.get("page_number", 0)),
                str(payload.get("output_dir", "")),
            )
            self._json(HTTPStatus.OK, result)
        except Exception as exc:
            logger.exception("pdf page analysis failed")
            self._json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})

    def _handle_preview_page(self) -> None:
        try:
            if self.headers.get("Content-Type", "").lower().startswith("application/pdf"):
                raw = self._read_body(MAX_PDF_BODY_BYTES)
                page_number = int(self.headers.get("X-Page-Number", "0"))
                output_dir = self.headers.get("X-Output-Dir", "")
                with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as handle:
                    handle.write(raw)
                    temporary_path = handle.name
                try:
                    result = preview_pdf_page(temporary_path, page_number, output_dir)
                finally:
                    os.unlink(temporary_path)
                self._json(HTTPStatus.OK, result)
                return
            payload = self._read_json()
            self._json(HTTPStatus.OK, preview_pdf_page(str(payload.get("pdf_path", "")), int(payload.get("page_number", 0)), str(payload.get("output_dir", ""))))
        except Exception as exc:
            logger.exception("pdf page preview failed")
            self._json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})

    def _handle_preview_document(self) -> None:
        try:
            raw = self._read_body(MAX_PDF_BODY_BYTES)
            output_dir = self.headers.get("X-Output-Dir", "")
            self._json(HTTPStatus.OK, preview_pdf_document(raw, output_dir))
        except Exception as exc:
            logger.exception("pdf document preview failed")
            self._json(HTTPStatus.BAD_REQUEST, {"error": str(exc)})

    def _read_body(self, maximum: int) -> bytes:
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            length = 0
        if length <= 0 or length > maximum:
            raise ValueError("request body must be between 1 byte and 250 MB")
        return self.rfile.read(length)

    def log_message(self, format: str, *args: object) -> None:
        logger.info("%s - %s", self.address_string(), format % args)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
    logger.info("journal preprocessor listening on port %d", PORT)
    server.serve_forever()
