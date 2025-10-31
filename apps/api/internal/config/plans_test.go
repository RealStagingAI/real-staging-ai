package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlans_GetPriceIDByCode(t *testing.T) {
	tests := []struct {
		name        string
		plans       Plans
		code        string
		expectedID  string
		expectedErr string
	}{
		{
			name: "success: returns free price ID",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			code:       "free",
			expectedID: "price_free_test",
		},
		{
			name: "success: returns pro price ID",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			code:       "pro",
			expectedID: "price_pro_test",
		},
		{
			name: "success: returns business price ID",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			code:       "business",
			expectedID: "price_business_test",
		},
		{
			name: "error: free price ID not configured",
			plans: Plans{
				FreePriceID:     "",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			code:        "free",
			expectedErr: "free plan price ID not configured",
		},
		{
			name: "error: pro price ID not configured",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "",
				BusinessPriceID: "price_business_test",
			},
			code:        "pro",
			expectedErr: "pro plan price ID not configured",
		},
		{
			name: "error: business price ID not configured",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "",
			},
			code:        "business",
			expectedErr: "business plan price ID not configured",
		},
		{
			name: "error: unknown plan code",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			code:        "unknown",
			expectedErr: "unknown plan code: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priceID, err := tt.plans.GetPriceIDByCode(tt.code)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Empty(t, priceID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, priceID)
			}
		})
	}
}

func TestPlans_GetAllPlans(t *testing.T) {
	plans := Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

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
}

func TestPlans_Validate(t *testing.T) {
	tests := []struct {
		name        string
		plans       Plans
		expectedErr string
	}{
		{
			name: "success: all price IDs configured",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			expectedErr: "",
		},
		{
			name: "error: free price ID missing",
			plans: Plans{
				FreePriceID:     "",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "price_business_test",
			},
			expectedErr: "STRIPE_PRICE_FREE environment variable is required",
		},
		{
			name: "error: pro price ID missing",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "",
				BusinessPriceID: "price_business_test",
			},
			expectedErr: "STRIPE_PRICE_PRO environment variable is required",
		},
		{
			name: "error: business price ID missing",
			plans: Plans{
				FreePriceID:     "price_free_test",
				ProPriceID:      "price_pro_test",
				BusinessPriceID: "",
			},
			expectedErr: "STRIPE_PRICE_BUSINESS environment variable is required",
		},
		{
			name: "error: multiple price IDs missing",
			plans: Plans{
				FreePriceID:     "",
				ProPriceID:      "",
				BusinessPriceID: "price_business_test",
			},
			expectedErr: "STRIPE_PRICE_FREE environment variable is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plans.Validate()

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_LoadWithPlans(t *testing.T) {
	// Save original env vars
	originalFree := os.Getenv("STRIPE_PRICE_FREE")
	originalPro := os.Getenv("STRIPE_PRICE_PRO")
	originalBusiness := os.Getenv("STRIPE_PRICE_BUSINESS")
	originalAppEnv := os.Getenv("APP_ENV")

	// Restore after test
	defer func() {
		if originalFree != "" {
			_ = os.Setenv("STRIPE_PRICE_FREE", originalFree)
		} else {
			_ = os.Unsetenv("STRIPE_PRICE_FREE")
		}
		if originalPro != "" {
			_ = os.Setenv("STRIPE_PRICE_PRO", originalPro)
		} else {
			_ = os.Unsetenv("STRIPE_PRICE_PRO")
		}
		if originalBusiness != "" {
			_ = os.Setenv("STRIPE_PRICE_BUSINESS", originalBusiness)
		} else {
			_ = os.Unsetenv("STRIPE_PRICE_BUSINESS")
		}
		if originalAppEnv != "" {
			_ = os.Setenv("APP_ENV", originalAppEnv)
		} else {
			_ = os.Unsetenv("APP_ENV")
		}
	}()

	t.Run("success: loads config with plan environment variables", func(t *testing.T) {
		t.Setenv("APP_ENV", "test")
		t.Setenv("STRIPE_PRICE_FREE", "price_free_test")
		t.Setenv("STRIPE_PRICE_PRO", "price_pro_test")
		t.Setenv("STRIPE_PRICE_BUSINESS", "price_business_test")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "price_free_test", cfg.Plans.FreePriceID)
		assert.Equal(t, "price_pro_test", cfg.Plans.ProPriceID)
		assert.Equal(t, "price_business_test", cfg.Plans.BusinessPriceID)
	})

	t.Run("error: missing required plan environment variables", func(t *testing.T) {
		t.Setenv("APP_ENV", "test")
		// Don't set the required price ID variables

		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid plans configuration")
	})

	t.Run("success: PlansConfig returns plans pointer", func(t *testing.T) {
		t.Setenv("APP_ENV", "test")
		t.Setenv("STRIPE_PRICE_FREE", "price_free_test")
		t.Setenv("STRIPE_PRICE_PRO", "price_pro_test")
		t.Setenv("STRIPE_PRICE_BUSINESS", "price_business_test")

		cfg, err := Load()
		require.NoError(t, err)

		plansConfig := cfg.PlansConfig()
		assert.NotNil(t, plansConfig)
		assert.Equal(t, cfg.Plans.FreePriceID, plansConfig.FreePriceID)
		assert.Equal(t, cfg.Plans.ProPriceID, plansConfig.ProPriceID)
		assert.Equal(t, cfg.Plans.BusinessPriceID, plansConfig.BusinessPriceID)
	})
}
