-- Seed initial pricing plans
-- Note: Replace the price_id values with your actual Stripe Price IDs from your Stripe Dashboard

INSERT INTO plans (code, price_id, monthly_limit) VALUES
  ('free', 'price_1SK67rLpUWppqPSl2XfvuIlh', 10),         -- Free tier: 10 images/month
  ('pro', 'price_1SJmy5LpUWppqPSlNElnvowM', 100),         -- Pro tier: 100 images/month
  ('business', 'price_1SJmyqLpUWppqPSlGhxfz2oQ', 500);    -- Business tier: 500 images/month

-- Set default plan price IDs from environment if needed
-- These should be updated via application configuration or manual SQL after deployment
