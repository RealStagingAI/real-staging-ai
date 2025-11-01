package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// DefaultUsageService implements UsageService using the database.
type DefaultUsageService struct {
	db     storage.Database
	config *config.Plans
}

// NewDefaultUsageService creates a new usage service.
func NewDefaultUsageService(db storage.Database, plans *config.Plans) UsageService {
	return &DefaultUsageService{
		db:     db,
		config: plans,
	}
}

// GetUsage returns the current usage statistics for a user.
func (s *DefaultUsageService) GetUsage(ctx context.Context, userID string) (*UsageStats, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	userUUID := pgtype.UUID{Bytes: uid, Valid: true}

	q := queries.New(s.db)

	// Get user's active plan (if they have a subscription)
	activePlan, err := q.GetUserActivePlan(ctx, userUUID)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}

	// If no active subscription from DB, try to find plan by subscription price ID
	// This handles the case where DB plans are out of sync with env vars
	var plan *queries.Plan
	hasSubscription := false
	if activePlan != nil {
		plan = activePlan
		hasSubscription = true
	} else {
		// Try to get subscription price ID and find matching plan from config
		subs, err := q.ListSubscriptionsByUserIDAndStatuses(ctx, queries.ListSubscriptionsByUserIDAndStatusesParams{
			UserID:  userUUID,
			Column2: []string{"active", "trialing"},
		})
		if err == nil && len(subs) > 0 {
			// Find the most recent active subscription (sorted by created_at desc)
			var mostRecentSub *queries.Subscription
			for _, sub := range subs {
				if mostRecentSub == nil || (sub.CreatedAt.Valid && mostRecentSub.CreatedAt.Valid && sub.CreatedAt.Time.After(mostRecentSub.CreatedAt.Time)) {
					mostRecentSub = sub
				}
			}
			
			if mostRecentSub != nil {
				fmt.Printf("DEBUG: Selected most recent subscription - ID: %s, PriceID: %s, Created: %s\n", 
					mostRecentSub.StripeSubscriptionID, mostRecentSub.PriceID.String, mostRecentSub.CreatedAt.Time)
				
				// User has active subscription, find plan by price ID using config
				subscriptionPriceID := mostRecentSub.PriceID.String
				for _, configPlan := range s.config.GetAllPlans() {
					if configPlan.PriceID == subscriptionPriceID {
						// Create a plan object from config
						plan = &queries.Plan{
							Code:         configPlan.Code,
							PriceID:      configPlan.PriceID,
							MonthlyLimit: configPlan.MonthlyLimit,
						}
						hasSubscription = true
						fmt.Printf("DEBUG: Matched plan - Code: %s, Limit: %d\n", plan.Code, plan.MonthlyLimit)
						break
					}
				}
				
				if plan == nil {
					fmt.Printf("DEBUG: No plan found for PriceID: %s\n", subscriptionPriceID)
				}
			}
		}

		// If still no plan found, use free plan from config
		if plan == nil {
			freePriceID, err := s.config.GetPriceIDByCode("free")
			if err != nil {
				return nil, errors.New("free plan price ID not configured")
			}
			plan = &queries.Plan{
				Code:         "free",
				PriceID:      freePriceID,
				MonthlyLimit: 100, // Updated free tier limit
			}
		}
	}

	// Calculate billing period from subscription (both free and paid)
	// All users (including free tier) should have a Stripe subscription that tracks their period
	now := time.Now().UTC()
	var periodStart, periodEnd time.Time

	// Try to get subscription period for all users (free and paid)
	subs, err := q.ListSubscriptionsByUserIDAndStatuses(ctx, queries.ListSubscriptionsByUserIDAndStatusesParams{
		UserID:  userUUID,
		Column2: []string{"active", "trialing"},
	})
	if err == nil && len(subs) > 0 {
		// Find the most recent active subscription for billing period
		var mostRecentSub *queries.Subscription
		for _, sub := range subs {
			if mostRecentSub == nil || (sub.CreatedAt.Valid && mostRecentSub.CreatedAt.Valid && sub.CreatedAt.Time.After(mostRecentSub.CreatedAt.Time)) {
				mostRecentSub = sub
			}
		}
		
		if mostRecentSub != nil && mostRecentSub.CurrentPeriodStart.Valid && mostRecentSub.CurrentPeriodEnd.Valid {
			// Use subscription period from Stripe (free tier subscriptions have $0.00 price)
			periodStart = mostRecentSub.CurrentPeriodStart.Time
			periodEnd = mostRecentSub.CurrentPeriodEnd.Time
		}
	} else {
		// Fallback to calendar month if no subscription found
		// (shouldn't happen for properly onboarded users)
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd = periodStart.AddDate(0, 1, 0)
	}

	// Count images created in this period
	imagesUsed, err := q.CountImagesCreatedInPeriod(ctx, queries.CountImagesCreatedInPeriodParams{
		UserID:      userUUID,
		CreatedAt:   pgtype.Timestamptz{Time: periodStart, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: periodEnd, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Calculate remaining images
	var remaining int32
	remaining = plan.MonthlyLimit - imagesUsed
	if remaining < 0 {
		remaining = 0
	}

	return &UsageStats{
		ImagesUsed:      imagesUsed,
		MonthlyLimit:    plan.MonthlyLimit,
		PlanCode:        plan.Code,
		PeriodStart:     periodStart.Format(time.RFC3339),
		PeriodEnd:       periodEnd.Format(time.RFC3339),
		HasSubscription: hasSubscription,
		RemainingImages: remaining,
	}, nil
}

// CanCreateImage checks if a user can create a new image based on their plan limits.
func (s *DefaultUsageService) CanCreateImage(ctx context.Context, userID string) (bool, error) {
	usage, err := s.GetUsage(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if user is under their limit
	return usage.ImagesUsed < usage.MonthlyLimit, nil
}

// GetPlanByCode returns plan details by plan code.
func (s *DefaultUsageService) GetPlanByCode(ctx context.Context, code string) (*PlanInfo, error) {
	if code == "" {
		return nil, errors.New("code cannot be empty")
	}

	// First try to get from database
	q := queries.New(s.db)
	plan, err := q.GetPlanByCode(ctx, code)
	if err == nil {
		return &PlanInfo{
			ID:           plan.ID.String(),
			Code:         plan.Code,
			PriceID:      plan.PriceID,
			MonthlyLimit: plan.MonthlyLimit,
		}, nil
	}

	// If not in database, get from config
	_, err = s.config.GetPriceIDByCode(code)
	if err != nil {
		return nil, err
	}

	for _, configPlan := range s.config.GetAllPlans() {
		if configPlan.Code == code {
			return &PlanInfo{
				ID:           "config-" + code, // Temporary ID
				Code:         configPlan.Code,
				PriceID:      configPlan.PriceID,
				MonthlyLimit: configPlan.MonthlyLimit,
			}, nil
		}
	}

	return nil, errors.New("plan not found")
}
