package billing

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/subscription"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
	stripeLib "github.com/real-staging-ai/api/internal/stripe"
	"github.com/real-staging-ai/api/internal/user"
)

// DefaultHandler implements the billing Handler by wrapping existing repositories
// and user resolution logic (Auth0 sub -> ensure users row).
type DefaultHandler struct {
	db              storage.Database
	usageService    UsageService
	stripeSecretKey string
	config          *config.Config
}

// NewDefaultHandler constructs a DefaultHandler.
func NewDefaultHandler(
	db storage.Database,
	usageService UsageService,
	stripeSecretKey string,
	cfg *config.Config,
) *DefaultHandler {
	return &DefaultHandler{
		db:              db,
		usageService:    usageService,
		stripeSecretKey: stripeSecretKey,
		config:          cfg,
	}
}

// ErrorResponse is a simple JSON error envelope for handler responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GetMySubscriptions returns the current user's subscriptions (paginated).
func (h *DefaultHandler) GetMySubscriptions(c echo.Context) error {
	limit, offset := h.parseLimitOffset(c)

	// No DB configured (e.g., special test mode) — return empty list gracefully.
	if h.db == nil {
		return c.JSON(http.StatusOK, ListResponse[SubscriptionDTO]{Items: []SubscriptionDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	var userID string
	if existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub); err != nil {
		// Create user on first access
		if newUser, createErr := uRepo.Create(c.Request().Context(), auth0Sub, "", "user"); createErr != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		} else {
			userID = newUser.ID.String()
		}
	} else {
		userID = existingUser.ID.String()
	}

	subRepo := stripeLib.NewSubscriptionsRepository(h.db)
	rows, err := subRepo.ListByUserID(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to list subscriptions",
		})
	}

	items := make([]SubscriptionDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, SubscriptionDTO{
			ID:                   uuidToString(r.ID),
			StripeSubscriptionID: r.StripeSubscriptionID,
			Status:               r.Status,
			PriceID:              textPtr(r.PriceID),
			CurrentPeriodStart:   timePtr(r.CurrentPeriodStart),
			CurrentPeriodEnd:     timePtr(r.CurrentPeriodEnd),
			CancelAt:             timePtr(r.CancelAt),
			CanceledAt:           timePtr(r.CanceledAt),
			CancelAtPeriodEnd:    r.CancelAtPeriodEnd,
			CreatedAt:            r.CreatedAt.Time,
			UpdatedAt:            r.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, ListResponse[SubscriptionDTO]{Items: items, Limit: limit, Offset: offset})
}

// GetMyInvoices returns the current user's invoices (paginated).
func (h *DefaultHandler) GetMyInvoices(c echo.Context) error {
	limit, offset := h.parseLimitOffset(c)

	// No DB configured (e.g., special test mode) — return empty list gracefully.
	if h.db == nil {
		return c.JSON(http.StatusOK, ListResponse[InvoiceDTO]{Items: []InvoiceDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	var userID string
	if existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub); err != nil {
		// Create user on first access
		if newUser, createErr := uRepo.Create(c.Request().Context(), auth0Sub, "", "user"); createErr != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		} else {
			userID = newUser.ID.String()
		}
	} else {
		userID = existingUser.ID.String()
	}

	invRepo := stripeLib.NewInvoicesRepository(h.db)
	rows, err := invRepo.ListByUserID(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to list invoices",
		})
	}

	items := make([]InvoiceDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, InvoiceDTO{
			ID:                   uuidToString(r.ID),
			StripeInvoiceID:      r.StripeInvoiceID,
			StripeSubscriptionID: textPtr(r.StripeSubscriptionID),
			Status:               r.Status,
			AmountDue:            r.AmountDue,
			AmountPaid:           r.AmountPaid,
			Currency:             textPtr(r.Currency),
			InvoiceNumber:        textPtr(r.InvoiceNumber),
			CreatedAt:            r.CreatedAt.Time,
			UpdatedAt:            r.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, ListResponse[InvoiceDTO]{Items: items, Limit: limit, Offset: offset})
}

// parseLimitOffset reads limit/offset from query params and applies defaults/caps.
func (h *DefaultHandler) parseLimitOffset(c echo.Context) (int32, int32) {
	limit := DefaultLimit
	offset := int32(0)

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= int(MaxLimit) {
			// #nosec G109,G115 -- Value is validated to be positive and within MaxLimit
			limit = int32(n)
		}
	}

	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 2147483647 {
			// #nosec G109,G115 -- Value is validated to fit in int32 range
			offset = int32(n)
		}
	}

	return limit, offset
}

// Helper mappers for sqlc/pgx types into DTO pointers.

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return pgUUIDToString(u)
}

// pgUUIDToString converts a pgtype.UUID to its canonical string form.
func pgUUIDToString(u pgtype.UUID) string {
	// pgtype.UUID.Bytes is a [16]byte; String() typically requires github.com/google/uuid.
	// Avoid importing another dep here by delegating to the repository’s string conversion,
	// but since we don't have it, reconstruct via the standard formatting.
	b := u.Bytes
	// Format as 8-4-4-4-12
	return formatUUIDBytes(b)
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func timePtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

// formatUUIDBytes formats a 16-byte UUID into a canonical string.
// This avoids importing github.com/google/uuid just for formatting.
func formatUUIDBytes(b [16]byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 36)

	writeByte := func(dst []byte, v byte) {
		dst[0] = hexdigits[v>>4]
		dst[1] = hexdigits[v&0x0f]
	}

	j := 0
	for i := 0; i < 16; i++ {
		switch i {
		case 4, 6, 8, 10:
			out[j] = '-'
			j++
		}
		writeByte(out[j:j+2], b[i])
		j += 2
	}
	return string(out)
}

// CreateCheckoutSession creates a Stripe Checkout Session for subscription signup.
// POST /api/v1/billing/create-checkout
func (h *DefaultHandler) CreateCheckoutSession(c echo.Context) error {
	var req struct {
		PriceID string `json:"price_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	if req.PriceID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "price_id is required",
		})
	}

	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get or create user
	uRepo := user.NewDefaultRepository(h.db)
	userRow, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		// Create user on first access
		_, createErr := uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
		if createErr != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		}
		// Get the newly created user
		userRow, err = uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user after creation",
			})
		}
	}

	// Set Stripe API key from config
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	// Create or get Stripe customer
	var customerID string
	if userRow.StripeCustomerID.Valid && userRow.StripeCustomerID.String != "" {
		customerID = userRow.StripeCustomerID.String
	} else {
		// Create new Stripe customer
		customerParams := &stripe.CustomerParams{
			Metadata: map[string]string{
				"user_id":   userRow.ID.String(),
				"auth0_sub": auth0Sub,
			},
		}
		cust, err := customer.New(customerParams)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: fmt.Sprintf("Failed to create Stripe customer: %v", err),
			})
		}
		customerID = cust.ID

		// Update user with Stripe customer ID
		if _, err := uRepo.UpdateStripeCustomerID(c.Request().Context(), userRow.ID.String(), customerID); err != nil {
			// Log but don't fail - customer is created
			fmt.Printf("Warning: failed to update user with Stripe customer ID: %v\n", err)
		}
	}

	// Create checkout session
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(baseURL + "/profile?checkout=success"),
		CancelURL:  stripe.String(baseURL + "/profile?checkout=canceled"),
	}

	// For free plans, don't require payment method collection
	if req.PriceID == h.config.Plans.FreePriceID {
		params.PaymentMethodCollection = stripe.String("off")
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to create checkout session: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": sess.URL,
	})
}

// CreatePortalSession creates a Stripe Customer Portal session for subscription management.
// POST /api/v1/billing/portal
func (h *DefaultHandler) CreatePortalSession(c echo.Context) error {
	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get user
	uRepo := user.NewDefaultRepository(h.db)
	existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	if !existingUser.StripeCustomerID.Valid || existingUser.StripeCustomerID.String == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "No payment method on file. Please subscribe first.",
		})
	}

	// Set Stripe API key from config
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(existingUser.StripeCustomerID.String),
		ReturnURL: stripe.String(baseURL + "/profile"),
	}

	sess, err := session.New(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to create portal session: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": sess.URL,
	})
}

// GetMyUsage returns the current user's usage statistics.
// GET /api/v1/billing/usage
func (h *DefaultHandler) GetMyUsage(c echo.Context) error {
	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get or create user
	uRepo := user.NewDefaultRepository(h.db)
	userRow, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		// Create user on first access
		_, createErr := uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
		if createErr != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		}
		// Get the newly created user
		userRow, err = uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user after creation",
			})
		}
	}

	// Get usage statistics
	usage, err := h.usageService.GetUsage(c.Request().Context(), userRow.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to get usage: %v", err),
		})
	}

	return c.JSON(http.StatusOK, usage)
}

// CreateSubscriptionWithElements creates a subscription and returns client secret for Elements confirmation
// POST /api/v1/billing/create-subscription-elements
func (h *DefaultHandler) CreateSubscriptionWithElements(c echo.Context) error {
	var req struct {
		PriceID string `json:"price_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get or create user
	uRepo := user.NewDefaultRepository(h.db)
	userRow, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	// Set Stripe API key
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	// Create or get Stripe customer
	var customerID string
	if userRow.StripeCustomerID.Valid && userRow.StripeCustomerID.String != "" {
		customerID = userRow.StripeCustomerID.String
	} else {
		customerParams := &stripe.CustomerParams{
			Metadata: map[string]string{
				"user_id":   userRow.ID.String(),
				"auth0_sub": auth0Sub,
			},
		}
		cust, err := customer.New(customerParams)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: fmt.Sprintf("Failed to create Stripe customer: %v", err),
			})
		}
		customerID = cust.ID

		// Update user with Stripe customer ID
		if _, err := uRepo.UpdateStripeCustomerID(c.Request().Context(), userRow.ID.String(), customerID); err != nil {
			fmt.Printf("Warning: failed to update user with Stripe customer ID: %v\n", err)
		}
	}

	// Create subscription with incomplete payment
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PriceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
		Expand: []*string{
			stripe.String("latest_invoice.payment_intent"),
		},
	}

	// For free plans, don't require payment method
	if req.PriceID == h.config.Plans.FreePriceID {
		subscriptionParams.PaymentBehavior = stripe.String("allow_incomplete")
	}

	subscription, err := subscription.New(subscriptionParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to create subscription: %v", err),
		})
	}

	// Extract client secret from the latest invoice
	var clientSecret string
	if subscription.LatestInvoice != nil && subscription.LatestInvoice.PaymentIntent != nil {
		clientSecret = subscription.LatestInvoice.PaymentIntent.ClientSecret
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscriptionId": subscription.ID,
		"clientSecret":   clientSecret,
	})
}

// GetPaymentMethods returns the customer's saved payment methods
// GET /api/v1/billing/payment-methods
func (h *DefaultHandler) GetPaymentMethods(c echo.Context) error {
	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get user
	uRepo := user.NewDefaultRepository(h.db)
	existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	if !existingUser.StripeCustomerID.Valid || existingUser.StripeCustomerID.String == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"paymentMethods": []interface{}{},
		})
	}

	// Set Stripe API key
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	// List payment methods
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(existingUser.StripeCustomerID.String),
		Type:     stripe.String("card"),
	}

	iterator := paymentmethod.List(params)
	var paymentMethods []interface{}

	for iterator.Next() {
		pm := iterator.PaymentMethod()
		paymentMethods = append(paymentMethods, map[string]interface{}{
			"id":   pm.ID,
			"type": pm.Type,
			"card": map[string]interface{}{
				"brand":    pm.Card.Brand,
				"last4":    pm.Card.Last4,
				"expMonth": pm.Card.ExpMonth,
				"expYear":  pm.Card.ExpYear,
			},
			"isDefault": pm.Metadata["is_default"] == "true",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"paymentMethods": paymentMethods,
	})
}

// UpgradeSubscription upgrades an existing subscription to a new price tier
// POST /api/v1/billing/upgrade-subscription
func (h *DefaultHandler) UpgradeSubscription(c echo.Context) error {
	var req struct {
		PriceID string `json:"price_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get user with subscription
	uRepo := user.NewDefaultRepository(h.db)
	existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	// Set Stripe API key
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	// Get user's existing subscription from database
	subRepo := stripeLib.NewSubscriptionsRepository(h.db)
	dbSubscriptions, err := subRepo.ListByUserID(c.Request().Context(), existingUser.ID.String(), 10, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve existing subscription",
		})
	}

	// Find active subscription
	var activeSubscription *queries.Subscription
	for _, sub := range dbSubscriptions {
		if sub.Status == "active" || sub.Status == "trialing" {
			activeSubscription = sub
			break
		}
	}

	if activeSubscription == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_active_subscription",
			Message: "No active subscription found to upgrade",
		})
	}

	// Retrieve the Stripe subscription
	stripeSubscription, err := subscription.Get(activeSubscription.StripeSubscriptionID, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to retrieve Stripe subscription: %v", err),
		})
	}

	// For free plans, downgrade immediately without payment
	if req.PriceID == h.config.Plans.FreePriceID {
		// Update subscription to free plan
		_, err = subscription.Update(stripeSubscription.ID, &stripe.SubscriptionParams{
			Items: []*stripe.SubscriptionItemsParams{
				{
					ID:    stripe.String(stripeSubscription.Items.Data[0].ID),
					Price: stripe.String(req.PriceID),
				},
			},
			ProrationBehavior: stripe.String("create_prorations"),
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: fmt.Sprintf("Failed to downgrade subscription: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Successfully downgraded to free plan",
		})
	}

	// For paid plans, create a subscription modification with payment
	updatedSubscription, err := subscription.Update(stripeSubscription.ID, &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(stripeSubscription.Items.Data[0].ID),
				Price: stripe.String(req.PriceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
		ProrationBehavior: stripe.String("create_prorations"),
		Expand: []*string{
			stripe.String("latest_invoice.payment_intent"),
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to upgrade subscription: %v", err),
		})
	}

	// Return client secret for payment confirmation
	var clientSecret string
	if updatedSubscription.LatestInvoice != nil {
		if updatedSubscription.LatestInvoice.PaymentIntent != nil {
			clientSecret = updatedSubscription.LatestInvoice.PaymentIntent.ClientSecret
		} else if updatedSubscription.LatestInvoice.Status == "paid" {
			// Subscription already paid/updated, no payment confirmation needed
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success":        true,
				"message":        "Subscription updated successfully",
				"subscriptionId": updatedSubscription.ID,
			})
		}
	}

	if clientSecret == "" {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "payment_failed",
			Message: "Failed to create payment intent for subscription upgrade",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"clientSecret":   clientSecret,
		"subscriptionId": updatedSubscription.ID,
	})
}

// CancelSubscription cancels the user's subscription
// POST /api/v1/billing/cancel-subscription
func (h *DefaultHandler) CancelSubscription(c echo.Context) error {
	// Resolve current user
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	// Get user
	uRepo := user.NewDefaultRepository(h.db)
	existingUser, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	// Get active subscription
	subRepo := stripeLib.NewSubscriptionsRepository(h.db)
	subs, err := subRepo.ListByUserID(c.Request().Context(), existingUser.ID.String(), 10, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to get subscriptions",
		})
	}

	if len(subs) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "No active subscription found",
		})
	}

	// Set Stripe API key
	stripe.Key = h.stripeSecretKey
	if stripe.Key == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "Stripe not configured",
		})
	}

	// Cancel subscription at period end using Update API
	_, err = subscription.Update(subs[0].StripeSubscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to cancel subscription: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription canceled successfully",
	})
}
