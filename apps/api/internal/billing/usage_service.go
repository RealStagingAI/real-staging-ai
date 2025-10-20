package billing

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out usage_service_mock.go . UsageService

// UsageService provides methods to check and enforce usage limits.
type UsageService interface {
	// GetUsage returns the current usage statistics for a user.
	// Returns usage count, monthly limit, plan code, and billing period dates.
	GetUsage(ctx context.Context, userID string) (*UsageStats, error)

	// CanCreateImage checks if a user can create a new image based on their plan limits.
	// Returns true if user is under their limit, false otherwise.
	CanCreateImage(ctx context.Context, userID string) (bool, error)

	// GetPlanByCode returns plan details by plan code (free, pro, business).
	GetPlanByCode(ctx context.Context, code string) (*PlanInfo, error)
}

// UsageStats represents a user's current usage statistics.
type UsageStats struct {
	ImagesUsed      int32  `json:"images_used"`      // Number of images created in current period
	MonthlyLimit    int32  `json:"monthly_limit"`    // Monthly limit for the plan
	PlanCode        string `json:"plan_code"`        // Plan code (free, pro, business)
	PeriodStart     string `json:"period_start"`     // ISO 8601 date of period start
	PeriodEnd       string `json:"period_end"`       // ISO 8601 date of period end
	HasSubscription bool   `json:"has_subscription"` // Whether user has active subscription
	RemainingImages int32  `json:"remaining_images"` // Remaining images in current period
}

// PlanInfo represents details about a subscription plan.
type PlanInfo struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	PriceID      string `json:"price_id"`
	MonthlyLimit int32  `json:"monthly_limit"`
}
