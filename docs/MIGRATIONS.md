# Database Migrations

This document explains how database migrations work in the Real Staging AI project.

## Overview

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations. Migrations are automatically applied in production before each deployment via Render's `preDeployCommand`.

## Directory Structure

```
infra/
└── migrations/
    ├── 0001_create_users_table.up.sql
    ├── 0001_create_users_table.down.sql
    ├── 0002_create_projects_table.up.sql
    ├── 0002_create_projects_table.down.sql
    └── ...
```

## Development

### Run Migrations Locally

```bash
# Apply all pending migrations
make migrate

# Rollback one migration
make migrate-down-dev
```

### Create a New Migration

```bash
# Create a new migration pair (up and down)
docker compose run --rm migrate create -ext sql -dir /migrations -seq <migration_name>
```

Example:
```bash
docker compose run --rm migrate create -ext sql -dir /migrations -seq add_user_avatar
```

This creates two files:
- `NNNN_add_user_avatar.up.sql` - Forward migration
- `NNNN_add_user_avatar.down.sql` - Rollback migration

## Production (Render)

### Automated Migrations

Migrations are **automatically applied** before each deployment via `preDeployCommand` in `render.yaml`:

```yaml
preDeployCommand: migrate -path /app/migrations -database $DATABASE_URL up
```

**How it works:**
1. You push code to GitHub
2. Render detects the push and starts a deploy
3. **Before** starting the new API containers, Render runs the migration command
4. Migrations apply in sequence
5. If migrations succeed, the new API version deploys
6. If migrations fail, deployment is aborted (zero-downtime)

### Manual Migrations (if needed)

If you need to manually run migrations in production:

```bash
# Connect to Render shell for realstaging-api service
# Then run:
migrate -path /app/migrations -database $DATABASE_URL up

# Or rollback:
migrate -path /app/migrations -database $DATABASE_URL down 1
```

### Monitoring

- Check migration status in Render deploy logs
- Each migration outputs: `OK   NNNN_migration_name.up.sql`
- Failures will show detailed error messages

## Best Practices

1. **Always create both up and down migrations** - even if you don't plan to rollback
2. **Test migrations locally first** - `make migrate` before pushing
3. **Keep migrations small and focused** - one logical change per migration
4. **Never modify existing migrations** - create new ones to fix issues
5. **Use transactions when possible** - wrap DDL in `BEGIN;` and `COMMIT;`
6. **Handle data migrations carefully** - consider impact on running services

## Migration File Format

### Up Migration
```sql
-- 0011_seed_plans_table.up.sql
BEGIN;

INSERT INTO plans (code, price_id, monthly_limit) VALUES
  ('free', 'price_1SKJ8mLkQ5x1VWxdP1dtxKK3', 10),
  ('pro', 'price_1SK67rLpUWppqPSl2XfvuIlh', 100),
  ('business', 'price_1SJOMjLkQ5x1VWxdUYOkNqI4', 500);

COMMIT;
```

### Down Migration
```sql
-- 0011_seed_plans_table.down.sql
BEGIN;

DELETE FROM plans WHERE code IN ('free', 'pro', 'business');

COMMIT;
```

## Troubleshooting

### Migration Failed in Production

1. Check Render deploy logs for error details
2. Fix the migration file (create a new one, don't edit existing)
3. Redeploy

### Dirty Database State

If a migration partially succeeds then fails:

```bash
# Force version (use with caution!)
migrate -path /app/migrations -database $DATABASE_URL force <version>

# Then fix and re-run
migrate -path /app/migrations -database $DATABASE_URL up
```

### Schema Out of Sync

```bash
# Check current migration version
migrate -path /app/migrations -database $DATABASE_URL version

# Check dirty state
psql $DATABASE_URL -c "SELECT * FROM schema_migrations;"
```

## References

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL best practices](https://www.postgresql.org/docs/current/ddl-schemas.html)
