-- Real Staging AI - Analytics Queries
-- Run these in psql to get immediate insights into your production data

-- Helper function to map price_id to plan names
CREATE OR REPLACE FUNCTION get_plan_name(price_id TEXT) RETURNS TEXT AS $$
BEGIN
  RETURN CASE price_id
    WHEN 'price_1SK67rLpUWppqPSl2XfvuIlh' THEN 'free'
    WHEN 'price_1SJmy5LpUWppqPSlNElnvowM' THEN 'pro'
    WHEN 'price_1SJmyqLpUWppqPSlGhxfz2oQ' THEN 'business'
    ELSE 'unknown'
  END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================================
-- USER OVERVIEW
-- ============================================================================

-- Total users and recent signups
SELECT 
  COUNT(*) as total_users,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_this_week,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as new_this_month,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '90 days') as new_this_quarter
FROM users;

-- Recent user signups (last 20)
SELECT 
  email,
  created_at,
  (SELECT get_plan_name(price_id) FROM subscriptions WHERE user_id = users.id AND status = 'active' LIMIT 1) as current_plan
FROM users 
ORDER BY created_at DESC 
LIMIT 20;

-- ============================================================================
-- SUBSCRIPTION & REVENUE OVERVIEW
-- ============================================================================

-- Active subscriptions by plan with MRR
SELECT 
  get_plan_name(price_id) as plan,
  COUNT(*) as active_subscriptions,
  CASE get_plan_name(price_id)
    WHEN 'free' THEN 0
    WHEN 'pro' THEN COUNT(*) * 29
    WHEN 'business' THEN COUNT(*) * 99
    ELSE 0
  END as monthly_revenue
FROM subscriptions 
WHERE status = 'active'
GROUP BY price_id
ORDER BY monthly_revenue DESC;

-- Total MRR
SELECT 
  SUM(CASE get_plan_name(price_id)
    WHEN 'pro' THEN 29 
    WHEN 'business' THEN 99 
    ELSE 0 
  END) as total_mrr
FROM subscriptions 
WHERE status = 'active';

-- Recent subscription changes
SELECT 
  get_plan_name(s.price_id) as plan,
  s.status,
  s.created_at,
  s.canceled_at,
  u.email
FROM subscriptions s
JOIN users u ON u.id = s.user_id
ORDER BY s.created_at DESC
LIMIT 20;

-- ============================================================================
-- IMAGE PROCESSING STATS
-- ============================================================================

-- Total images and recent activity
SELECT 
  COUNT(*) as total_images,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as images_this_week,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as images_this_month,
  COUNT(*) FILTER (WHERE status = 'ready') as completed_images,
  COUNT(*) FILTER (WHERE status = 'error') as error_images,
  COUNT(*) FILTER (WHERE status = 'processing') as currently_processing,
  COUNT(*) FILTER (WHERE status = 'queued') as queued_images
FROM images
WHERE deleted_at IS NULL;

-- Most popular room types (last 30 days)
SELECT 
  room_type,
  COUNT(*) as count,
  ROUND(AVG(EXTRACT(EPOCH FROM (updated_at - created_at))), 2) as avg_processing_seconds
FROM images
WHERE status = 'ready'
  AND created_at > NOW() - INTERVAL '30 days'
GROUP BY room_type
ORDER BY count DESC;

-- Most popular styles (last 30 days)
SELECT 
  style,
  COUNT(*) as count
FROM images
WHERE status = 'ready'
  AND created_at > NOW() - INTERVAL '30 days'
GROUP BY style
ORDER BY count DESC;

-- Room type + style combinations
SELECT 
  room_type,
  style,
  COUNT(*) as count
FROM images
WHERE status = 'ready'
  AND created_at > NOW() - INTERVAL '30 days'
GROUP BY room_type, style
ORDER BY count DESC
LIMIT 15;

-- ============================================================================
-- TOP USERS BY ACTIVITY
-- ============================================================================

-- Top 20 users by total image count
SELECT 
  u.email,
  u.created_at as signup_date,
  COALESCE(get_plan_name(s.price_id), 'free') as plan,
  COUNT(i.id) as total_images,
  COUNT(i.id) FILTER (WHERE i.created_at > NOW() - INTERVAL '30 days') as images_last_30d,
  COUNT(i.id) FILTER (WHERE i.created_at > NOW() - INTERVAL '7 days') as images_last_7d,
  MAX(i.created_at) as last_image_date
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id AND s.status = 'active'
LEFT JOIN projects p ON p.user_id = u.id
LEFT JOIN images i ON i.project_id = p.id AND i.deleted_at IS NULL
GROUP BY u.id, u.email, u.created_at, s.price_id
HAVING COUNT(i.id) > 0
ORDER BY total_images DESC
LIMIT 20;

-- Power users (>10 images in last 30 days)
SELECT 
  u.email,
  COALESCE(get_plan_name(s.price_id), 'free') as plan,
  COUNT(i.id) as images_last_30d
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id AND s.status = 'active'
LEFT JOIN projects p ON p.user_id = u.id
LEFT JOIN images i ON i.project_id = p.id AND i.created_at > NOW() - INTERVAL '30 days' AND i.deleted_at IS NULL
GROUP BY u.id, u.email, s.price_id
HAVING COUNT(i.id) > 10
ORDER BY images_last_30d DESC;

-- ============================================================================
-- USER ENGAGEMENT & RETENTION
-- ============================================================================

-- Users by plan with usage stats
SELECT 
  COALESCE(get_plan_name(s.price_id), 'free') as plan,
  COUNT(DISTINCT u.id) as user_count,
  COUNT(i.id) as total_images,
  ROUND(COALESCE(AVG(user_images.image_count), 0), 2) as avg_images_per_user
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id AND s.status = 'active'
LEFT JOIN projects p ON p.user_id = u.id
LEFT JOIN images i ON i.project_id = p.id AND i.deleted_at IS NULL
LEFT JOIN (
  SELECT p.user_id, COUNT(*) as image_count
  FROM images i
  JOIN projects p ON p.id = i.project_id
  WHERE i.deleted_at IS NULL
  GROUP BY p.user_id
) user_images ON user_images.user_id = u.id
GROUP BY s.price_id
ORDER BY 
  CASE get_plan_name(s.price_id)
    WHEN 'business' THEN 1
    WHEN 'pro' THEN 2
    ELSE 3
  END;

-- Inactive users (no images in last 30 days but have subscription)
SELECT 
  u.email,
  get_plan_name(s.price_id) as plan,
  s.created_at as subscription_start,
  MAX(i.created_at) as last_image_date,
  EXTRACT(DAY FROM (NOW() - MAX(i.created_at))) as days_since_last_image
FROM users u
JOIN subscriptions s ON s.user_id = u.id AND s.status = 'active'
LEFT JOIN projects p ON p.user_id = u.id
LEFT JOIN images i ON i.project_id = p.id AND i.deleted_at IS NULL
WHERE get_plan_name(s.price_id) != 'free'
GROUP BY u.id, u.email, s.price_id, s.created_at
HAVING MAX(i.created_at) IS NULL OR MAX(i.created_at) < NOW() - INTERVAL '30 days'
ORDER BY days_since_last_image DESC NULLS FIRST;

-- ============================================================================
-- GROWTH TRENDS
-- ============================================================================

-- Daily signups (last 30 days)
SELECT 
  DATE(created_at) as signup_date,
  COUNT(*) as new_users
FROM users
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY signup_date DESC;

-- Daily image processing (last 30 days)
SELECT 
  DATE(created_at) as process_date,
  COUNT(*) as images_processed,
  COUNT(*) FILTER (WHERE status = 'ready') as successful,
  COUNT(*) FILTER (WHERE status = 'error') as errors,
  COUNT(*) FILTER (WHERE status = 'processing') as still_processing
FROM images
WHERE created_at > NOW() - INTERVAL '30 days'
  AND deleted_at IS NULL
GROUP BY DATE(created_at)
ORDER BY process_date DESC;

-- Weekly subscription changes (last 12 weeks)
SELECT 
  DATE_TRUNC('week', created_at) as week,
  get_plan_name(price_id) as plan,
  COUNT(*) as new_subscriptions
FROM subscriptions
WHERE created_at > NOW() - INTERVAL '12 weeks'
GROUP BY DATE_TRUNC('week', created_at), price_id
ORDER BY week DESC, plan;

-- ============================================================================
-- CHURN ANALYSIS
-- ============================================================================

-- Recent cancellations (last 30 days)
SELECT 
  DATE(s.canceled_at) as cancellation_date,
  get_plan_name(s.price_id) as plan,
  u.email,
  s.created_at as subscription_start,
  EXTRACT(DAY FROM (s.canceled_at - s.created_at)) as days_subscribed
FROM subscriptions s
JOIN users u ON u.id = s.user_id
WHERE s.canceled_at > NOW() - INTERVAL '30 days'
ORDER BY s.canceled_at DESC;

-- Cancellation summary by plan
SELECT 
  get_plan_name(price_id) as plan,
  COUNT(*) as total_cancellations,
  ROUND(AVG(EXTRACT(DAY FROM (canceled_at - created_at))), 2) as avg_days_before_cancel
FROM subscriptions
WHERE canceled_at > NOW() - INTERVAL '90 days'
GROUP BY price_id
ORDER BY total_cancellations DESC;

-- ============================================================================
-- QUICK SUMMARY DASHBOARD
-- ============================================================================

-- Complete overview (copy all of this for one query)
WITH user_stats AS (
  SELECT 
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_this_week
  FROM users
),
subscription_stats AS (
  SELECT 
    COUNT(*) FILTER (WHERE get_plan_name(price_id) = 'free' AND status = 'active') as free_users,
    COUNT(*) FILTER (WHERE get_plan_name(price_id) = 'pro' AND status = 'active') as pro_users,
    COUNT(*) FILTER (WHERE get_plan_name(price_id) = 'business' AND status = 'active') as business_users,
    SUM(CASE get_plan_name(price_id) WHEN 'pro' THEN 29 WHEN 'business' THEN 99 ELSE 0 END) as mrr
  FROM subscriptions
  WHERE status = 'active'
),
image_stats AS (
  SELECT 
    COUNT(*) as total_images,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as images_this_week,
    COUNT(*) FILTER (WHERE status = 'error') as error_images
  FROM images
  WHERE deleted_at IS NULL
)
SELECT 
  us.total_users,
  us.new_this_week,
  ss.free_users,
  ss.pro_users,
  ss.business_users,
  ss.mrr,
  im.total_images,
  im.images_this_week,
  im.error_images
FROM user_stats us, subscription_stats ss, image_stats im;
