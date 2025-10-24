# Image Deduplication Design

## Problem Statement

When users test the same original image with multiple styles, the current 1:1 storage model creates duplicate copies of the original image in S3. This wastes storage and increases costs.

## Solution: Content-Addressable Storage

Implement content-based deduplication using cryptographic hashing to store each unique original image only once.

## Architecture

### Storage Layout

```
# Current (wasteful)
uploads/user-123/kitchen-abc123.jpg          # Original upload 1
uploads/user-123/kitchen-def456.jpg          # Same image, different upload
staged/abc12345/abc12345-678-staged.jpg      # Styled version 1
staged/def45678/def45678-901-staged.jpg      # Styled version 2

# Proposed (optimized)
originals/a1/a1b2c3d4e5...hash              # Unique original (stored once)
staged/abc12345/abc12345-678-staged.jpg     # Styled version 1
staged/def45678/def45678-901-staged.jpg     # Styled version 2
```

### Data Model

#### New Table: `original_images`

```sql
CREATE TABLE original_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hex
  s3_key TEXT NOT NULL,                      -- originals/{hash[:2]}/{hash}
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(50) NOT NULL,
  width INTEGER,
  height INTEGER,
  reference_count INTEGER NOT NULL DEFAULT 0, -- For cleanup
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_original_images_hash ON original_images(content_hash);
CREATE INDEX idx_original_images_ref_count ON original_images(reference_count);
```

#### Modified Table: `images`

```sql
-- Add new column
ALTER TABLE images ADD COLUMN original_image_id UUID REFERENCES original_images(id) ON DELETE RESTRICT;

-- Keep original_url temporarily for backward compatibility during migration
-- Will be removed after full migration
ALTER TABLE images ALTER COLUMN original_url DROP NOT NULL;

-- Add index for foreign key
CREATE INDEX idx_images_original_image_id ON images(original_image_id);
```

## Implementation Phases

### Phase 1: Infrastructure Setup

1. **Create migration** for `original_images` table
2. **Add hash computation** utility (SHA-256)
3. **Create repository layer** for `original_images` CRUD operations
4. **Update S3 service** with deduplication-aware upload method

### Phase 2: Upload Flow Changes

**Current Upload Flow:**

```
Client → Presigned URL → S3 uploads/{user}/{file}-{uuid}.ext
     ↓
   CreateImage(original_url)
```

**New Upload Flow:**

```
Client uploads → Temporary location
     ↓
API downloads & hashes
     ↓
Check if hash exists in original_images
     ↓
If exists: Reuse existing
If new: Upload to originals/{hash[:2]}/{hash}
     ↓
CreateImage(original_image_id, ref to hash)
Increment reference_count
```

### Phase 3: Worker Changes

Update worker to fetch original from new location:

```go
// Before
originalImage, err := s.DownloadFromS3(ctx, extractS3KeyFromURL(req.OriginalURL))

// After
original, err := s.originalRepo.GetByID(ctx, req.OriginalImageID)
originalImage, err := s.DownloadFromS3(ctx, original.S3Key)
```

### Phase 4: Migration Strategy

1. **Dual-write period**: Support both old and new columns
2. **Background job**: Hash and deduplicate existing images
3. **Verification**: Ensure all images reference `original_images`
4. **Cleanup**: Remove deprecated `original_url` column

## API Changes

### Upload Endpoint

**Before:**

```go
POST /api/v1/projects/{id}/images/upload
Response: {
  "upload_url": "https://...",
  "file_key": "uploads/{user}/{file}-{uuid}.ext",
  "expires_in": 900
}
```

**After:**

```go
POST /api/v1/projects/{id}/images/upload
Request: {
  "filename": "kitchen.jpg",
  "content_type": "image/jpeg",
  "file_size": 2048000,
  "content_hash": "a1b2c3d4..." // Client computes SHA-256
}
Response: {
  "upload_url": "https://...",
  "file_key": "temp/{uuid}",     // Temp location
  "expires_in": 900,
  "deduplication_check": false   // Or true if hash exists
}
```

Alternative: Keep presigned upload, then have a "finalize" endpoint that hashes and deduplicates.

### Create Image Endpoint

**Before:**

```go
POST /api/v1/projects/{id}/images
{
  "original_url": "s3://bucket/uploads/...",
  "room_type": "living_room",
  "style": "modern"
}
```

**After:**

```go
POST /api/v1/projects/{id}/images
{
  "file_key": "temp/{uuid}",      // From upload response
  "room_type": "living_room",
  "style": "modern"
}
```

Backend logic:

1. Download from temp location
2. Compute hash
3. Check `original_images` for existing hash
4. Reuse or create new `original_images` record
5. Create `images` record with `original_image_id`
6. Delete temp file

## Code Organization

```
apps/api/internal/
  originalimage/
    repository.go          # CRUD for original_images
    default_repository.go
    service.go            # Hash, deduplicate, upload logic
    default_service.go
    repository_mock.go
    service_mock.go

  image/
    repository.go         # Update CreateImage signature
    default_repository.go # Add original_image_id handling
    service.go           # Update upload flow

  hash/
    hash.go              # SHA-256 utilities
    hash_test.go

apps/worker/internal/
  staging/
    default_service.go   # Update to use original_images
```

## Hash Utilities

```go
package hash

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
)

// ComputeSHA256 computes SHA-256 hash of reader content
func ComputeSHA256(r io.Reader) (string, error) {
    h := sha256.New()
    if _, err := io.Copy(h, r); err != nil {
        return "", err
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeSHA256FromBytes computes SHA-256 from byte slice
func ComputeSHA256FromBytes(data []byte) string {
    h := sha256.Sum256(data)
    return hex.EncodeToString(h[:])
}
```

## Reference Counting & Cleanup

### Increment References

```sql
-- When creating new image referencing existing original
UPDATE original_images
SET reference_count = reference_count + 1,
    updated_at = NOW()
WHERE id = $1;
```

### Decrement References

```sql
-- When soft-deleting or hard-deleting an image
UPDATE original_images
SET reference_count = reference_count - 1,
    updated_at = NOW()
WHERE id = $1;
```

### Cleanup Job

```go
// Delete unreferenced originals (reference_count = 0)
// Run periodically (e.g., daily)
func CleanupOrphanedOriginals(ctx context.Context, olderThanDays int) error {
    // Find originals with reference_count = 0 for X days
    // Delete from S3
    // Delete from original_images table
}
```

## Benefits

1. **Storage savings**: ~50-90% for users who test multiple styles
2. **Faster uploads**: If hash exists, skip S3 upload entirely
3. **Cost reduction**: Less S3 storage and transfer costs
4. **Integrity**: Content-addressable storage ensures data integrity

## Risks & Mitigations

| Risk                 | Mitigation                                                       |
| -------------------- | ---------------------------------------------------------------- |
| Hash collisions      | SHA-256 has negligible collision probability (~2^-128)           |
| Orphaned files       | Reference counting + cleanup job                                 |
| Migration downtime   | Dual-write period, background migration                          |
| Increased complexity | Comprehensive tests, documentation                               |
| Performance impact   | Hash computation is fast (~50MB/s); consider client-side hashing |

## Testing Strategy

1. **Unit tests**: Hash computation, deduplication logic
2. **Integration tests**: Full upload → deduplicate → process flow
3. **Load tests**: Performance impact of hashing
4. **Migration tests**: Verify backward compatibility during transition

## Rollout Plan

1. **Week 1**: Implement infrastructure (tables, repos, hash utilities)
2. **Week 2**: Update upload flow with deduplication
3. **Week 3**: Update worker, add tests
4. **Week 4**: Deploy with feature flag, monitor metrics
5. **Week 5**: Run migration for existing images
6. **Week 6**: Enable for all users, remove old code

## Metrics to Track

- Deduplication rate (% of uploads that reuse existing originals)
- Storage savings (GB saved)
- Upload latency (impact of hashing)
- Error rates during migration
- Reference count accuracy

## Alternative Approaches Considered

### 1. Client-Side Hashing

**Pros**: Offload computation to client
**Cons**: Trust issues, client library complexity

### 2. Perceptual Hashing (pHash)

**Pros**: Detect similar (not just identical) images
**Cons**: Higher false positive rate, more complex

### 3. Post-Upload Deduplication

**Pros**: Simpler upload flow
**Cons**: Temporary duplicates, more storage churn

**Decision**: Content-addressable storage with server-side SHA-256 is the best balance of simplicity, reliability, and performance.

## Future Enhancements

1. **Lazy migration**: Deduplicate on-demand rather than background job
2. **Multi-region support**: Replicate originals across regions
3. **Tiered storage**: Move cold originals to Glacier
4. **Client-side hashing**: Optional optimization for power users
