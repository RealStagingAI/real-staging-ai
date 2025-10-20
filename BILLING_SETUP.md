# Billing & Usage Tracking Setup Guide

This guide walks through setting up the new billing and usage tracking system.

## ✅ What Was Implemented

### Backend (API)

1. **Database Schema**
   - Migration `0009_seed_plans_table.up.sql` adds plan data
   - Queries in `apps/api/internal/storage/queries/usage.sql`

2. **Usage Service**
   - `apps/api/internal/billing/usage_service.go` (interface)
   - `apps/api/internal/billing/default_usage_service.go` (implementation)
   - Tracks images per billing period
   - Enforces monthly limits

3. **API Endpoints**
   - `GET /api/v1/billing/usage` - Current usage statistics
   - All existing billing endpoints remain functional

4. **Image Creation Enforcement**
   - `apps/api/internal/image/default_handler.go` updated
   - Checks limits before allowing image creation
   - Returns `402 Payment Required` when limit exceeded

### Frontend (Web)

1. **Billing Page**
   - `/billing` - New dedicated billing and usage page
   - Shows usage progress, current plan, billing dates
   - Upgrade options for free users
   - Link to Stripe Customer Portal

2. **Navigation**
   - Added "Billing" link to main navigation
   - Accessible to all authenticated users

3. **Upload Page**
   - Already had subscription checking
   - Now enforced on backend as well

## 🚀 Deployment Steps

### 1. Update Stripe Price IDs

The migration includes placeholder price IDs. Update them to match your Stripe Dashboard:

```sql
-- Run this AFTER deploying the migration
UPDATE plans SET price_id = 'price_YOUR_ACTUAL_FREE_PRICE' WHERE code = 'free';
UPDATE plans SET price_id = 'price_YOUR_ACTUAL_PRO_PRICE' WHERE code = 'pro';
UPDATE plans SET price_id = 'price_YOUR_ACTUAL_BUSINESS_PRICE' WHERE code = 'business';
```

Or update the migration file before deploying:
```bash
# Edit: infra/migrations/0009_seed_plans_table.up.sql
# Replace placeholder price IDs with your actual Stripe Price IDs
```

### 2. Run Database Migration

```bash
# Local
make migrate-up

# Render (via Settings → Environment → Run Migration)
# Or via render CLI
```

### 3. Set Environment Variables

**API (Render Web Service)**:
- `STRIPE_SECRET_KEY` - Already configured ✅
- `STRIPE_WEBHOOK_SECRET` - Already configured ✅
- `FRONTEND_URL` - Already configured ✅

**Web (Render Static Site)**:
- `NEXT_PUBLIC_STRIPE_PRICE_PRO=price_xxx` (optional, for upgrade buttons)
- `NEXT_PUBLIC_STRIPE_PRICE_BUSINESS=price_yyy` (optional)

### 4. Deploy Services

```bash
# Deploy API (auto-deploys on push to main)
git push origin main

# Web will auto-deploy as well
```

### 5. Verify Free Tier Limits

1. Create a test user
2. Upload 10 images (free limit)
3. Try to upload an 11th image
4. Should see: "You have reached your monthly image limit"

### 6. Test Subscription Flow

1. Subscribe to Pro plan via `/billing` page
2. Verify you can upload > 10 images
3. Check `/billing` page shows Pro plan and usage

## 🔧 Configuration

### Plan Limits

Current limits (change in migration before deploying):
- Free: 10 images/month
- Pro: 100 images/month  
- Business: 500 images/month

### Stripe Plans

You need to create these in Stripe Dashboard:

1. **Pro Monthly** ($29/month)
   - Create Product: "Pro Plan"
   - Create Price: Recurring, monthly, $29
   - Copy Price ID → use in migration

2. **Business Monthly** ($99/month)
   - Create Product: "Business Plan"
   - Create Price: Recurring, monthly, $99
   - Copy Price ID → use in migration

3. **Free** (optional)
   - Can use a placeholder like `price_free`
   - Or create a $0 price in Stripe

### Webhook Configuration

Existing webhook should handle these events (already configured):
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

No additional webhook configuration needed ✅

## 🧪 Testing

### Manual Testing Checklist

- [ ] Free user can create 10 images
- [ ] Free user blocked at 11th image
- [ ] Error message is clear and helpful
- [ ] Billing page shows correct usage (X/10 images used)
- [ ] Billing page shows correct period dates
- [ ] Subscribe to Pro via billing page
- [ ] After subscribing, can create > 10 images
- [ ] Billing page shows Pro plan details
- [ ] Usage resets at start of new billing period
- [ ] Stripe Customer Portal link works
- [ ] Can cancel subscription via portal
- [ ] After cancellation, reverts to free tier at period end

### API Testing

```bash
# Get usage (requires auth token)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.yourapp.com/api/v1/billing/usage

# Expected response:
# {
#   "images_used": 5,
#   "monthly_limit": 10,
#   "plan_code": "free",
#   "remaining_images": 5,
#   ...
# }

# Try to create image when at limit (should fail with 402)
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"xxx","original_url":"s3://..."}' \
  https://api.yourapp.com/api/v1/images

# Expected response when over limit:
# HTTP 402 Payment Required
# {
#   "error": "usage_limit_exceeded",
#   "message": "You have reached your monthly image limit..."
# }
```

## 🐛 Known Issues / TODOs

### Test Failures

The following test files need to be updated with new function signatures:

**Billing Tests** (need `UsageService` mock parameter):
- `apps/api/internal/billing/default_handler_test.go` (15 failures)

**Image Handler Tests** (need `UsageChecker` and `UserRepository` parameters):
- `apps/api/internal/image/default_handler_test.go` (5 failures)
- `apps/api/internal/image/default_handler_batch_test.go` (5 failures)

**Fix Pattern**:
```go
// Before
handler := NewDefaultHandler(db)

// After (billing)
usageService := &UsageServiceMock{...}
handler := NewDefaultHandler(db, usageService)

// After (image)
usageChecker := &UsageCheckerMock{...}
userRepo := &user.RepositoryMock{...}
handler := NewDefaultHandler(service, usageChecker, userRepo)
```

### Missing Features

1. **Usage Alerts** - Email users at 80%, 90%, 100% of limit
2. **Usage History** - Chart showing usage over time
3. **Overage Handling** - Allow pay-per-image beyond limit
4. **Grace Period** - Allow a few images over limit with warning
5. **Admin Dashboard** - View all users' usage and plans

## 📊 Monitoring

### Key Metrics to Track

1. **Usage API Latency**
   - Target: < 200ms p95
   - Monitor in Render or Grafana

2. **402 Error Rate**
   - Track how often users hit limits
   - High rate = users need to upgrade

3. **Conversion Rate**
   - Free users → Paid subscribers
   - Track in Stripe Dashboard

4. **Image Creation Rate**
   - By plan tier
   - Helps with capacity planning

### Queries for Debugging

```sql
-- Count users by plan
SELECT p.code, COUNT(DISTINCT s.user_id)
FROM plans p
LEFT JOIN subscriptions s ON s.price_id = p.price_id
  AND s.status IN ('active', 'trialing')
GROUP BY p.code;

-- Find users near their limit
SELECT u.id, COUNT(i.id) as images_used
FROM users u
JOIN projects pr ON pr.user_id = u.id
JOIN images i ON i.project_id = pr.id
WHERE i.created_at >= DATE_TRUNC('month', NOW())
GROUP BY u.id
HAVING COUNT(i.id) >= 8  -- 80% of free tier limit
ORDER BY images_used DESC;

-- Check plan data
SELECT * FROM plans;
```

## 🔐 Security Notes

- ✅ Limits enforced server-side (not client-side)
- ✅ Usage checks use authenticated user from JWT
- ✅ No way to bypass limits via API
- ✅ Stripe webhooks validate signatures
- ⚠️ Rate-limit the billing endpoints in production
- ⚠️ Monitor for abuse (rapid account creation)

## 📞 Support

### Common User Questions

**Q: Why can't I upload images?**
A: Check `/billing` page - you may have reached your monthly limit. Upgrade to Pro or wait for period reset.

**Q: When does my usage reset?**
A: Free tier: 1st of each month. Paid: Your subscription renewal date (shown on `/billing` page).

**Q: How do I cancel my subscription?**
A: Go to `/billing` → "Manage Subscription" → Stripe portal → Cancel.

**Q: What happens after I cancel?**
A: You keep Pro benefits until end of current period, then revert to Free tier.

## 📚 Related Documentation

- [Full Billing Documentation](./apps/docs/docs/features/billing-and-usage.md)
- [Stripe Integration](./apps/docs/docs/integrations/stripe.md)
- [API Reference](./web/api/v1/oas3.yaml)
