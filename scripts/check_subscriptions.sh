#!/bin/bash
# Check subscriptions in the database

DATABASE_URL="${DATABASE_URL:-postgresql://realstaging_db_user:xO6fcDNxOmaX10NCYh51AEpcMBXNjNJR@dpg-d3q4ue8gjchc73b28njg-a.oregon-postgres.render.com/realstaging_db}"

echo "=== Checking Users ==="
psql "$DATABASE_URL" -c "SELECT id, auth0_sub, stripe_customer_id, created_at FROM users ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "=== Checking Subscriptions ==="
psql "$DATABASE_URL" -c "SELECT id, user_id, stripe_subscription_id, status, created_at FROM subscriptions ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "=== Checking Processed Events Schema ==="
psql "$DATABASE_URL" -c "\d processed_events"

echo ""
echo "=== Checking Processed Events (Last 10) ==="
psql "$DATABASE_URL" -c "SELECT stripe_event_id, type, received_at FROM processed_events ORDER BY received_at DESC LIMIT 10;"

echo ""
echo "=== User-Subscription Join ==="
psql "$DATABASE_URL" -c "SELECT u.auth0_sub, u.stripe_customer_id, s.stripe_subscription_id, s.status, s.created_at FROM users u LEFT JOIN subscriptions s ON u.id = s.user_id ORDER BY s.created_at DESC LIMIT 5;"
