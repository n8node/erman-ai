.PHONY: dev prod down migrate migrate-down test lint logs logs-prod \
        psql wp-cli deploy deploy-dev backup-db backup-wp status ssl-renew setup first-deploy

first-deploy:
	bash scripts/server-first-deploy.sh

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d

down:
	docker compose down

migrate:
	docker compose exec backend goose -dir ./migrations postgres "$$DATABASE_URL" up

migrate-down:
	docker compose exec backend goose -dir ./migrations postgres "$$DATABASE_URL" down

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
	docker compose -f docker-compose.yml -f docker-compose.prod.yml ps

ssl-renew:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml exec nginx certbot renew
	docker compose -f docker-compose.yml -f docker-compose.prod.yml exec nginx nginx -s reload

setup:
	bash scripts/setup.sh

%:
	@:
