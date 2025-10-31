-- Sync plans table with environment variables
-- This migration ensures plans table has the correct price IDs
-- It uses environment variables as defaults but allows manual override

-- Insert or update plans with current price IDs
-- These should match the environment variables:
-- STRIPE_PRICE_FREE, STRIPE_PRICE_PRO, STRIPE_PRICE_BUSINESS

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('free', 'price_1SK67rLpUWppqPSl2XfvuIlh', 100)
ON CONFLICT (code) 
DO UPDATE SET 
  price_id = EXCLUDED.price_id,
  monthly_limit = EXCLUDED.monthly_limit;

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('pro', 'price_1SJmy5LpUWppqPSlNElnvowM', 100)
ON CONFLICT (code) 
DO UPDATE SET 
  price_id = EXCLUDED.price_id,
  monthly_limit = EXCLUDED.monthly_limit;

INSERT INTO plans (code, price_id, monthly_limit) 
VALUES 
  ('business', 'price_1SJmyqLpUWppqPSlGhxfz2oQ', 500)
ON CONFLICT (code) 
DO UPDATE SET 
  price_id = EXCLUDED.price_id,
  monthly_limit = EXCLUDED.monthly_limit;
