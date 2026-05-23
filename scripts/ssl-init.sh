#!/bin/bash
set -e

source .env

DOMAIN="${DOMAIN:-erman.ai}"
EMAIL="${CERTBOT_EMAIL:-hello@erman.ai}"

echo "=== Requesting SSL certificate for $DOMAIN ==="

docker compose -f docker-compose.yml -f docker-compose.prod.yml stop nginx

certbot certonly --standalone \
  -d "$DOMAIN" \
  -d "www.$DOMAIN" \
  --email "$EMAIL" \
  --agree-tos \
  --no-eff-email \
  --non-interactive

mkdir -p nginx/ssl
cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/fullchain.pem
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/privkey.pem

docker compose -f docker-compose.yml -f docker-compose.prod.yml start nginx

echo "=== SSL configured! Uncomment HTTPS block in nginx/conf.d/default.conf ==="
