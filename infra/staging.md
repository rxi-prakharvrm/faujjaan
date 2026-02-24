# Staging deployment (free-tier friendly)

Target stack:

- **DB**: Supabase Postgres
- **API**: Google Cloud Run (container)
- **Web**: Vercel (Next.js)

## 1) Supabase Postgres

1. Create a Supabase project and grab the **connection string** for Postgres.
2. Apply migrations:

```bash
export DATABASE_URL="postgres://..."
./infra/migrate.sh up
```

## 2) Cloud Run (Go API)

### Build & deploy

From the repo root:

```bash
gcloud auth login
gcloud config set project <your-gcp-project>

gcloud run deploy clothes-shop-api \
  --source ./api \
  --region <region> \
  --allow-unauthenticated \
  --set-env-vars "DATABASE_URL=$DATABASE_URL,JWT_SECRET=<strong-secret>,AUTO_MIGRATE=0,DEV_ALLOW_ALL_CORS=false,ALLOWED_CORS_ORIGIN=<vercel-url>,RAZORPAY_KEY_ID=<id>,RAZORPAY_KEY_SECRET=<secret>,RAZORPAY_WEBHOOK_SECRET=<webhook-secret>"
```

Notes:

- `AUTO_MIGRATE=0` on Cloud Run. Run migrations via `./infra/migrate.sh up` during releases instead.
- Set `ALLOWED_CORS_ORIGIN` to your Vercel domain (e.g. `https://yourapp.vercel.app`).

### Observability

- Cloud Run automatically ships stdout/stderr to **Cloud Logging**.
- The API includes request logging middleware (method/path/status/latency).

## 3) Vercel (Next.js web)

1. Import the `web/` folder as a Vercel project.
2. Set environment variables:

- `NEXT_PUBLIC_API_BASE_URL`: your Cloud Run API URL (e.g. `https://clothes-shop-api-xxxx.a.run.app`)
- `API_SERVER_BASE_URL`: same as above (used for server-side rendering)

3. Deploy.

## 4) Razorpay webhook

In Razorpay Dashboard:

- Create a webhook pointing to `POST /v1/webhooks/razorpay` on your API, e.g.:
  - `https://<cloud-run-url>/v1/webhooks/razorpay`
- Set webhook secret to match `RAZORPAY_WEBHOOK_SECRET`.
- Enable events:
  - `payment.authorized`
  - `payment.captured`
  - `payment.failed`

