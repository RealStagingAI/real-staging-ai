package billing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// DefaultUsageService implements UsageService using the database.
type DefaultUsageService struct {
	db storage.Database
}

// NewDefaultUsageService creates a new usage service.
func NewDefaultUsageService(db storage.Database) UsageService {
	return &DefaultUsageService{db: db}
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

	// If no active subscription, use free plan
	var plan *queries.Plan
	hasSubscription := false
	if activePlan != nil {
		plan = activePlan
		hasSubscription = true
	} else {
		// Get free plan
		freePlan, err := q.GetPlanByCode(ctx, "free")
		if err != nil {
			return nil, err
		}
		plan = freePlan
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
	if err == nil && len(subs) > 0 && subs[0].CurrentPeriodStart.Valid && subs[0].CurrentPeriodEnd.Valid {
		// Use subscription period from Stripe (free tier subscriptions have $0.00 price)
		periodStart = subs[0].CurrentPeriodStart.Time
		periodEnd = subs[0].CurrentPeriodEnd.Time
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

	q := queries.New(s.db)
	plan, err := q.GetPlanByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return &PlanInfo{
		ID:           plan.ID.String(),
		Code:         plan.Code,
		PriceID:      plan.PriceID,
		MonthlyLimit: plan.MonthlyLimit,
	}, nil
}
