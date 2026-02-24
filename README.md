# Clothes Shop (MVP)

Monorepo for a production-shaped ecommerce MVP:

- `web/`: Next.js storefront + admin UI
- `api/`: Go API
- `docs/`: requirements + OpenAPI
- `infra/`: local/dev/deploy notes

## Local development (Docker Compose)

1) Copy env file and adjust as needed:

```bash
cp .env.example .env
```

2) Start services:

```bash
docker compose up --build
```

3) URLs (defaults):

- Web: `http://localhost:3001`
- API: `http://localhost:8081` (health: `/healthz`)

## Quick commands

```bash
docker compose down -v
```

