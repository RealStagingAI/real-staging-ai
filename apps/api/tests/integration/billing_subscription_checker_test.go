//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/billing"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestDefaultSubscriptionChecker_Integration_HasActiveSubscription(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	q := queries.New(db)

	checker := billing.NewDefaultSubscriptionChecker(db)

	t.Run("success: user with no subscription returns false", func(t *testing.T) {
		// Create a test user
		userUUID := uuid.New()
		_, err := q.CreateUser(ctx, queries.CreateUserParams{
			Auth0Sub:         "test|nosub|" + userUUID.String(),
			StripeCustomerID: pgtype.Text{String: "cus_nosub_" + userUUID.String(), Valid: true},
			Role:             "user",
		})
		require.NoError(t, err)

		// Check subscription status
		hasSubscription, err := checker.HasActiveSubscription(ctx, userUUID.String())
		require.NoError(t, err)
		require.False(t, hasSubscription)
	})

	t.Run("success: user with active subscription returns true", func(t *testing.T) {
		// Create a test user
		userUUID := uuid.New()
		userRow, err := q.CreateUser(ctx, queries.CreateUserParams{
			Auth0Sub:         "test|active|" + userUUID.String(),
			StripeCustomerID: pgtype.Text{String: "cus_active_" + userUUID.String(), Valid: true},
			Role:             "user",
		})
		require.NoError(t, err)

		// Create an active subscription using raw SQL since we don't have the CreateSubscription method
		subUUID := uuid.New()
		_, err = db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, stripe_subscription_id, price_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, subUUID, userRow.ID, "sub_active_"+userUUID.String(), "price_test_free", "active")
		require.NoError(t, err)

		// Check subscription status using the actual user ID from database
		dbUserID := uuid.UUID(userRow.ID.Bytes)
		hasSubscription, err := checker.HasActiveSubscription(ctx, dbUserID.String())
		require.NoError(t, err)
		require.True(t, hasSubscription)
	})

	t.Run("success: user with trialing subscription returns true", func(t *testing.T) {
		// Create a test user
		userUUID := uuid.New()
		userRow, err := q.CreateUser(ctx, queries.CreateUserParams{
			Auth0Sub:         "test|trialing|" + userUUID.String(),
			StripeCustomerID: pgtype.Text{String: "cus_trialing_" + userUUID.String(), Valid: true},
			Role:             "user",
		})
		require.NoError(t, err)

		// Create a trialing subscription using raw SQL
		subUUID := uuid.New()
		_, err = db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, stripe_subscription_id, price_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, subUUID, userRow.ID, "sub_trial_"+userUUID.String(), "price_test_pro", "trialing")
		require.NoError(t, err)

		// Check subscription status using the actual user ID from database
		dbUserID := uuid.UUID(userRow.ID.Bytes)
		hasSubscription, err := checker.HasActiveSubscription(ctx, dbUserID.String())
		require.NoError(t, err)
		require.True(t, hasSubscription)
	})

	t.Run("success: user with only canceled subscription returns false", func(t *testing.T) {
		// Create a test user
		userUUID := uuid.New()
		userRow, err := q.CreateUser(ctx, queries.CreateUserParams{
			Auth0Sub:         "test|canceled|" + userUUID.String(),
			StripeCustomerID: pgtype.Text{String: "cus_canceled_" + userUUID.String(), Valid: true},
			Role:             "user",
		})
		require.NoError(t, err)

		// Create a canceled subscription using raw SQL
		subUUID := uuid.New()
		_, err = db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, stripe_subscription_id, price_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, subUUID, userRow.ID, "sub_canceled_"+userUUID.String(), "price_test_business", "canceled")
		require.NoError(t, err)

		// Check subscription status using the actual user ID from database
		dbUserID := uuid.UUID(userRow.ID.Bytes)
		hasSubscription, err := checker.HasActiveSubscription(ctx, dbUserID.String())
		require.NoError(t, err)
		require.False(t, hasSubscription)
	})

	t.Run("success: user with multiple subscriptions returns true if any are active", func(t *testing.T) {
		// Create a test user
		userUUID := uuid.New()
		userRow, err := q.CreateUser(ctx, queries.CreateUserParams{
			Auth0Sub:         "test|multiple|" + userUUID.String(),
			StripeCustomerID: pgtype.Text{String: "cus_multiple_" + userUUID.String(), Valid: true},
			Role:             "user",
		})
		require.NoError(t, err)

		// Create multiple subscriptions - one canceled, one active
		canceledUUID := uuid.New()
		activeUUID := uuid.New()
		
		_, err = db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, stripe_subscription_id, price_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, canceledUUID, userRow.ID, "sub_canceled_"+userUUID.String(), "price_test_free", "canceled")
		require.NoError(t, err)

		_, err = db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, stripe_subscription_id, price_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, activeUUID, userRow.ID, "sub_active_"+userUUID.String(), "price_test_pro", "active")
		require.NoError(t, err)

		// Check subscription status using the actual user ID from database
		dbUserID := uuid.UUID(userRow.ID.Bytes)
		hasSubscription, err := checker.HasActiveSubscription(ctx, dbUserID.String())
		require.NoError(t, err)
		require.True(t, hasSubscription)
	})
}
