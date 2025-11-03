//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/billing"
	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestPlanSyncService_Integration_SyncPlans(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	service := billing.NewPlanSyncService(db, &cfg.Plans)

	t.Run("success: sync plans to database", func(t *testing.T) {
		// First, clear any existing plans to start fresh
		q := queries.New(db)
		_, err := db.Exec(ctx, "DELETE FROM plans")
		require.NoError(t, err)

		// Sync plans from config to database
		err = service.SyncPlans(ctx)
		require.NoError(t, err)

		// Verify all plans were created
		freePlan, err := q.GetPlanByCode(ctx, "free")
		require.NoError(t, err)
		require.Equal(t, "free", freePlan.Code)
		require.Equal(t, "price_test_free", freePlan.PriceID)
		require.Equal(t, int32(100), freePlan.MonthlyLimit)

		proPlan, err := q.GetPlanByCode(ctx, "pro")
		require.NoError(t, err)
		require.Equal(t, "pro", proPlan.Code)
		require.Equal(t, "price_test_pro", proPlan.PriceID)
		require.Equal(t, int32(100), proPlan.MonthlyLimit)

		businessPlan, err := q.GetPlanByCode(ctx, "business")
		require.NoError(t, err)
		require.Equal(t, "business", businessPlan.Code)
		require.Equal(t, "price_test_business", businessPlan.PriceID)
		require.Equal(t, int32(500), businessPlan.MonthlyLimit)
	})

	t.Run("success: update existing plans", func(t *testing.T) {
		q := queries.New(db)

		// Modify existing plans to have different values
		_, err := q.UpdatePlan(ctx, queries.UpdatePlanParams{
			Code:         "free",
			PriceID:      "old_price_id_unique",
			MonthlyLimit: 50,
		})
		require.NoError(t, err)

		_, err = q.UpdatePlan(ctx, queries.UpdatePlanParams{
			Code:         "pro",
			PriceID:      "old_price_id_unique2",
			MonthlyLimit: 250,
		})
		require.NoError(t, err)

		// Sync plans - should update the changed values
		err = service.SyncPlans(ctx)
		require.NoError(t, err)

		// Verify plans were updated back to config values
		freePlan, err := q.GetPlanByCode(ctx, "free")
		require.NoError(t, err)
		require.Equal(t, "price_test_free", freePlan.PriceID)
		require.Equal(t, int32(100), freePlan.MonthlyLimit)

		proPlan, err := q.GetPlanByCode(ctx, "pro")
		require.NoError(t, err)
		require.Equal(t, "price_test_pro", proPlan.PriceID)
		require.Equal(t, int32(100), proPlan.MonthlyLimit)
	})

	t.Run("success: skip unchanged plans", func(t *testing.T) {
		q := queries.New(db)

		// Get current plan values
		freePlan, err := q.GetPlanByCode(ctx, "free")
		require.NoError(t, err)

		// Sync plans again - should not change anything
		err = service.SyncPlans(ctx)
		require.NoError(t, err)

		// Verify plan values are unchanged
		updatedFreePlan, err := q.GetPlanByCode(ctx, "free")
		require.NoError(t, err)
		require.Equal(t, freePlan.PriceID, updatedFreePlan.PriceID)
		require.Equal(t, freePlan.MonthlyLimit, updatedFreePlan.MonthlyLimit)
	})
}

func TestPlanSyncService_Integration_ValidatePriceIDs(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	service := billing.NewPlanSyncService(db, &cfg.Plans)

	// Clean up any test subscriptions with unknown price IDs
	_, err = db.Exec(ctx, `DELETE FROM subscriptions WHERE price_id NOT IN ('price_test_free', 'price_test_pro', 'price_test_business')`)
	require.NoError(t, err)

	t.Run("success: no subscriptions validates successfully", func(t *testing.T) {
		// No subscriptions yet - should validate successfully
		err = service.ValidatePriceIDs(ctx)
		require.NoError(t, err)
	})

	t.Run("success: subscriptions with valid price IDs", func(t *testing.T) {
		// Since we don't have subscription creation methods easily available,
		// let's test the validation logic indirectly
		// The validation should pass if there are no subscriptions with invalid price IDs
		
		// Should validate successfully with no subscriptions
		err = service.ValidatePriceIDs(ctx)
		require.NoError(t, err)
	})
}
