# CDN Revert Plan - Switch to Render Edge Caching

## Objective
Remove all Cloudflare CDN implementation and switch to Render Edge Caching for simpler, more reliable image delivery.

## Current CDN Implementation Analysis

### Files to Remove (Cloudflare-specific)
- `cloudflare-cdn-worker/` - Entire Cloudflare Worker directory
- `docs/CLOUDFLARE_CDN_SETUP.md` - Cloudflare setup documentation
- `docs/CLOUDFLARE_WORKER_AUTH_CDN.md` - Worker authentication documentation
- `apps/api/internal/cdn/` - CDN handler package (3 files)

### Files to Modify
1. **Backend API (`apps/api/`)**
   - `internal/config/config.go` - Remove CDN.URL config
   - `internal/http/server.go` - Remove CDN route and handler initialization
   - `internal/auth/middleware.go` - Clean up any CDN-specific auth logic
   - Any test files referencing CDN

2. **Frontend Web (`apps/web/`)**
   - Image components using CDN URLs - revert to presigned S3 URLs
   - Any localStorage caching related to CDN
   - Environment variables referencing CDN

3. **Configuration**
   - `.env` files - Remove CDN_URL variables
   - `config/` - Remove CDN configuration sections

4. **Documentation**
   - Update any docs that reference CDN

## Git History CDN Commits to Revert

### Phase 1: Recent Debug/Fix Commits (Last 7 commits)
- `dff5034` - debug: add INFO logs after type assertion to track execution
- `361c3ba` - debug: add granular step-by-step logging to pinpoint 401 source
- `6b45617` - fix: line length linting issue
- `2e9e665` - debug: add detailed execution tracing to CDN handler
- `2596045` - refactor: use dependency injection for logger throughout codebase (KEEP - unrelated to CDN)
- `797208e` - refactor: replace fmt.Printf with structured logging package (KEEP - unrelated to CDN)
- `3698e64` - fix: format CDN handler code

### Phase 2: CDN Feature Development Commits
- `c366f4b` - debug: add CDN handler request tracing
- `6ba074d` - debug: add detailed JWT validation error logging
- `eb2c705` - debug: log audience mismatch
- `519f93d` - debug: add logging to CDN token fetch
- `b670af4` - fix(web): add access token to CDN URLs for img tag authentication
- `db91317` - fix(cdn): extract JWT from session context instead of request headers
- `5a64a45` - feat(web): switch from S3 presigned URLs to CDN for image viewing
- `0810b4b` - refactor(api): move CDN handler to dedicated package
- `fd1898f` - feat(web): integrate Cloudflare CDN worker for image delivery
- `a353fd5` - refactor(api): use config store instead of os.Getenv for WORKER_SECRET
- `7d7f47a` - fix(test): update Worker tests to expect 401 for auth failures
- `db67352` - fix(cdn): update Wrangler installation to use npm instead of Homebrew
- `fc11f8e` - docs(cdn): add comprehensive testing guide for Worker and API
- `4c6e4dd` - test(cdn): add comprehensive tests for CDN worker and ownership handler
- `7519c3f` - feat(cdn): implement Cloudflare Worker for authenticated CDN
- `4ba9361` - docs(infra): add secure Cloudflare Worker CDN implementation for private images

## Execution Plan

### Step 1: Identify Clean Revert Point
Commit `58cccfa` (feat(web): implement lazy loading and localStorage caching for images page) is the last commit before CDN work began. This is our baseline.

### Step 2: Create Feature Branch
```bash
git checkout -b revert-cloudflare-cdn
```

### Step 3: Interactive Rebase to Remove CDN Commits
We'll use `git rebase -i` to cleanly remove all CDN-related commits from history.

```bash
# Start interactive rebase from commit before first CDN commit
git rebase -i 58cccfa

# In the editor, we'll mark CDN commits to drop:
# - Change 'pick' to 'drop' (or just delete the line) for these commits:
#   - All commits from 4ba9361 through dff5034 that are CDN-related
# - Keep commits that are unrelated (like 2596045, 797208e)

# After rebase completes, we'll need to:
# 1. Keep the logging refactor commits (2596045, 797208e)
# 2. Drop all CDN-related commits
# 3. Resolve any conflicts if they arise
```

**Commits to DROP in rebase:**
- dff5034 - debug: add INFO logs after type assertion to track execution
- 361c3ba - debug: add granular step-by-step logging to pinpoint 401 source
- 6b45617 - fix: line length linting issue
- 2e9e665 - debug: add detailed execution tracing to CDN handler
- 3698e64 - fix: format CDN handler code
- c366f4b - debug: add CDN handler request tracing
- 6ba074d - debug: add detailed JWT validation error logging
- eb2c705 - debug: log audience mismatch
- 519f93d - debug: add logging to CDN token fetch
- b670af4 - fix(web): add access token to CDN URLs
- db91317 - fix(cdn): extract JWT from session context
- 5a64a45 - feat(web): switch from S3 presigned URLs to CDN
- 0810b4b - refactor(api): move CDN handler to dedicated package
- fd1898f - feat(web): integrate Cloudflare CDN worker
- a353fd5 - refactor(api): use config store for WORKER_SECRET
- 7d7f47a - fix(test): update Worker tests
- db67352 - fix(cdn): update Wrangler installation
- fc11f8e - docs(cdn): add testing guide
- 4c6e4dd - test(cdn): add tests
- 7519c3f - feat(cdn): implement Cloudflare Worker
- 4ba9361 - docs(infra): add Cloudflare Worker CDN docs

**Commits to KEEP:**
- 2596045 - refactor: use dependency injection for logger
- 797208e - refactor: replace fmt.Printf with structured logging
- c54b503 - fix(test): fixed failing integration test
- 79a30aa - fix(api): skip S3 bucket creation in production
- e7d9f13 - fix(web): prevent infinite request loop
- 0d19be7 - fix(billing): implement soft delete
- b0f3dcd - fix(web): prevent infinite loop in image URL prefetching

### Step 4: Manual Cleanup
After reverts, manually clean up any remaining artifacts:

```bash
# Remove Cloudflare Worker directory
rm -rf cloudflare-cdn-worker/

# Remove CDN documentation
rm docs/CLOUDFLARE_CDN_SETUP.md docs/CLOUDFLARE_WORKER_AUTH_CDN.md

# Remove CDN handler package
rm -rf apps/api/internal/cdn/

# Remove HAR file
rm real-staging.ai.har
```

### Step 5: Restore S3 Presigned URL Usage in Web
The CDN usage is in `apps/web/app/images/page.tsx`:
- Function `getCdnUrl()` - needs to be reverted to use presigned URLs
- Update image URL mapping to use `/api/v1/images/:id/presign?kind=original|staged`
- Remove access token fetching logic (`/auth/access-token`)
- Keep localStorage caching but adapt it for presigned URLs

### Step 6: Update Configuration
1. Remove CDN config from `apps/api/internal/config/config.go`
2. Remove CDN routes from `apps/api/internal/http/server.go`
3. Remove CDN environment variables from config files
4. Update `.env.example` to remove CDN_URL

### Step 7: Database Migration (if needed)
Check if any CDN-specific tables or columns were added:
```bash
# Review migrations
ls -la infra/migrations/
```

If CDN-specific database changes exist, create a rollback migration.

### Step 8: Update Tests
1. Remove CDN-related integration tests
2. Restore any modified test expectations
3. Run full test suite:
```bash
make test
make test-integration
```

### Step 9: Update Documentation
1. Update README if it mentions CDN
2. Update any architecture docs
3. Add note about using Render Edge Caching

### Step 10: Implement Render Edge Caching
Add proper caching headers to presigned S3 URL responses:

```go
// apps/api/internal/http/image_handlers.go
// Add Cache-Control headers to presigned URL responses
c.Response().Header().Set("Cache-Control", "public, max-age=3600")
```

## Verification Checklist

- [ ] All Cloudflare Worker files removed
- [ ] CDN handler package removed
- [ ] CDN routes removed from API server
- [ ] Frontend uses presigned S3 URLs
- [ ] All tests pass (`make test && make test-integration`)
- [ ] Linting passes (`make lint`)
- [ ] No CDN references in config files
- [ ] Documentation updated
- [ ] Manual testing: Image viewing works in web app
- [ ] Manual testing: Image upload and download work
- [ ] Check Render Edge Caching headers are present

## Rollback Plan (if needed)
If this revert causes issues, we can revert the revert:

```bash
git revert <revert-commit-sha>
```

## Benefits of Render Edge Caching
1. **Simpler**: No separate Worker to maintain
2. **Fewer moving parts**: One less service to debug
3. **Built-in**: Render handles caching automatically
4. **Sufficient**: Edge caching provides most CDN benefits
5. **Private**: S3 bucket stays private, presigned URLs control access

## Next Steps After Revert
1. Test image delivery works correctly
2. Monitor response times
3. Verify presigned URLs are cached at edge
4. Consider adding cache headers if needed
5. Update monitoring/alerts to remove CDN checks
