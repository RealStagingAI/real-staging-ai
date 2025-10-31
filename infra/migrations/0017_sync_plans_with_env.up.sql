-- Sync plans table with environment variables
-- This migration ensures plans table has the basic structure
-- The actual price ID synchronization is handled by the application's PlanSyncService

-- Ensure basic plans exist with safe placeholder price IDs
-- These will be updated by the application on startup

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('free', 'placeholder_free_price_id', 100)
ON CONFLICT (code) 
DO UPDATE SET 
  monthly_limit = EXCLUDED.monthly_limit;

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('pro', 'placeholder_pro_price_id', 100)
ON CONFLICT (code) 
DO UPDATE SET 
  monthly_limit = EXCLUDED.monthly_limit;

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('business', 'placeholder_business_price_id', 500)
ON CONFLICT (code) 
DO UPDATE SET 
  monthly_limit = EXCLUDED.monthly_limit;
