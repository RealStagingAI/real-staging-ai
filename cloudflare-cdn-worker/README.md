# Real Staging AI - CDN Worker

Cloudflare Worker that provides authenticated CDN access to private Backblaze B2 images.

## Features

- ✅ **JWT Authentication** - Validates Auth0 tokens using JWKS
- ✅ **Ownership Verification** - Ensures users can only access their own images
- ✅ **Edge Caching** - User-specific caching at Cloudflare edge locations
- ✅ **AWS Signature V4** - Secure access to private B2 bucket
- ✅ **CORS Support** - Handles cross-origin requests
- ✅ **Global CDN** - Fast delivery from 200+ cities worldwide

## Prerequisites

- Cloudflare account (free tier works)
- Wrangler CLI installed: `npm install -g wrangler`
- Backblaze B2 credentials (keyID and applicationKey)
- Auth0 domain and audience configured

## Setup

### 1. Install Dependencies

```bash
cd cloudflare-cdn-worker
npm install
```

### 2. Login to Cloudflare

```bash
wrangler login
```

### 3. Configure Secrets

Set the required secrets:

```bash
# B2 Access Key ID
npm run secret:b2-key
# Enter your B2 keyID when prompted

# B2 Secret Access Key
npm run secret:b2-secret
# Enter your B2 applicationKey when prompted

# Worker Secret (for internal API auth)
npm run secret:worker
# Enter a strong random string when prompted
```

### 4. Update wrangler.toml

Edit `wrangler.toml` and verify:
- `AUTH0_DOMAIN` - Your Auth0 domain
- `AUTH0_AUDIENCE` - Your API audience
- `API_BASE_URL` - Your backend API URL
- `B2_BUCKET_NAME` - Your B2 bucket name
- `B2_ENDPOINT` - Your B2 S3-compatible endpoint
- `B2_REGION` - Your B2 region

### 5. Deploy Worker

```bash
npm run deploy
```

Output will show your worker URL:
```
Published realstaging-cdn-worker
https://realstaging-cdn-worker.workers.dev
```

### 6. Configure Custom Domain

1. Go to Cloudflare Dashboard → **Workers & Pages**
2. Click on `realstaging-cdn-worker`
3. Go to **Triggers** → **Custom Domains**
4. Click **Add Custom Domain**
5. Enter: `cdn.real-staging.ai`
6. Click **Add Custom Domain**

Cloudflare automatically provisions SSL certificate.

## Usage

### Request Format

```
GET https://cdn.real-staging.ai/images/{imageId}/{kind}
Authorization: Bearer {jwt-token}
```

**Parameters:**
- `imageId` - UUID of the image
- `kind` - Either `original` or `staged`

### Example

```bash
# Fetch staged image
curl -H "Authorization: Bearer eyJ..." \
  https://cdn.real-staging.ai/images/abc-123-def/staged
```

### Frontend Integration

```typescript
// Next.js Image component
<NextImage
  src={`https://cdn.real-staging.ai/images/${imageId}/staged`}
  alt="Property"
  width={800}
  height={600}
  loading="lazy"
  unoptimized // Worker handles optimization
/>

// Add Authorization header via fetch
const response = await fetch(
  `https://cdn.real-staging.ai/images/${imageId}/staged`,
  {
    headers: {
      Authorization: `Bearer ${accessToken}`
    }
  }
);
```

## Response Headers

- `X-Cache-Status` - `HIT` (cached) or `MISS` (fetched from B2)
- `X-Worker-Version` - Worker version
- `Cache-Control` - `private, max-age=3600`
- `Vary` - `Authorization` (user-specific caching)

## Development

### Local Development

```bash
npm run dev
```

Access at: `http://localhost:8787`

### View Logs

```bash
npm run tail
```

### Test Authentication

```bash
# Get a valid JWT token from Auth0
TOKEN="your-jwt-token"

# Test worker
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8787/images/test-image-id/original
```

## Security

### Implemented

- ✅ JWT signature verification (RS256)
- ✅ Token expiry validation
- ✅ Audience claim validation
- ✅ User ownership checks via API
- ✅ AWS Signature V4 for B2 access
- ✅ User-specific cache keys
- ✅ Private B2 bucket (no public access)

### Recommendations

- Set specific CORS origin (change from `*`)
- Implement rate limiting per user
- Add request logging and monitoring
- Rotate WORKER_SECRET regularly

## API Requirements

The worker requires your backend API to implement:

### GET /v1/images/:id/owner

**Headers:**
- `X-User-ID` - User ID from JWT
- `X-Image-Kind` - `original` or `staged`
- `X-Internal-Auth` - Worker secret for verification

**Response:**
```json
{
  "image_id": "abc-123",
  "owner_id": "auth0|user123",
  "has_access": true,
  "s3_key": "images/abc-123/original.jpg"
}
```

## Monitoring

### Cloudflare Dashboard

1. Go to **Workers & Pages** → `realstaging-cdn-worker`
2. Click **Metrics** tab
3. Monitor:
   - Requests per second
   - Success rate
   - CPU time
   - Errors

### Cache Performance

Check `X-Cache-Status` header:
- `HIT` - Served from cache (fast!)
- `MISS` - Fetched from B2 (first request)

Target cache hit rate: >95%

## Cost

### Free Tier
- 100,000 requests/day
- 10ms CPU time per request
- Unlimited bandwidth

### Paid (if needed)
- $5/month base
- $0.50 per million requests beyond free tier

**Expected cost for typical usage: $0** (within free tier)

## Troubleshooting

### Images Not Loading

1. Check worker logs: `npm run tail`
2. Verify secrets are set: `wrangler secret list`
3. Test JWT token validity
4. Check API ownership endpoint

### 401 Unauthorized

- Token expired
- Invalid token signature
- Wrong audience claim
- JWKS fetch failed

### 403 Forbidden

- User doesn't own the image
- Ownership endpoint returned false

### 404 Not Found

- Image doesn't exist
- S3 key not found in B2
- Invalid path format

### 500 Internal Server Error

- B2 credentials invalid
- AWS signature error
- API endpoint down

## Updating

To update the worker:

```bash
# Make changes to src/index.ts
# Deploy
npm run deploy
```

Changes take effect immediately (no downtime).

## Rollback

To rollback to previous version:

```bash
wrangler rollback
```

## License

MIT
