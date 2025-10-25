FROM metabase/metabase:v0.51.2

# Set environment variables for PostgreSQL as the application database
ENV MB_DB_TYPE=postgres
ENV MB_DB_DBNAME=metabase
ENV MB_DB_PORT=5432
ENV MB_DB_USER=postgres
ENV MB_DB_PASS=postgres
ENV MB_DB_HOST=postgres

# Expose Metabase port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1
