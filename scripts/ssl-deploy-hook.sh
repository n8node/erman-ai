#!/usr/bin/env bash
set -euo pipefail

cd /opt/erman-ai
# shellcheck source=lib/env.sh
source /opt/erman-ai/scripts/lib/env.sh

DOMAIN="$(env_get DOMAIN erman.ai)"
COMPOSE="docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml"

cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/fullchain.pem
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/privkey.pem
chmod 644 nginx/ssl/fullchain.pem
chmod 600 nginx/ssl/privkey.pem

$COMPOSE exec nginx nginx -s reload

echo "SSL certificates renewed and nginx reloaded"
