# Migrations

This repo stores SQL migrations in `api/migrations/`.

In dev, the API will run migrations automatically when `AUTO_MIGRATE=1`.

Files:

- `000001_init.*.sql`: base schema
- `000002_seed_dev.*.sql`: dev seed (admin user + sample products)

