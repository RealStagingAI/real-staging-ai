# Real Staging AI - Analytics Dashboard (Metabase)


Metabase is an open-source business intelligence tool that provides a visual interface for exploring your data.

## Overview

The analytics dashboard provides:
- **User Analytics**: Signups, active users, user cohorts
- **Revenue Metrics**: MRR, subscriptions by plan, churn analysis
- **Usage Statistics**: Image processing, popular room types/styles
- **Growth Trends**: Daily/weekly charts for key metrics
- **Custom SQL Queries**: Full access to write custom queries

## Local Development

### Start Metabase

```bash
# Start all services including Metabase (recommended)
make up

# Or manually start Metabase
docker compose up -d postgres  # Start Postgres first
make setup-metabase            # Ensure metabase database exists
docker compose up -d metabase  # Start Metabase
```

Metabase will be available at: **http://localhost:3001**

**Note**: The `make up` command automatically ensures the `metabase` database exists before starting Metabase.

### First-Time Setup

On first launch, Metabase will guide you through initial setup:

1. **Create Admin Account**
   - Email: your email
   - Password: choose a secure password
   - Company name: Real Staging AI

2. **Connect to Database** (already configured via environment variables)
   - Type: PostgreSQL
   - Host: `postgres`
   - Port: `5432`
   - Database: `realstaging`
   - Username: `postgres`
   - Password: `postgres`

3. **Data Preferences**
   - Skip or configure as preferred

## Pre-Built Queries

The analytics queries from `scripts/analytics-queries.sql` can be used as Metabase native queries:

### Quick Dashboard Queries

**User Overview:**
```sql
SELECT 
  COUNT(*) as total_users,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_this_week,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as new_this_month
FROM users;
```

**Revenue Summary:**
```sql
SELECT 
  get_plan_name(price_id) as plan,
  COUNT(*) as active_subscriptions,
  CASE get_plan_name(price_id)
    WHEN 'pro' THEN COUNT(*) * 29
    WHEN 'business' THEN COUNT(*) * 99
    ELSE 0
  END as monthly_revenue
FROM subscriptions 
WHERE status = 'active'
GROUP BY price_id;
```

**Image Processing Stats:**
```sql
SELECT 
  COUNT(*) as total_images,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as images_this_week,
  COUNT(*) FILTER (WHERE status = 'ready') as completed_images,
  COUNT(*) FILTER (WHERE status = 'error') as error_images
FROM images
WHERE deleted_at IS NULL;
```

## Creating Dashboards

### Recommended Dashboards

1. **Executive Dashboard**
   - Total users card
   - MRR card
   - Images processed this month
   - User growth chart (30 days)
   - Revenue trend chart

2. **Product Analytics**
   - Most popular room types (pie chart)
   - Most popular styles (bar chart)
   - Processing time trends
   - Error rate over time

3. **User Engagement**
   - Active users by plan
   - Images per user average
   - Power users list
   - Inactive paid users (churn risk)

### Creating a Dashboard

1. Click **"New"** → **"Dashboard"**
2. Add questions (queries) to the dashboard
3. Arrange and resize visualizations
4. Set auto-refresh intervals
5. Share with team members

## Metabase Features

### Native SQL Queries

Write custom SQL queries with full PostgreSQL syntax:
- Use the `get_plan_name()` helper function for plan names
- Join through `projects` table to connect users and images
- Filter soft-deletes with `WHERE deleted_at IS NULL`

### Question Builder

Visual query builder for non-SQL users:
- Drag and drop interface
- Automatic chart suggestions
- Drill-down capabilities

### Filters

Add dashboard filters for:
- Date ranges
- User plans
- Room types
- Styles

### Alerts

Set up email/Slack alerts:
- MRR drops below threshold
- Error rate exceeds limit
- No signups in 24 hours
- High churn rate

## Production Deployment

### Environment Variables

For production, use secure credentials:

```yaml
# Render deployment
services:
  - type: web
    name: metabase
    env: docker
    dockerfilePath: apps/analytics/Dockerfile
    envVars:
      - key: MB_DB_TYPE
        value: postgres
      - key: MB_DB_HOST
        fromService:
          name: postgres
          type: postgres
          property: host
      - key: MB_DB_PORT
        fromService:
          name: postgres
          type: postgres
          property: port
      - key: MB_DB_DBNAME
        value: metabase
      - key: MB_DB_USER
        fromService:
          name: postgres
          type: postgres
          property: user
      - key: MB_DB_PASS
        fromService:
          name: postgres
          type: postgres
          property: password
          sync: false
```

### Authentication

For production:
1. Set up Google SSO or LDAP authentication
2. Configure user permissions (View/Edit/Admin)
3. Restrict database access to read-only users
4. Enable audit logging

### Performance

For large datasets:
- Enable query caching
- Set up scheduled question runs
- Use materialized views for complex queries
- Index frequently queried columns

## Backup & Restore

Metabase stores all dashboards and questions in its PostgreSQL database (`metabase` database).

**Backup:**
```bash
docker exec -t postgres pg_dump -U postgres metabase > metabase_backup.sql
```

**Restore:**
```bash
cat metabase_backup.sql | docker exec -i postgres psql -U postgres metabase
```

## Troubleshooting

### Metabase Won't Start

Check logs:
```bash
docker compose logs metabase
```

Common issues:
- PostgreSQL not ready (wait for healthcheck)
- Port 3000 already in use
- Database connection refused

### Can't Connect to Application Database

Verify connection from Metabase container:
```bash
docker exec -it real-staging-ai-metabase-1 \
  psql -h postgres -U postgres -d realstaging
```

### Slow Queries

- Check query execution plan with `EXPLAIN ANALYZE`
- Add indexes to frequently filtered columns
- Use materialized views for complex aggregations
- Limit date ranges in queries

## Resources

- [Metabase Documentation](https://www.metabase.com/docs/latest/)
- [SQL Query Best Practices](https://www.metabase.com/learn/sql-questions/)
- [Dashboard Design Guide](https://www.metabase.com/learn/dashboards/)
- [Analytics Queries](../../scripts/analytics-queries.sql)

## Next Steps

1. ✅ Start Metabase with `make up`
2. ✅ Complete initial setup at http://localhost:3001
3. ✅ Create your first question using SQL queries
4. ✅ Build an Executive Dashboard
5. ✅ Set up email alerts for key metrics
6. ✅ Share dashboards with your team
