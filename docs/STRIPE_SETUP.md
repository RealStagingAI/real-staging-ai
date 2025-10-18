# Stripe Integration Setup Guide

This guide explains how to configure Stripe for your Real Staging AI application using the centralized `config/*.yml` system.

## Quick Start

1. **Get your Stripe test API key** from https://dashboard.stripe.com/test/apikeys
2. **Update `config/dev.yml`** with your key (already done! ✅)
3. **Restart the API:** `make down && make up`
4. **Test the integration** at http://localhost:3000/profile

---

## Configuration System

Real Staging AI uses a centralized config system located in `config/*.yml`:

- `config/shared.yml` - Shared settings across all environments
- `config/dev.yml`    - Development-specific settings (Docker Compose)
- `config/local.yml`  - Local machine settings (running outside Docker)
- `config/prod.yml`   - Production settings

### Stripe Configuration in `config/dev.yml`

You've already added the Stripe configuration:

```yaml
stripe:
  secret_key: sk_test_...
  webhook_secret: 
  frontend_url: http://localhost:3000
```

✅ **This is the recommended approach!** All config in one place.

---

## How It Works

The configuration system uses environment variables that are automatically loaded:

1. **Config file is read** → `apps/api/internal/config/config.go` loads `config/dev.yml`
2. **Environment variables are exported** → Values from YAML can override with env vars
3. **API reads environment** → `os.Getenv("STRIPE_SECRET_KEY")` gets the value

### Docker Compose Integration

The `docker-compose.yml` is already configured to pass Stripe environment variables:

```yaml
api:
  environment:
    - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY:-sk_test_placeholder}
    - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET:-}
    - FRONTEND_URL=${FRONTEND_URL:-http://localhost:3000}
```

**How it works:**
- If you set `STRIPE_SECRET_KEY` in your shell → Docker uses that
- If not set → Docker uses the placeholder value
- **Recommended:** Keep values in `config/dev.yml` and remove env overrides

---

## Setup Steps

### 1. Get Your Stripe API Keys

#### For Development (Test Mode)

1. Go to https://dashboard.stripe.com/test/apikeys
2. Click "Reveal test key" for the **Secret key**
3. Copy the key (starts with `sk_test_...`)

#### For Production

1. Go to https://dashboard.stripe.com/apikeys  
2. Copy the **Live Secret key** (starts with `sk_live_...`)
3. **NEVER commit live keys to git!**

### 2. Update Configuration

**Option A: Use `config/dev.yml` (Recommended)**

Edit `config/dev.yml` and update the `secret_key`:

```yaml
stripe:
  secret_key: sk_test_YOUR_ACTUAL_KEY_HERE
  webhook_secret:   # Leave empty for now
  frontend_url: http://localhost:3000
```

**Option B: Use Environment Variables**

```bash
# Add to your shell profile or .env file
export STRIPE_SECRET_KEY=sk_test_YOUR_KEY_HERE
export FRONTEND_URL=http://localhost:3000
```

### 3. Create Stripe Products & Prices

1. Go to https://dashboard.stripe.com/test/products
2. Click **"+ Add product"**

#### Pro Plan
- **Name:** Pro Plan
- **Description:** 100 images/month with priority processing  
- **Pricing:** $29.00 USD / month (recurring)
- Click **"Save product"**
- **Copy the Price ID** (e.g., `price_1ABC123xyz`)

#### Business Plan
- **Name:** Business Plan
- **Description:** 500 images/month with fastest processing
- **Pricing:** $99.00 USD / month (recurring)
- Click **"Save product"**
- **Copy the Price ID** (e.g., `price_1XYZ789def`)

### 4. Update Frontend with Price IDs

Edit `apps/web/app/profile/page.tsx` (around line 121-126):

```typescript
const priceIds = {
  pro: 'price_YOUR_PRO_PRICE_ID',       // Replace with real ID from Stripe
  business: 'price_YOUR_BUSINESS_PRICE_ID', // Replace with real ID from Stripe
};
```

### 5. Start the Application

```bash
# Restart to pick up new configuration
make down && make up

# The API logs will show:
# "Loaded configuration for environment: dev"
```

---

## Test the Integration

### 1. Access Profile Page

Visit http://localhost:3000/profile

### 2. Select a Pricing Tier

- Click on **Pro** or **Business** to select it
- The selected tier will be highlighted with a checkmark

### 3. Subscribe

1. Click **"Subscribe Now"**
2. You should be redirected to Stripe Checkout
3. Use Stripe test card: **`4242 4242 4242 4242`**
   - Expiry: Any future date
   - CVC: Any 3 digits  
   - ZIP: Any 5 digits
4. Click "Subscribe"
5. You'll be redirected to `/profile?checkout=success`

### 4. Verify Subscription

- The profile page should now show "Active Subscription"
- Click **"Manage Subscription & Payment Methods"**
- You'll be redirected to Stripe Customer Portal
- Try updating payment method or canceling

---

## Webhook Configuration (Optional)

Webhooks allow Stripe to notify your backend when events occur (e.g., payment succeeded, subscription canceled).

### For Local Development

Use the Stripe CLI to forward webhooks to localhost:

```bash
# Install Stripe CLI: https://stripe.com/docs/stripe-cli
brew install stripe/stripe-cli/stripe

# Login to your Stripe account
stripe login

# Forward webhooks to your local API
stripe listen --forward-to localhost:8080/api/v1/stripe/webhook

# Copy the webhook signing secret (starts with whsec_...)
# Add to config/dev.yml:
```

```yaml
stripe:
  secret_key: sk_test_...
  webhook_secret: whsec_YOUR_WEBHOOK_SECRET  # From stripe listen
  frontend_url: http://localhost:3000
```

### For Production

1. Go to https://dashboard.stripe.com/webhooks
2. Click **"+ Add endpoint"**
3. Set **Endpoint URL:** `https://yourdomain.com/api/v1/stripe/webhook`
4. Select events:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.paid`
   - `invoice.payment_failed`
5. Click **"Add endpoint"**
6. Copy the **Signing secret** → Add to `config/prod.yml`

---

## Troubleshooting

### 503 Error: "Stripe not configured"

**Problem:** API can't find `STRIPE_SECRET_KEY`

**Solution:**
1. Check `config/dev.yml` has the correct key
2. Restart Docker: `make down && make up`
3. Check API logs: `docker logs virtual-staging-ai-api-1`

```bash
# Verify environment variable in container
docker exec virtual-staging-ai-api-1 env | grep STRIPE
```

### "Invalid API Key" Error

**Problem:** Stripe key is wrong or expired

**Solution:**
1. Go to https://dashboard.stripe.com/test/apikeys  
2. Reveal and copy the Secret key again
3. Update `config/dev.yml`
4. Restart: `make down && make up`

### Pricing Tiers Not Clickable

**Problem:** Frontend not allowing tier selection

**Solution:**
1. Make sure you pulled the latest frontend code
2. Check browser console for errors
3. Hard refresh: Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)

### Webhook Signature Verification Failed

**Problem:** Webhook secret doesn't match

**Solution:**
1. If using Stripe CLI: Copy `whsec_` from `stripe listen` output
2. If using dashboard: Copy from webhook endpoint settings
3. Update `config/dev.yml` with the secret
4. Restart: `make down && make up`

---

## Configuration Reference

### Complete Stripe Configuration

```yaml
# config/dev.yml
stripe:
  secret_key: sk_test_YOUR_KEY_HERE
  webhook_secret: whsec_YOUR_SECRET_HERE  # Optional for local
  frontend_url: http://localhost:3000
```

### Environment Variable Overrides

You can override any config value with environment variables:

```bash
export STRIPE_SECRET_KEY=sk_test_override
export STRIPE_WEBHOOK_SECRET=whsec_override
export FRONTEND_URL=http://localhost:3001
```

### Production Configuration

```yaml
# config/prod.yml
stripe:
  secret_key: sk_live_YOUR_LIVE_KEY  # NEVER commit this!
  webhook_secret: whsec_YOUR_PROD_SECRET
  frontend_url: https://app.yourdomain.com
```

**Security Note:** For production, use environment variables or a secrets manager (e.g., AWS Secrets Manager, HashiCorp Vault) instead of storing live keys in YAML files.

---

## Next Steps

1. ✅ Configure Stripe test mode
2. ✅ Test complete checkout flow  
3. ✅ Test customer portal
4. ⬜ Set up webhook endpoint (optional)
5. ⬜ Test with real pricing tiers
6. ⬜ Deploy to production with live keys

---

## Support

- **Stripe Documentation:** https://stripe.com/docs
- **Stripe Test Cards:** https://stripe.com/docs/testing#cards
- **Stripe API Reference:** https://stripe.com/docs/api

**Questions?** Check the logs:
```bash
# API logs
docker logs -f virtual-staging-ai-api-1

# All services
docker-compose logs -f
```
