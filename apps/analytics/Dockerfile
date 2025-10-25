FROM metabase/metabase:v0.51.2

# Metabase requires environment variables to be set at runtime
# These are configured in:
# - docker-compose.yml (local development)
# - render.yaml (production)
#
# Required environment variables:
# - MB_DB_TYPE: Database type (postgres)
# - MB_DB_HOST: Database host
# - MB_DB_PORT: Database port
# - MB_DB_DBNAME: Database name (metabase)
# - MB_DB_USER: Database user
# - MB_DB_PASS: Database password
#
# Optional environment variables:
# - MB_SITE_URL: Public URL for Metabase
# - MB_SITE_NAME: Display name
# - MB_ENABLE_EMBEDDING: Enable iframe embedding
# - MB_EMBEDDING_SECRET_KEY: Secret for signed embeds

# Expose Metabase port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1
