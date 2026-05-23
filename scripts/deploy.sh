#!/bin/bash
set -e

SERVER="${DEPLOY_SERVER:-deploy@185.77.231.245}"
PROJECT_DIR="/opt/erman-ai"
BRANCH="${1:-main}"

echo "=== Deploying branch: $BRANCH ==="

git push origin "$BRANCH"

ssh "$SERVER" << EOF
  set -e
  cd $PROJECT_DIR

  echo ">>> Pulling latest code..."
  git fetch origin
  git checkout $BRANCH
  git pull origin $BRANCH

  echo ">>> Building and starting containers..."
  docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d

  echo ">>> Running migrations..."
  docker compose -f docker-compose.yml -f docker-compose.prod.yml exec -T backend \
    goose -dir ./migrations postgres "\$DATABASE_URL" up

  echo ">>> Cleaning old images..."
  docker image prune -f

  echo ">>> Health check..."
  sleep 5
  curl -sf http://127.0.0.1/health || curl -sf https://erman.ai/health || echo "WARNING: Health check failed!"

  echo ">>> Done!"
EOF

echo "=== Deploy complete ==="
