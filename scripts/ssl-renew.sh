#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck disable=SC1091
source .env

COMPOSE="docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml"

mkdir -p certbot-webroot

certbot renew --webroot -w "$(pwd)/certbot-webroot"

bash scripts/ssl-deploy-hook.sh

echo "Renewal complete"
