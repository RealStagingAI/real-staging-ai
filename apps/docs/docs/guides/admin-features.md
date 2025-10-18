# Admin Features Guide

Guide for administrators to manage Real Staging AI system settings, AI models, and maintenance tasks.

## Overview

Admin features allow authorized users to:

- **Switch AI models** - Select which model processes images
- **Manage system settings** - Configure application behavior
- **Reconcile storage** - Synchronize database with S3

All admin endpoints require authentication and are accessed under `/api/v1/admin`.

## Authentication & Authorization

### Current Implementation

Admin endpoints require **Auth0 JWT authentication** via the standard `Authorization: Bearer <token>` header.

**Access Control:**
- ✅ Authentication required (JWT token)
- ⏳ Role-based authorization (not yet implemented)
- ⚠️ **Security Note:** Currently, any authenticated user can access admin endpoints

### Planned: Role-Based Access Control

Future releases will implement role-based authorization:

```go
// Planned middleware
func RequireAdminRole(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        user := getUserFromContext(c)
        if user.Role != "admin" {
            return echo.NewHTTPError(http.StatusForbidden, "Admin access required")
        }
        return next(c)
    }
}
```

**User Roles** (in database):
- `user` - Regular user (default)
- `admin` - Administrator with access to admin endpoints

**Setting User Role:**
```sql
-- Promote user to admin
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';

-- Demote admin to user
UPDATE users SET role = 'user' WHERE email = 'user@example.com';
```

### Security Recommendations

Until role-based authorization is implemented:

1. **Use feature flags** - Keep admin endpoints disabled in production:
   ```bash
   RECONCILE_ENABLED=0  # Disable reconciliation endpoint
   ```

2. **Firewall rules** - Restrict `/api/v1/admin/*` at infrastructure level

3. **API Gateway** - Use AWS API Gateway, Cloudflare, or nginx to block admin routes

4. **Monitor access** - Alert on admin endpoint usage

## AI Model Management

Real Staging AI supports multiple AI models for image processing. Admins can switch between models to balance quality, speed, and cost.

### Available Models

| Model ID | Name | Description | Speed | Quality | Cost |
|----------|------|-------------|-------|---------|------|
| `qwen/qwen-image-edit` | Qwen Image Edit | Fast editing optimized for virtual staging | ⚡⚡⚡ | ⭐⭐⭐ | $ |
| `black-forest-labs/flux-kontext-max` | Flux Kontext Max | High-quality with advanced context understanding | ⚡⚡ | ⭐⭐⭐⭐⭐ | $$$ |

**Model Characteristics:**

**Qwen Image Edit:**
- Requires input image (no text-to-image)
- Optimized for speed and cost
- Best for: High-volume staging jobs
- Processing time: ~5-10 seconds

**Flux Kontext Max:**
- Supports text-to-image and image-to-image
- Superior quality and context understanding
- Best for: Premium results, complex scenes
- Processing time: ~20-30 seconds

### List All Models

**GET /api/v1/admin/models**

Returns all available AI models with their current status.

**Request:**
```bash
curl -X GET https://api.realstaging.ai/api/v1/admin/models \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "models": [
    {
      "id": "qwen/qwen-image-edit",
      "name": "Qwen Image Edit",
      "description": "Fast image editing model optimized for virtual staging. Requires input image.",
      "version": "v1",
      "is_active": true
    },
    {
      "id": "black-forest-labs/flux-kontext-max",
      "name": "Flux Kontext Max",
      "description": "High-quality image generation and editing with advanced context understanding. Supports both text-to-image and image-to-image.",
      "version": "v1",
      "is_active": false
    }
  ]
}
```

### Get Active Model

**GET /api/v1/admin/models/active**

Returns the currently active AI model used for all new image processing jobs.

**Request:**
```bash
curl -X GET https://api.realstaging.ai/api/v1/admin/models/active \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "id": "qwen/qwen-image-edit",
  "name": "Qwen Image Edit",
  "description": "Fast image editing model optimized for virtual staging. Requires input image.",
  "version": "v1",
  "is_active": true
}
```

**Error Cases:**
- `404 Not Found` - Active model not found in registry (data corruption)
- `500 Internal Server Error` - Database error

### Switch Active Model

**PUT /api/v1/admin/models/active**

Updates the active AI model. All new jobs will use this model.

**Request:**
```bash
curl -X PUT https://api.realstaging.ai/api/v1/admin/models/active \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "value": "black-forest-labs/flux-kontext-max"
  }'
```

**Request Body:**
```json
{
  "value": "qwen/qwen-image-edit"  // Model ID to activate
}
```

**Response:**
```json
{
  "message": "Active model updated successfully",
  "model_id": "qwen/qwen-image-edit"
}
```

**Error Cases:**
- `400 Bad Request` - Invalid model ID or missing `value` field
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database error

**Important Notes:**
- ✅ Change takes effect immediately for new jobs
- ⚠️ Existing jobs continue with their original model
- 📝 Change is logged with admin user ID and timestamp

**Example Workflow:**
```bash
# 1. Check current model
CURRENT=$(curl -s https://api.realstaging.ai/api/v1/admin/models/active \
  -H "Authorization: Bearer $TOKEN" | jq -r '.id')
echo "Current model: $CURRENT"

# 2. List available models
curl -s https://api.realstaging.ai/api/v1/admin/models \
  -H "Authorization: Bearer $TOKEN" | jq '.models[] | "\(.id) - \(.name)"'

# 3. Switch to high-quality model
curl -X PUT https://api.realstaging.ai/api/v1/admin/models/active \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value": "black-forest-labs/flux-kontext-max"}'

# 4. Verify change
curl -s https://api.realstaging.ai/api/v1/admin/models/active \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Settings Management

System settings control application behavior. Settings are stored in the database and can be updated at runtime without redeployment.

### Settings Structure

```typescript
interface Setting {
  key: string;           // Unique identifier (e.g., "active_model")
  value: string;         // Setting value (always string, parse as needed)
  description?: string;  // Human-readable description
  updated_at: string;    // ISO 8601 timestamp
  updated_by?: string;   // User UUID who made the change
}
```

### Common Settings

| Key | Description | Example Value | Type |
|-----|-------------|---------------|------|
| `active_model` | Currently active AI model | `qwen/qwen-image-edit` | string |
| `max_image_size_mb` | Maximum upload size in MB | `10` | number |
| `default_timeout_seconds` | Job timeout | `300` | number |
| `maintenance_mode` | Enable maintenance mode | `true` or `false` | boolean |

### List All Settings

**GET /api/v1/admin/settings**

Returns all system settings.

**Request:**
```bash
curl -X GET https://api.realstaging.ai/api/v1/admin/settings \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "settings": [
    {
      "key": "active_model",
      "value": "qwen/qwen-image-edit",
      "description": "Currently active AI model for image processing",
      "updated_at": "2025-10-17T18:30:00Z",
      "updated_by": "550e8400-e29b-41d4-a716-446655440000"
    },
    {
      "key": "max_image_size_mb",
      "value": "10",
      "description": "Maximum image upload size in megabytes",
      "updated_at": "2025-10-15T10:00:00Z",
      "updated_by": null
    }
  ]
}
```

### Get Specific Setting

**GET /api/v1/admin/settings/:key**

Returns a single setting by key.

**Request:**
```bash
curl -X GET https://api.realstaging.ai/api/v1/admin/settings/active_model \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "key": "active_model",
  "value": "qwen/qwen-image-edit",
  "description": "Currently active AI model for image processing",
  "updated_at": "2025-10-17T18:30:00Z",
  "updated_by": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Cases:**
- `404 Not Found` - Setting key doesn't exist

### Update Setting

**PUT /api/v1/admin/settings/:key**

Updates a setting value.

**Request:**
```bash
curl -X PUT https://api.realstaging.ai/api/v1/admin/settings/max_image_size_mb \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "value": "20"
  }'
```

**Request Body:**
```json
{
  "value": "new_value"  // New setting value (always string)
}
```

**Response:**
```json
{
  "message": "Setting updated successfully",
  "key": "max_image_size_mb",
  "value": "20"
}
```

**Error Cases:**
- `400 Bad Request` - Missing `value` field
- `401 Unauthorized` - Missing or invalid authentication token

**Important Notes:**
- ✅ All values stored as strings (parse to number/boolean as needed)
- ✅ Changes logged with admin user ID
- ⚠️ No validation on values (application must handle invalid values)
- 📝 Create new settings by updating non-existent keys

## Storage Reconciliation

Storage reconciliation synchronizes the database with S3, detecting orphaned or missing images.

### Reconcile Images

**POST /api/v1/admin/reconcile/images**

Scans S3 bucket and database to find discrepancies.

**Feature Flag:** Requires `RECONCILE_ENABLED=1` environment variable.

**Request:**
```bash
curl -X POST "https://api.realstaging.ai/api/v1/admin/reconcile/images?dry_run=true&limit=100" \
  -H "Authorization: Bearer <token>"
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `dry_run` | boolean | `false` | Preview changes without modifying database |
| `limit` | integer | `100` | Max images to process |
| `concurrency` | integer | `10` | Parallel S3 requests |
| `batch_size` | integer | `50` | Database batch size |

**Response:**
```json
{
  "status": "received",
  "dry_run": true,
  "config": {
    "limit": 100,
    "concurrency": 10,
    "batch_size": 50
  }
}
```

**Error Cases:**
- `503 Service Unavailable` - Feature flag `RECONCILE_ENABLED` not set
- `400 Bad Request` - Invalid query parameters

**What It Does:**
1. Lists all images in S3 bucket
2. Compares with database records
3. Detects:
   - **Orphaned S3 files** - In S3 but not in database
   - **Missing S3 files** - In database but not in S3
   - **Corrupted records** - Database/S3 mismatch
4. Optionally fixes discrepancies (when `dry_run=false`)

**Example:**
```bash
# Dry run to preview issues
curl -X POST "https://api.realstaging.ai/api/v1/admin/reconcile/images?dry_run=true&limit=1000" \
  -H "Authorization: Bearer $TOKEN"

# Check logs for results
docker logs virtual-staging-ai-api-1 | grep reconcile

# Actually fix issues
curl -X POST "https://api.realstaging.ai/api/v1/admin/reconcile/images?limit=1000" \
  -H "Authorization: Bearer $TOKEN"
```

For detailed information, see [Storage Reconciliation Guide](../operations/reconciliation.md).

## Admin UI (Future)

A web-based admin panel is planned for easier management.

**Planned Features:**
- 📊 Dashboard with system metrics
- 🎛️ Model switcher with preview
- ⚙️ Settings editor with validation
- 🔍 Job monitoring and debugging
- 👥 User management
- 📈 Analytics and reporting

**Access:** `/admin/settings` (URL reserved)

## Best Practices

### Model Selection

**When to use Qwen Image Edit:**
- High-volume processing
- Cost-sensitive applications
- Fast turnaround required
- Good quality acceptable

**When to use Flux Kontext Max:**
- Premium tier customers
- Marketing materials
- Complex scene understanding needed
- Quality is priority over speed

### Settings Management

1. **Document changes** - Keep a changelog of setting updates
2. **Test first** - Use staging environment to verify setting changes
3. **Backup values** - Save old values before updating
4. **Validate** - Application should handle invalid values gracefully

### Reconciliation

1. **Always dry run first** - Preview changes before applying
2. **Off-peak hours** - Run during low traffic to avoid performance impact
3. **Small batches** - Start with `limit=100`, increase gradually
4. **Monitor progress** - Watch logs and metrics during reconciliation
5. **Schedule regularly** - Weekly reconciliation catches drift early

## Security Checklist

Before enabling admin features in production:

- [ ] Implement role-based access control (check `user.role == 'admin'`)
- [ ] Add admin middleware to all `/admin/*` routes
- [ ] Set up monitoring/alerting for admin endpoint access
- [ ] Restrict network access to admin endpoints (firewall/API gateway)
- [ ] Rotate admin user tokens regularly
- [ ] Audit admin actions (log all changes with user ID)
- [ ] Disable unused feature flags (`RECONCILE_ENABLED=0`)
- [ ] Review admin user list monthly
- [ ] Require MFA for admin accounts (Auth0 configuration)

## Troubleshooting

### "Admin access required" (Future)

**Cause:** User doesn't have `admin` role

**Fix:**
```sql
-- Check user role
SELECT id, email, role FROM users WHERE email = 'user@example.com';

-- Grant admin role
UPDATE users SET role = 'admin' WHERE email = 'user@example.com';
```

### "Service unavailable" on reconciliation

**Cause:** `RECONCILE_ENABLED` not set

**Fix:**
```bash
# Set environment variable
export RECONCILE_ENABLED=1

# Restart API
docker compose restart api
```

### Model switch not taking effect

**Issue:** New jobs still use old model

**Check:**
```bash
# Verify active model setting
curl -s https://api.realstaging.ai/api/v1/admin/settings/active_model \
  -H "Authorization: Bearer $TOKEN" | jq

# Check database directly
psql $DATABASE_URL -c "SELECT * FROM settings WHERE key = 'active_model';"
```

**Common causes:**
- Worker using cached setting (restart worker)
- Database transaction not committed
- Wrong environment (staging vs production)

### Can't update setting

**Error:** `404 Not Found`

**Cause:** Setting doesn't exist yet

**Solution:** Create it by updating (same endpoint creates if missing):
```bash
curl -X PUT https://api.realstaging.ai/api/v1/admin/settings/new_setting \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value": "initial_value"}'
```

## API Reference

### Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/models` | List all AI models |
| GET | `/admin/models/active` | Get active model |
| PUT | `/admin/models/active` | Switch active model |
| GET | `/admin/settings` | List all settings |
| GET | `/admin/settings/:key` | Get specific setting |
| PUT | `/admin/settings/:key` | Update setting |
| POST | `/admin/reconcile/images` | Reconcile S3 storage |

### Authentication

All endpoints require:
```
Authorization: Bearer <jwt_token>
```

Get token for testing:
```bash
make token
```

## Related Documentation

- [Authentication Guide](authentication.md) - Auth0 setup and JWT tokens
- [Storage Reconciliation](../operations/reconciliation.md) - Detailed reconciliation guide
- [Configuration Guide](configuration.md) - Environment variables and settings
- [Adding AI Models](adding-models.md) - How to integrate new models
