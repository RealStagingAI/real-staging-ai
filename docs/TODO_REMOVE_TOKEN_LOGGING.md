# TODO: Remove Token Logging (Security Issue)

**CRITICAL**: Remove debugging token logs added in commit `6ba074d`

## Files to Update

### apps/api/internal/auth/middleware.go
Remove all `fmt.Printf` statements that log JWT validation errors. These currently log:
- Token validation failures
- Audience/issuer mismatches  
- Public key fetch failures

The error handler also logs the full error which may contain token details.

## Why This Matters
Logging authentication errors in detail can expose:
- Token structure and claims
- Validation logic that attackers could exploit
- Timing information for brute force attacks

## What to Keep
- Generic "unauthorized" responses to clients
- Internal error tracking without token details (if needed for monitoring)

## When to Remove
After diagnosing the current CDN 401 issue.
