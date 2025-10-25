-- Create metabase database for Metabase application metadata
-- This is separate from the realstaging database which Metabase will query
--
-- USAGE:
-- 1. Automatically created on first Docker container setup via docker-entrypoint-initdb.d
-- 2. For existing setups, run: make setup-metabase
-- 3. Or manually: docker compose exec postgres psql -U postgres -c "CREATE DATABASE metabase;"
--
-- Note: This is NOT a regular migration because CREATE DATABASE cannot run in a transaction
-- and golang-migrate runs all migrations in transactions by default.

CREATE DATABASE metabase;
