# Backblaze B2 Storage Setup Guide

Complete guide for configuring Backblaze B2 object storage for Real Staging AI production deployment.

## Overview

Backblaze B2 provides S3-compatible object storage at a fraction of AWS S3 costs. This guide walks you through setting up B2 for storing original images, staged images, and thumbnails.

### Why Backblaze B2?

- **Cost-Effective**: $0.005/GB vs AWS S3's $0.023/GB (5x cheaper)
- **S3-Compatible**: Drop-in replacement for AWS S3
- **Generous Egress**: First 3x storage included free
- **Simple Pricing**: No complex tiers or hidden costs
- **Reliable**: 99.9% uptime SLA with 11 nines durability

### Cost Comparison

**1TB Storage + Moderate Usage:**
- Backblaze B2: ~$5-10/month
- AWS S3: ~$25-50/month
- **Savings: ~$15-40/month (75-80%)**

---

## Prerequisites

- [x] Valid email address
- [x] Credit card (for account verification - free tier available)
- [x] Production domain name (for CORS configuration)

---

## Step 1: Create Backblaze Account

1. **Go to Backblaze B2**: [https://www.backblaze.com/b2/sign-up.html](https://www.backblaze.com/b2/sign-up.html)

2. **Fill out registration form:**
   - Email address
   - Password (strong password required)
   - Company name (or personal name)

3. **Verify email address** and complete account setup

4. **Add payment method:**
   - Go to **My Settings** → **Billing**
   - Add credit/debit card
   - **Note**: Free tier includes 10GB storage + 1GB daily download

5. **Enable B2 Cloud Storage:**
   - Click **B2 Cloud Storage** in left sidebar
   - Accept terms of service
   - Click **Enable**

---

## Step 2: Create Production Bucket

1. **Go to Buckets:**
   - Click **Buckets** in left sidebar
   - Click **Create a Bucket**

2. **Configure Bucket:**

   **Bucket Name:** `realstaging-prod`
   - Must be globally unique
   - Use lowercase letters, numbers, hyphens only
   - Alternative: `yourcompany-realstaging`

   **Files in Bucket:** Select **Private**
   - Requires authentication to access
   - Presigned URLs will still work

   **Default Encryption:** **Disable** (recommended for S3 compatibility)

   **Object Lock:** **Disable** (not needed)

3. **Note the bucket region** (e.g., `us-west-004`, `eu-central-003`)

### Bucket Structure

The application automatically creates this structure:

```
realstaging-prod/
├── uploads/           # Original uploaded images
├── staged/            # AI-processed staged images
└── thumbnails/        # Image thumbnails
```

---

## Step 3: Configure CORS (Required)

CORS is **required** for presigned upload URLs to work from your web application.

1. **Select Your Bucket** → **Bucket Settings** tab

2. **Add CORS Rules:**

### Production CORS Configuration

```json
[
  {
    "corsRuleName": "allowProductionUploads",
    "allowedOrigins": [
      "https://real-staging.ai",
      "https://realstaging-api.onrender.com"
    ],
    "allowedOperations": [
      "s3_put",
      "s3_get",
      "s3_head"
    ],
    "allowedHeaders": ["*"],
    "exposeHeaders": ["ETag", "x-amz-request-id"],
    "maxAgeSeconds": 3600
  }
]
```

3. **Click "Update Bucket"** to save

---

## Step 4: Create Application Key

Application keys provide programmatic access to your B2 bucket.

1. **Go to App Keys** → **Add a New Application Key**

2. **Configure:**
   - **Name**: `realstaging-prod-api`
   - **Bucket Access**: Select **Specific Bucket** → `realstaging-prod`
   - **Type of Access**: **Read and Write**
   - **File name prefix**: Leave empty
   - **Duration**: **Indefinite**

3. **Click "Create New Key"**

### Save Your Credentials

**CRITICAL**: The `applicationKey` is shown **only once**!

You will see:

1. **keyID** (always visible)
   ```
   Example: 005a1b2c3d4e5f6000000000a
   ```
   - This is your **S3_ACCESS_KEY**

2. **applicationKey** (shown only once)
   ```
   Example: K005abcdefGHIJKlmnopQRSTuvwxyz1234567890
   ```
   - This is your **S3_SECRET_KEY**
   - **Copy immediately** and store securely

**Save in password manager:**
```
B2_KEY_ID=005a1b2c3d4e5f6000000000a
B2_APPLICATION_KEY=K005abcdefGHIJKlmnopQRSTuvwxyz1234567890
```

---

## Step 5: Get Configuration Values

Gather these values for your deployment:

1. **Bucket Name**: `realstaging-prod`
2. **Bucket Region**: `us-west-004` (from bucket details)
3. **S3 Endpoint**: `https://s3.us-west-004.backblazeb2.com`
4. **Access Key** (keyID): From Step 4
5. **Secret Key** (applicationKey): From Step 4

### Configuration Summary

```bash
S3_ENDPOINT=https://s3.us-west-004.backblazeb2.com
S3_REGION=us-west-004
S3_BUCKET=realstaging-prod
S3_ACCESS_KEY=<your-keyID>
S3_SECRET_KEY=<your-applicationKey>
S3_USE_PATH_STYLE=false
```

**Important**: `S3_USE_PATH_STYLE=false` is required for B2.

---

## Step 6: Configure in Render

1. **Go to Render Dashboard**: [https://dashboard.render.com](https://dashboard.render.com)

2. **Configure API Service:**
   - Open `realstaging-api` service
   - Go to **Environment** tab
   - Add these variables:
     - `S3_ENDPOINT`: `https://s3.us-west-004.backblazeb2.com`
     - `S3_REGION`: `us-west-004`
     - `S3_BUCKET`: `realstaging-prod`
     - `S3_ACCESS_KEY`: `<your-keyID>`
     - `S3_SECRET_KEY`: `<your-applicationKey>`
     - `S3_USE_PATH_STYLE`: `false`

3. **Configure Worker Service:**
   - Open `realstaging-worker` service
   - Add the same S3 variables

4. **Save and Redeploy**

---

## Step 7: Test Your Configuration

### Test with AWS CLI

```bash
# Install AWS CLI
brew install awscli  # macOS

# Configure
aws configure set aws_access_key_id <your-keyID>
aws configure set aws_secret_access_key <your-applicationKey>

# Test: List bucket
aws s3 ls s3://realstaging-prod \
  --endpoint-url=https://s3.us-west-004.backblazeb2.com

# Test: Upload file
echo "test" > test.txt
aws s3 cp test.txt s3://realstaging-prod/test.txt \
  --endpoint-url=https://s3.us-west-004.backblazeb2.com

# Test: Delete file
aws s3 rm s3://realstaging-prod/test.txt \
  --endpoint-url=https://s3.us-west-004.backblazeb2.com
```

### Test Application

1. **Log into your application**
2. **Create a new project**
3. **Upload an image**
4. **Verify in B2 dashboard:**
   - Buckets → realstaging-prod → Browse Files
   - Should see `uploads/{project_id}/{image_id}.jpg`

---

## Troubleshooting

### Authentication Failures (403/401)

**Causes:**
- Incorrect keyID or applicationKey
- Key doesn't have bucket access
- Key has been revoked

**Solutions:**
- Verify credentials in environment variables
- Check key exists in B2 App Keys
- Verify key has access to correct bucket

### CORS Errors

**Symptom:** Browser console shows CORS error

**Causes:**
- CORS rules not configured
- Domain not in `allowedOrigins`
- Using HTTP instead of HTTPS

**Solutions:**
- Check CORS rules in Bucket Settings
- Verify your domain is listed exactly
- Ensure using HTTPS in allowedOrigins

### Bucket Not Found (404)

**Causes:**
- Incorrect bucket name (case-sensitive)
- Wrong region in endpoint
- Typo in endpoint URL

**Solutions:**
- Copy exact bucket name from B2 dashboard
- Verify endpoint matches bucket region
- Format: `https://s3.{region}.backblazeb2.com`

### Slow Uploads

**Causes:**
- Geographic distance from B2 datacenter
- Large file sizes

**Solutions:**
- Choose bucket region closest to users
- Consider Cloudflare CDN
- Implement upload progress indicators

---

## Security Best Practices

### Bucket Security

✅ **Do:**
- Keep bucket Private
- Use presigned URLs for access
- Restrict application keys to specific buckets

❌ **Don't:**
- Make bucket Public
- Share direct file URLs
- Grant unnecessary permissions

### Key Management

✅ **Do:**
- Create separate keys per environment
- Use descriptive key names
- Store keys in password manager
- Rotate keys periodically (every 90 days)

❌ **Don't:**
- Hardcode keys in source code
- Commit keys to git
- Share keys via email/chat
- Use production keys in development

### CORS Configuration

✅ **Production:**
- Whitelist only production domains
- Use HTTPS origins only
- Limit to required operations

❌ **Avoid:**
- Wildcard origins (`"*"`)
- HTTP origins
- Overly permissive operations

---

## Cost Optimization

### Understanding Pricing

**Storage**: $0.005/GB per month
**Downloads**: First 3× storage free per day
**API Calls**: Generous free tier

### Examples

**Small (100 GB):**
- Storage: $0.50/month
- Downloads: ~150 GB free/day
- **Cost: ~$0.50/month**

**Medium (500 GB):**
- Storage: $2.50/month
- Downloads: ~1.5 TB free/day
- **Cost: ~$2.50-3/month**

**Large (2 TB):**
- Storage: $10/month
- Downloads: ~6 TB free/day
- **Cost: ~$10-12/month**

### Optimization Tips

1. **Use thumbnails** instead of full images
2. **Enable lifecycle policies** to delete old files
3. **Consider Cloudflare CDN** for caching
4. **Monitor usage** in B2 dashboard

---

## Production Checklist

- [ ] B2 account created and verified
- [ ] Production bucket created (`realstaging-prod`)
- [ ] Bucket set to Private
- [ ] CORS rules configured for production domain
- [ ] Application key created with restricted access
- [ ] Credentials saved securely
- [ ] Environment variables configured in Render
- [ ] `S3_USE_PATH_STYLE` set to `false`
- [ ] Test upload completed successfully
- [ ] Test download completed successfully
- [ ] No CORS errors in browser
- [ ] Files visible in B2 dashboard

---

## Support and Resources

### Documentation
- [B2 Documentation](https://www.backblaze.com/b2/docs/)
- [S3 Compatible API](https://www.backblaze.com/b2/docs/s3_compatible_api.html)
- [CORS Configuration](https://www.backblaze.com/b2/docs/cors_rules.html)

### Support
- Email: help@backblaze.com
- Phone: +1 (650) 352-3738

### Internal Resources
- [Configuration Guide](configuration.md)
- [Deployment Guide](../operations/deployment.md)
- [Production Checklist](../operations/production-checklist.md)

---

## Summary

**Configuration:**
```bash
S3_ENDPOINT=https://s3.us-west-004.backblazeb2.com
S3_REGION=us-west-004
S3_BUCKET=realstaging-prod
S3_ACCESS_KEY=<your-keyID>
S3_SECRET_KEY=<your-applicationKey>
S3_USE_PATH_STYLE=false
```

**Cost**: $5-20/month (75-80% savings vs AWS S3)

**Setup Time**: 15-30 minutes
