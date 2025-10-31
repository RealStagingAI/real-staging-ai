package billing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// PlanSyncService keeps the plans table in sync with environment configuration
type PlanSyncService struct {
	db     storage.Database
	config *config.Plans
}

// NewPlanSyncService creates a new plan sync service
func NewPlanSyncService(db storage.Database, plans *config.Plans) *PlanSyncService {
	return &PlanSyncService{
		db:     db,
		config: plans,
	}
}

// SyncPlans ensures the database plans match the environment configuration
func (s *PlanSyncService) SyncPlans(ctx context.Context) error {
	q := queries.New(s.db)

	// Get all plans from environment config
	envPlans := s.config.GetAllPlans()

	for _, envPlan := range envPlans {
		// Check if plan exists in database
		dbPlan, err := q.GetPlanByCode(ctx, envPlan.Code)
		if err != nil {
			// Plan doesn't exist, create it
			planUUID := uuid.New()
			_, err = q.CreatePlan(ctx, queries.CreatePlanParams{
				ID:           pgtype.UUID{Bytes: planUUID, Valid: true},
				Code:         envPlan.Code,
				PriceID:      envPlan.PriceID,
				MonthlyLimit: envPlan.MonthlyLimit,
			})
			if err != nil {
				return fmt.Errorf("failed to create plan %s: %w", envPlan.Code, err)
			}
		} else if dbPlan.PriceID != envPlan.PriceID || dbPlan.MonthlyLimit != envPlan.MonthlyLimit {
			// Plan exists, update if price ID or limit changed
			_, err = q.UpdatePlan(ctx, queries.UpdatePlanParams{
				Code:         envPlan.Code,
				PriceID:      envPlan.PriceID,
				MonthlyLimit: envPlan.MonthlyLimit,
			})
			if err != nil {
				return fmt.Errorf("failed to update plan %s: %w", envPlan.Code, err)
			}
		}
	}

	return nil
}

// ValidatePriceIDs checks that all active subscriptions have valid price IDs
func (s *PlanSyncService) ValidatePriceIDs(ctx context.Context) error {
	q := queries.New(s.db)

	// Get all active subscriptions
	subs, err := q.ListAllActiveSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active subscriptions: %w", err)
	}

	envPlans := s.config.GetAllPlans()
	priceIDToCode := make(map[string]string)
	for _, plan := range envPlans {
		priceIDToCode[plan.PriceID] = plan.Code
	}

	for _, sub := range subs {
		subscriptionPriceID := sub.PriceID.String
		if _, exists := priceIDToCode[subscriptionPriceID]; !exists {
			return fmt.Errorf("active subscription %s has unknown price_id: %s",
				sub.ID.String(), subscriptionPriceID)
		}
	}

	return nil
}
