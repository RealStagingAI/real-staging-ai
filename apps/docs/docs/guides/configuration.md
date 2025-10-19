# Configuration

This document provides a detailed explanation of all the environment variables used in the Real Staging AI project.

## API Service (`api`)

| Variable                      | Description                                                                                                                                                                                 | Required | Default Value                   |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------- |
| `APP_ENV`                     | The application environment (`dev`, `test`, `production`).                                                                                                                                  | No       | `dev`                           |
| `PORT`                        | The port the API server listens on.                                                                                                                                                         | No       | `8080`                          |
| **Database**                  |                                                                                                                                                                                             |          |                                 |
| `DATABASE_URL`                | Full Postgres DSN. If set, takes precedence over PG\* vars below. Format: `postgres://user:pass@host:port/db?sslmode=disable`                                                               | Yes*     |                                 |
| `PGHOST`                      | The hostname of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                                    | Yes*     | `postgres`                      |
| `PGPORT`                      | The port of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                                        | Yes*     | `5432`                          |
| `PGUSER`                      | The username for the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                                   | Yes*     | `postgres`                      |
| `PGPASSWORD`                  | The password for the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                                   | Yes*     | `postgres`                      |
| `PGDATABASE`                  | The name of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                                        | Yes*     | `realstaging`                   |
| `PGSSLMODE`                   | Postgres SSL mode when constructing DSN from PG\* vars (`disable`, `require`, `verify-ca`, `verify-full`).                                                                                  | No       | `disable`                       |
| **Auth0**                     |                                                                                                                                                                                             |          |                                 |
| `AUTH0_DOMAIN`                | Your Auth0 domain (e.g., `your-tenant.us.auth0.com`). Required for authentication.                                                                                                         | Yes      |                                 |
| `AUTH0_AUDIENCE`              | The audience for your Auth0 API (e.g., `https://api.yourdomain.com`). Required for token validation.                                                                                       | Yes      | `https://api.realstaging.local` |
| **Redis**                     |                                                                                                                                                                                             |          |                                 |
| `REDIS_ADDR`                  | The address of the Redis server. Format: `host:port` or `redis://host:port`. Required for job queue and SSE.                                                                               | Yes      | `redis:6379`                    |
| **Job Queue**                 |                                                                                                                                                                                             |          |                                 |
| `JOB_QUEUE_NAME`              | Default Asynq queue name used by the API enqueuer.                                                                                                                                          | No       | `default`                       |
| **Stripe**                    |                                                                                                                                                                                             |          |                                 |
| `STRIPE_SECRET_KEY`           | Stripe secret key for server-side operations. **CRITICAL for production - required for payment processing.**                                                                                | Yes      |                                 |
| `STRIPE_WEBHOOK_SECRET`       | Required in non-dev environments; used to verify Stripe webhooks. **CRITICAL for production security - webhook verification will fail without this.**                                       | Yes*     |                                 |
| `STRIPE_PUBLISHABLE_KEY`      | Stripe publishable key for frontend integration. Optional, mainly for documentation.                                                                                                        | No       |                                 |
| **S3 Storage**                |                                                                                                                                                                                             |          |                                 |
| `S3_ENDPOINT`                 | The endpoint of the S3-compatible storage. For Backblaze B2, use `https://s3.{region}.backblazeb2.com` (e.g., `https://s3.us-west-004.backblazeb2.com`). For local dev, use MinIO endpoint. | Yes      | `http://minio:9000`             |
| `S3_REGION`                   | The region of the S3 bucket. For Backblaze B2, use the bucket's region code (e.g., `us-west-004`).                                                                                          | Yes      | `us-west-1`                     |
| `S3_BUCKET` or `S3_BUCKET_NAME` | The name of the S3 bucket. Both variables are supported; `S3_BUCKET` is checked first.                                                                                                    | Yes      | `real-staging`                  |
| `S3_ACCESS_KEY`               | The access key for the S3 bucket. For Backblaze B2, this is the `keyID` from your application key.                                                                                          | Yes      | `minioadmin`                    |
| `S3_SECRET_KEY`               | The secret key for the S3 bucket. For Backblaze B2, this is the `applicationKey` from your application key.                                                                                 | Yes      | `minioadmin`                    |
| `S3_USE_PATH_STYLE`           | Whether to use path-style addressing for S3. Set to `false` for Backblaze B2 and AWS S3, `true` for MinIO.                                                                                  | No       | `true`                          |
| `S3_PUBLIC_ENDPOINT`          | Public/base endpoint to use when presigning URLs (ensures browser-accessible host); when set, presigners use this host. Optional.                                                           | No       |                                 |
| **Frontend**                  |                                                                                                                                                                                             |          |                                 |
| `FRONTEND_URL`                | The URL of your frontend application. Used for CORS and redirect URLs in Stripe checkout.                                                                                                   | Yes      | `http://localhost:3000`         |
| **Observability**             |                                                                                                                                                                                             |          |                                 |
| `LOG_LEVEL`                   | Logging level (`debug`, `info`, `warn`, `error`).                                                                                                                                           | No       | `info`                          |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector. Optional for tracing.                                                                                                                          | No       | `http://otel:4318`              |
| **Feature Flags**             |                                                                                                                                                                                             |          |                                 |
| `RECONCILE_ENABLED`           | Enable reconciliation endpoint (`1` to enable). Used for storage reconciliation operations.                                                                                                 | No       | (disabled)                      |

## Worker Service (`worker`)

| Variable                      | Description                                                                                                                                                          | Required | Default Value       |
| ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------- |
| `APP_ENV`                     | The application environment (`dev`, `test`, `production`).                                                                                                           | No       | `dev`               |
| **Database**                  |                                                                                                                                                                      |          |                     |
| `DATABASE_URL`                | Full Postgres DSN. If set, takes precedence over PG\* vars below. Format: `postgres://user:pass@host:port/db?sslmode=disable`                                       | Yes*     |                     |
| `PGHOST`                      | The hostname of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                             | Yes*     | `postgres`          |
| `PGPORT`                      | The port of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                 | Yes*     | `5432`              |
| `PGUSER`                      | The username for the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                            | Yes*     | `postgres`          |
| `PGPASSWORD`                  | The password for the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                            | Yes*     | `postgres`          |
| `PGDATABASE`                  | The name of the PostgreSQL database. Used if `DATABASE_URL` not set.                                                                                                 | Yes*     | `realstaging`       |
| `PGSSLMODE`                   | Postgres SSL mode when constructing DSN from PG\* vars (`disable`, `require`, `verify-ca`, `verify-full`).                                                           | No       | `disable`           |
| **Redis**                     |                                                                                                                                                                      |          |                     |
| `REDIS_ADDR`                  | The address of the Redis server. Format: `host:port` or `redis://host:port`. Required for job queue and event publishing.                                           | Yes      | `redis:6379`        |
| **Job Queue**                 |                                                                                                                                                                      |          |                     |
| `JOB_QUEUE_NAME`              | Default Asynq queue name to listen on. Must match queue name used by API.                                                                                            | No       | `default`           |
| `WORKER_CONCURRENCY`          | Number of concurrent workers processing jobs.                                                                                                                        | No       | `5`                 |
| **Replicate AI**              |                                                                                                                                                                      |          |                     |
| `REPLICATE_API_TOKEN`         | Replicate API token for AI image processing. **CRITICAL for production - worker cannot process images without this.**                                                | Yes      |                     |
| **S3 Storage**                |                                                                                                                                                                      |          |                     |
| `S3_ENDPOINT`                 | The endpoint of the S3-compatible storage. For Backblaze B2, use `https://s3.{region}.backblazeb2.com`. For local dev, use MinIO endpoint.                          | Yes      | `http://minio:9000` |
| `S3_REGION`                   | The region of the S3 bucket. For Backblaze B2, use the bucket's region code (e.g., `us-west-004`).                                                                   | Yes      | `us-west-1`         |
| `S3_BUCKET` or `S3_BUCKET_NAME` | The name of the S3 bucket. Both variables are supported; `S3_BUCKET` is checked first.                                                                             | Yes      | `real-staging`      |
| `S3_ACCESS_KEY`               | The access key for the S3 bucket. For Backblaze B2, this is the `keyID` from your application key.                                                                   | Yes      | `minioadmin`        |
| `S3_SECRET_KEY`               | The secret key for the S3 bucket. For Backblaze B2, this is the `applicationKey` from your application key.                                                          | Yes      | `minioadmin`        |
| `S3_USE_PATH_STYLE`           | Whether to use path-style addressing for S3. Set to `false` for Backblaze B2 and AWS S3, `true` for MinIO.                                                           | No       | `true`              |
| `S3_PUBLIC_ENDPOINT`          | Public/base endpoint to use when presigning URLs. Optional.                                                                                                          | No       |                     |
| **Observability**             |                                                                                                                                                                      |          |                     |
| `LOG_LEVEL`                   | Logging level (`debug`, `info`, `warn`, `error`).                                                                                                                    | No       | `info`              |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector. Optional for tracing.                                                                                                   | No       | `http://otel:4318`  |

## Environment-Specific Notes

### Development (`APP_ENV=dev`)
- Uses MinIO for local S3 storage (`S3_ENDPOINT=http://minio:9000`)
- Uses local PostgreSQL and Redis via Docker Compose
- `S3_USE_PATH_STYLE=true` for MinIO compatibility
- `STRIPE_WEBHOOK_SECRET` is optional (not enforced)
- Auth0 can use test domain

### Production (`APP_ENV=production`)
- **Must use** Backblaze B2 or AWS S3 (`S3_ENDPOINT=https://s3.{region}.backblazeb2.com`)
- **Must set** `S3_USE_PATH_STYLE=false` for B2/S3
- **Must set** `STRIPE_WEBHOOK_SECRET` (enforced, API fails closed without it)
- **Must set** `REPLICATE_API_TOKEN` (worker cannot process without it)
- **Must use** `PGSSLMODE=require` for Render PostgreSQL
- **Must set** Auth0 production domain and audience
- **Must set** `FRONTEND_URL` to production frontend URL

### Testing (`APP_ENV=test`)
- Uses LocalStack for S3 in integration tests
- Special S3 configuration for test environment
- Skips certain validations for testing

## Security Notes

### Required in Production

**Stripe Configuration:**
- `STRIPE_SECRET_KEY`: Required for all payment operations
- `STRIPE_WEBHOOK_SECRET`: Required in non-dev environments. The API will fail closed (HTTP 503) if it is missing.
  - Webhook verification uses HMAC-SHA256 with timestamp tolerance (default 5m)
  - Requests with invalid signatures are rejected (HTTP 401)

**Auth0 Configuration:**
- `AUTH0_DOMAIN`: Required for JWT validation
- `AUTH0_AUDIENCE`: Required for token audience validation
- Without these, all authenticated requests will fail

**Replicate AI:**
- `REPLICATE_API_TOKEN`: Required for worker to process images
- Worker will fail jobs without valid token

### Secret Rotation

**Stripe Webhooks:**
1. Generate a new webhook secret in Stripe Dashboard
2. Deploy the new secret as environment variable
3. Roll out to all environments
4. Verify with test event from Stripe CLI
5. Remove old secret after confirmation

**Backblaze B2 Keys:**
1. Create new application key in B2 dashboard
2. Update `S3_ACCESS_KEY` and `S3_SECRET_KEY`
3. Deploy to all services (API and Worker)
4. Verify uploads/downloads work
5. Delete old application key in B2

### Never Commit Secrets

❌ **Do NOT commit:**
- Production `DATABASE_URL`
- `STRIPE_SECRET_KEY` or `STRIPE_WEBHOOK_SECRET`
- `REPLICATE_API_TOKEN`
- `S3_ACCESS_KEY` or `S3_SECRET_KEY`
- `AUTH0_CLIENT_SECRET` (if using M2M tokens)

✅ **Use instead:**
- Render dashboard environment variables
- CI/CD secrets (GitHub Actions secrets)
- Secret management services (AWS Secrets Manager, etc.)
- Local `.env` files (gitignored)

## Configuration Precedence

Configuration is loaded in this order (later sources override earlier):

1. `config/shared.yml` - Shared defaults
2. `config/{APP_ENV}.yml` - Environment-specific config
3. `secrets.yml` - Local secrets file (if present, gitignored)
4. **Environment variables** - Highest precedence

For production deployments, use environment variables exclusively.
