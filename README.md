# BE Assignment (Go + Gin + Postgres)

This service supports ordering a limited-stock product and running a background settlement job.

## Quick start

1. Start Postgres via Docker Compose

```bash
DATABASE_URL=postgres://postgres:password@db:5432/appdb?sslmode=disable docker compose up -d
```

2. Apply migrations (using psql)

```bash
# Ensure psql is installed; adjust host/port/user/password if needed
psql postgresql://postgres:password@localhost:5432/appdb -f migrations/001_init.sql
psql postgresql://postgres:password@localhost:5432/appdb -f migrations/002_seed.sql
psql postgresql://postgres:password@localhost:5432/appdb -f migrations/003_seed_transactions.sql
psql postgresql://postgres:password@localhost:5432/appdb -f migrations/004_add_job_range.sql
```

3. Run the server

```bash
# Install deps and start
go mod tidy
go run ./...
```

Server runs on :8080 by default.

## Endpoints

- POST `/orders` → create an order: `{ "product_id":1, "quantity":1, "buyer_id":"user-123" }`
- GET `/orders/:id` → fetch order details
- POST `/jobs/settlement` → start a settlement job: `{ "from":"2025-01-01", "to":"2025-01-31" }`
- GET `/jobs/:id` → job status
- POST `/jobs/:id/cancel` → request cancel
- Download CSV when completed via `download_url` in job status

## Configuration

Environment variables:

- `PORT` (default `8080`)
- `DATABASE_URL` (default `postgres://postgres:password@localhost:5432/appdb?sslmode=disable`)
- `WORKERS` (default `8`) number of job workers used for settlement fan-out

## Notes

- Concurrency-safe ordering uses a DB transaction with `SELECT ... FOR UPDATE` on the product row to avoid overselling.
- Settlement job processes ~1M rows via batching and workers; CSV is written to `/tmp/settlements/<job_id>.csv` and exposed under `/downloads`.
