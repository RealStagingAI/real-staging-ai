# Soft Delete Implementation for Billing Protection

## Problem

**Critical billing vulnerability discovered:** Users could game the system by:
1. Creating images (counts toward monthly limit)
2. Downloading the processed images
3. Deleting images (usage count decreased!)
4. Repeating the cycle indefinitely

The root cause: Usage was calculated with `COUNT(*) FROM images WHERE created_at BETWEEN ...`, and deleting images **hard deleted** them from the database, reducing the count.

## Solution

Implemented **soft deletes** to ensure usage tracking is immutable:

### Changes Made

1. **Migration `0012_add_soft_delete_to_images`**
   - Added `deleted_at TIMESTAMPTZ` column to `images` table
   - Added indexes for efficient querying of non-deleted images
   - Soft-deleted images remain in database for accurate usage tracking

2. **Updated Queries (`apps/api/internal/storage/queries/images.sql`)**
   - Added `SoftDeleteImage` query: `UPDATE images SET deleted_at = NOW() WHERE id = $1`
   - Updated all SELECT queries to filter: `WHERE deleted_at IS NULL`
   - **Important:** `CountImagesCreatedInPeriod` does NOT filter deleted_at - counts ALL images

3. **Updated Repository**
   - `DeleteImage()` now calls `SoftDeleteImage()` instead of hard DELETE
   - All other queries filter out soft-deleted images

4. **Updated Tests**
   - Fixed all repository tests to expect `deleted_at` column
   - Updated delete tests to expect UPDATE instead of DELETE

### Behavior

**Before:**
- User creates 10 images → usage = 10
- User deletes 5 images → usage = 5 ❌ (allows gaming)

**After:**
- User creates 10 images → usage = 10
- User deletes 5 images → usage = 10 ✅ (images soft-deleted, still count)

### Database Schema

```sql
ALTER TABLE images ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_images_deleted_at ON images (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_images_project_deleted ON images (project_id, deleted_at);
```

### API Behavior

- `GET /api/v1/images` - Only returns non-deleted images
- `DELETE /api/v1/images/:id` - Soft deletes (sets deleted_at)
- `/api/v1/billing/usage` - Counts ALL images including deleted ones

### Hard Deletes

Hard deletes are still used for:
- Cleanup operations (`DeleteStuckQueuedImages`)
- Cascade deletions (`DeleteImagesByProjectID`)
- These are administrative/cleanup operations, not user-facing

## Security Impact

This change prevents users from exploiting the billing system by repeatedly creating and deleting images to stay under their monthly limit while actually generating unlimited images.
