-- Seed initial pricing plans

INSERT INTO plans (code, price_id, monthly_limit) VALUES
  ('free', 'price_1SKJ8mLkQ5x1VWxdP1dtxKK3', 10),         -- Free tier: 10 images/month
  ('pro', 'price_1SJOLOLkQ5x1VWxdO06cPbj1', 100),         -- Pro tier: 100 images/month
  ('business', 'price_1SJOMjLkQ5x1VWxdUYOkNqI4', 500);    -- Business tier: 500 images/month

-- Set default plan price IDs from environment if needed
-- These should be updated via application configuration or manual SQL after deployment
