#!/usr/bin/env bash
# First deploy on a fresh Ubuntu VPS. Run as root or user with sudo + docker.
# Usage: curl -fsSL ... | bash   OR   bash scripts/server-first-deploy.sh
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/opt/erman-ai}"
REPO="${REPO:-https://github.com/n8node/erman-ai.git}"
BRANCH="${BRANCH:-main}"
DOMAIN="${DOMAIN:-erman.ai}"

echo "=== Erman AI — first deploy ==="
echo "Project: $PROJECT_DIR"
echo "Domain:  $DOMAIN ($(getent hosts "$DOMAIN" 2>/dev/null | awk '{print $1}' || echo 'DNS?'))"

if ! command -v docker >/dev/null 2>&1; then
  echo ">>> Installing Docker..."
  curl -fsSL https://get.docker.com | sh
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "ERROR: docker compose plugin not found"
  exit 1
fi

if [ ! -d "$PROJECT_DIR/.git" ]; then
  echo ">>> Cloning repository..."
  sudo mkdir -p "$(dirname "$PROJECT_DIR")"
  sudo git clone --branch "$BRANCH" "$REPO" "$PROJECT_DIR"
  sudo chown -R "$USER:$USER" "$PROJECT_DIR" 2>/dev/null || true
fi

cd "$PROJECT_DIR"
git fetch origin
git checkout "$BRANCH"
git pull origin "$BRANCH"

if [ ! -f .env ]; then
  echo ">>> Creating .env with generated secrets..."
  cp .env.example .env
  PG_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
  WP_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
  WP_ROOT=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
  JWT=$(openssl rand -base64 48 | tr -d '\n')
  SALT=$(openssl rand -base64 24 | tr -d '\n')

  sed -i "s|change_me_postgres|$PG_PASS|g" .env
  sed -i "s|change_me_wp_root|$WP_ROOT|g" .env
  sed -i "s|change_me_wp|$WP_PASS|g" .env
  sed -i "s|change_me_jwt_secret_min_32_chars|$JWT|g" .env
  sed -i "s|change_me_api_key_salt|$SALT|g" .env
  sed -i "s|postgres://erman_ai:change_me_postgres@|postgres://erman_ai:${PG_PASS}@|g" .env
  sed -i "s|ENVIRONMENT=development|ENVIRONMENT=production|g" .env

  echo ""
  echo ">>> .env created. Add OPENROUTER_API_KEY later if needed."
  echo ">>> Save these credentials securely!"
fi

chmod +x scripts/*.sh 2>/dev/null || true

echo ">>> Building and starting containers (this may take 10–15 min)..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d

echo ">>> Waiting for postgres..."
for i in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U erman_ai -d erman_ai >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo ">>> Running migrations..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec -T backend \
  goose -dir ./migrations postgres "$(
    grep '^DATABASE_URL=' .env | cut -d= -f2-
  )" up

echo ">>> Container status:"
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps

echo ""
echo ">>> Health check (HTTP):"
sleep 5
if curl -sf "http://127.0.0.1/health"; then
  echo ""
  echo "OK — backend is up"
else
  echo "WARNING: health check failed — check: docker compose logs backend nginx"
fi

echo ""
echo "=== Done ==="
echo "  Landing:   http://${DOMAIN}/          (WordPress setup on first visit)"
echo "  Dashboard: http://${DOMAIN}/dashboard"
echo "  Health:    http://${DOMAIN}/health"
echo ""
echo "Next steps:"
echo "  1. Complete WordPress install at http://${DOMAIN}/"
echo "  2. SSL: ./scripts/ssl-init.sh (after DNS points to this server)"
echo "  3. Superadmin: docker compose exec backend ./server --seed-admin --email=admin@erman.ai --password=YOUR_PASSWORD"
echo "     (seed-admin CLI — Phase 2)"
