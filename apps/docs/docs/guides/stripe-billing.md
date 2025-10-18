# Stripe Billing Integration Guide

Real Staging AI uses Stripe for subscription billing, payment processing, and customer management.

## Overview

The billing system provides:

- **Subscription Management** - Recurring monthly billing with multiple pricing tiers
- **Self-Service Portal** - Customers can update payment methods and manage subscriptions
- **Webhook Events** - Real-time synchronization between Stripe and your database
- **Customer Portal** - Stripe-hosted interface for subscription management

### Architecture

```
┌─────────────┐     ┌──────────┐     ┌──────────────┐
│   Web App   │────▶│   API    │────▶│   Stripe     │
│  (React)    │     │  (Go)    │     │  (Payments)  │
└─────────────┘     └──────────┘     └──────────────┘
       │                  ▲                   │
       │                  │                   │
       │                  └───────────────────┘
       │                   Webhooks (async)
       │
       └──▶ Polling (waits for webhook)
```

**Flow:**
1. User selects pricing tier on profile page
2. Frontend calls `/api/v1/billing/create-checkout`
3. API creates Stripe Checkout session
4. User redirected to Stripe (hosted checkout)
5. User completes payment with credit card
6. Stripe redirects back to app with `?checkout=success`
7. Frontend polls `/api/v1/billing/subscriptions` endpoint
8. Stripe webhook fires → API saves subscription to database
9. Frontend receives subscription data → shows "Active Subscription"

## Prerequisites

- Stripe account ([stripe.com](https://stripe.com))
- Stripe CLI for local webhook forwarding ([stripe.com/docs/stripe-cli](https://stripe.com/docs/stripe-cli))
- Database with subscriptions table (run migrations)

## Setting Up Stripe

### Step 1: Create Stripe Account

1. Sign up at [stripe.com](https://stripe.com)
2. Complete account verification
3. Enable **Test Mode** (toggle in top-right of dashboard)

!!! note "Test vs Live Mode"
    Always use Test Mode for development. Live Mode requires full business verification and processes real payments.

### Step 2: Get API Keys

1. Navigate to [Developers → API Keys](https://dashboard.stripe.com/test/apikeys)
2. Copy your **Secret key** (starts with `sk_test_...`)
3. Never commit this to git - store in `apps/api/secrets.yml`

**Add to `apps/api/secrets.yml`:**
```yaml
stripe:
  secret_key: sk_test_YOUR_KEY_HERE
  webhook_secret: # Leave empty for now
```

!!! warning "Security"
    - Never commit API keys to version control
    - Use different keys for development, staging, and production
    - Rotate keys if accidentally exposed
    - Production keys start with `sk_live_` and require extra caution

### Step 3: Create Products & Prices

Products represent your subscription tiers. Prices define the billing amount and interval.

#### Create Pro Plan

1. Go to [Products](https://dashboard.stripe.com/test/products)
2. Click **+ Add product**
3. Configure:
   - **Name**: `Pro Plan`
   - **Description**: `100 images/month with priority processing`
   - **Pricing Model**: `Standard pricing`
   - **Price**: `$29.00`
   - **Billing Period**: `Monthly`
   - **Currency**: `USD`
4. Click **Add product**
5. **Copy the Price ID** (e.g., `price_1ABC...`) - you'll need this for the frontend

#### Create Business Plan

1. Click **+ Add product**
2. Configure:
   - **Name**: `Business Plan`
   - **Description**: `500 images/month with fastest processing`
   - **Pricing Model**: `Standard pricing`
   - **Price**: `$99.00`
   - **Billing Period**: `Monthly`
   - **Currency**: `USD`
3. Click **Add product**
4. **Copy the Price ID** (e.g., `price_1XYZ...`)

!!! tip "Multiple Prices"
    You can add multiple prices to a product for different billing intervals (monthly, yearly) or currencies. Each price gets a unique ID.

### Step 4: Configure Frontend Price IDs

Edit `apps/web/app/profile/page.tsx` and update the price IDs (around line 120):

```typescript
const priceIds = {
  pro: 'price_YOUR_PRO_PRICE_ID',       // From Step 3
  business: 'price_YOUR_BUSINESS_PRICE_ID', // From Step 3
};
```

### Step 5: Install Stripe CLI

The Stripe CLI forwards webhook events from Stripe to your local development server.

**macOS/Linux:**
```bash
brew install stripe/stripe-cli/stripe
```

**Windows:**
```powershell
scoop install stripe
```

**Verify installation:**
```bash
stripe --version
```

### Step 6: Forward Webhooks to Localhost

1. **Login to Stripe:**
   ```bash
   stripe login
   ```
   This opens a browser to authenticate with your Stripe account.

2. **Forward webhooks:**
   ```bash
   stripe listen --forward-to localhost:8080/api/v1/stripe/webhook
   ```

3. **Copy the webhook signing secret** from the output:
   ```
   > Ready! You are using Stripe API Version [2024-10-28]. Your webhook signing secret is whsec_a5420e9d... (^C to quit)
   ```

4. **Add to `apps/api/secrets.yml`:**
   ```yaml
   stripe:
     secret_key: sk_test_51SD3dJ...
     webhook_secret: whsec_a5420e9d...  # From stripe listen
   ```

5. **Rebuild API:**
   ```bash
   docker compose up --build -d api
   ```

!!! important "Keep Stripe CLI Running"
    The `stripe listen` command must stay running while you're testing. Open a dedicated terminal window for it.

### Step 7: Configure Webhook Events (Production Only)

For production deployments, configure webhooks in the Stripe dashboard:

1. Go to [Developers → Webhooks](https://dashboard.stripe.com/webhooks)
2. Click **+ Add endpoint**
3. Configure:
   - **Endpoint URL**: `https://yourdomain.com/api/v1/stripe/webhook`
   - **Description**: `Real Staging AI Production Webhook`
   - **Events to send**:
     - `checkout.session.completed`
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
     - `invoice.paid`
     - `invoice.payment_failed`
4. Click **Add endpoint**
5. Click **Reveal** next to **Signing secret**
6. Add to production secrets (never commit):
   ```yaml
   stripe:
     secret_key: sk_live_YOUR_PRODUCTION_KEY
     webhook_secret: whsec_YOUR_PRODUCTION_SECRET
   ```

## Configuration Reference

### Environment Variables

The API reads Stripe configuration from environment variables and secrets files:

**Priority (highest to lowest):**
1. Environment variables (e.g., `STRIPE_SECRET_KEY`)
2. `apps/api/secrets.yml` (git-ignored)
3. `config/dev.yml` (committed, non-sensitive defaults)

### Complete Configuration

**`apps/api/secrets.yml`** (never commit):
```yaml
stripe:
  secret_key: sk_test_51SD3dJLkQ5x1VWxd...
  webhook_secret: whsec_a5420e9d3fdd3fb1...
```

**`config/dev.yml`** (safe to commit):
```yaml
stripe:
  frontend_url: http://localhost:3000
```

**`docker-compose.yml`** (environment variable overrides):
```yaml
services:
  api:
    environment:
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY:-sk_test_placeholder}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET:-}
      - FRONTEND_URL=${FRONTEND_URL:-http://localhost:3000}
```

### Docker Compose Variables

You can override config with environment variables:

```bash
# Export variables in your shell
export STRIPE_SECRET_KEY=sk_test_override
export FRONTEND_URL=http://localhost:3001

# Start services (will use exported variables)
docker compose up -d
```

## API Endpoints

### POST /api/v1/billing/create-checkout

Creates a Stripe Checkout session for subscription signup.

**Request:**
```json
{
  "price_id": "price_1SJOLOLkQ5x1VWxd..."
}
```

**Response:**
```json
{
  "url": "https://checkout.stripe.com/c/pay/cs_test_..."
}
```

**Flow:**
1. Validates user is authenticated (JWT middleware)
2. Gets or creates user in database
3. Creates or retrieves Stripe customer
4. Creates checkout session with:
   - Success URL: `/profile?checkout=success`
   - Cancel URL: `/profile?checkout=canceled`
5. Returns checkout URL for redirect

**Implementation:** `apps/api/internal/billing/default_handler.go`

### POST /api/v1/billing/portal

Creates a Stripe Customer Portal session for subscription management.

**Request:** (empty body)

**Response:**
```json
{
  "url": "https://billing.stripe.com/p/session/test_..."
}
```

**Flow:**
1. Validates user is authenticated
2. Requires existing Stripe customer (must have subscribed before)
3. Creates portal session with return URL: `/profile`
4. Returns portal URL for redirect

**What customers can do in portal:**
- Update payment methods
- View invoices and payment history
- Cancel subscription
- Update billing information

**Implementation:** `apps/api/internal/billing/default_handler.go`

### GET /api/v1/billing/subscriptions

Lists all active subscriptions for the current user.

**Request:** (no body)

**Response:**
```json
{
  "items": [
    {
      "id": "uuid-...",
      "status": "active",
      "priceId": "price_1SJOLO...",
      "currentPeriodEnd": "2025-11-17T18:30:00Z"
    }
  ]
}
```

**Implementation:** `apps/api/internal/billing/default_handler.go`

### POST /api/v1/stripe/webhook

Receives webhook events from Stripe (public endpoint, no auth required).

**Events handled:**
- `checkout.session.completed` - Links Stripe customer to user
- `customer.subscription.created` - Saves new subscription
- `customer.subscription.updated` - Updates subscription status
- `customer.subscription.deleted` - Marks subscription as canceled
- `invoice.paid` - Records successful payment
- `invoice.payment_failed` - Records failed payment

**Verification:**
- Validates webhook signature using `webhook_secret`
- Rejects requests with invalid signatures (prevents spoofing)

**Implementation:** `apps/api/internal/stripe/default_handler.go`

## Database Schema

### Subscriptions Table

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,  -- 'active', 'canceled', 'past_due', etc.
    price_id TEXT,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_stripe_id ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
```

**Migration:** `infra/migrations/006_add_stripe_tables.up.sql`

### Users Table (Stripe Customer Link)

```sql
ALTER TABLE users ADD COLUMN stripe_customer_id TEXT UNIQUE;
CREATE INDEX idx_users_stripe_customer ON users(stripe_customer_id);
```

This links local users to Stripe customers for subscription management.

### Invoices Table

```sql
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_invoice_id TEXT NOT NULL UNIQUE,
    amount_paid BIGINT NOT NULL,  -- Amount in cents
    currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL,  -- 'paid', 'open', 'void', 'uncollectible'
    invoice_pdf TEXT,  -- URL to PDF
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Migration:** `infra/migrations/007_add_invoices_table.up.sql`

## Testing

### Test with Stripe Test Cards

Stripe provides test credit cards that simulate different scenarios:

**Success:**
- **4242 4242 4242 4242** - Succeeds
- Expiry: Any future date
- CVC: Any 3 digits
- ZIP: Any 5 digits

**Specific scenarios:**
- **4000 0025 0000 3155** - Requires 3D Secure authentication
- **4000 0000 0000 9995** - Declined (insufficient funds)
- **4000 0000 0000 0341** - Declined (charge exceeds limit)

Full list: [stripe.com/docs/testing#cards](https://stripe.com/docs/testing#cards)

### Testing Checkout Flow

1. **Start services:**
   ```bash
   docker compose up -d
   stripe listen --forward-to localhost:8080/api/v1/stripe/webhook
   ```

2. **Open app:**
   ```bash
   open http://localhost:3000/profile
   ```

3. **Select pricing tier:**
   - Click on **Pro** or **Business** card
   - Verify checkmark appears on selected tier

4. **Click "Subscribe Now"**
   - Redirected to Stripe Checkout
   - Enter test card: `4242 4242 4242 4242`
   - Complete checkout

5. **Verify webhook received:**
   In Stripe CLI terminal, you should see:
   ```
   2025-10-17 18:30:00   --> checkout.session.completed [evt_...]
   2025-10-17 18:30:01   --> customer.subscription.created [evt_...]
   ```

6. **Verify subscription appears:**
   - Redirected to `/profile?checkout=success`
   - Shows "Payment successful! Activating..."
   - After ~2 seconds: "Subscription activated successfully!"
   - "Active Subscription" card displays with plan details

### Testing Customer Portal

1. **With active subscription, click "Manage Subscription"**
2. **Redirected to Stripe Customer Portal**
3. **Available actions:**
   - Update payment method
   - Cancel subscription
   - View invoices
   - Download receipts

### Manual Database Testing

If webhooks aren't working, you can manually insert a test subscription:

```sql
-- Get your user ID
SELECT id, auth0_sub FROM users;

-- Insert test subscription
INSERT INTO subscriptions (
  user_id,
  stripe_subscription_id,
  status,
  price_id,
  current_period_end
) VALUES (
  'YOUR_USER_ID',
  'sub_test_manual',
  'active',
  'price_1SJOLOLkQ5x1VWxd...',
  NOW() + INTERVAL '30 days'
);
```

Refresh the profile page - subscription should appear.

## Troubleshooting

### 503 Error: "Stripe not configured"

**Cause:** API can't find `STRIPE_SECRET_KEY`

**Fix:**
1. Verify key in `apps/api/secrets.yml`
2. Restart API: `docker compose up --build -d api`
3. Check logs: `docker logs virtual-staging-ai-api-1`

### Webhook Signature Verification Failed

**Cause:** `webhook_secret` doesn't match Stripe CLI or dashboard

**Fix:**
1. Restart `stripe listen` and copy new secret
2. Update `apps/api/secrets.yml` with new secret
3. Rebuild API: `docker compose up --build -d api`

### No Subscription After Checkout

**Possible causes:**

1. **Webhooks not forwarded:**
   - Is `stripe listen` running?
   - Check terminal for webhook events
   - Verify URL: `localhost:8080/api/v1/stripe/webhook`

2. **Database tables missing:**
   ```bash
   # Check if subscriptions table exists
   docker exec virtual-staging-ai-postgres-1 psql -U postgres -d postgres -c "\dt"
   
   # Run migrations if needed
   make migrate-up
   ```

3. **Webhook secret mismatch:**
   - Check API logs for "signature verification failed"
   - Update secret in `apps/api/secrets.yml`

### Polling Times Out

**Cause:** Webhook processing took >20 seconds (rare)

**Fix:** Refresh the page manually. The subscription should appear.

**Prevention:** Ensure database is fast and not overloaded.

### Test vs Live Mode Confusion

**Symptoms:**
- Test cards don't work
- Products not appearing
- "No such customer" errors

**Fix:**
1. Check Stripe dashboard - toggle in top-right shows current mode
2. Ensure API keys match mode:
   - Test keys: `sk_test_...`, `whsec_...` (local dev)
   - Live keys: `sk_live_...`, `whsec_...` (production only)

## Production Checklist

Before deploying to production:

- [ ] Use live API keys (`sk_live_...`)
- [ ] Configure webhook endpoint in Stripe dashboard (not CLI)
- [ ] Use production webhook secret
- [ ] Store secrets in secure vault (not files)
- [ ] Enable HTTPS on webhook endpoint
- [ ] Create real products with actual prices
- [ ] Update frontend with live price IDs
- [ ] Test with real credit cards (your own)
- [ ] Set up monitoring for webhook failures
- [ ] Configure email notifications for failed payments
- [ ] Review Stripe tax settings if applicable
- [ ] Enable 3D Secure for fraud prevention
- [ ] Set up invoice email templates
- [ ] Test subscription cancellation flow
- [ ] Verify refund process works

## Security Best Practices

1. **Never commit secrets:**
   ```gitignore
   # .gitignore
   apps/api/secrets.yml
   apps/worker/secrets.yml
   *.env
   ```

2. **Rotate keys if exposed:**
   - Roll keys in Stripe dashboard
   - Update all environments
   - Monitor for unauthorized usage

3. **Validate webhook signatures:**
   - Always verify `Stripe-Signature` header
   - Reject requests with invalid signatures
   - Log verification failures

4. **Use environment-specific keys:**
   - Development: Test keys
   - Staging: Separate test keys
   - Production: Live keys (restricted access)

5. **Implement proper error handling:**
   - Don't expose Stripe errors to users
   - Log detailed errors internally
   - Return generic user-facing messages

6. **Rate limit webhook endpoint:**
   - Prevent abuse
   - Stripe retries failed webhooks
   - Use exponential backoff

## Additional Resources

- **Stripe Documentation:** [stripe.com/docs](https://stripe.com/docs)
- **Test Cards:** [stripe.com/docs/testing](https://stripe.com/docs/testing)
- **Webhooks Guide:** [stripe.com/docs/webhooks](https://stripe.com/docs/webhooks)
- **Customer Portal:** [stripe.com/docs/billing/subscriptions/customer-portal](https://stripe.com/docs/billing/subscriptions/customer-portal)
- **Stripe CLI:** [stripe.com/docs/stripe-cli](https://stripe.com/docs/stripe-cli)
- **Go SDK:** [github.com/stripe/stripe-go](https://github.com/stripe/stripe-go)

## Related Guides

- [Authentication](authentication.md) - User authentication with Auth0
- [Configuration](configuration.md) - Environment variables and settings
- [Testing](testing.md) - Unit and integration testing
- [Local Development](local-development.md) - Development environment setup
