# Operations

Guides for deploying, monitoring, and maintaining Real Staging AI in production.

## Deployment & Maintenance

- **[Deployment Guide](deployment.md)** - Production deployment on Render
- **[Production Checklist](production-checklist.md)** - Complete deployment checklist
- **[Database Migrations](migrations.md)** - Schema migration management
- **[Storage Reconciliation](reconciliation.md)** - Database and S3 consistency checks
- **[Monitoring](monitoring.md)** - Observability and alerting
- **[Visibility & Analytics](visibility-analytics.md)** - Business intelligence and user insights

## Topics

### Deployment

Deploy Real Staging AI to Render with Backblaze B2 storage, including complete pre-deployment and post-deployment checklists.

[Read the deployment guide →](deployment.md)

### Database Migrations

Manage database schema changes using golang-migrate with automated and manual migration strategies.

[Read the migrations guide →](migrations.md)

### Storage Reconciliation

Maintain consistency between database records and S3 objects with automated reconciliation tools.

[Read the reconciliation guide →](reconciliation.md)

### Monitoring

Set up comprehensive observability with OpenTelemetry, metrics, traces, and structured logging.

[Read the monitoring guide →](monitoring.md)

### Visibility & Analytics

Gain insights into your users, subscriptions, revenue, and product usage with business analytics dashboards and admin tools.

[Read the visibility guide →](visibility-analytics.md)

## Quick Reference

| Topic                      | Key Points                                             |
| -------------------------- | ------------------------------------------------------ |
| **Deployment**             | Render, Docker, Backblaze B2                           |
| **Production Checklist**   | Infrastructure setup, verification, post-deployment    |
| **Migrations**             | golang-migrate, automated & manual strategies          |
| **Reconciliation**         | S3/DB consistency, orphan detection                    |
| **Monitoring**             | OTEL, metrics, traces, logs                            |
| **Visibility & Analytics** | User insights, revenue tracking, business intelligence |
