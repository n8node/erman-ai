.PHONY: dev prod down migrate migrate-down test lint logs logs-prod \
        psql wp-cli deploy deploy-dev backup-db backup-wp status ssl-renew setup first-deploy

COMPOSE := docker compose --env-file .env
COMPOSE_PROD := $(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml

first-deploy:
	bash scripts/server-first-deploy.sh

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

prod:
	$(COMPOSE_PROD) up --build -d

down:
	$(COMPOSE) down

migrate:
	$(COMPOSE_PROD) exec -T backend sh -c 'goose -dir ./migrations postgres "$$DATABASE_URL" up'

migrate-down:
	$(COMPOSE_PROD) exec -T backend sh -c 'goose -dir ./migrations postgres "$$DATABASE_URL" down'

test:
	cd backend && go test ./...

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

logs:
	docker compose logs -f

logs-prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

psql:
	docker compose exec postgres psql -U erman_ai erman_ai

wp-cli:
	docker compose exec wordpress wp --allow-root $(filter-out $@,$(MAKECMDGOALS))

deploy:
	bash scripts/deploy.sh main

deploy-dev:
	bash scripts/deploy.sh develop

backup-db:
	docker compose exec postgres pg_dump -U erman_ai erman_ai | gzip > backups/pg_$$(date +%Y%m%d_%H%M%S).sql.gz

backup-wp:
	docker compose exec mysql mysqldump -u wordpress -p$$WP_DB_PASSWORD wordpress | gzip > backups/wp_$$(date +%Y%m%d_%H%M%S).sql.gz

status:
	$(COMPOSE_PROD) ps

ssl-renew:
	$(COMPOSE_PROD) exec nginx certbot renew
	$(COMPOSE_PROD) exec nginx nginx -s reload

setup:
	bash scripts/setup.sh

%:
	@:
