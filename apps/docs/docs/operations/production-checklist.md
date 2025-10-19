# Production Deployment Checklist

Complete checklist for deploying Real Staging AI to production on Render with Backblaze B2.

## Pre-Deployment

### Infrastructure Setup

- [ ] **Render Account**
  - [x] Sign up at [render.com](https://render.com)
  - [ ] Add payment method
  - [x] Verify account

- [ ] **Backblaze B2 Storage**
  - [ ] Create account at [backblaze.com](https://www.backblaze.com/b2/cloud-storage.html)
  - [ ] Add payment method
  - [ ] Create bucket: `realstaging-prod` (Private)
  - [ ] Note bucket region (e.g., `us-west-004`)
  - [ ] Create Application Key with bucket access
  - [ ] Save `keyID` and `applicationKey` securely
  - [ ] Configure CORS for your frontend domain

- [ ] **Auth0 Configuration**
  - [x] Create production application
  - [ ] Configure allowed callback URLs for production
  - [ ] Configure allowed logout URLs
  - [ ] Configure allowed web origins
  - [ ] Enable refresh token rotation
  - [ ] Save Domain and Audience values

- [ ] **Stripe Configuration**
  - [x] Complete business verification
  - [x] Switch to Live Mode
  - [x] Get live API keys (secret and publishable)
  - [x] Create webhook endpoint for production URL
  - [x] Save webhook secret
  - [x] Configure products and pricing
  - [ ] Test checkout flow in live mode

- [ ] **Replicate Account**
  - [ ] Create account at [replicate.com](https://replicate.com)
  - [ ] Get API token
  - [ ] Add payment method
  - [ ] Verify model access (`qwen/qwen-image-edit`)

### Code Preparation

- [ ] **Update Configuration**
  - [ ] Review `render.yaml` in repository root
  - [ ] Update region to your preferred location
  - [ ] Update S3 endpoint to match B2 bucket region
  - [ ] Update S3 bucket name
  - [ ] Update frontend URL
  - [ ] Commit changes to main branch

- [ ] **Version Control**
  - [ ] Tag release version: `git tag -a v1.0.0 -m "Production release"`
  - [ ] Push tags: `git push origin --tags`
  - [ ] Ensure all changes are committed and pushed

- [ ] **Docker Images**
  - [ ] Test local Docker builds: `make build-api` and `make build-worker`
  - [ ] Verify Dockerfiles are production-ready
  - [ ] Check for hardcoded development values

## Deployment

### Render Setup

- [x] **Connect Repository**
  - [x] Log into Render dashboard
  - [x] Click "New" → "Blueprint"
  - [x] Connect GitHub repository
  - [x] Select main branch
  - [x] Render detects `render.yaml`
  - [x] Click "Apply"

- [x] **Configure Secrets (API Service)**
  - [x] `AUTH0_DOMAIN`: `your-tenant.us.auth0.com`
  - [x] `AUTH0_AUDIENCE`: `https://api.yourdomain.com`
  - [x] `S3_ACCESS_KEY`: B2 keyID
  - [x] `S3_SECRET_KEY`: B2 applicationKey
  - [x] `STRIPE_SECRET_KEY`: `sk_live_...`
  - [x] `STRIPE_WEBHOOK_SECRET`: `whsec_...`
  - [x] `REPLICATE_API_TOKEN`: `r8_...`

- [x] **Configure Secrets (Worker Service)**
  - [x] `S3_ACCESS_KEY`: B2 keyID
  - [x] `S3_SECRET_KEY`: B2 applicationKey
  - [x] `REPLICATE_API_TOKEN`: `r8_...`

- [ ] **Database Setup**
- [ ] Get DATABASE_URL from Render dashboard:
  - Go to `realstaging-db` → Connect → Copy Internal Database URL
  - [ ] **Run Migrations (Choose ONE method):**
    
    **Option A: From Local Machine (Recommended)**
    ```bash
    # Export the DATABASE_URL from Render
    export DATABASE_URL="postgres://realstaging_db_user:xxx@xxx.oregon-postgres.render.com/realstaging_db"
    
    # Run migrations using Docker
    docker run --rm \
      -v $(pwd)/infra/migrations:/migrations \
      migrate/migrate \
      -path /migrations \
      -database "$DATABASE_URL" \
      up
    
    # Verify
    docker run --rm \
      -v $(pwd)/infra/migrations:/migrations \
      migrate/migrate \
      -path /migrations \
      -database "$DATABASE_URL" \
      version
    ```
    
    **Option B: Add to render.yaml** (Future deployments)
    - See [Running Migrations on Render](#running-migrations-on-render) section below

### Verification

- [x] **Health Checks**
  - [x] Check API health: `curl https://api.real-staging.ai/health`
  - [x] Verify database connection
  - [x] Verify Redis connection
  - [x] Check service logs for errors

- [x] **Functional Testing**
  - [x] Test user signup/login via Auth0
  - [x] Test S3 presigned upload to B2
  - [x] Test image creation and job queueing
  - [x] Test worker processing
  - [x] Verify staged images uploaded to B2
  - [x] Test subscription checkout
  - [x] Test Stripe webhook processing

## Post-Deployment

### Domain Configuration

- [x] **Custom Domain (Optional)**
  - [x] Add custom domain in Render dashboard
  - [x] Configure DNS records
  - [x] Wait for SSL certificate provisioning
  - [x] Test HTTPS access

- [x] **Update External Services**
  - [x] Update Auth0 callback URLs to production domain
  - [x] Update Stripe webhook URLs to production domain
  - [x] Update frontend environment variables
  - [x] Test OAuth flow end-to-end

### Monitoring & Alerts

- [ ] **Render Monitoring**
  - [ ] Enable metrics in Render dashboard
  - [ ] Set up alerts for service down
  - [ ] Set up alerts for high error rate
  - [ ] Set up alerts for high response time
  - [ ] Configure notification channels (email, Slack)

- [ ] **Log Management**
  - [ ] Review log output in Render dashboard
  - [ ] Configure log retention
  - [ ] Set up log alerts for critical errors

### Security

- [ ] **Secrets Rotation**
  - [ ] Document secret rotation procedures
  - [ ] Set calendar reminder for quarterly rotation
  - [ ] Test rotation process in staging first

- [ ] **Access Control**
  - [ ] Limit Render dashboard access
  - [ ] Limit Auth0 admin access
  - [ ] Limit Stripe admin access
  - [ ] Document who has access to what

- [ ] **Security Headers**
  - [ ] Verify HTTPS is enforced
  - [ ] Check CORS configuration
  - [ ] Review security headers

### Backup & Recovery

- [ ] **Database Backups**
  - [ ] Verify Render automatic backups are enabled
  - [ ] Set backup retention period
  - [ ] Document restore procedures
  - [ ] Test restore process

- [ ] **Disaster Recovery Plan**
  - [ ] Document rollback procedures
  - [ ] Document data recovery procedures
  - [ ] Keep copy of critical credentials offline
  - [ ] Test recovery procedures

### Performance

- [ ] **Baseline Metrics**
  - [ ] Record initial response times
  - [ ] Record initial error rates
  - [ ] Record initial job processing times
  - [ ] Set performance targets

- [ ] **Optimization**
  - [ ] Review database query performance
  - [ ] Check connection pool settings
  - [ ] Monitor worker concurrency
  - [ ] Plan for scaling thresholds

### Documentation

- [ ] **Runbook**
  - [ ] Document common operations
  - [ ] Document troubleshooting steps
  - [ ] Document escalation procedures
  - [ ] Document on-call procedures

- [ ] **Update Documentation**
  - [ ] Update deployment docs with actual values
  - [ ] Document any issues encountered
  - [ ] Document solutions and workarounds
  - [ ] Update architecture diagrams if needed   ```

## Ongoing Maintenance

### Weekly

- [ ] Review error logs
- [ ] Check service health metrics
- [ ] Review billing and costs
- [ ] Monitor disk usage

### Monthly

- [ ] Review security logs
- [ ] Check for service updates
- [ ] Review backup integrity
- [ ] Review cost optimization opportunities

### Quarterly

- [ ] Rotate secrets
- [ ] Review and update documentation
- [ ] Review access control lists
- [ ] Performance review and optimization
- [ ] Capacity planning review

## Running Migrations on Render

### Why Migrations Aren't Automated

The current Docker setup doesn't include the `migrate` binary in the API container because:
1. The API Dockerfile only includes compiled Go binaries (`api-server`, `reconcile`)
2. Migration files are in `infra/migrations`, not in `apps/api`
3. Keeping the production image minimal (security best practice)

### Migration Strategies

#### Option 1: Manual Migrations from Local Machine (Current)

**Pros:**
- Simple and reliable
- Full control over timing
- Can review migrations before applying
- No changes to production infrastructure

**Cons:**
- Requires manual intervention for each deployment
- Requires VPN or database to accept external connections

**Steps:**
```bash
# 1. Get DATABASE_URL from Render dashboard
export DATABASE_URL="postgres://user:pass@host.oregon-postgres.render.com/dbname"

# 2. Run migrations
docker run --rm \
  -v $(pwd)/infra/migrations:/migrations \
  migrate/migrate \
  -path /migrations \
  -database "$DATABASE_URL" \
  up

# 3. Verify
docker run --rm \
  -v $(pwd)/infra/migrations:/migrations \
  migrate/migrate \
  -path /migrations \
  -database "$DATABASE_URL" \
  version
```

#### Option 2: Include Migrate Binary in API Docker Image

Add the migrate binary to your API container for on-demand migrations.

**Update `apps/api/Dockerfile`:**
```dockerfile
# ---- Builder ----
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Install migrate
RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | \
    tar xvz && \
    mv migrate /usr/local/bin/migrate

# ... rest of your build steps ...

# ---- Runner ----
FROM alpine:latest

WORKDIR /app

# Copy migrate binary
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy migrations
COPY ../../infra/migrations /app/migrations

# Copy compiled binaries
COPY --from=builder /api-server /app/api-server
COPY --from=builder /reconcile /app/reconcile

EXPOSE 8080
ENTRYPOINT ["/app/api-server"]
```

**Then run migrations via Render Shell:**
```bash
# In Render dashboard: realstaging-api → Shell
migrate -path /app/migrations -database $DATABASE_URL up
```

**Pros:**
- Can run migrations directly from Render shell
- No external access needed
- Migrations bundled with application

**Cons:**
- Increases Docker image size slightly
- Requires rebuilding and redeploying to update migrations
- Still requires manual execution

#### Option 3: Separate Migration Job Service

Create a dedicated service for running migrations.

**Create `apps/migrate/Dockerfile`:**
```dockerfile
FROM migrate/migrate:v4.17.0

WORKDIR /migrations

# Copy all migration files
COPY infra/migrations ./

# Default command
CMD ["migrate", "-path", "/migrations", "-database", "${DATABASE_URL}", "up"]
```

**Add to `render.yaml`:**
```yaml
services:
  # ... existing services ...

  # Migration Job (manual trigger)
  - type: worker
    name: realstaging-migrations
    runtime: docker
    dockerfilePath: ./apps/migrate/Dockerfile
    dockerContext: .
    region: oregon
    plan: starter
    numInstances: 0  # Don't run automatically
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: realstaging-db
          property: connectionString
```

**Run migrations:**
```bash
# Manually scale up to run migrations
# Render dashboard: realstaging-migrations → Scale to 1 instance
# Wait for it to complete
# Scale back down to 0
```

**Pros:**
- Dedicated migration infrastructure
- Can be triggered on-demand
- Clear separation of concerns

**Cons:**
- Additional service to manage
- Still somewhat manual (scale up/down)
- Worker services run continuously unless numInstances=0

#### Option 4: Pre-Deploy Hook with CI/CD

Run migrations automatically in your CI/CD pipeline before deploying.

**GitHub Actions Example:**
```yaml
name: Deploy to Render

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Migrations
        run: |
          docker run --rm \
            -v $(pwd)/infra/migrations:/migrations \
            migrate/migrate \
            -path /migrations \
            -database "${{ secrets.DATABASE_URL }}" \
            up
      
      - name: Trigger Render Deploy
        run: |
          curl -X POST "${{ secrets.RENDER_DEPLOY_HOOK }}"
```

**Pros:**
- Fully automated
- Migrations run before deployment
- No manual intervention needed
- Audit trail in CI/CD logs

**Cons:**
- Requires CI/CD setup
- Requires DATABASE_URL in CI/CD secrets
- Render needs to accept connections from CI/CD IPs

### Recommended Approach

**For initial setup:** Use **Option 1** (manual from local machine)
- Quickest to get started
- No infrastructure changes needed
- Works immediately

**For ongoing deployments:** Consider **Option 4** (CI/CD) or **Option 2** (include in Docker)
- Option 4 is best for teams with CI/CD already
- Option 2 is simpler if you don't have CI/CD

### Migration Best Practices

1. **Always test migrations locally first:**
   ```bash
   make migrate  # Run against local dev database
   ```

2. **Make migrations backward-compatible:**
   - Add columns as nullable first
   - Remove columns in separate migration later
   - Don't break running application versions

3. **Backup before migrations:**
   ```bash
   # Render auto-backs up daily, but you can trigger manual backup
   # Dashboard → realstaging-db → Backups → Create Backup
   ```

4. **Monitor migrations:**
   ```bash
   # Check migration status
   docker run --rm \
     -v $(pwd)/infra/migrations:/migrations \
     migrate/migrate \
     -path /migrations \
     -database "$DATABASE_URL" \
     version
   ```

5. **Have rollback plan:**
   ```bash
   # Rollback last migration
   docker run --rm \
     -v $(pwd)/infra/migrations:/migrations \
     migrate/migrate \
     -path /migrations \
     -database "$DATABASE_URL" \
     down 1
   ```

## Rollback Plan

If deployment fails or issues arise:

1. **Immediate Actions**
   - [ ] Stop new deployments
   - [ ] Assess impact and severity
   - [ ] Communicate to stakeholders

2. **Rollback Steps**
   - [ ] Revert to previous Render deployment
   - [ ] Or roll back git commit and redeploy
   - [ ] Roll back database migrations if needed
   - [ ] Verify services are healthy
   - [ ] Test critical user flows

3. **Post-Incident**
   - [ ] Document what went wrong
   - [ ] Identify root cause
   - [ ] Create action items to prevent recurrence
   - [ ] Update runbook with lessons learned

## Cost Monitoring

### Expected Monthly Costs

- **Render Services**: ~$31/month (starter tier)
  - API: $7/month
  - Worker: $7/month
  - PostgreSQL: $7/month
  - Redis: $10/month

- **Backblaze B2**: ~$5-20/month
  - Storage: $0.005/GB
  - Bandwidth: First 3x storage free
  - API calls: Generous free tier

- **Replicate AI**: Variable
  - ~$0.011 per image processed
  - Estimate based on expected volume

- **Total Infrastructure**: ~$40-60/month (before AI costs)

### Cost Alerts

- [ ] Set up billing alerts in Render
- [ ] Set up billing alerts in Backblaze
- [ ] Monitor Replicate usage and costs
- [ ] Review monthly cost reports

## Success Criteria

Deployment is successful when:

- [x] All services are running and healthy
- [x] Health endpoints return 200 OK
- [x] Users can sign up and log in
- [x] Users can upload images
- [x] Images are processed successfully
- [x] Users can subscribe and checkout
- [x] Stripe webhooks are processing correctly
- [x] No critical errors in logs
- [x] Response times are acceptable (<500ms P95)
- [x] Monitoring and alerts are configured
- [x] Backups are enabled and tested
- [x] Documentation is updated

---

## Additional Resources

- [Deployment Guide](deployment.md) - Detailed deployment instructions
- [Configuration Guide](../guides/configuration.md) - Environment variables
- [Monitoring Guide](monitoring.md) - Observability setup
- [Stripe Integration](../guides/stripe-billing.md) - Payment processing
- [Auth0 Setup](../guides/authentication.md) - Authentication configuration

## Support

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review service logs in Render dashboard
3. Check [Render Status](https://status.render.com/)
4. Check [Backblaze Status](https://status.backblaze.com/)
5. Review this checklist for missed steps
