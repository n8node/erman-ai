# Промпт для Cursor — Фаза 1: Erman AI

Вставь в Cursor:

---

Прочитай все правила в `.cursor/rules/` — там 5 файлов, описывающих проект Erman AI:
- `erman-ai-rules.mdc` — архитектура, стек, критические правила
- `erman-ai-project.mdc` — продукт, роли, инструменты, биллинг
- `erman-ai-design.mdc` — дизайн-система, цвета, компоненты
- `erman-ai-deploy.mdc` — деплой, Git, инфраструктура
- `erman-ai-llm.mdc` — LLM, воркеры, приватность
- `erman-ai-backlog.mdc` — backlog, директивы, roadmap

Выполни **Фазу 1: Скелет и инфраструктура.** Создай ВСЁ за один проход:

---

**1. Backend (Go):**
- Go-модуль `github.com/{username}/erman-ai` (замени {username} на мой GitHub username)
- Зависимости: chi, pgx/v5, goose, caarlos0/env, go-playground/validator, golang-jwt/jwt
- `cmd/server/main.go` — точка входа, запуск HTTP-сервера, подключение к Postgres, запуск миграций
- `internal/config/config.go` — чтение конфигурации из env
- `internal/server/server.go` — chi-роутер, middleware, маунт handler'ов
- `internal/middleware/logging.go` — slog middleware
- `internal/middleware/cors.go` — CORS
- `internal/middleware/auth.go` — JWT auth middleware
- `internal/handler/health.go` — `GET /health` → `{"status":"ok","version":"0.1.0"}`
- `internal/repository/postgres.go` — подключение через pgx pool
- `Dockerfile` (production, multi-stage, alpine)
- `Dockerfile.dev` (с air для hot reload)
- `.air.toml`
- `migrations/001_init.up.sql` — таблицы: users, api_keys, plans, subscriptions, tool_runs, usage_log, tool_usage_counters
- `migrations/001_init.down.sql`

**2. Frontend (Next.js):**
- Next.js 15, App Router, TypeScript, Tailwind CSS
- `next.config.ts`: `basePath: '/dashboard'`, `output: 'standalone'`
- shadcn/ui, lucide-react, next-intl (EN + RU)
- `src/app/layout.tsx` — минималистичный layout, светлая тема, Inter
- `src/app/page.tsx` — редирект на `/dashboard`
- `src/app/dashboard/page.tsx` — "Erman AI Dashboard — Coming soon"
- `Dockerfile` и `Dockerfile.dev`

**3. WordPress:**
- `wordpress/Dockerfile` (wordpress:6-php8.3-apache + wp-cli)

**4. Nginx:**
- `nginx/Dockerfile` (nginx:1.27-alpine)
- `nginx/nginx.conf` — базовый (worker_processes auto, gzip on)
- `nginx/conf.d/default.conf` — роутинг:
  - `/api/` → backend:8080 (timeout 120s, body 50M)
  - `/health` → backend
  - `/dashboard` → frontend:3000 (websocket upgrade)
  - `/_next/` → frontend (cache 365d)
  - `/wp-admin/`, `/wp-content/`, `/wp-includes/`, `/wp-json/`, `/wp-login.php` → wordpress
  - `/` → wordpress
  - SSL server block — закомментированный, готовый к включению (домен: erman.ai)
- `nginx/ssl/.gitkeep`

**5. Docker Compose:**
- `docker-compose.yml` — nginx, backend, frontend, wordpress, postgres:16, mysql:8
- `docker-compose.prod.yml` — replicas, resource limits, logging, nats, dragonfly, minio (закомментированы)
- Порты наружу ТОЛЬКО у nginx (80, 443)
- Healthcheck для postgres и mysql
- Volumes: postgres_data, mysql_data, wp_uploads

**6. Инфраструктурные файлы:**
- `.env.example` — ВСЕ переменные (Postgres, MySQL/WP, OpenRouter, Backend, Frontend, Workers, MinIO, Domain)
- `.gitignore` — полный (Go, Node, Docker, .env, ssl, uploads, backups, IDE, OS, artifacts)
- `Makefile` — команды: dev, prod, down, migrate, migrate-down, test, lint, logs, logs-prod, psql, wp-cli, deploy, backup-db, backup-wp, status, ssl-renew
- `scripts/deploy.sh` — скрипт деплоя на сервер (плейсхолдер SERVER IP, домен erman.ai)
- `scripts/backup.sh` — скрипт бэкапов
- `scripts/setup.sh` — первичная настройка (директории, chmod)
- `scripts/ssl-init.sh` — скрипт получения Let's Encrypt сертификата
- `scripts/ssl-deploy-hook.sh` — хук для автопродления
- `backups/.gitkeep`

**7. Git инициализация:**
- `git init`
- Первый коммит: `feat: initial project structure (Phase 1)`
- Покажи команды для remote и push на GitHub

---

**Критерий успеха Фазы 1:**
1. `cp .env.example .env` + заполнить пароли → `make dev` → все 6 контейнеров стартуют
2. `curl http://localhost/health` → `{"status":"ok","version":"0.1.0"}`
3. `http://localhost/dashboard` → Next.js страница
4. `http://localhost/` → WordPress setup
5. `git log` показывает первый коммит

Не пропускай ни один файл. После завершения покажи:
- Список всех созданных файлов
- Команды для запуска: `make dev`
- Команды для git push на GitHub
- Команды для первого деплоя на сервер
