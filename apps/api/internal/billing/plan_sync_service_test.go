package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
)

func TestNewPlanSyncService(t *testing.T) {
	db := &storage.DatabaseMock{}
	plans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewPlanSyncService(db, plans)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, plans, service.config)
}

func TestPlanSyncService_ConfigMethods(t *testing.T) {
	plans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	t.Run("GetPriceIDByCode returns correct price IDs", func(t *testing.T) {
		freePriceID, err := plans.GetPriceIDByCode("free")
		assert.NoError(t, err)
		assert.Equal(t, "price_free_test", freePriceID)

		proPriceID, err := plans.GetPriceIDByCode("pro")
		assert.NoError(t, err)
		assert.Equal(t, "price_pro_test", proPriceID)

		businessPriceID, err := plans.GetPriceIDByCode("business")
		assert.NoError(t, err)
		assert.Equal(t, "price_business_test", businessPriceID)

		_, err = plans.GetPriceIDByCode("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown plan code")
	})

	t.Run("GetAllPlans returns all plan definitions", func(t *testing.T) {
		allPlans := plans.GetAllPlans()
		assert.Len(t, allPlans, 3)

		// Check free plan
		freePlan := allPlans[0]
		assert.Equal(t, "free", freePlan.Code)
		assert.Equal(t, "price_free_test", freePlan.PriceID)
		assert.Equal(t, int32(100), freePlan.MonthlyLimit)

		// Check pro plan
		proPlan := allPlans[1]
		assert.Equal(t, "pro", proPlan.Code)
		assert.Equal(t, "price_pro_test", proPlan.PriceID)
		assert.Equal(t, int32(100), proPlan.MonthlyLimit)

		// Check business plan
		businessPlan := allPlans[2]
		assert.Equal(t, "business", businessPlan.Code)
		assert.Equal(t, "price_business_test", businessPlan.PriceID)
		assert.Equal(t, int32(500), businessPlan.MonthlyLimit)
	})

	t.Run("Validate checks all price IDs are set", func(t *testing.T) {
		err := plans.Validate()
		assert.NoError(t, err)

		// Test with missing free price ID
		invalidPlans := &config.Plans{
			FreePriceID:     "",
			ProPriceID:      "price_pro_test",
			BusinessPriceID: "price_business_test",
		}
		err = invalidPlans.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "STRIPE_PRICE_FREE environment variable is required")
	})
}
