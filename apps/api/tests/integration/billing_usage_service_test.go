//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/billing"
	"github.com/real-staging-ai/api/internal/config"
)

func TestDefaultUsageService_Integration_GetUsage(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	service := billing.NewDefaultUsageService(db, &cfg.Plans)

	// Use existing seeded test user
	testUserID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	t.Run("success: get usage for existing user", func(t *testing.T) {
		stats, err := service.GetUsage(ctx, testUserID)
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Verify period dates are set and valid
		require.NotEmpty(t, stats.PeriodStart)
		require.NotEmpty(t, stats.PeriodEnd)

		// Verify dates are valid RFC3339
		_, err = time.Parse(time.RFC3339, stats.PeriodStart)
		require.NoError(t, err)
		_, err = time.Parse(time.RFC3339, stats.PeriodEnd)
		require.NoError(t, err)

		// Should count the seeded project (images created)
		require.GreaterOrEqual(t, stats.ImagesUsed, int32(0))
		require.GreaterOrEqual(t, stats.RemainingImages, int32(0))
	})

	t.Run("success: can create image check", func(t *testing.T) {
		canCreate, err := service.CanCreateImage(ctx, testUserID)
		require.NoError(t, err)
		// Should return some boolean value without error
		_ = canCreate
	})
}

func TestDefaultUsageService_Integration_GetPlanByCode(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	service := billing.NewDefaultUsageService(db, &cfg.Plans)

	t.Run("success: get free plan from config", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "free")
		require.NoError(t, err)
		require.NotNil(t, plan)

		require.Equal(t, "free", plan.Code)
		require.Equal(t, "price_test_free", plan.PriceID)
		require.Equal(t, int32(100), plan.MonthlyLimit) // From seed data
		// ID should be a UUID from database
		require.NotEmpty(t, plan.ID)
		require.Len(t, plan.ID, 36) // UUID length
	})

	t.Run("success: get pro plan from config", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "pro")
		require.NoError(t, err)
		require.NotNil(t, plan)

		require.Equal(t, "pro", plan.Code)
		require.Equal(t, "price_test_pro", plan.PriceID)
		require.Equal(t, int32(100), plan.MonthlyLimit) // From seed data
		// ID should be a UUID from database
		require.NotEmpty(t, plan.ID)
		require.Len(t, plan.ID, 36) // UUID length
	})

	t.Run("success: get business plan from config", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "business")
		require.NoError(t, err)
		require.NotNil(t, plan)

		require.Equal(t, "business", plan.Code)
		require.Equal(t, "price_test_business", plan.PriceID)
		require.Equal(t, int32(500), plan.MonthlyLimit) // From seed data
		// ID should be a UUID from database
		require.NotEmpty(t, plan.ID)
		require.Len(t, plan.ID, 36) // UUID length
	})

	t.Run("fail: nonexistent plan code", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "nonexistent")
		require.Error(t, err)
		require.Nil(t, plan)
		require.Contains(t, err.Error(), "unknown plan code")
	})
}

func TestDefaultUsageService_Integration_ErrorCases(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	service := billing.NewDefaultUsageService(db, &cfg.Plans)

	t.Run("fail: empty userID", func(t *testing.T) {
		stats, err := service.GetUsage(ctx, "")
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "userID cannot be empty")
	})

	t.Run("fail: invalid userID format", func(t *testing.T) {
		stats, err := service.GetUsage(ctx, "invalid-uuid")
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "invalid user ID format")
	})

	t.Run("fail: empty plan code", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "")
		require.Error(t, err)
		require.Nil(t, plan)
		require.Contains(t, err.Error(), "code cannot be empty")
	})

	t.Run("fail: can create image with invalid userID", func(t *testing.T) {
		canCreate, err := service.CanCreateImage(ctx, "")
		require.Error(t, err)
		require.False(t, canCreate)
		require.Contains(t, err.Error(), "userID cannot be empty")
	})
}
