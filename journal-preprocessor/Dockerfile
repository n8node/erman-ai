FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
  PORT=8090 \
  OMP_THREAD_LIMIT=1 \
  OMP_NUM_THREADS=1 \
  TESSERACT_NUM_THREADS=1 \
  FLAGS_use_mkldnn=0 \
  PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK=True

WORKDIR /app

RUN apt-get update \
 && apt-get install -y --no-install-recommends libgl1 tesseract-ocr tesseract-ocr-rus tesseract-ocr-eng \
 && rm -rf /var/lib/apt/lists/* \
 && pip install --no-cache-dir opencv-python-headless pillow PyMuPDF pytesseract \
      paddlepaddle paddleocr rapidocr_onnxruntime \
 && useradd --create-home --uid 10001 app

COPY --chown=app:app app.py /app/app.py

VOLUME ["/app/data/geological-journal"]

# The backend and preprocessor share page artifacts in a Docker volume. Backend
# creates document directories as root, so running this internal worker as the
# unprivileged app user prevents traversal of existing page directories. The
# HTTP service is not published outside the Compose network.
USER root

EXPOSE 8090

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8090/health', timeout=2)"]

CMD ["python", "/app/app.py"]
