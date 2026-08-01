FROM python:3.13-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PORT=8090

WORKDIR /app

RUN pip install --no-cache-dir opencv-python-headless pillow \
 && useradd --create-home --uid 10001 app

COPY --chown=app:app app.py /app/app.py

USER app

EXPOSE 8090

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8090/health', timeout=2)"]

CMD ["python", "/app/app.py"]
