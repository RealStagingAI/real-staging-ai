# Cloudflare CDN Setup for Backblaze B2

This guide walks you through setting up a **FREE** global CDN for your Backblaze B2 storage using Cloudflare.

## Benefits

- 💰 **FREE** - Cloudflare CDN free tier + B2 Bandwidth Alliance = $0 CDN costs
- 🌍 **Global** - Edge caching in 200+ cities worldwide
- ⚡ **Fast** - 50-200ms load times (vs 500-2000ms from B2 direct)
- 📈 **99% cache hit rate** after first load
- 🔒 **HTTPS** - Automatic SSL/TLS encryption

## Prerequisites

- ✅ Backblaze B2 account with bucket (`realstaging-prod`)
- ✅ Cloudflare account (free tier)
- ✅ Custom domain (e.g., `real-staging.ai`)

---

## ⚠️ IMPORTANT: Security Consideration

**This guide describes the simple public bucket approach. However, for production use with private user images, see:**

📖 **[CLOUDFLARE_WORKER_AUTH_CDN.md](./CLOUDFLARE_WORKER_AUTH_CDN.md)** - Secure authenticated CDN using Cloudflare Workers

The Worker approach keeps your bucket private and enforces user authentication and ownership checks.

---

## Step 1: Configure Backblaze B2 Bucket

### Option A: Make Bucket Public (⚠️ NOT SECURE FOR PRIVATE IMAGES)

**Use only for:**
- Public marketing images
- Non-sensitive content
- Content that doesn't require access control

1. Log in to [Backblaze B2 Console](https://secure.backblaze.com/b2_buckets.htm)
2. Find your bucket: `realstaging-prod`
3. Click **Bucket Settings**
4. Under **Files in Bucket are**: Select **Public**
5. Click **Update Bucket**

**Pros:**
- Simple setup
- No authentication needed
- Direct CDN caching

**Cons:**
- ❌ **Anyone with URL can access images**
- ❌ No access control
- ❌ Cannot revoke access
- ❌ Compliance issues (GDPR, etc.)

### Option B: Keep Bucket Private + Cloudflare Worker (✅ RECOMMENDED)

**See [CLOUDFLARE_WORKER_AUTH_CDN.md](./CLOUDFLARE_WORKER_AUTH_CDN.md) for full implementation.**

Keep bucket private and use Cloudflare Workers to:
- Authenticate users via JWT tokens
- Enforce image ownership
- Cache authorized responses at edge
- Provide secure, fast global delivery

**Pros:**
- ✅ Images remain private
- ✅ Full access control
- ✅ User-specific authorization
- ✅ CDN caching benefits
- ✅ GDPR compliant

**Cons:**
- More complex setup (worth it for security)

---

## Step 2: Get B2 Bucket URL

1. In Backblaze B2 console, go to **Bucket Details**
2. Note the **Friendly URL**:
   ```
   https://f004.backblazeb2.com/file/realstaging-prod/
   ```
3. Or use the S3-compatible endpoint:
   ```
   https://s3.us-west-004.backblazeb2.com/realstaging-prod/
   ```

---

## Step 3: Configure Cloudflare DNS

### 3.1 Add Domain to Cloudflare

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Click **Add a Site**
3. Enter your domain: `real-staging.ai`
4. Choose **Free** plan
5. Follow steps to update nameservers at your domain registrar

### 3.2 Create CDN Subdomain

1. In Cloudflare Dashboard → **DNS** → **Records**
2. Click **Add Record**
3. Configure:
   - **Type**: `CNAME`
   - **Name**: `cdn` (creates `cdn.real-staging.ai`)
   - **Target**: `f004.backblazeb2.com`
   - **Proxy status**: ✅ **Proxied** (orange cloud)
   - **TTL**: Auto
4. Click **Save**

---

## Step 4: Configure Cloudflare Page Rules (Free Tier)

Page Rules optimize caching behavior.

1. Go to **Rules** → **Page Rules**
2. Click **Create Page Rule**
3. Configure:

**Page Rule #1: Cache Everything**
```
URL Match: cdn.real-staging.ai/file/realstaging-prod/*
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  - Browser Cache TTL: 1 hour
```

**Page Rule #2: Bypass cache for staging images (if needed)**
```
URL Match: cdn.real-staging.ai/file/realstaging-prod/staging/*
Settings:
  - Cache Level: Bypass
```

**Free Tier Limit**: 3 page rules (you have 2 left)

---

## Step 5: Configure CORS (If Needed)

If you get CORS errors in browser:

1. Go to Backblaze B2 → **Bucket Settings** → **CORS Rules**
2. Add:
```json
{
  "corsRuleName": "allow-cloudflare-cdn",
  "allowedOrigins": [
    "https://real-staging.ai",
    "https://cdn.real-staging.ai"
  ],
  "allowedHeaders": ["*"],
  "allowedOperations": ["b2_download_file_by_name"],
  "exposeHeaders": [],
  "maxAgeSeconds": 3600
}
```

---

## Step 6: Update Application Code

### 6.1 Add Environment Variable

Add to Render dashboard and local `.env.local`:

```bash
# Cloudflare CDN URL (optional, falls back to S3_ENDPOINT)
S3_CDN_URL=https://cdn.real-staging.ai/file/realstaging-prod
```

Update `render.yaml`:
```yaml
- key: S3_CDN_URL
  value: https://cdn.real-staging.ai/file/realstaging-prod
```

### 6.2 Update Backend to Support CDN URLs

The presigned URL logic needs to optionally return CDN URLs for **completed** images.

**For images that are done processing:**
- Use CDN URL (cached, fast)

**For uploads and in-progress images:**
- Use presigned S3 URL (direct, not cached)

---

## Step 7: Test CDN Setup

### 7.1 Test DNS Resolution
```bash
dig cdn.real-staging.ai +short
# Should show Cloudflare IP addresses
```

### 7.2 Test CDN Response
```bash
# Upload a test image to B2 first, then:
curl -I https://cdn.real-staging.ai/file/realstaging-prod/test.jpg

# Look for Cloudflare headers:
# cf-cache-status: HIT (cached) or MISS (first request)
# cf-ray: <ray-id>
# server: cloudflare
```

### 7.3 Check Cache Status

First request:
```
cf-cache-status: MISS
```

Second request (should be cached):
```
cf-cache-status: HIT
```

---

## Step 8: Monitor Performance

### Cloudflare Analytics (Free)

1. Cloudflare Dashboard → **Analytics** → **Traffic**
2. Monitor:
   - Cache hit rate (target: >95%)
   - Bandwidth saved
   - Response time

### Expected Metrics

**Before CDN:**
- Load time: 500-2000ms
- Cache hit rate: 0%
- Bandwidth cost: Full B2 egress

**After CDN:**
- Load time: 50-200ms ✅
- Cache hit rate: 95-99% ✅
- Bandwidth cost: ~$0 ✅

---

## Step 9: Add Cache-Control Headers (Backend)

Update S3 uploads to include proper cache headers.

See implementation in: `apps/api/internal/http/upload_handler.go`

```go
CacheControl: aws.String("public, max-age=31536000, immutable")
```

This tells Cloudflare and browsers to cache for 1 year (images are immutable).

---

## Troubleshooting

### Images Not Loading

1. **Check bucket is public** (if using public bucket method)
2. **Verify DNS**: `dig cdn.real-staging.ai`
3. **Check CORS** if seeing browser errors
4. **Check Cloudflare proxy** is enabled (orange cloud)

### Cache Not Working

1. **Check Page Rules** are configured correctly
2. **Verify URL path** matches Page Rule pattern
3. **Check cf-cache-status header**:
   ```bash
   curl -I https://cdn.real-staging.ai/file/realstaging-prod/yourfile.jpg | grep cf-cache-status
   ```

### CORS Errors

1. Add CORS rules to B2 bucket
2. Ensure Cloudflare is proxied (orange cloud)
3. Check `Access-Control-Allow-Origin` header

---

## Cost Breakdown

| Service | Cost | Notes |
|---------|------|-------|
| Backblaze B2 Storage | $6/TB/month | You're already paying this |
| B2 → Cloudflare Bandwidth | **FREE** | Bandwidth Alliance |
| Cloudflare CDN | **FREE** | Free tier unlimited |
| **Total CDN Cost** | **$0** | 🎉 |

---

## Alternative: Cloudflare R2 (Future Consideration)

Cloudflare R2 is an S3-compatible storage that's already CDN-enabled:
- **Free egress** to internet
- **Lower storage cost** than B2 ($0.015/GB vs B2's $0.006/GB)
- **Simpler setup** (no separate CDN config needed)

**Migration path:** B2 → Cloudflare R2 (if you want to simplify further)

---

## Next Steps

After CDN is working:

1. ✅ Monitor cache hit rates in Cloudflare Analytics
2. ✅ Add Cache-Control headers to all uploads
3. ✅ Consider image optimization (WebP conversion)
4. ✅ Test from different geographic locations
5. ✅ Update documentation with new CDN URLs

---

## References

- [Backblaze + Cloudflare Integration](https://www.backblaze.com/b2/docs/cloudflare.html)
- [Cloudflare Page Rules](https://developers.cloudflare.com/rules/page-rules/)
- [Cloudflare Cache](https://developers.cloudflare.com/cache/)
