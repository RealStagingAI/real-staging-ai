-- Seed data for the plans table (required for usage tracking)
INSERT INTO plans (code, price_id, monthly_limit) VALUES
('free', 'price_free_test', 10),
('pro', 'price_pro_test', 100),
('business', 'price_business_test', 500)
ON CONFLICT (code) DO NOTHING;

-- Seed data for the users table
INSERT INTO users (id, auth0_sub, stripe_customer_id, role) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'auth0|testuser', 'cus_test', 'user');

-- Seed data for the projects table
INSERT INTO projects (id, user_id, name) VALUES
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test Project 1');
