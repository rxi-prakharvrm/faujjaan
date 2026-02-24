.PHONY: dev down reset-db

dev:
	docker compose up --build

down:
	docker compose down

reset-db:
	docker compose down -v

