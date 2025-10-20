-- Remove seed data for pricing plans
DELETE FROM plans WHERE code IN ('free', 'pro', 'business');
