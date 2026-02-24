# Production cutover checklist

## Before go-live

- **Secrets**
  - Generate a strong `JWT_SECRET` (do not reuse staging).
  - Use production Razorpay keys and webhook secret.
  - Ensure secrets are stored only in Vercel/Cloud Run secret managers.

- **Database**
  - Confirm migrations are applied on production DB (`./infra/migrate.sh up`).
  - Enable daily backups / PITR (Supabase: verify backup plan).
  - Verify least-privileged DB credentials where possible.

- **CORS**
  - Set `DEV_ALLOW_ALL_CORS=false`.
  - Set `ALLOWED_CORS_ORIGIN` to the production domain.

- **Razorpay**
  - Webhook points to production API endpoint.
  - Webhook secret matches `RAZORPAY_WEBHOOK_SECRET`.
  - Verify events enabled: `payment.authorized`, `payment.captured`, `payment.failed`.

- **Operational readiness**
  - Confirm `/healthz` is reachable.
  - Confirm admin login works.
  - Create a test order end-to-end (Razorpay test mode), verify webhook updates order/payment.

## Launch day

- Deploy API (Cloud Run) + Web (Vercel) to production.
- Switch Razorpay keys from test → live (if applicable).
- Update DNS / domain mappings.
- Monitor Cloud Run logs for 4xx/5xx spikes.

## After go-live

- Rotate the seeded admin credentials (create a real admin user; disable/replace the dev seed user).
- Add rate limiting + stricter request validation for public endpoints.
- Add alerting (Cloud Monitoring) on 5xx error rate + latency.

