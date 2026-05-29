.PHONY: up down down-volumes logs logs-api logs-ingester restart rebuild \
        up-prod down-prod logs-prod

# ── Dev ───────────────────────────────────────────────────────────────────────

up:
	docker compose -f compose.yml up --build -d

down:
	docker compose -f compose.yml down

logs-api:
	docker compose -f compose.yml logs -f api

logs-ingester:
	docker compose -f compose.yml logs -f ingester

restart:
	docker compose -f compose.yml restart

rebuild:
	docker compose -f compose.yml up --build --force-recreate -d
