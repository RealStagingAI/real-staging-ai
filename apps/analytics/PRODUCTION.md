# Metabase Production Deployment

This guide covers deploying Metabase to Render as a private internal service, accessible through your Next.js web app.

## Architecture

```
┌─────────────────┐
│   Public Web    │  https://real-staging.ai
│   (Next.js)     │
└────────┬────────┘
         │ Authenticated users
         │ access /admin/analytics
         ▼
┌─────────────────┐
│   Proxy Route   │  /admin/analytics/**
│  (API Route)    │  
└────────┬────────┘
         │ Internal network
         │ (not public)
         ▼
┌─────────────────┐
│    Metabase     │  Private service
│   (Internal)    │  realstaging-metabase.onrender.com
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │  realstaging-db
│  (metabase DB)  │
└─────────────────┘
```

## Deployment Steps

### 1. Create Metabase Database

The `metabase` database must exist in your PostgreSQL instance before deploying.

**Option A: Using Render Dashboard**

```bash
# Connect to your Render PostgreSQL database
# Get connection string from Render dashboard

psql "postgresql://username:password@host/database"

# Create metabase database
CREATE DATABASE metabase;
```

**Option B: Using Render Shell**

```bash
# Open shell to your API service
# Navigate to: Render Dashboard → realstaging-api → Shell

# Connect to database
psql $DATABASE_URL

# Create metabase database
CREATE DATABASE metabase;
```

### 2. Deploy Metabase Service

The Metabase service is already configured in `render.yaml`. Deploy via:

**Option A: Git Push (Automatic)**
```bash
git push origin main
# Render auto-deploys from main branch
```

**Option B: Manual Deploy**
```bash
# From Render Dashboard:
# Services → realstaging-metabase → Manual Deploy → Deploy latest commit
```

### 3. Set Environment Variables

In Render Dashboard, navigate to: **realstaging-metabase → Environment**

Set the following secret:

```
MB_EMBEDDING_SECRET_KEY=<generate-random-32-char-string>
```

Generate with:
```bash
openssl rand -hex 32
```

This is required if you plan to use embedded Metabase dashboards (optional).

### 4. Verify Deployment

1. **Check Metabase service status:**
   - Render Dashboard → realstaging-metabase
   - Should show "Live" status

2. **Check web app proxy:**
   - Visit: https://real-staging.ai/admin/analytics
   - Should redirect to login if not authenticated
   - After login, should load Metabase iframe

3. **Complete Metabase setup:**
   - First visit will show setup wizard
   - Create admin account
   - Database connection is pre-configured
   - Skip data collection preferences

## Configuration

### Environment Variables

All configured automatically via `render.yaml`:

| Variable | Source | Purpose |
|----------|--------|---------|
| `MB_DB_TYPE` | Static | PostgreSQL |
| `MB_DB_HOST` | Database | Postgres host |
| `MB_DB_PORT` | Database | Postgres port |
| `MB_DB_DBNAME` | Static | `metabase` |
| `MB_DB_USER` | Database | Database user |
| `MB_DB_PASS` | Database | Database password |
| `MB_SITE_URL` | Static | https://real-staging.ai/admin/analytics |
| `MB_SITE_NAME` | Static | Real Staging AI Analytics |
| `METABASE_INTERNAL_URL` | Service | Internal URL (for proxy) |

### Security

**✅ Secure by Default:**
- Metabase service is **NOT publicly accessible**
- All access goes through Next.js proxy at `/admin/analytics`
- Authentication required via Auth0
- TODO: Add RBAC to restrict to admin users only

**🔒 To Add RBAC:**

In `apps/web/app/admin/analytics/[[...path]]/route.ts`:

```typescript
// Uncomment RBAC check:
const userRole = session.user['https://real-staging.ai/roles'];
if (!userRole?.includes('admin')) {
  return NextResponse.json(
    { error: 'Forbidden: Admin access required' },
    { status: 403 }
  );
}
```

And in Auth0:
1. Set up custom claims: https://auth0.com/docs/secure/tokens/json-web-tokens/create-custom-claims
2. Add roles to user tokens
3. Assign admin role to users

## Maintenance

### Viewing Logs

```bash
# From Render Dashboard:
# Services → realstaging-metabase → Logs
```

### Restarting Service

```bash
# From Render Dashboard:
# Services → realstaging-metabase → Manual Deploy → Restart
```

### Updating Metabase

Update the version in `apps/analytics/Dockerfile`:

```dockerfile
FROM metabase/metabase:v0.56.11  # Update version here
```

Then deploy via git push or manual deploy.

### Backup Metabase Configuration

Metabase stores all dashboards, questions, and settings in the `metabase` database.

**Backup:**
```bash
# Render provides automatic database backups
# Or manually:
pg_dump -U username -h host metabase > metabase_backup.sql
```

**Restore:**
```bash
psql -U username -h host metabase < metabase_backup.sql
```

## Troubleshooting

### Metabase Won't Start

**Error**: `database "metabase" does not exist`
- **Fix**: Follow Step 1 to create the database

**Error**: `Connection refused`
- **Fix**: Check that PostgreSQL service is running and healthy

### Proxy Not Working

**Error**: `502 Bad Gateway`
- **Check**: Metabase service is running (Render dashboard)
- **Check**: `METABASE_INTERNAL_URL` env var is set on web service
- **Check**: Web service has restarted after adding env var

**Error**: `401 Unauthorized`
- **Check**: User is logged in
- **Check**: Auth0 session is valid

### Analytics Page Shows Blank

**Check browser console** for errors:
- CORS errors → Metabase proxy issue
- 403/401 → Authentication issue
- Network errors → Metabase service down

## Cost

**Render Starter Plan:**
- Metabase service: $7/month (Starter plan)
- Database storage: Included in realstaging-db
- Bandwidth: Included

**Scaling:**
- Upgrade to Standard ($25/month) for:
  - Better performance
  - More concurrent users
  - Higher memory limits

## Monitoring

### Key Metrics

Monitor in Render Dashboard → Metrics:

1. **CPU Usage** - Should be <70% normally
2. **Memory Usage** - Should be <80% normally  
3. **Response Time** - Should be <2s for most queries
4. **Error Rate** - Should be <1%

### Alerts

Set up alerts in Render:
- High CPU/Memory usage
- Service restarts
- Failed health checks

## Next Steps

1. ✅ Complete Metabase first-time setup
2. ✅ Import queries from `scripts/analytics-queries.sql`
3. ✅ Create your first dashboard
4. ⏱️ Add RBAC to restrict access to admins
5. ⏱️ Set up Metabase alerts for key metrics
6. ⏱️ Create embedded dashboards for users (optional)

## Resources

- [Metabase Documentation](https://www.metabase.com/docs/latest/)
- [Render Private Services](https://docs.render.com/private-services)
- [Next.js API Routes](https://nextjs.org/docs/app/building-your-application/routing/route-handlers)
