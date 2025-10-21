# Deployment Guide - Cloudflare CDN Worker

## Quick Start

### 1. Install Wrangler CLI

Wrangler is installed via npm (no longer available via Homebrew):

```bash
# Install globally
npm install -g wrangler

# Verify installation
wrangler --version
```

### 2. Install Project Dependencies

```bash
cd cloudflare-cdn-worker
npm install
```

### 3. Login to Cloudflare

```bash
wrangler login
```

This will open a browser to authenticate with Cloudflare.

### 4. Set Secrets

```bash
# Set B2 credentials
npm run secret:b2-key
# Enter your B2 keyID

npm run secret:b2-secret
# Enter your B2 applicationKey

# Set worker secret (generate a strong random string)
npm run secret:worker
# Enter the same value as WORKER_SECRET in your API environment
```

### 5. Update Configuration

Edit `wrangler.toml` and verify:
- `AUTH0_DOMAIN` - Should match your Auth0 domain
- `AUTH0_AUDIENCE` - Should match your API audience  
- `API_BASE_URL` - Should be your production API URL
- `B2_BUCKET_NAME` - Your B2 bucket name
- `B2_ENDPOINT` - Your B2 S3 endpoint
- `B2_REGION` - Your B2 region

### 6. Deploy Worker

```bash
npm run deploy
```

Output:
```
Published realstaging-cdn-worker
https://realstaging-cdn-worker.workers.dev
```

### 7. Configure Custom Domain

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Navigate to **Workers & Pages**
3. Click on `realstaging-cdn-worker`
4. Go to **Triggers** tab
5. Click **Add Custom Domain**
6. Enter: `cdn.real-staging.ai`
7. Click **Add Custom Domain**

Cloudflare will:
- Create DNS record automatically
- Provision SSL certificate
- Enable CDN caching

### 8. Update Backend API

Set the `WORKER_SECRET` environment variable in your API:

**Render Dashboard:**
1. Go to your API service
2. Environment → Add Environment Variable
3. Key: `WORKER_SECRET`
4. Value: (same value you set in step 3)

**Local `.env`:**
```bash
WORKER_SECRET=your-secret-here
```

### 9. Test

```bash
# Get a JWT token from Auth0
TOKEN="your-jwt-token"

# Test ownership endpoint
curl -H "Authorization: Bearer $TOKEN" \
  https://cdn.real-staging.ai/images/test-image-id/original

# Check response headers
curl -I -H "Authorization: Bearer $TOKEN" \
  https://cdn.real-staging.ai/images/test-image-id/staged
```

Look for:
- `X-Cache-Status: MISS` (first request)
- `X-Cache-Status: HIT` (subsequent requests)
- `X-Worker-Version: 1.0`

## Updating

To deploy changes:

```bash
# Make changes to src/index.ts
npm run deploy
```

Changes are live immediately with zero downtime.

## Rollback

```bash
wrangler rollback
```

## Monitoring

### View Logs

```bash
npm run tail
```

### Cloudflare Dashboard

1. Go to **Workers & Pages** → `realstaging-cdn-worker`
2. Click **Metrics** tab
3. Monitor:
   - Requests/second
   - Success rate
   - CPU time
   - Errors

## Troubleshooting

### Worker Not Working

1. Check secrets are set:
   ```bash
   wrangler secret list
   ```

2. View logs:
   ```bash
   npm run tail
   ```

3. Verify API endpoint is accessible

### Images Not Loading

1. Test ownership endpoint directly
2. Check JWT token is valid
3. Verify B2 credentials
4. Check worker logs for errors

### Cache Not Working

1. Verify `Cache-Control` headers
2. Check `Vary: Authorization` header
3. Test with different users

## Cost

**Free Tier (100,000 requests/day):**
- Most applications stay within free tier
- $0/month

**Paid (if needed):**
- $5/month base
- $0.50 per million requests beyond free tier

## Security

### Rotate Secrets

```bash
# Generate new WORKER_SECRET
openssl rand -hex 32

# Update in worker
npm run secret:worker

# Update in API
# Set WORKER_SECRET in Render dashboard
```

### Update B2 Credentials

```bash
npm run secret:b2-key
npm run secret:b2-secret
```

## Performance

Target metrics:
- Cache hit rate: >95%
- Response time (cache hit): <50ms
- Response time (cache miss): <500ms
- Success rate: >99.9%

## Next Steps

1. ✅ Monitor cache hit rates
2. ✅ Update frontend to use CDN URLs
3. ✅ Add rate limiting (if needed)
4. ✅ Consider image optimization (WebP conversion)
