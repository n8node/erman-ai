#!/bin/bash
# One-off fix when wp-content/uploads is not writable (theme/media install fails).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE="docker compose"
if [ -f docker-compose.prod.yml ] && [ "${USE_PROD_COMPOSE:-}" = "1" ]; then
  COMPOSE="docker compose -f docker-compose.yml -f docker-compose.prod.yml"
fi

echo ">>> Fixing WordPress wp-content permissions..."
$COMPOSE exec -T -u root wordpress bash -c '
  for dir in uploads themes plugins upgrade cache; do
    mkdir -p "/var/www/html/wp-content/${dir}"
    chown -R www-data:www-data "/var/www/html/wp-content/${dir}"
    chmod -R ug+rwX "/var/www/html/wp-content/${dir}"
  done
  echo "Done. uploads owner:"
  ls -la /var/www/html/wp-content/uploads
'

echo ">>> Fix complete."
