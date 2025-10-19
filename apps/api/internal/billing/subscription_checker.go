package billing

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out subscription_checker_mock.go . SubscriptionChecker

// SubscriptionChecker provides methods to check user subscription status.
type SubscriptionChecker interface {
	// HasActiveSubscription checks if a user has an active paid subscription.
	// Returns true if the user has an active or trialing subscription.
	// Active subscription statuses include: "active" and "trialing".
	HasActiveSubscription(ctx context.Context, userID string) (bool, error)
}
