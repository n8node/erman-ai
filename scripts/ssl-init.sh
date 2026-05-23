#!/usr/bin/env bash
# Obtain Let's Encrypt certificate and start nginx with HTTPS.
# Run on server: cd /opt/erman-ai && bash scripts/ssl-init.sh
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
  echo "ERROR: .env not found"
  exit 1
fi

# shellcheck source=lib/env.sh
source "$(dirname "$0")/lib/env.sh"

DOMAIN="$(env_get DOMAIN erman.ai)"
EMAIL="$(env_get CERTBOT_EMAIL hello@erman.ai)"
COMPOSE="docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml"

echo "=== SSL init for $DOMAIN ==="

if command -v apt-get >/dev/null 2>&1; then
  DEBIAN_FRONTEND=noninteractive apt-get update -qq
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq certbot
fi

if ! command -v certbot >/dev/null 2>&1; then
  echo "ERROR: certbot not installed. Run: apt install certbot"
  exit 1
fi

mkdir -p nginx/ssl certbot-webroot

echo ">>> Stopping nginx (free port 80 for certbot)..."
$COMPOSE stop nginx

echo ">>> Requesting certificate..."
certbot certonly --standalone \
  -d "$DOMAIN" \
  -d "www.$DOMAIN" \
  --email "$EMAIL" \
  --agree-tos \
  --no-eff-email \
  --non-interactive \
  --preferred-challenges http

echo ">>> Installing certificate to nginx/ssl/"
cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/fullchain.pem
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/privkey.pem
chmod 644 nginx/ssl/fullchain.pem
chmod 600 nginx/ssl/privkey.pem

echo ">>> Starting nginx with HTTPS..."
$COMPOSE up -d --build nginx

sleep 3
echo ""
echo ">>> nginx config test:"
$COMPOSE exec nginx nginx -t

echo ""
echo ">>> Health check:"
if curl -sf "https://$DOMAIN/health"; then
  echo ""
else
  echo "WARNING: https health check failed"
  echo "--- nginx logs ---"
  $COMPOSE logs --tail 40 nginx
  echo "--- ssl files ---"
  ls -la nginx/ssl/ || true
fi
curl -sfI "http://$DOMAIN/health" | head -1 || true

echo ""
echo "=== SSL ready ==="
echo "  https://$DOMAIN/"
echo "  https://$DOMAIN/dashboard"
echo ""
echo "Auto-renewal cron (optional):"
echo "  0 3 * * * /opt/erman-ai/scripts/ssl-renew.sh >> /var/log/erman-ai-ssl-renew.log 2>&1"
