# BE Assignment (Go + Gin + Postgres)

This service supports ordering a limited-stock product and running a background settlement job.

## Quick start

Start the entire stack (Postgres + auto-migrations + app) with one command:

```bash
DATABASE_URL=postgres://postgres:password@db:5432/appdb?sslmode=disable docker compose up -d
```

What this does:

- Brings up Postgres on localhost:5432
- Runs database migrations automatically using the `migrate` service
- Starts the app on http://localhost:8080

Optional:

- View logs: `docker compose logs -f app`

To stop everything:

```bash
docker compose down
```

If you prefer running locally without Docker, you can still:

```bash
go mod tidy
go run .
```

The server runs on :8080 by default.

## Postman collection

You can import the Postman collection to try the endpoints quickly.

Steps:

1. Import `Indico Case Study.postman_collection.json` into Postman.
2. Set the collection or environment variable `url` to `http://localhost:8080`.
3. Use the requests in order:
   - `health` (GET) → should return status ok
   - `orders` (POST) → creates an order for the seeded product
   - `jobs settlement` (POST) → starts a settlement job for a date range
   - `job detail` (GET) → replace the job id in the path with the id returned from the previous step
   - `downloads` (GET) → once the job completes, download the CSV at the `download_url` returned by `job detail`

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
- Settlement job processes ~1M rows via batching and workers; CSV is written to `./tmp/settlements/<job_id>.csv` on the host (mounted into the container at `/app/tmp/settlements`) and is exposed under `/downloads`.
