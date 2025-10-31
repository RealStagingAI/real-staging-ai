-- Revert plans to original seed values
-- This restores the original price IDs from migration 0011

UPDATE plans 
SET price_id = 'price_1SKJ8mLkQ5x1VWxdP1dtxKK3', monthly_limit = 10
WHERE code = 'free';

UPDATE plans 
SET price_id = 'price_1SJOLOLkQ5x1VWxdO06cPbj1', monthly_limit = 100
WHERE code = 'pro';

UPDATE plans 
SET price_id = 'price_1SJOMjLkQ5x1VWxdUYOkNqI4', monthly_limit = 500
WHERE code = 'business';
