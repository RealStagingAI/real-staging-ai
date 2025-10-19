# Stripe Webhook Setup Guide

Complete guide for configuring Stripe webhooks for Real Staging AI production deployment.

## Overview

Stripe webhooks notify your application about events that happen in your Stripe account, such as successful payments, subscription changes, and customer updates. This guide shows you exactly which events to configure and how to set up the webhook endpoint securely.

## Prerequisites

- ✅ Stripe account in **Live Mode**
- ✅ Production API deployed and accessible (e.g., `https://realstaging-api.onrender.com`)
- ✅ SSL/TLS certificate configured (automatic with Render)

---

## Step 1: Navigate to Webhooks in Stripe Dashboard

1. **Log into Stripe Dashboard**: [https://dashboard.stripe.com](https://dashboard.stripe.com)
2. **Switch to Live Mode**: Toggle in the top-right corner (important!)
3. **Go to Webhooks**: 
   - Click **Developers** in the left sidebar
   - Click **Webhooks**
4. **Click "Add endpoint"** button

---

## Step 2: Configure Endpoint URL

**Endpoint URL Format:**
```
https://your-api-domain.com/api/v1/stripe/webhook
```

**Examples:**
- Render default: `https://realstaging-api.onrender.com/api/v1/stripe/webhook`
- Custom domain: `https://api.yourdomain.com/api/v1/stripe/webhook`

**Important Notes:**
- ✅ Must use HTTPS (not HTTP)
- ✅ Must be publicly accessible
- ✅ Must match your production API URL exactly
- ❌ Do NOT use localhost or development URLs for live webhooks

---

## Step 3: Select Events to Listen To

Real Staging AI requires **9 specific webhook events** to function properly. Select **exactly these events**:

### Required Events (9 total)

#### Checkout Events (1)
- [x] **`checkout.session.completed`**
  - **Purpose**: Links Stripe customers to users after successful checkout
  - **Critical**: Required for subscription activation

#### Customer Events (3)
- [x] **`customer.created`**
  - **Purpose**: Logs when new customers are created
  - **Optional**: Currently logs only, future expansion planned

- [x] **`customer.updated`**
  - **Purpose**: Tracks customer detail changes
  - **Optional**: Currently logs only, future expansion planned

- [x] **`customer.deleted`**
  - **Purpose**: Handles customer deletion
  - **Optional**: Currently logs only, future expansion planned

#### Subscription Events (3)
- [x] **`customer.subscription.created`**
  - **Purpose**: Activates new subscriptions in database
  - **Critical**: Required for subscription features

- [x] **`customer.subscription.updated`**
  - **Purpose**: Syncs subscription changes (upgrades, downgrades, renewals)
  - **Critical**: Required for subscription state management

- [x] **`customer.subscription.deleted`**
  - **Purpose**: Deactivates canceled subscriptions
  - **Critical**: Required for subscription cancellation

#### Invoice Events (2)
- [x] **`invoice.payment_succeeded`**
  - **Purpose**: Records successful payments and creates invoice records
  - **Critical**: Required for billing history

- [x] **`invoice.payment_failed`**
  - **Purpose**: Handles failed payments and alerts
  - **Critical**: Required for payment failure handling

---

## Step 4: Configure Event Filters

### How to Select Events in Stripe Dashboard

**Option 1: Select Events (Recommended)**

1. In the "Add endpoint" dialog, under "Events to send", choose **"Select events"**
2. Use the search box to find each event by name
3. Check the box next to each of the 9 events listed above
4. Verify you have exactly **9 events selected**

**Option 2: Send All Events (Not Recommended)**

- Sends all Stripe events (100+ events)
- Creates unnecessary load and logs
- Harder to debug specific issues
- Use only for testing or if you plan to handle more events

### Event Selection Screenshot Reference

```
┌─────────────────────────────────────────────┐
│ Select events to send                       │
├─────────────────────────────────────────────┤
│ Search: "checkout"                          │
│                                             │
│ ☑ checkout.session.completed                │
│                                             │
├─────────────────────────────────────────────┤
│ Search: "customer"                          │
│                                             │
│ ☑ customer.created                          │
│ ☑ customer.updated                          │
│ ☑ customer.deleted                          │
│ ☑ customer.subscription.created             │
│ ☑ customer.subscription.updated             │
│ ☑ customer.subscription.deleted             │
│                                             │
├─────────────────────────────────────────────┤
│ Search: "invoice"                           │
│                                             │
│ ☑ invoice.payment_succeeded                 │
│ ☑ invoice.payment_failed                    │
│                                             │
└─────────────────────────────────────────────┘

Selected: 9 events
```

---

## Step 5: Configure Webhook Settings

### API Version

- **Recommended**: Use the **latest API version** (default)
- **Alternative**: Pin to a specific version if needed (e.g., `2024-10-28`)
- Your application handles webhooks using standard Stripe JSON format

### Description (Optional)

Add a descriptive note for your reference:
```
Production webhook for Real Staging AI - Handles subscriptions and payments
```

---

## Step 6: Add the Endpoint

1. **Review Your Configuration:**
   - Endpoint URL: `https://your-api-domain.com/api/v1/stripe/webhook`
   - Events: 9 selected (see list above)
   - API version: Latest (or pinned version)

2. **Click "Add endpoint"**

3. **Copy the Signing Secret:**
   - Stripe will show: `whsec_...` (long string)
   - **CRITICAL**: This is shown only once!
   - Click the copy icon or select and copy manually
   - Store in a secure location immediately

---

## Step 7: Configure Webhook Secret in Your Application

### For Render Deployment

1. **Go to Render Dashboard**: [https://dashboard.render.com](https://dashboard.render.com)

2. **Navigate to API Service**:
   - Find `realstaging-api` service
   - Click to open service details

3. **Go to Environment Variables**:
   - Click **"Environment"** in the left sidebar
   - Or go to **Settings** → **Environment**

4. **Add Webhook Secret**:
   - Click **"Add Environment Variable"**
   - Key: `STRIPE_WEBHOOK_SECRET`
   - Value: `whsec_...` (paste the secret from Stripe)
   - Click **"Save Changes"**

5. **Redeploy** (if auto-deploy is disabled):
   - Go to **"Manual Deploy"** 
   - Click **"Deploy latest commit"**

### For Other Platforms

Set the environment variable `STRIPE_WEBHOOK_SECRET` with your signing secret:

```bash
# Example for Docker/Docker Compose
STRIPE_WEBHOOK_SECRET=whsec_your_actual_secret_here

# Example for Kubernetes
kubectl create secret generic stripe-webhook \
  --from-literal=secret=whsec_your_actual_secret_here
```

---

## Step 8: Test the Webhook

### Using Stripe Dashboard

1. **Go to your webhook endpoint** in Stripe Dashboard
2. Click **"Send test webhook"** button
3. Select an event (e.g., `checkout.session.completed`)
4. Click **"Send test webhook"**
5. Check the **Response** tab for status code **200 OK**

### Expected Response

**Success (200 OK):**
```json
{
  "status": "received"
}
```

**Duplicate Event (200 OK):**
```json
{
  "status": "duplicate"
}
```

**Signature Verification Failed (401 Unauthorized):**
```json
{
  "error": "unauthorized",
  "message": "Invalid webhook signature"
}
```

### Check Application Logs

In Render dashboard (or your platform's logs):

**Successful webhook:**
```
Received Stripe webhook event: checkout.session.completed (ID: evt_...)
Checkout completed - Customer: cus_..., Payment Status: paid
```

**Signature verification failure:**
```
Stripe signature verification failed: no matching signature
```

---

## Step 9: Verify Event Processing

### Test Real Events

**Create a Test Subscription:**

1. Use Stripe test cards (in test mode first):
   - `4242 4242 4242 4242` - Succeeds
   - Any future date, any CVC

2. Complete a checkout session

3. Check that webhooks are received:
   - `checkout.session.completed` - Links customer
   - `customer.subscription.created` - Creates subscription record
   - `invoice.payment_succeeded` - Records payment

### Database Verification

Check that data was persisted correctly:

```sql
-- Check subscriptions table
SELECT * FROM subscriptions 
WHERE stripe_subscription_id = 'sub_...' 
ORDER BY updated_at DESC;

-- Check invoices table
SELECT * FROM invoices 
WHERE stripe_invoice_id = 'in_...' 
ORDER BY created_at DESC;

-- Check users table for stripe_customer_id
SELECT id, email, stripe_customer_id 
FROM users 
WHERE stripe_customer_id = 'cus_...';
```

---

## Security Best Practices

### 1. Always Verify Signatures

✅ **The application automatically verifies signatures** when `STRIPE_WEBHOOK_SECRET` is set.

**How it works:**
- Every webhook includes a `Stripe-Signature` header
- Application computes HMAC-SHA256 of the payload
- Compares computed signature with Stripe's signature
- Rejects requests with invalid or missing signatures
- Enforces 5-minute timestamp tolerance window

### 2. Idempotency Protection

✅ **The application tracks processed events** to prevent duplicate processing.

**How it works:**
- Each webhook event has a unique ID (e.g., `evt_1ABC...`)
- Before processing, checks `processed_events` table
- If already processed, returns `{"status": "duplicate"}`
- After successful processing, records event ID
- Prevents double-charging or duplicate actions

### 3. Environment Separation

- ✅ Use **separate webhook endpoints** for test and live modes
- ✅ Use **different signing secrets** for each environment
- ✅ Never use test mode secrets in production
- ✅ Never expose signing secrets in logs or error messages

### 4. Error Handling

The application returns appropriate HTTP status codes:

| Status Code | Meaning | Stripe Action |
|------------|---------|---------------|
| `200 OK` | Event processed successfully | No retry |
| `200 OK` (duplicate) | Event already processed | No retry |
| `401 Unauthorized` | Invalid signature | No retry (security issue) |
| `400 Bad Request` | Invalid payload/JSON | No retry (client error) |
| `500 Internal Error` | Processing error | Automatic retry |
| `503 Service Unavailable` | Webhook secret not configured | No retry |

**Stripe's Retry Behavior:**
- Retries failed webhooks automatically
- Uses exponential backoff
- Continues retrying for up to 3 days
- Sends email alerts after repeated failures

---

## Troubleshooting

### Common Issues

#### 1. Signature Verification Failures

**Symptom:** Webhooks return 401 Unauthorized

**Causes:**
- Incorrect `STRIPE_WEBHOOK_SECRET` value
- Using test mode secret in live mode (or vice versa)
- Secret not configured as environment variable
- Secret contains extra spaces or newlines

**Solutions:**
```bash
# Verify secret is set correctly
echo $STRIPE_WEBHOOK_SECRET

# Should start with whsec_
# Should be exactly as shown in Stripe dashboard
# No quotes, spaces, or newlines

# In Render dashboard:
# - Go to Environment variables
# - Find STRIPE_WEBHOOK_SECRET
# - Click Edit
# - Verify value matches Stripe exactly
# - Save and redeploy
```

#### 2. Events Not Being Received

**Symptom:** No webhook logs in application

**Causes:**
- Incorrect endpoint URL
- Firewall blocking Stripe IPs
- SSL certificate issues
- Service not running

**Solutions:**
```bash
# Test endpoint is reachable
curl https://realstaging-api.onrender.com/health

# Should return:
# {"status":"healthy","database":"connected","redis":"connected"}

# Test webhook endpoint (will fail signature check but confirms reachability)
curl -X POST https://realstaging-api.onrender.com/api/v1/stripe/webhook \
  -H "Content-Type: application/json" \
  -d '{"type":"test.event"}'

# Should return 401 or 400 (endpoint is reachable)
```

#### 3. Events Received But Not Processed

**Symptom:** Logs show event received but no database changes

**Causes:**
- Database connection issues
- Missing required fields in webhook payload
- Application errors during processing

**Solutions:**
```bash
# Check application logs for errors
# Look for lines containing:
# - "Error handling checkout.session.completed"
# - "Failed to upsert subscription"
# - "Failed to upsert invoice"

# Check database connection
# In Render shell:
psql $DATABASE_URL -c "SELECT 1;"

# Check that webhook event was recorded
psql $DATABASE_URL -c "SELECT * FROM processed_events ORDER BY created_at DESC LIMIT 5;"
```

#### 4. Duplicate Events

**Symptom:** Same event processed multiple times

**Causes:**
- Idempotency check not working
- Database issues with `processed_events` table
- Multiple webhook endpoints configured

**Solutions:**
```bash
# Verify only one webhook endpoint in Stripe dashboard
# Check processed_events table exists
psql $DATABASE_URL -c "\d processed_events;"

# Check recent processed events
psql $DATABASE_URL -c "
  SELECT event_id, event_type, created_at 
  FROM processed_events 
  ORDER BY created_at DESC 
  LIMIT 10;
"
```

---

## Webhook Event Reference

### Complete Event Details

#### checkout.session.completed

**When it fires:**
- User completes checkout (payment or free trial)
- Payment is successful

**What the app does:**
- Links Stripe customer ID to user account
- Uses `client_reference_id` to find user (Auth0 sub)
- Updates `users.stripe_customer_id`

**Payload excerpt:**
```json
{
  "type": "checkout.session.completed",
  "data": {
    "object": {
      "id": "cs_...",
      "customer": "cus_...",
      "client_reference_id": "auth0|...",
      "payment_status": "paid",
      "subscription": "sub_..."
    }
  }
}
```

#### customer.subscription.created

**When it fires:**
- New subscription is created
- Usually after checkout or API call

**What the app does:**
- Creates subscription record in database
- Stores subscription ID, status, price ID
- Links to user via customer ID

**Important fields:**
- `id`: Subscription ID
- `status`: `active`, `trialing`, etc.
- `current_period_start/end`: Billing period
- `items.data[0].price.id`: Price ID

#### customer.subscription.updated

**When it fires:**
- Subscription changes (upgrade, downgrade, renewal)
- Status changes (active → past_due)
- Billing cycle updates

**What the app does:**
- Updates subscription record with new data
- Syncs status, price, period dates
- Handles plan changes

#### customer.subscription.deleted

**When it fires:**
- Subscription is canceled
- Either immediate or at period end

**What the app does:**
- Marks subscription as `canceled`
- Records cancellation timestamp
- Preserves subscription history

**Important fields:**
- `canceled_at`: When cancellation happened
- `cancel_at_period_end`: If access continues until period ends

#### invoice.payment_succeeded

**When it fires:**
- Recurring payment succeeds
- Initial subscription payment succeeds

**What the app does:**
- Creates invoice record in database
- Records amount paid, currency
- Links to subscription and user

**Important fields:**
- `id`: Invoice ID
- `amount_paid`: Amount in cents
- `currency`: e.g., "usd"
- `subscription`: Subscription ID

#### invoice.payment_failed

**When it fires:**
- Payment fails (insufficient funds, expired card)
- Retry payment fails

**What the app does:**
- Creates invoice record with `failed` status
- Enables dunning and notification logic
- Tracks failed payment attempts

---

## Monitoring and Maintenance

### Regular Checks

**Weekly:**
- [ ] Review webhook delivery success rate in Stripe dashboard
- [ ] Check for any failed webhooks
- [ ] Verify no signature verification errors in logs

**Monthly:**
- [ ] Review processed_events table size
- [ ] Consider archiving old events (optional)
- [ ] Check for any unhandled event types

### Webhook Analytics in Stripe

**Access Webhook Logs:**
1. Stripe Dashboard → Developers → Webhooks
2. Click on your webhook endpoint
3. View recent deliveries

**Key Metrics:**
- **Success rate**: Should be >99%
- **Response time**: Should be <1 second
- **Failed deliveries**: Investigate if >5 in a row

### Alerting

Set up alerts for:
- Multiple consecutive webhook failures
- Sudden increase in signature verification errors
- Webhook endpoint unreachable
- Processing errors in application logs

---

## Production Checklist

Before going live with webhooks:

- [ ] Webhook endpoint configured for production URL
- [ ] All 9 required events selected
- [ ] Using **Live Mode** in Stripe (not test mode)
- [ ] `STRIPE_WEBHOOK_SECRET` set in production environment
- [ ] Secret verified (no spaces, correct format)
- [ ] Test webhook sent successfully (200 OK)
- [ ] Real subscription test completed end-to-end
- [ ] Database shows subscription and invoice records
- [ ] Application logs show successful event processing
- [ ] No signature verification errors
- [ ] Monitoring/alerting configured
- [ ] Documented secret storage location

---

## Advanced Topics

### Webhook Ordering

**Note:** Stripe does not guarantee webhook delivery order.

**Implications:**
- `customer.subscription.updated` may arrive before `customer.subscription.created`
- Application uses upsert logic to handle out-of-order events
- Idempotency prevents duplicate processing

### Webhook Retries

Stripe retries failed webhooks:
- Immediately
- 1 hour later
- 3 hours later
- 6 hours later
- 12 hours later
- 24 hours later
- 48 hours later
- 72 hours later

**Best practices:**
- Return 2xx quickly (within 1 second)
- Process complex logic asynchronously if needed
- Always return 200 even for duplicate events

### Multiple Webhook Endpoints

You can create multiple endpoints:
- Separate endpoints for different environments
- Separate endpoints for different event groups
- Backup endpoints for redundancy

**Not recommended unless needed** - adds complexity

### Webhook Secret Rotation

To rotate webhook secrets:
1. Create new webhook endpoint with new URL
2. Configure new secret in application
3. Test new endpoint
4. Disable old endpoint
5. Remove old secret from application

See [Security: Stripe Webhooks](../security/stripe-webhooks.md) for detailed rotation procedures.

---

## Support and Resources

### Documentation
- [Stripe Webhooks Docs](https://stripe.com/docs/webhooks)
- [Stripe Event Types](https://stripe.com/docs/api/events/types)
- [Testing Webhooks](https://stripe.com/docs/webhooks/test)

### Internal Resources
- [Stripe Billing Guide](stripe-billing.md)
- [Stripe Webhook Security](../security/stripe-webhooks.md)
- [Configuration Guide](configuration.md)

### Troubleshooting
- Check application logs in Render dashboard
- Review webhook logs in Stripe dashboard
- Verify environment variables are set correctly
- Test with Stripe CLI for local debugging

---

## Summary

**Required Webhook Events (9):**
1. `checkout.session.completed` ⭐ Critical
2. `customer.created`
3. `customer.updated`
4. `customer.deleted`
5. `customer.subscription.created` ⭐ Critical
6. `customer.subscription.updated` ⭐ Critical
7. `customer.subscription.deleted` ⭐ Critical
8. `invoice.payment_succeeded` ⭐ Critical
9. `invoice.payment_failed` ⭐ Critical

**Endpoint URL:**
```
https://your-api-domain.com/api/v1/stripe/webhook
```

**Required Environment Variable:**
```bash
STRIPE_WEBHOOK_SECRET=whsec_...
```

**Success Indicator:**
- Webhook test returns 200 OK
- Events appear in application logs
- Database records created for subscriptions/invoices
- No signature verification errors
