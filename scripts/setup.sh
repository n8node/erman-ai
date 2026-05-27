#!/bin/bash
set -e

echo "=== Erman AI setup ==="

mkdir -p backups nginx/ssl wordpress/uploads artifacts
chmod 775 wordpress/uploads 2>/dev/null || true
chmod +x scripts/*.sh 2>/dev/null || true

if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example — fill in secrets before make dev"
fi

echo "=== Setup complete ==="
