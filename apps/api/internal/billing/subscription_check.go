package billing

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// SubscriptionChecker provides methods to check user subscription status.
type SubscriptionChecker interface {
	// HasActiveSubscription checks if a user has an active paid subscription.
	// Returns true if the user has an active or trialing subscription.
	HasActiveSubscription(ctx context.Context, userID string) (bool, error)
}

// DefaultSubscriptionChecker implements SubscriptionChecker using the database.
type DefaultSubscriptionChecker struct {
	db storage.Database
}

// NewSubscriptionChecker creates a new subscription checker.
func NewSubscriptionChecker(db storage.Database) SubscriptionChecker {
	return &DefaultSubscriptionChecker{db: db}
}

// HasActiveSubscription checks if a user has an active paid subscription.
// Active subscription statuses include: "active" and "trialing".
func (s *DefaultSubscriptionChecker) HasActiveSubscription(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, errors.New("userID cannot be empty")
	}

	q := queries.New(s.db)

	// Query for active subscriptions
	// According to Stripe docs, valid active statuses are:
	// - "active": Subscription is active and paid
	// - "trialing": Subscription is in trial period
	//
	// We explicitly exclude:
	// - "incomplete": Payment failed during creation
	// - "incomplete_expired": Incomplete subscription expired
	// - "past_due": Payment failed but subscription still active (grace period)
	// - "canceled": Subscription has been canceled
	// - "unpaid": Payment failed and no grace period
	//
	// Note: You may want to include "past_due" if you want to allow grace period access.
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}
	userUUID := pgtype.UUID{Bytes: uid, Valid: true}

	subs, err := q.ListSubscriptionsByUserIDAndStatuses(ctx, queries.ListSubscriptionsByUserIDAndStatusesParams{
		UserID:  userUUID,
		Column2: []string{"active", "trialing"},
	})
	if err != nil {
		return false, err
	}

	// If user has at least one active or trialing subscription, they have access
	return len(subs) > 0, nil
}
