# Image Loading & Caching Optimization

## Current Issues

1. **Presigned URLs fetched on every page load** - Each image requires an API call to `/v1/images/:id/presign`
2. **No lazy loading** - All images load immediately when page renders
3. **No CDN** - Images served directly from Backblaze B2 S3 endpoint
4. **Limited browser caching** - Presigned URLs expire, preventing long-term caching

## Optimization Strategy

### 1. Lazy Loading (IMPLEMENT NOW ✅)

**Benefits:**
- Only load images when they enter viewport
- Reduces initial page load time
- Saves bandwidth for users who don't scroll

**Implementation:**
- Use Next.js `<Image>` component with `loading="lazy"`
- Intersection Observer for custom lazy loading
- Progressive loading with blur placeholders

### 2. Browser Caching Options

#### Option A: LocalStorage for Presigned URLs (SHORT-TERM ✅)
**Pros:**
- Free
- Fast subsequent loads
- No infrastructure changes

**Cons:**
- 5-10MB limit
- URLs expire (need refresh logic)
- Only helps individual users

**Implementation:**
```typescript
// Cache presigned URLs with expiry
const CACHE_KEY = 'realstaging_image_urls';
const CACHE_TTL = 3600000; // 1 hour

function getCachedUrl(imageId: string, kind: 'original' | 'staged') {
  const cache = JSON.parse(localStorage.getItem(CACHE_KEY) || '{}');
  const cached = cache[`${imageId}_${kind}`];
  if (cached && Date.now() < cached.expiry) {
    return cached.url;
  }
  return null;
}
```

#### Option B: Service Worker + Cache API (MEDIUM-TERM)
**Pros:**
- Can cache actual image bytes
- Works offline
- No size limit (respects disk space)

**Cons:**
- More complex setup
- PWA considerations

### 3. CDN Options

#### ⭐ **RECOMMENDED: Backblaze B2 + Cloudflare CDN**

**Pricing:**
- **B2 Storage**: $6/TB/month (you're already using this)
- **B2 → Cloudflare CDN**: **FREE** (Bandwidth Alliance partnership)
- **Cloudflare CDN**: **FREE** tier available

**Benefits:**
- Global edge caching
- HTTPS included
- Image optimization/transformations
- DDoS protection
- Near-instant cache hits after first load

**Setup Steps:**
1. Enable B2 public buckets
2. Add Cloudflare in front of B2
3. Update S3_ENDPOINT to Cloudflare Workers URL
4. Set Cache-Control headers on uploads

**Example:**
```bash
# Instead of: https://s3.us-west-004.backblazeb2.com
# Use: https://cdn.real-staging.ai (via Cloudflare)
```

#### Alternative: Render CDN
**Pricing:**
- Not available on free/starter tiers
- **$20+/month** for CDN addon

**Verdict:** ❌ Not worth it - use B2 + Cloudflare instead

### 4. Image Optimization Techniques

#### A. Next.js Image Component
```tsx
<Image
  src={imageUrl}
  alt="Property"
  width={800}
  height={600}
  loading="lazy"
  placeholder="blur"
  blurDataURL={lowResPreview}
  quality={85}
/>
```

#### B. Progressive JPEG/WebP
- Serve WebP for modern browsers (30% smaller)
- Progressive JPEG for older browsers
- Cloudflare can do this automatically

#### C. Responsive Images
```tsx
<Image
  srcSet="
    image-400w.jpg 400w,
    image-800w.jpg 800w,
    image-1200w.jpg 1200w
  "
  sizes="(max-width: 640px) 400px, (max-width: 1024px) 800px, 1200px"
/>
```

### 5. Caching Headers

Update S3 upload to include proper Cache-Control:

```go
// In upload handler
_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(key),
    Body:   file,
    CacheControl: aws.String("public, max-age=31536000, immutable"), // 1 year
    ContentType: aws.String(contentType),
})
```

## Implementation Priority

### Phase 1: Quick Wins (1-2 hours)
1. ✅ Add lazy loading with Intersection Observer
2. ✅ LocalStorage caching for presigned URLs
3. ✅ Add loading states and blur placeholders

### Phase 2: CDN Setup (2-4 hours)
1. ✅ Configure Cloudflare CDN for B2
2. ✅ Update S3_ENDPOINT to use CDN
3. ✅ Add Cache-Control headers on uploads
4. ✅ Test cache performance

### Phase 3: Advanced (Optional)
1. Service Worker for offline support
2. WebP/AVIF format conversion
3. Image resizing on-the-fly

## Expected Performance Improvements

**Before:**
- Initial page load: ~5-10 API calls for presigned URLs
- Every image: Full S3 download from Oregon
- No caching between sessions

**After Phase 1:**
- Initial visible images only: ~2-3 API calls
- LocalStorage: ~80% cache hit rate for returning users
- Lazy load: Only load images user actually sees

**After Phase 2:**
- CDN edge cache: ~99% cache hit rate globally
- Load time: ~50-200ms (vs 500-2000ms from S3)
- Bandwidth cost: Near zero (Cloudflare free tier)

## Monitoring

Track these metrics:
```typescript
// Add to your analytics
const imageMetrics = {
  presignedUrlCacheHits: 0,
  presignedUrlCacheMisses: 0,
  lazyLoadedImages: 0,
  cdnCacheHits: 0 // from Cloudflare headers
};
```

## References

- [Backblaze + Cloudflare Integration](https://www.backblaze.com/b2/docs/cloudflare.html)
- [Next.js Image Optimization](https://nextjs.org/docs/app/building-your-application/optimizing/images)
- [Web.dev Image Performance](https://web.dev/fast/#optimize-your-images)
