# Marketing Images Strategy

## Overview
This document outlines the strategy for hosting marketing images used in the landing page before/after sliders.

## Architecture

### Current Setup
- **Marketing Images**: Local images in `/public/images/marketing/`
- **User Images**: S3/Backblaze with pre-signed URLs
- **Static Assets**: Local `/public` folder

### Image Strategy: Local-Only

**Location**: `apps/web/public/images/marketing/`

**Pros**:
- ✅ Full control over image quality and optimization
- ✅ No external dependencies or rate limits
- ✅ Better SEO (same domain)
- ✅ Faster loading (same origin)
- ✅ Consistent branding
- ✅ Simple and reliable

**Cons**:
- ❌ Increases build size
- ❌ Need to manage image updates
- ❌ Storage costs at scale (minimal for marketing images)

## Implementation

### Configuration
Images are configured in:
```typescript
// apps/web/lib/marketingImages.ts
export const marketingImages: MarketingImage[] = [
  {
    title: "Bedroom Transformation",
    beforeSrc: "/images/marketing/bedroom-before.jpg",
    afterSrc: "/images/marketing/bedroom-after.jpg",
    beforeAlt: "Empty bedroom with neutral walls",
    afterAlt: "Professionally staged bedroom with cozy furnishings"
  }
  // Add more images as needed
];
```

### Adding New Images

1. **Add images to directory**:
   ```bash
   # Copy your before/after images
   cp before.jpg apps/web/public/images/marketing/new-room-before.jpg
   cp after.jpg apps/web/public/images/marketing/new-room-after.jpg
   ```

2. **Update configuration**:
   ```typescript
   // apps/web/lib/marketingImages.ts
   export const marketingImages: MarketingImage[] = [
     // ... existing images
     {
       title: "New Room Type",
       beforeSrc: "/images/marketing/new-room-before.jpg",
       afterSrc: "/images/marketing/new-room-after.jpg",
       beforeAlt: "Empty new room",
       afterAlt: "Staged new room"
     }
   ];
   ```

3. **Update landing page** (optional):
   The landing page shows filtered images. Update the filter in `app/page.tsx`:
   ```typescript
   {getMarketingImages().filter(image => 
     image.title === "Bedroom Transformation" || image.title === "New Room Type"
   ).map((image, index) => (
     // ... slider component
   ))}
   ```

### Image Cropping & Normalization

For perfect before/after comparison, images should have identical dimensions and framing. Use the provided scripts to normalize your images:

```bash
# Crop and normalize ALL marketing images at once
./scripts/crop-all-marketing-images.sh

# Crop and normalize specific images
./scripts/crop-marketing-images.sh

# Advanced alignment tools
./scripts/align-marketing-images.sh manual
```

**What the script does:**
- Converts all images to JPEG format for consistency
- Crops both images to identical 1200x900 dimensions (4:3 aspect ratio)
- Uses center gravity to maintain the most important content
- Optimizes for web (quality 85, metadata stripped)
- Creates backups of original images

**Manual cropping (if needed):**
```bash
# Using ImageMagick directly
magick before.jpg -gravity center -resize 1200x900^ -extent 1200x900 -quality 85 before-cropped.jpg
magick after.jpg -gravity center -resize 1200x900^ -extent 1200x900 -quality 85 after-cropped.jpg
```

**Tips for best results:**
- Use the staged (after) photo as the reference for composition
- Ensure both photos are taken from the same angle and height
- Leave some extra space in original photos for cropping flexibility
- Test the slider at different screen sizes after cropping

### Advanced Image Alignment

For pixel-perfect alignment, use the advanced alignment tools:

```bash
# Manual alignment with visual guides
./scripts/align-marketing-images.sh manual

# Test automatic alignment positions
./scripts/align-marketing-images.sh auto

# Create perspective correction guides
./scripts/align-marketing-images.sh perspective
```

**What the alignment tools provide:**
- **Visual guides**: Grid overlays to identify misaligned elements
- **Multiple alignment options**: Tests different crop positions automatically
- **Perspective correction**: Guides for fixing angle differences
- **Side-by-side comparisons**: Easy visual comparison of alignment options

**Professional alignment tips:**
- Focus on aligning permanent architectural features (walls, windows, doors)
- Don't worry about furniture alignment (it should be different!)
- Use a tripod for consistent camera positioning
- Mark camera position on the floor for repeatable shots
- Consider using reference points in the room for alignment

**Visual enhancements in the slider:**
- Subtle overlay gradient to blend minor misalignments
- Smooth transition effects
- Responsive design for all screen sizes

### Next.js Configuration

The `next.config.js` is configured for:
- S3/Backblaze domains for user images
- Local images work automatically (no configuration needed)

## Deployment Strategy

### All Environments
- Uses local images consistently
- No external dependencies
- Same behavior across development, staging, and production

## Performance Considerations

### Image Loading
- Next.js Image component with lazy loading
- Responsive sizing for different viewports
- Automatic optimization

### Build Impact
- Marketing images add to build size
- Recommended to keep images optimized
- Monitor bundle size as image collection grows

## Maintenance

### Regular Tasks
1. **Review image performance**: Check Core Web Vitals
2. **Update images**: Seasonal or promotional changes
3. **Optimize new images**: Follow specifications
4. **Monitor build size**: Track impact on deployment

### Image Guidelines
- **Resolution**: 1200px width minimum
- **Quality**: Professional real estate photography
- **Content**: Clear before/after transformations
- **Branding**: Consistent with company style

## Troubleshooting

### Images Not Loading
1. Verify image files exist in `public/images/marketing/`
2. Check file names match configuration
3. Review browser console for 404 errors
4. Ensure images are optimized for web

### Performance Issues
1. Optimize image sizes
2. Check image compression
3. Monitor Core Web Vitals
4. Consider lazy loading for many images

### Build Size Concerns
1. Optimize image compression
2. Review image necessity
3. Consider image count limits
4. Monitor deployment metrics

## File Management

### Directory Structure
```
apps/web/public/images/marketing/
├── bedroom-before.jpg
├── bedroom-after.jpg
├── living-room-before.jpg
├── living-room-after.jpg
└── ... (add more as needed)
```

### Naming Convention
- Use descriptive names: `{room}-{state}.jpg`
- State: `before` or `after`
- Room: `bedroom`, `living-room`, `kitchen`, etc.

## Cost Analysis

### Local Images
- **Storage**: Included in build size
- **Bandwidth**: Included in hosting
- **Maintenance**: Developer time
- **Total Cost**: Minimal at current scale

## Best Practices

### Image Selection
1. **High quality**: Professional photography
2. **Clear transformation**: Obvious before/after difference
3. **Realistic**: Representative of actual staging results
4. **Consistent style**: Similar lighting and angles

### Performance
1. **Optimize first**: Always optimize before adding
2. **Monitor regularly**: Check performance metrics
3. **Test thoroughly**: Verify loading across devices
4. **Update strategically**: Don't change too frequently

This simplified approach provides reliable, professional marketing images with minimal complexity and maximum control over the user experience.
