# Testing Guide - Cloudflare CDN Worker

## Overview

Comprehensive test coverage for both the Cloudflare Worker and backend API ownership handler.

## Backend API Tests

**File**: `apps/api/internal/http/image_ownership_handler_test.go`

### Test Coverage: 11 Tests ✅

#### Success Cases (3 tests)
1. **Owned Original Image** - Returns ownership info with S3 key for owned original images
2. **Owned Staged Image** - Returns ownership info with S3 key for owned staged images  
3. **Non-Owned Image** - Returns `has_access: false` without S3 key for non-owned images

#### Failure Cases (8 tests)
4. **Missing Worker Secret** - Returns 401 when WORKER_SECRET not configured
5. **Invalid Internal Auth** - Returns 401 when X-Internal-Auth header is wrong
6. **Missing Image ID** - Returns 400 when image ID parameter is missing
7. **Invalid Image ID Format** - Returns 400 when image ID is not a valid UUID
8. **Missing User ID Header** - Returns 400 when X-User-ID header is missing
9. **Invalid Image Kind** - Returns 400 when kind is not 'original' or 'staged'
10. **Image Not Found** - Returns 404 when image doesn't exist in database
11. **Database Error** - Returns 500 when project ownership query fails

### Helper Function Tests
- **extractS3KeyFromURL** - Tests S3 key extraction from various URL formats
  - Path-style URLs
  - URLs with subdirectories
  - Fallback to last slash
  - Different bucket names

### Running Backend Tests

```bash
# Run all ownership handler tests
cd apps/api
go test ./internal/http -run TestGetImageOwnerHandler -v

# Run with coverage
go test ./internal/http -run TestGetImageOwnerHandler -cover

# Run all HTTP tests
go test ./internal/http -v
```

### Expected Output

```
=== RUN   TestGetImageOwnerHandler
=== RUN   TestGetImageOwnerHandler/success:_returns_ownership_info_with_s3_key_for_owned_original_image
=== RUN   TestGetImageOwnerHandler/success:_returns_ownership_info_with_s3_key_for_owned_staged_image
=== RUN   TestGetImageOwnerHandler/success:_returns_has_access_false_for_non-owned_image
=== RUN   TestGetImageOwnerHandler/fail:_missing_worker_secret
=== RUN   TestGetImageOwnerHandler/fail:_invalid_internal_auth
=== RUN   TestGetImageOwnerHandler/fail:_missing_image_id
=== RUN   TestGetImageOwnerHandler/fail:_invalid_image_id_format
=== RUN   TestGetImageOwnerHandler/fail:_missing_user_id_header
=== RUN   TestGetImageOwnerHandler/fail:_invalid_image_kind
=== RUN   TestGetImageOwnerHandler/fail:_image_not_found
=== RUN   TestGetImageOwnerHandler/fail:_project_ownership_query_fails
--- PASS: TestGetImageOwnerHandler (0.00s)
    --- PASS: TestGetImageOwnerHandler/success:_returns_ownership_info_with_s3_key_for_owned_original_image (0.00s)
    --- PASS: TestGetImageOwnerHandler/success:_returns_ownership_info_with_s3_key_for_owned_staged_image (0.00s)
    --- PASS: TestGetImageOwnerHandler/success:_returns_has_access_false_for_non-owned_image (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_missing_worker_secret (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_invalid_internal_auth (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_missing_image_id (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_invalid_image_id_format (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_missing_user_id_header (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_invalid_image_kind (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_image_not_found (0.00s)
    --- PASS: TestGetImageOwnerHandler/fail:_project_ownership_query_fails (0.00s)
PASS
```

---

## Cloudflare Worker Tests

**File**: `cloudflare-cdn-worker/src/index.test.ts`

### Test Coverage: 20+ Tests

#### Request Validation
- Reject non-GET requests (405)
- Handle OPTIONS requests for CORS preflight (204)
- Reject missing Authorization header (401)
- Reject malformed Authorization header (401)
- Reject invalid path formats (400)
- Reject invalid kind parameter (400)

#### Path Parsing
- Parse valid image path with 'original' kind
- Parse valid image path with 'staged' kind

#### CORS Headers
- Include CORS headers in all responses
- Handle CORS preflight with proper headers
- Set Access-Control-Allow-Origin
- Set Access-Control-Max-Age

#### Error Handling
- Return 500 for unexpected errors
- Handle malformed JWT tokens gracefully

#### Response Headers
- Set Content-Type to application/json for errors

#### Security
- Require Bearer token format
- Validate JWT structure (3 parts)
- Enforce HTTPS URLs

#### Cache Behavior
- Use Cache API with Authorization header
- Set proper cache headers (Cache-Control, Vary)

#### Environment Configuration
- Validate all required environment variables present

### Setup Worker Tests

```bash
cd cloudflare-cdn-worker

# Install dependencies
npm install

# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

### Expected Output

```
 RUN  v1.0.0

 ✓ cloudflare-cdn-worker/src/index.test.ts (20+)
   ✓ Cloudflare Worker CDN
     ✓ Request Validation (6)
     ✓ Path Parsing (2)
     ✓ CORS Headers (2)
     ✓ Error Handling (1)
     ✓ Response Headers (1)
   ✓ Helper Functions (2)
   ✓ Environment Configuration (1)
   ✓ Cache Behavior (2)
   ✓ Security (3)

 Test Files  1 passed (1)
      Tests  20+ passed (20+)
   Duration  100ms
```

---

## Test Philosophy

### Backend Tests
- **Unit tests** - Mock all dependencies (database, image service)
- **Table-driven** - Use subtests for clarity
- **Comprehensive** - Cover success, failure, and edge cases
- **Fast** - No external dependencies, runs in milliseconds

### Worker Tests
- **Integration-style** - Test request/response flow
- **Validation-focused** - Ensure proper error handling
- **Security-first** - Verify auth and CORS behavior
- **Environment-aware** - Validate configuration

## Coverage Goals

- **Backend**: >90% coverage on ownership handler
- **Worker**: >80% coverage on request handling
- **Critical paths**: 100% coverage (auth, ownership checks)

## Continuous Integration

Tests run automatically on:
- Every commit (via pre-commit hook with `make test`)
- Pull requests (via GitHub Actions)
- Before deployment (via `make test-integration`)

## Adding New Tests

### Backend Test Template

```go
t.Run("test name", func(t *testing.T) {
    // Arrange
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    
    // Act
    err := handler(c)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
})
```

### Worker Test Template

```typescript
it('should test behavior', async () => {
    const request = new Request('https://cdn.example.com/path', {
        method: 'GET',
        headers: { 'Authorization': 'Bearer token' }
    });

    const response = await worker.fetch(request, mockEnv, mockCtx);

    expect(response.status).toBe(200);
});
```

## References

- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Vitest Documentation](https://vitest.dev/)
- [Cloudflare Workers Testing](https://developers.cloudflare.com/workers/testing/)
