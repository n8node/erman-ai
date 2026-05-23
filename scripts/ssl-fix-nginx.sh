#!/usr/bin/env bash
# Rebuild nginx when SSL certs already exist (skip certbot).
set -euo pipefail

cd "$(dirname "$0")/.."

COMPOSE="docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml"

if [[ ! -f nginx/ssl/fullchain.pem ]]; then
  echo "No certificate found. Run: bash scripts/ssl-init.sh"
  exit 1
fi

$COMPOSE up -d --build nginx
sleep 2
$COMPOSE exec nginx nginx -t
curl -sf "https://erman.ai/health" && echo "" || $COMPOSE logs --tail 30 nginx
