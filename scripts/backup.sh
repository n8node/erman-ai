#!/bin/bash
set -e

BACKUP_DIR="/opt/erman-ai/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

docker compose -f /opt/erman-ai/docker-compose.yml \
  -f /opt/erman-ai/docker-compose.prod.yml \
  exec -T postgres pg_dump -U erman_ai erman_ai \
  | gzip > "$BACKUP_DIR/pg_$DATE.sql.gz"

docker compose -f /opt/erman-ai/docker-compose.yml \
  -f /opt/erman-ai/docker-compose.prod.yml \
  exec -T mysql mysqldump -u wordpress -p"$WP_DB_PASSWORD" wordpress \
  | gzip > "$BACKUP_DIR/wp_$DATE.sql.gz"

docker compose -f /opt/erman-ai/docker-compose.yml \
  -f /opt/erman-ai/docker-compose.prod.yml \
  exec -T affine-postgres \
  pg_dump -U "${AFFINE_DB_USER:-affine}" "${AFFINE_DB_NAME:-affine}" \
  | gzip > "$BACKUP_DIR/affine_pg_$DATE.sql.gz"

docker compose -f /opt/erman-ai/docker-compose.yml \
  -f /opt/erman-ai/docker-compose.prod.yml \
  run --rm --no-deps \
  -v "$BACKUP_DIR:/backup" \
  affine sh -c "tar -czf /backup/affine_files_$DATE.tar.gz -C /root/.affine storage config"

find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $DATE"
