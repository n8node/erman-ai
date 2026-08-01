import io
import json
import logging
import math
import os
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import cv2
import numpy as np
from PIL import Image, ImageOps, UnidentifiedImageError


MAX_BODY_BYTES = 10 * 1024 * 1024
MAX_IMAGE_PIXELS = 40_000_000
MAX_OUTPUT_DIMENSION = 6000
PORT = int(os.getenv("PORT", "8090"))

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
    if lines is None:
        return image, 0.0

    angles: list[float] = []
    for x1, y1, x2, y2 in lines[:, 0][:200]:
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
        if self.path != "/health":
            self._json(HTTPStatus.NOT_FOUND, {"error": "not found"})
            return
        self._json(HTTPStatus.OK, {"status": "ok"})

    def do_POST(self) -> None:
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

    def log_message(self, format: str, *args: object) -> None:
        logger.info("%s - %s", self.address_string(), format % args)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
    logger.info("journal preprocessor listening on port %d", PORT)
    server.serve_forever()
