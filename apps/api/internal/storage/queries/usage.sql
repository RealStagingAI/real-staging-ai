-- name: CountImagesCreatedInPeriod :one
-- Count how many images a user created within a specific date range
-- IMPORTANT: This counts ALL images (including soft-deleted) to prevent gaming the system
-- Users cannot reduce their usage count by deleting images
SELECT COUNT(*)::int
FROM images i
JOIN projects p ON i.project_id = p.id
WHERE p.user_id = $1
  AND i.created_at >= $2
  AND i.created_at < $3;

-- name: GetPlanByCode :one
-- Get a plan by its code (free, pro, business, etc.)
SELECT *
FROM plans
WHERE code = $1;

-- name: GetPlanByPriceID :one
-- Get a plan by its Stripe price ID
SELECT *
FROM plans
WHERE price_id = $1;

-- name: GetUserActivePlan :one
-- Get the user's current active plan based on their subscription
-- Returns the plan for active/trialing subscriptions, or NULL if no active subscription
SELECT p.*
FROM plans p
JOIN subscriptions s ON s.price_id = p.price_id
WHERE s.user_id = $1
  AND s.status IN ('active', 'trialing')
ORDER BY s.created_at DESC
LIMIT 1;

-- name: ListAllPlans :many
-- List all available plans
SELECT *
FROM plans
ORDER BY monthly_limit ASC;

-- name: CreatePlan :one
-- Create a new plan
INSERT INTO plans (id, code, price_id, monthly_limit)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePlan :one
-- Update an existing plan
UPDATE plans 
SET price_id = $2, monthly_limit = $3
WHERE code = $1
RETURNING *;

-- name: ListAllActiveSubscriptions :many
-- List all active subscriptions (for validation)
SELECT *
FROM subscriptions
WHERE status IN ('active', 'trialing');
