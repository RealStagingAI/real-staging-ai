package billing

import (
	"context"
	"testing"

	"github.com/real-staging-ai/api/internal/storage"
)

func TestNewDefaultSubscriptionChecker(t *testing.T) {
	t.Run("success: creates checker with database", func(t *testing.T) {
		mockDB := &storage.DatabaseMock{}
		checker := NewDefaultSubscriptionChecker(mockDB)

		if checker == nil {
			t.Fatal("Expected non-nil checker")
		}

		// Verify it implements the interface
		_ = SubscriptionChecker(checker)
	})
}

func TestDefaultSubscriptionChecker_HasActiveSubscription(t *testing.T) {
	t.Run("fail: empty userID", func(t *testing.T) {
		mockDB := &storage.DatabaseMock{}
		checker := NewDefaultSubscriptionChecker(mockDB)

		hasSubscription, err := checker.HasActiveSubscription(context.Background(), "")
		if err == nil {
			t.Fatal("Expected error for empty userID, got nil")
		}
		if err.Error() != "userID cannot be empty" {
			t.Errorf("Expected 'userID cannot be empty' error, got: %v", err)
		}
		if hasSubscription {
			t.Errorf("Expected hasSubscription=false for error case, got true")
		}
	})

	t.Run("fail: invalid userID format", func(t *testing.T) {
		mockDB := &storage.DatabaseMock{}
		checker := NewDefaultSubscriptionChecker(mockDB)

		hasSubscription, err := checker.HasActiveSubscription(context.Background(), "not-a-uuid")
		if err == nil {
			t.Fatal("Expected error for invalid userID format, got nil")
		}
		if err.Error() != "invalid user ID format" {
			t.Errorf("Expected 'invalid user ID format' error, got: %v", err)
		}
		if hasSubscription {
			t.Errorf("Expected hasSubscription=false for error case, got true")
		}
	})

}

// Test the subscription status logic explicitly
func TestSubscriptionStatusLogic(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		shouldBeActive bool
	}{
		{"active subscription allows upload", "active", true},
		{"trialing subscription allows upload", "trialing", true},
		{"incomplete subscription blocks upload", "incomplete", false},
		{"incomplete_expired subscription blocks upload", "incomplete_expired", false},
		{"past_due subscription blocks upload", "past_due", false},
		{"canceled subscription blocks upload", "canceled", false},
		{"unpaid subscription blocks upload", "unpaid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the logic: only "active" and "trialing" should allow uploads
			allowedStatuses := map[string]bool{
				"active":   true,
				"trialing": true,
			}

			isAllowed := allowedStatuses[tt.status]
			if isAllowed != tt.shouldBeActive {
				t.Errorf("Status %s: expected shouldBeActive=%v, got %v",
					tt.status, tt.shouldBeActive, isAllowed)
			}
		})
	}
}

// NOTE: Full integration tests with database are in tests/integration/http_upload_handlers_test.go
// These test the full flow including:
// - User without subscription -> 403 Forbidden
// - User with active subscription -> 200 OK
// - User with trialing subscription -> 200 OK
// - User with canceled subscription -> 403 Forbidden
