# Image Deduplication: Next Steps

## What Was Completed

### ✅ Core Infrastructure
1. **Database Schema**
   - Migration files: `0014_create_original_images_table.{up,down}.sql`
   - `original_images` table with content-addressable storage
   - `images.original_image_id` foreign key
   - Reference counting with `reference_count` column
   - Indexes for performance

2. **SQL Queries**
   - `apps/api/internal/storage/queries/original_images.sql` - Full CRUD for originals
   - `apps/api/internal/storage/queries/images.sql` - Added FK retrieval queries

3. **Hash Utilities**
   - `apps/api/internal/hash/hash.go` - SHA-256 computation
   - `apps/api/internal/hash/hash_test.go` - Comprehensive tests

4. **Repository Layer**
   - `apps/api/internal/originalimage/repository.go` - Interface
   - `apps/api/internal/originalimage/default_repository.go` - Implementation

5. **Service Layer**
   - `apps/api/internal/originalimage/service.go` - Business logic
   - **Key feature**: `DecrementReferenceAndCleanup` method that:
     - Decrements reference count
     - Deletes from S3 when count reaches 0
     - Deletes from DB when count reaches 0

6. **Integration in Image Service**
   - Updated `DefaultService` struct with `OriginalImageService` dependency
   - Updated `DeleteImage` method to:
     - Get `original_image_id` before deletion
     - Soft delete the image
     - **Call `DecrementReferenceAndCleanup` to remove original if last reference**
   - Added `OriginalImageService` interface to avoid circular deps

### ✅ Reference Counting Flow

**When an image is deleted:**
```
1. Retrieve original_image_id from images table
2. Soft delete the image (marks deleted_at)
3. Call DecrementReferenceAndCleanup(original_image_id)
   a. Get original image record
   b. Decrement reference_count
   c. If count reaches 0:
      - Delete file from S3
      - Delete record from database
   d. Return true if deleted, false otherwise
4. Log result
```

**Safety features:**
- Errors in cleanup don't fail the deletion
- Orphaned files can be cleaned up by batch job
- Reference count uses GREATEST() to prevent negative values
- Soft delete preserves billing/usage data

## What Needs To Be Done

### 1. Run Code Generation ⚠️

```bash
cd /Users/jasonadams/code/github/virtual-staging-ai
make generate
```

This will:
- Generate Go code from SQL queries (sqlc)
- Generate mocks for all interfaces (moq)
- Fix most of the current lint errors

### 2. Fix Test Files ⚠️

After running `make generate`, update test files to pass `nil` for the new parameter:

**Pattern to find and fix:**
```go
// OLD
service := image.NewDefaultService(cfg, imageRepo, jobRepo)

// NEW
service := image.NewDefaultService(cfg, imageRepo, jobRepo, nil)
```

**Files to update:**
- `apps/api/internal/image/default_service_test.go` (8 occurrences)
- `apps/api/tests/integration/e2e_sse_test.go` (1 occurrence)

For tests that specifically test deletion, create a mock:
```go
originalImageSvc := &image.OriginalImageServiceMock{
    DecrementReferenceAndCleanupFunc: func(ctx context.Context, id string) (bool, error) {
        return true, nil  // Or false, or error based on test case
    },
}
service := image.NewDefaultService(cfg, imageRepo, jobRepo, originalImageSvc)
```

### 3. Wire Up Services in main.go ⚠️

**File**: `apps/api/cmd/api/main.go`

```go
// After creating s3Service and before imageService:

// Create original image repository and service
originalImageRepo := originalimage.NewDefaultRepository(db)
originalImageService := originalimage.NewDefaultService(originalImageRepo, s3Service)

// Update image service creation
imageService := image.NewDefaultService(cfg, imageRepo, jobRepo, originalImageService)
```

### 4. Run Pre-Commit Workflow ⚠️

Per `AGENTS.md`, run these commands in order:

```bash
# 1. Generate code
make generate

# 2. Lint (must pass with 0 issues)
make lint

# 3. Unit tests (all must pass)
make test

# 4. Integration tests (all must pass)
make test-integration
```

### 5. Add Tests for New Functionality ⚠️

**Test file**: `apps/api/internal/originalimage/default_service_test.go`

Test cases needed:
```go
func TestDecrementReferenceAndCleanup_WithMultipleReferences(t *testing.T) {
    // Given: Original with reference_count = 3
    // When: DecrementReferenceAndCleanup called
    // Then: Count becomes 2, original NOT deleted, returns (false, nil)
}

func TestDecrementReferenceAndCleanup_WithLastReference(t *testing.T) {
    // Given: Original with reference_count = 1
    // When: DecrementReferenceAndCleanup called
    // Then: Original deleted from S3 and DB, returns (true, nil)
}

func TestCleanupOrphanedOriginals(t *testing.T) {
    // Given: Multiple originals with reference_count = 0
    // When: CleanupOrphanedOriginals called
    // Then: All orphaned originals deleted
}
```

**Test file**: `apps/api/internal/image/default_service_test.go`

Update `TestDefaultService_DeleteImage`:
```go
func TestDefaultService_DeleteImage_WithOriginalImageCleanup(t *testing.T) {
    // Test that DeleteImage calls DecrementReferenceAndCleanup
    // Verify both when original is deleted and when it's not
}
```

### 6. Integration Test (Optional but Recommended) ⚠️

**Test file**: `apps/api/tests/integration/image_deduplication_test.go`

```go
func TestImageDeduplication_EndToEnd(t *testing.T) {
    // 1. Create original_images record
    // 2. Create 3 images referencing same original (ref count = 3)
    // 3. Delete first image → verify ref count = 2, original exists
    // 4. Delete second image → verify ref count = 1, original exists
    // 5. Delete third image → verify ref count = 0, original deleted from S3 and DB
}
```

### 7. Database Migration ⚠️

**Before deployment:**

```bash
# Development
make migrate-up

# Test
make migrate-test-up

# Production (via CI/CD or manual)
# Run migration 0014_create_original_images_table.up.sql
```

### 8. Create Cleanup Job (Future Enhancement)

**File**: `apps/api/cmd/cleanup/main.go` or add to existing cron jobs

```go
func cleanupOrphanedOriginals(ctx context.Context) {
    originalImageService := // ... initialize service
    
    olderThan := 24 * time.Hour  // Grace period
    limit := 100
    
    deleted, err := originalImageService.CleanupOrphanedOriginals(ctx, olderThan, limit)
    if err != nil {
        log.Error(ctx, "cleanup failed", "error", err)
        return
    }
    
    log.Info(ctx, "cleanup completed", "deleted", deleted)
}
```

Schedule to run daily via cron or task scheduler.

## Verification Checklist

After completing above steps:

- [ ] `make generate` runs successfully
- [ ] `make lint` passes with 0 issues  
- [ ] `make test` passes all unit tests
- [ ] `make test-integration` passes all integration tests
- [ ] Migration applied to dev database
- [ ] Manual test: Delete image and verify original cleaned up
- [ ] Manual test: Delete one of multiple images and verify original remains

## Current Status

**Phase**: Core implementation complete, needs wiring and testing

**Estimate**: 2-3 hours to complete all remaining steps

**Blockers**: None - all code is written, just needs integration and testing

## Key Design Decisions

1. **Minimal interface in image service**: Avoids circular dependency by defining `OriginalImageService` interface locally
2. **Graceful degradation**: Cleanup errors don't fail deletions
3. **Safety first**: Reference counting prevents premature deletion
4. **Backward compatible**: Works with images that don't have `original_image_id` (legacy)
5. **Idempotent**: Multiple decrements handled safely with GREATEST()

## Summary

The core infrastructure for image deduplication with automatic cleanup is **complete**. When you delete the last staged image that references an original, the original image will be automatically deleted from both S3 storage and the database, saving storage costs while maintaining data integrity through reference counting.

Next step: Run `make generate` and fix the resulting test compilation errors.
