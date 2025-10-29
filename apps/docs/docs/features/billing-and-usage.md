# Billing and Usage Tracking

This document describes the billing, subscription management, and usage tracking features.

## Overview

The platform implements a tiered subscription model with usage-based limits:

- **Free Tier**: 10 images/month
- **Pro Tier**: 100 images/month ($29/month)
- **Business Tier**: 500 images/month ($99/month)

Usage is tracked per billing period, and limits are enforced at image creation time.

## Architecture

### Database Schema

**Plans Table** (`plans`)
- `id`: UUID primary key
- `code`: Plan code (free, pro, business)
- `price_id`: Stripe Price ID
- `monthly_limit`: Monthly image limit

**Subscriptions Table** (`subscriptions`)
- Links users to their Stripe subscriptions
- Tracks billing period dates
- Stores subscription status

**Images Table** (`images`)
- Each image creation counts toward usage
- Linked to projects, which are linked to users

### Services

**UsageService** (`apps/api/internal/billing/default_usage_service.go`)
- `GetUsage`: Returns current usage statistics for a user
- `CanCreateImage`: Checks if user can create a new image
- `GetPlanByCode`: Retrieves plan information

**SubscriptionChecker** (`apps/api/internal/billing/default_subscription_checker.go`)
- `HasActiveSubscription`: Checks if user has active paid subscription

### API Endpoints

#### GET /api/v1/billing/usage

Returns current usage statistics for the authenticated user.

**Response:**
```json
{
  "images_used": 5,
  "monthly_limit": 10,
  "plan_code": "free",
  "period_start": "2025-10-01T00:00:00Z",
  "period_end": "2025-11-01T00:00:00Z",
  "has_subscription": false,
  "remaining_images": 5
}
```

#### GET /api/v1/billing/subscriptions

Returns active subscriptions for the user.

#### GET /api/v1/billing/invoices

Returns invoice history for the user.

#### POST /api/v1/billing/create-checkout

Creates a Stripe Checkout session for subscription signup.

**Request:**
```json
{
  "price_id": "price_pro_monthly"
}
```

**Response:**
```json
{
  "url": "https://checkout.stripe.com/..."
}
```

#### POST /api/v1/billing/portal

Creates a Stripe Customer Portal session for subscription management.

**Response:**
```json
{
  "url": "https://billing.stripe.com/..."
}
```

## Usage Enforcement

### Image Creation Flow

1. User attempts to create image via `POST /api/v1/images`
2. Handler checks usage limits via `UsageChecker.CanCreateImage()`
3. If limit exceeded, returns `402 Payment Required`:
   ```json
   {
     "error": "usage_limit_exceeded",
     "message": "You have reached your monthly image limit. Please upgrade your plan to continue."
   }
   ```
4. If under limit, image creation proceeds normally

### Billing Period Calculation

- **Free users**: Calendar month (1st to last day of month)
- **Paid users**: Subscription period from Stripe
  - Uses `current_period_start` and `current_period_end` from subscription
  - Falls back to calendar month if subscription data unavailable

### Usage Counting

Images are counted using:
```sql
SELECT COUNT(*)
FROM images i
JOIN projects p ON i.project_id = p.id
WHERE p.user_id = $1
  AND i.created_at >= $2  -- period_start
  AND i.created_at < $3   -- period_end
```

## Frontend Integration

### Billing Page (`/billing`)

The billing page (`apps/web/app/billing/page.tsx`) displays:

- **Current Usage**: Progress bar showing images used vs. limit
- **Current Plan**: Active subscription details and period dates
- **Upgrade Options**: Available plans (if on free tier)
- **Manage Subscription**: Button to access Stripe Customer Portal

### Upload Page Protection

The upload page checks subscription status:
```typescript
const res = await apiFetch('/v1/billing/subscriptions')
const activeSubscription = res.items?.some(
  sub => sub.status === "active" || sub.status === "trialing"
)
```

If no active subscription, a banner prompts the user to subscribe.

## Configuration

### Environment Variables

**API** (`.env` or Render environment variables):
```bash
STRIPE_SECRET_KEY=sk_test_...        # Stripe API secret key
STRIPE_WEBHOOK_SECRET=whsec_...      # Stripe webhook signing secret
FRONTEND_URL=https://app.example.com # Frontend URL for redirects
```

**Web** (`.env.local`):
```bash
NEXT_PUBLIC_STRIPE_PRICE_PRO=price_pro_monthly
NEXT_PUBLIC_STRIPE_PRICE_BUSINESS=price_business_monthly
```

### Seeding Plans

Plans are seeded via migration `0009_seed_plans_table.up.sql`:

```sql
INSERT INTO plans (code, price_id, monthly_limit) VALUES
  ('free', 'price_free', 10),
  ('pro', 'price_pro_monthly', 100),
  ('business', 'price_business_monthly', 999999);
```

**Important**: Update the `price_id` values to match your actual Stripe Price IDs after deployment.

## Stripe Integration

### Webhooks

Stripe webhooks update subscription state:
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

Handler: `apps/api/internal/stripe/default_handler.go`

### Customer Portal

Users can manage their subscriptions via Stripe Customer Portal:
- Update payment method
- Cancel subscription
- View invoices
- Download receipts

## Testing

### Unit Tests

Usage service tests should mock the database and verify:
- Correct usage counting within billing periods
- Proper limit enforcement
- Plan lookups
- Edge cases (no subscription, etc.)

### Integration Tests

Test the complete flow:
1. Create user
2. Create images up to limit
3. Verify next creation fails with 402
4. Add subscription
5. Verify creation succeeds

## Monitoring

### Key Metrics

- **Usage by Plan**: Track average usage per plan tier
- **Limit Exceedances**: Count how often users hit limits
- **Conversion Rate**: Free → Paid subscription rate
- **Churn Rate**: Subscription cancellations

### Dashboards

Monitor in Render or your observability platform:
- Image creation rate
- Subscription events
- Usage API latency
- Failed payment webhooks

## Future Enhancements

- **Usage Alerts**: Email users at 80% and 100% of limit
- **Overage Charges**: Allow pay-per-image beyond limit
- **Annual Plans**: Discounted annual billing
- **Enterprise Plans**: Custom limits and pricing
- **Usage Analytics**: Detailed breakdown in billing page
- **Grace Period**: Allow X images over limit before hard block

## Troubleshooting

### User Can't Create Images (Free Tier)

1. Check usage count: `SELECT COUNT(*) FROM images WHERE ...`
2. Verify plan limit: `SELECT * FROM plans WHERE code = 'free'`
3. Check billing period dates
4. Review error logs for enforcement logic

### Subscription Not Recognized

1. Verify webhook delivery in Stripe Dashboard
2. Check `subscriptions` table for user
3. Ensure `price_id` matches between Stripe and `plans` table
4. Review processed events for idempotency

### Usage Count Incorrect

1. Verify billing period calculation
2. Check for timezone issues (all dates are UTC)
3. Ensure images aren't double-counted
4. Review query logic in `CountImagesCreatedInPeriod`

## Security Considerations

- Never trust client-side usage checks
- Always enforce limits on server side
- Validate Stripe webhook signatures
- Use HTTPS for all redirect URLs
- Don't expose Stripe secret keys in frontend
- Rate-limit billing API endpoints

## Related Documentation

- [Stripe Billing](../guides/stripe-billing.md)
- [API Reference](../api-reference/openapi.md)
- [Database Schema](../architecture/database.md)
- [Production Checklist](../operations/production-checklist.md)
