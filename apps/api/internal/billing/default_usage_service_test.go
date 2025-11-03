package billing

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestNewDefaultUsageService(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	defaultService, ok := service.(*DefaultUsageService)
	if !ok {
		t.Fatal("Expected service to be *DefaultUsageService")
	}

	if defaultService.db != mockDB {
		t.Error("Expected db to be set correctly")
	}

	if defaultService.config != testPlans {
		t.Error("Expected config to be set correctly")
	}
}

func TestDefaultUsageService_GetUsage_validation(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans)
	ctx := context.Background()

	t.Run("fail: empty userID", func(t *testing.T) {
		stats, err := service.GetUsage(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty userID, got nil")
		}
		if err.Error() != "userID cannot be empty" {
			t.Errorf("Expected 'userID cannot be empty' error, got: %v", err)
		}
		if stats != nil {
			t.Error("Expected stats to be nil for error case")
		}
	})

	t.Run("fail: invalid userID format", func(t *testing.T) {
		stats, err := service.GetUsage(ctx, "not-a-uuid")
		if err == nil {
			t.Fatal("Expected error for invalid userID format, got nil")
		}
		if err.Error() != "invalid user ID format" {
			t.Errorf("Expected 'invalid user ID format' error, got: %v", err)
		}
		if stats != nil {
			t.Error("Expected stats to be nil for error case")
		}
	})
}

func TestDefaultUsageService_CanCreateImage_validation(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans)
	ctx := context.Background()

	t.Run("fail: empty userID", func(t *testing.T) {
		canCreate, err := service.CanCreateImage(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty userID, got nil")
		}
		if err.Error() != "userID cannot be empty" {
			t.Errorf("Expected 'userID cannot be empty' error, got: %v", err)
		}
		if canCreate {
			t.Error("Expected canCreate to be false for error case")
		}
	})

	t.Run("fail: invalid userID format", func(t *testing.T) {
		canCreate, err := service.CanCreateImage(ctx, "not-a-uuid")
		if err == nil {
			t.Fatal("Expected error for invalid userID format, got nil")
		}
		if err.Error() != "invalid user ID format" {
			t.Errorf("Expected 'invalid user ID format' error, got: %v", err)
		}
		if canCreate {
			t.Error("Expected canCreate to be false for error case")
		}
	})
}

func TestDefaultUsageService_GetPlanByCode_validation(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans)
	ctx := context.Background()

	t.Run("fail: empty code", func(t *testing.T) {
		plan, err := service.GetPlanByCode(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty code, got nil")
		}
		if err.Error() != "code cannot be empty" {
			t.Errorf("Expected 'code cannot be empty' error, got: %v", err)
		}
		if plan != nil {
			t.Error("Expected plan to be nil for error case")
		}
	})
}

func TestDefaultUsageService_getFreePlan(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans).(*DefaultUsageService)

	plan, hasSubscription, err := service.getFreePlan()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if plan == nil {
		t.Fatal("Expected plan to be returned")
	}

	if plan.Code != "free" {
		t.Errorf("Expected plan code 'free' but got '%s'", plan.Code)
	}

	if plan.PriceID != "price_free_test" {
		t.Errorf("Expected price ID 'price_free_test' but got '%s'", plan.PriceID)
	}

	if plan.MonthlyLimit != 100 {
		t.Errorf("Expected monthly limit 100 but got %d", plan.MonthlyLimit)
	}

	if hasSubscription {
		t.Error("Expected hasSubscription to be false for free plan")
	}
}

func TestDefaultUsageService_getFreePlan_missingConfig(t *testing.T) {
	// Test with empty plans config
	mockDB := &storage.DatabaseMock{}
	emptyPlans := &config.Plans{
		FreePriceID: "", // Empty free price ID
	}

	service := NewDefaultUsageService(mockDB, emptyPlans).(*DefaultUsageService)

	plan, hasSubscription, err := service.getFreePlan()

	if err == nil {
		t.Error("Expected error for missing free plan config")
		return
	}

	if plan != nil {
		t.Error("Expected plan to be nil when config is missing")
	}

	if hasSubscription {
		t.Error("Expected hasSubscription to be false when config is missing")
	}

	if err.Error() != "free plan price ID not configured" {
		t.Errorf("Expected specific error message but got: %v", err)
	}
}

func TestDefaultUsageService_findMostRecentSubscription(t *testing.T) {
	mockDB := &storage.DatabaseMock{}
	testPlans := &config.Plans{
		FreePriceID:     "price_free_test",
		ProPriceID:      "price_pro_test",
		BusinessPriceID: "price_business_test",
	}

	service := NewDefaultUsageService(mockDB, testPlans).(*DefaultUsageService)

	t.Run("empty list returns nil", func(t *testing.T) {
		result := service.findMostRecentSubscription([]*queries.Subscription{})
		if result != nil {
			t.Errorf("Expected nil but got %+v", result)
		}
	})

	t.Run("single subscription returns that one", func(t *testing.T) {
		now := time.Now().UTC()
		subs := []*queries.Subscription{
			{
				StripeSubscriptionID: "sub_1",
				CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			},
		}

		result := service.findMostRecentSubscription(subs)
		if result == nil {
			t.Error("Expected subscription but got nil")
			return
		}

		if result.StripeSubscriptionID != "sub_1" {
			t.Errorf("Expected sub_1 but got %s", result.StripeSubscriptionID)
		}
	})

	t.Run("returns most recent subscription", func(t *testing.T) {
		now := time.Now().UTC()
		subs := []*queries.Subscription{
			{
				StripeSubscriptionID: "sub_old",
				CreatedAt:            pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
			},
			{
				StripeSubscriptionID: "sub_new",
				CreatedAt:            pgtype.Timestamptz{Time: now.Add(-1 * time.Hour), Valid: true},
			},
			{
				StripeSubscriptionID: "sub_oldest",
				CreatedAt:            pgtype.Timestamptz{Time: now.Add(-3 * time.Hour), Valid: true},
			},
		}

		result := service.findMostRecentSubscription(subs)
		if result == nil {
			t.Error("Expected subscription but got nil")
			return
		}

		if result.StripeSubscriptionID != "sub_new" {
			t.Errorf("Expected sub_new but got %s", result.StripeSubscriptionID)
		}
	})

	t.Run("handles nil CreatedAt gracefully", func(t *testing.T) {
		now := time.Now().UTC()
		subs := []*queries.Subscription{
			{
				StripeSubscriptionID: "sub_with_time",
				CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			},
			{
				StripeSubscriptionID: "sub_no_time",
				CreatedAt:            pgtype.Timestamptz{Valid: false},
			},
		}

		result := service.findMostRecentSubscription(subs)
		if result == nil {
			t.Error("Expected subscription but got nil")
			return
		}

		if result.StripeSubscriptionID != "sub_with_time" {
			t.Errorf("Expected sub_with_time but got %s", result.StripeSubscriptionID)
		}
	})
}
