#!/bin/bash
set -e

source /opt/erman-ai/.env
DOMAIN="${DOMAIN:-erman.ai}"

cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" /opt/erman-ai/nginx/ssl/fullchain.pem
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" /opt/erman-ai/nginx/ssl/privkey.pem

docker compose -f /opt/erman-ai/docker-compose.yml \
  -f /opt/erman-ai/docker-compose.prod.yml \
  exec nginx nginx -s reload
