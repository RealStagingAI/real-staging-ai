# Image Deduplication Integration Status

## Completed Components

### 1. Database Schema ✅
- **Migration**: `0014_create_original_images_table.up.sql` and `.down.sql`
- **Tables**: `original_images` table with content-addressable storage
- **Columns**: Added `original_image_id` FK to `images` table
- **Indexes**: Hash lookup, reference count, and FK indexes
- **Triggers**: Auto-update `updated_at` timestamp

### 2. SQL Queries ✅
- **Location**: `apps/api/internal/storage/queries/original_images.sql`
- **Queries**:
  - `CreateOriginalImage`: Insert new original with reference_count=1
  - `GetOriginalImageByID`: Fetch by UUID
  - `GetOriginalImageByHash`: Lookup for deduplication
  - `IncrementReferenceCount`: Add reference
  - `DecrementReferenceCount`: Remove reference (with GREATEST to prevent negative)
  - `ListOrphanedOriginalImages`: Find zero-ref originals for cleanup
  - `DeleteOriginalImage`: Hard delete original
  - `GetOriginalImageStats`: Metrics for monitoring

- **Location**: `apps/api/internal/storage/queries/images.sql` (additions)
  - `GetOriginalImageIDForImage`: Retrieve FK before deletion
  - `GetOriginalImageIDsForProject`: Bulk FK retrieval

### 3. Hash Utilities ✅
- **Location**: `apps/api/internal/hash/hash.go`
- **Functions**:
  - `ComputeSHA256(io.Reader)`: Stream-based hashing
  - `ComputeSHA256FromBytes([]byte)`: Direct byte hashing
  - `ValidateHash(string)`: Format validation
- **Tests**: Comprehensive unit tests in `hash_test.go`

### 4. Repository Layer ✅
- **Location**: `apps/api/internal/originalimage/`
- **Files**:
  - `repository.go`: Repository interface
  - `default_repository.go`: PostgreSQL implementation
- **Methods**: All CRUD operations for original images

### 5. Service Layer ✅
- **Location**: `apps/api/internal/originalimage/service.go`
- **Key Method**: `DecrementReferenceAndCleanup(ctx, originalImageID) (bool, error)`
  - Decrements reference count
  - **Deletes from S3 if count reaches 0**
  - **Deletes from database if count reaches 0**
  - Returns `true` if original was deleted
- **Additional Methods**:
  - `CleanupOrphanedOriginals`: Batch cleanup job
  - `GetStats`: Monitoring metrics

## Integration Points Needed

### 1. Update Image Service DeleteImage ⚠️

**Current Implementation** (simplified):
```go
func (s *DefaultService) DeleteImage(ctx context.Context, imageID string) error {
    return s.imageRepo.DeleteImage(ctx, imageID)  // Soft delete only
}
```

**Required Implementation**:
```go
func (s *DefaultService) DeleteImage(ctx context.Context, imageID string) error {
    // 1. Get original_image_id BEFORE soft-deleting
    originalImageID, err := s.imageRepo.GetOriginalImageID(ctx, imageID)
    if err != nil {
        return fmt.Errorf("failed to get original image ID: %w", err)
    }

    // 2. Soft delete the image
    if err := s.imageRepo.DeleteImage(ctx, imageID); err != nil {
        return fmt.Errorf("failed to delete image: %w", err)
    }

    // 3. Decrement reference and cleanup if this was the last reference
    if originalImageID != "" {
        deleted, err := s.originalImageService.DecrementReferenceAndCleanup(ctx, originalImageID)
        if err != nil {
            // Log error but don't fail the deletion
            // The orphaned original can be cleaned up later by batch job
            log.Warn(ctx, "failed to decrement original reference", "error", err)
        } else if deleted {
            log.Info(ctx, "deleted unreferenced original image", "original_id", originalImageID)
        }
    }

    return nil
}
```

**Changes Needed**:
```go
// In DefaultService struct
type DefaultService struct {
    imageRepo            Repository
    jobRepo              job.Repository
    enqueuer             queue.Enqueuer
    originalImageService originalimage.Service  // ADD THIS
}

// In NewDefaultService constructor
func NewDefaultService(
    cfg *config.Config,
    imageRepo Repository,
    jobRepo job.Repository,
    originalImageService originalimage.Service,  // ADD THIS
) *DefaultService {
    // ... existing code ...
    return &DefaultService{
        imageRepo:            imageRepo,
        jobRepo:              jobRepo,
        enqueuer:             enq,
        originalImageService: originalImageService,  // ADD THIS
    }
}
```

### 2. Update Image Service Constructor Calls ⚠️

**Files to Update**:
1. `apps/api/cmd/api/main.go` - Wire up services
2. `apps/api/tests/integration/e2e_sse_test.go` - Test setup
3. All test files in `apps/api/internal/image/*_test.go` - Mock setup

**Example**:
```go
// In main.go
originalImageRepo := originalimage.NewDefaultRepository(db)
originalImageService := originalimage.NewDefaultService(originalImageRepo, s3Service)
imageService := image.NewDefaultService(cfg, imageRepo, jobRepo, originalImageService)
```

### 3. Bulk Deletion for Projects ⚠️

When deleting a project, all images are deleted. Need to handle reference counting:

**Current**: `DeleteImagesByProjectID` does hard delete
**Required**: Loop through images, decrement each original's reference

```go
func (s *DefaultService) DeleteImagesByProject(ctx context.Context, projectID string) error {
    // 1. Get all original_image_ids for the project
    originalIDs, err := s.imageRepo.GetOriginalImageIDsForProject(ctx, projectID)
    if err != nil {
        return fmt.Errorf("failed to get original IDs: %w", err)
    }

    // 2. Delete all images in project
    if err := s.imageRepo.DeleteImagesByProjectID(ctx, projectID); err != nil {
        return fmt.Errorf("failed to delete images: %w", err)
    }

    // 3. Decrement references for all originals
    for _, originalID := range originalIDs {
        _, _ = s.originalImageService.DecrementReferenceAndCleanup(ctx, originalID)
    }

    return nil
}
```

### 4. Background Cleanup Job (Optional but Recommended) ⚠️

Create a periodic job to clean up orphaned originals:

```go
// apps/api/cmd/cleanup/main.go or similar
func cleanupOrphans(ctx context.Context, svc originalimage.Service) {
    olderThan := 24 * time.Hour  // Grace period
    limit := 100

    deleted, err := svc.CleanupOrphanedOriginals(ctx, olderThan, limit)
    if err != nil {
        log.Error(ctx, "cleanup failed", "error", err)
        return
    }

    log.Info(ctx, "cleanup completed", "deleted", deleted)
}
```

## Testing Requirements

### Unit Tests Needed ⚠️
1. `originalimage/default_service_test.go`
   - Test `DecrementReferenceAndCleanup` with count > 1 (no deletion)
   - Test `DecrementReferenceAndCleanup` with count = 1 (deletion)
   - Test `CleanupOrphanedOriginals`

2. `image/default_service_test.go` (update existing)
   - Test `DeleteImage` decrements original reference
   - Test `DeleteImage` when original is deleted
   - Test `DeleteImage` handles missing original_image_id gracefully

### Integration Tests Needed ⚠️
1. Full flow: Upload → Deduplicate → Process → Delete all → Verify original deleted
2. Multiple styles: Upload → Create 3 images with same original → Delete 2 → Verify original remains → Delete last → Verify original deleted

## Pre-Commit Checklist (from AGENTS.md)

Before committing, run in order:

```bash
# 1. Generate code (sqlc, mocks)
make generate

# 2. Run linter (must pass with 0 issues)
make lint

# 3. Run unit tests (all must pass)
make test

# 4. Run integration tests (all must pass)
make test-integration
```

## Deployment Steps

1. **Database Migration**: Run `0014_create_original_images_table.up.sql`
2. **Deploy Code**: Services are backward compatible (original_url still supported)
3. **Backfill Data**: Migrate existing images to use original_images table
4. **Enable Feature**: Start using deduplication for new uploads
5. **Monitor**: Track deduplication rate and storage savings

## Monitoring Metrics

Use `GetOriginalImageStats` to track:
- `TotalCount`: Number of unique originals
- `TotalSize`: Storage used by originals
- `OrphanedCount`: Images awaiting cleanup
- `OrphanedSize`: Reclaimable storage
- `AvgReferences`: Deduplication effectiveness

## Critical Safety Features

✅ **Reference Counting**: Prevents premature deletion
✅ **Soft Delete on Images**: Preserves billing/usage data
✅ **Hard Delete on Originals**: Reclaims storage when safe
✅ **Graceful Degradation**: Errors in cleanup don't fail deletions
✅ **Idempotency**: Multiple decrements handled with GREATEST()
✅ **Cleanup Job**: Recovers from any missed deletions

## Summary

**Completed**: Core infrastructure (database, queries, repositories, services, utilities)
**Remaining**: Wire up services in main.go and update DeleteImage implementation
**Effort**: ~1-2 hours to complete integration + testing

The design ensures **all originals are automatically deleted when their last referencing staged image is removed**, with safety mechanisms to prevent data loss.
