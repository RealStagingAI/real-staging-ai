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

	plan, hasSubscription, err := s.resolveUserPlan(ctx, q, userUUID)
	if err != nil {
		return nil, err
	}

	periodStart, periodEnd, err := s.getBillingPeriod(ctx, q, userUUID)
	if err != nil {
		return nil, err
	}

	imagesUsed, err := q.CountImagesCreatedInPeriod(ctx, queries.CountImagesCreatedInPeriodParams{
		UserID:      userUUID,
		CreatedAt:   pgtype.Timestamptz{Time: periodStart, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: periodEnd, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	remaining := plan.MonthlyLimit - imagesUsed
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

// resolveUserPlan determines the user's current plan and subscription status
func (s *DefaultUsageService) resolveUserPlan(
	ctx context.Context,
	q *queries.Queries,
	userUUID pgtype.UUID,
) (*queries.Plan, bool, error) {
	// Get user's active plan from database
	activePlan, err := q.GetUserActivePlan(ctx, userUUID)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, false, err
	}

	if activePlan != nil {
		return activePlan, true, nil
	}

	// Try to find plan by subscription price ID
	plan, hasSubscription, err := s.findPlanBySubscription(ctx, q, userUUID)
	if err != nil {
		return nil, false, err
	}

	if plan != nil {
		return plan, hasSubscription, nil
	}

	// Fallback to free plan
	return s.getFreePlan()
}

// findPlanBySubscription looks up the user's plan based on their active subscription
func (s *DefaultUsageService) findPlanBySubscription(
	ctx context.Context,
	q *queries.Queries,
	userUUID pgtype.UUID,
) (*queries.Plan, bool, error) {
	subs, err := q.ListSubscriptionsByUserIDAndStatuses(ctx, queries.ListSubscriptionsByUserIDAndStatusesParams{
		UserID:  userUUID,
		Column2: []string{"active", "trialing"},
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return nil, false, nil
	}

	mostRecentSub := s.findMostRecentSubscription(subs)
	if mostRecentSub == nil {
		return nil, false, nil
	}

	fmt.Printf("DEBUG: Selected most recent subscription - ID: %s, PriceID: %s, Created: %s\n",
		mostRecentSub.StripeSubscriptionID, mostRecentSub.PriceID.String, mostRecentSub.CreatedAt.Time)

	subscriptionPriceID := mostRecentSub.PriceID.String
	for _, configPlan := range s.config.GetAllPlans() {
		if configPlan.PriceID == subscriptionPriceID {
			plan := &queries.Plan{
				Code:         configPlan.Code,
				PriceID:      configPlan.PriceID,
				MonthlyLimit: configPlan.MonthlyLimit,
			}
			fmt.Printf("DEBUG: Matched plan - Code: %s, Limit: %d\n", plan.Code, plan.MonthlyLimit)
			return plan, true, nil
		}
	}

	fmt.Printf("DEBUG: No plan found for PriceID: %s\n", subscriptionPriceID)
	return nil, false, nil
}

// findMostRecentSubscription returns the most recent subscription from a list
func (s *DefaultUsageService) findMostRecentSubscription(subs []*queries.Subscription) *queries.Subscription {
	var mostRecentSub *queries.Subscription
	for _, sub := range subs {
		isMoreRecent := mostRecentSub == nil ||
			(sub.CreatedAt.Valid && mostRecentSub.CreatedAt.Valid &&
				sub.CreatedAt.Time.After(mostRecentSub.CreatedAt.Time))
		if isMoreRecent {
			mostRecentSub = sub
		}
	}
	return mostRecentSub
}

// getFreePlan returns the default free plan configuration
func (s *DefaultUsageService) getFreePlan() (*queries.Plan, bool, error) {
	freePriceID, err := s.config.GetPriceIDByCode("free")
	if err != nil {
		return nil, false, errors.New("free plan price ID not configured")
	}
	plan := &queries.Plan{
		Code:         "free",
		PriceID:      freePriceID,
		MonthlyLimit: 100, // Updated free tier limit
	}
	return plan, false, nil
}

// getBillingPeriod determines the current billing period for a user
func (s *DefaultUsageService) getBillingPeriod(
	ctx context.Context,
	q *queries.Queries,
	userUUID pgtype.UUID,
) (time.Time, time.Time, error) {
	now := time.Now().UTC()

	// Try to get subscription period for all users (free and paid)
	subs, err := q.ListSubscriptionsByUserIDAndStatuses(ctx, queries.ListSubscriptionsByUserIDAndStatusesParams{
		UserID:  userUUID,
		Column2: []string{"active", "trialing"},
	})
	if err == nil && len(subs) > 0 {
		mostRecentSub := s.findMostRecentSubscription(subs)
		if mostRecentSub != nil && mostRecentSub.CurrentPeriodStart.Valid && mostRecentSub.CurrentPeriodEnd.Valid {
			// Use subscription period from Stripe
			return mostRecentSub.CurrentPeriodStart.Time, mostRecentSub.CurrentPeriodEnd.Time, nil
		}
	}

	// Fallback to calendar month if no subscription found
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)
	return periodStart, periodEnd, nil
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
