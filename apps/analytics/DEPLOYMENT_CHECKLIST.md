# Metabase Production Deployment Checklist

Quick checklist for deploying Metabase to Render.

## Pre-Deployment

- [ ] Commit and push all changes to `main` branch
- [ ] Verify `render.yaml` includes `realstaging-metabase` service
- [ ] Verify Next.js proxy routes exist at `apps/web/app/admin/analytics`

## Database Setup (One-Time)

- [ ] Connect to production PostgreSQL database
- [ ] Run: `CREATE DATABASE metabase;`
- [ ] Verify database created: `\l` in psql

**Quick command:**
```bash
# Get connection string from Render dashboard
psql "postgresql://user:pass@host/database"

# Create database
CREATE DATABASE metabase;

# Verify
\l
```

## Render Configuration

- [ ] Push code to main branch (Render auto-deploys)
- [ ] Wait for `realstaging-metabase` service to build and deploy
- [ ] Set `MB_EMBEDDING_SECRET_KEY` in Render dashboard:
  - Navigate to: realstaging-metabase → Environment
  - Add: `MB_EMBEDDING_SECRET_KEY = <32-char-random-string>`
  - Generate: `openssl rand -hex 32`

- [ ] Restart Metabase service after setting env var

## Web App Configuration

- [ ] Verify `METABASE_INTERNAL_URL` is set on `realstaging-web`:
  - Should be auto-configured from render.yaml
  - Points to internal URL of realstaging-metabase

- [ ] Restart `realstaging-web` if needed

## First-Time Setup

- [ ] Visit: https://real-staging.ai/admin/analytics
- [ ] Complete Metabase setup wizard:
  - Create admin account
  - Database connection (pre-configured, just verify)
  - Skip data collection preferences
  
- [ ] Add application database as data source:
  - Settings → Admin → Databases → Add Database
  - Type: PostgreSQL
  - Name: "Real Staging AI"
  - Host: (from Render dashboard - realstaging-db host)
  - Port: 5432
  - Database: realstaging
  - Username: (from Render dashboard)
  - Password: (from Render dashboard)
  - Save

- [ ] Test connection and sync schema

## Import Analytics Queries

- [ ] Open `scripts/analytics-queries.sql` locally
- [ ] Create new questions in Metabase using SQL from the file:
  - Total Users
  - MRR
  - Active Subscriptions
  - Top Users
  - Image Processing Stats
  - etc.

## Create First Dashboard

- [ ] Create "Executive Dashboard"
- [ ] Add key metrics cards:
  - Total Users
  - New Users This Week
  - MRR
  - Active Subscriptions
  - Images Processed This Week

## Verification

- [ ] Access /admin/analytics without authentication → Redirects to login ✅
- [ ] Access /admin/analytics after login → Shows Metabase ✅
- [ ] Can run SQL queries ✅
- [ ] Can create/save questions ✅
- [ ] Can create dashboards ✅
- [ ] Dashboard auto-refreshes ✅

## Post-Deployment

- [ ] Add navigation link to admin menu (optional)
- [ ] Set up RBAC to restrict to admin users (TODO in code)
- [ ] Create alerts for key metrics (in Metabase)
- [ ] Share dashboards with team members

## Troubleshooting

**Metabase won't start:**
```bash
# Check logs
# Render Dashboard → realstaging-metabase → Logs

# Common fix: Ensure metabase database exists
```

**502 Bad Gateway on proxy:**
```bash
# Check Metabase service is running
# Check METABASE_INTERNAL_URL is set on web service
# Restart web service
```

**Can't connect to database in Metabase:**
```bash
# Get connection details from Render dashboard
# realstaging-db → Connection Details
# Use Internal Database URL (not external)
```

## Rollback Plan

If something goes wrong:

1. Remove Metabase service from `render.yaml`
2. Push to main
3. Delete `/admin/analytics` routes from web app
4. Push to main
5. Service will be removed on next deploy

The `metabase` database will persist, so you can redeploy later without losing data.

---

**See [`PRODUCTION.md`](./PRODUCTION.md) for detailed deployment guide.**
