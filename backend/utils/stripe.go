package utils

import (
	"backend/configs"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v80"
	portalSession "github.com/stripe/stripe-go/v80/billingportal/session"
	"github.com/stripe/stripe-go/v80/checkout/session"
	"github.com/stripe/stripe-go/v80/coupon"
	"github.com/stripe/stripe-go/v80/customer"
	"github.com/stripe/stripe-go/v80/invoice"
	"github.com/stripe/stripe-go/v80/invoiceitem"
	"github.com/stripe/stripe-go/v80/paymentintent"
	"github.com/stripe/stripe-go/v80/paymentmethod"
	"github.com/stripe/stripe-go/v80/price"
	"github.com/stripe/stripe-go/v80/product"
	"github.com/stripe/stripe-go/v80/subscription"
	"github.com/stripe/stripe-go/v80/webhook"
)

// request payload
type checkoutRequest struct {
	SubscriptionPlanId string `json:"subscriptionPlanId"`
}

// auto reload request payload
type AutoReloadRequest struct {
	AutoReload     bool   `json:"autoReload"`
	AutoReloadSlug string `json:"autoReloadSlug"`
}

// simplified structs for DB rows
type User struct {
	UserID     string
	OwnerEmail string
	OwnerName  string
	OrgID      string
	OrgName    string
}

type OrgProfile struct {
	Slug           string
	StripeID       sql.NullString
	SubscriptionID sql.NullString
}

type PlanPriceResult struct {
	Price           int `json:"price"`
	Credit          int `json:"credit"`
	BalanceCredit   int `json:"balanceCredit"`
	Discount        int `json:"discount"`
	DiscountedPrice int `json:"discountedPrice"`
}

type SubscriptionPlan struct {
	ID                 string
	Name               string
	Price              float64
	CustomerType       string // "INDIVIDUAL" | "ENTERPRISE"
	BillingCycle       string // "month" | "quarter" | "year"
	ValueScore         float64
	BillingType        string
	Credits            int
	Details            json.RawMessage
	ConcurrentSessions int
	MaxSessionDuration int
}

type StripePlan struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	DisplayName        string          `json:"displayName"`
	Description        string          `json:"description"`
	UserType           string          `json:"userType"`
	BillingCycle       string          `json:"billingCycle"`
	Price              float64         `json:"price"`
	Credits            int             `json:"credits"`
	BillingType        string          `json:"billingType"`
	Slug               string          `json:"slug"`
	ValueScore         float64         `json:"valueScore"`
	ConcurrentSessions int             `json:"concurrentSessions"`
	MaxSessionDuration int             `json:"maxSessionDuration"`
	Details            json.RawMessage `json:"details"`
}

type License struct {
	ID                 string
	SubscriptionPlanId string
	PurchaseType       string
	SubscriptionId     string
	CreatedAt          time.Time
	BillingCycle       string
	BalanceCredit      int
	TotalCredit        int
	Price              float64
	StartDate          time.Time
	EndDate            time.Time
	Name               string
	DisplayName        string
}

type PurchaseLog struct {
	ID               string          `json:"id"`
	OrganizationID   string          `json:"organizationId"`
	CreatedAt        time.Time       `json:"createdAt"`
	CreditValue      int             `json:"creditValue"`
	MetaData         json.RawMessage `json:"metaData"`
	PaymentID        string          `json:"paymentId"`
	UserID           string          `json:"userId"`
	Price            float64         `json:"price"`
	SubscriptionID   string          `json:"subscriptionId"`
	Slug             string          `json:"slug"`
	CurrentPeriodEnd *time.Time      `json:"currentPeriodEnd"`
	InvoiceURL       string          `json:"invoiceURL"`
	Name             string          `json:"name"`
	DisplayName      string          `json:"displayName"`
}

type StripeCreditLimit struct {
	ID                 string     `json:"id"`
	OrganizationID     string     `json:"organizationId"`
	ConcurrentSessions int        `json:"concurrentSessions"`
	TotalCredit        int        `json:"totalCredit"`
	BalanceCredit      int        `json:"balanceCredit"`
	PurchasedCredit    int        `json:"purchasedCredit"`
	CreatedAt          *time.Time `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
	MaxSessionDuration int        `json:"maxSessionDuration"`
	SubscriptionID     string     `json:"subscriptionId"`
	StripeID           string     `json:"stripeId"`
	Slug               string     `json:"slug"`
	CancelledAt        *time.Time `json:"cancelledAt"`
	BillingCycle       string     `json:"billingCycle"`
	StartsAt           *time.Time `json:"startsAt"`
	ExpiresAt          *time.Time `json:"expiresAt"`
	AutoReload         bool       `json:"autoReload"`
	AutoReloadSlug     string     `json:"autoReloadSlug"`
}

func FormatWithCommas(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// helper: write json
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func GetUserFromRequest(r *http.Request) (*User, error) {
	keyid := r.Header.Get("X-API-Key")

	if keyid == "" {
		return nil, errors.New("unauthenticated")
	}
	var org struct {
		Org_ID    string
		Owner     string
		Org_Name  string
		User_ID   string
		OwnerName string
	}
	query := `
       SELECT DISTINCT
			ak.user_id AS user_id,
			ws.organization_id AS organization_id,
			org.owner AS owner_email,
			org.name AS org_name,
			u.first_name as owner_name
		FROM api_keys ak
		JOIN workspaces ws ON ws.id = ak.workspace_id
		JOIN organizations org ON org.id = ws.organization_id
		JOIN users u ON u.email = org.owner
		where ak.key_hash = $1`

	err := DB.QueryRow(query, keyid).Scan(&org.User_ID, &org.Org_ID, &org.Owner, &org.Org_Name, &org.OwnerName)
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid API key")
	}
	if err != nil {
		log.Printf("error fetching organization from api_key_id: %v", err)
		return nil, errors.New("failed to authenticate")
	}
	return &User{UserID: org.User_ID, OwnerEmail: org.Owner, OrgID: org.Org_ID, OrgName: org.Org_Name, OwnerName: org.OwnerName}, nil
}

func CreateAutoReloadCredits() {
	log.Println("Starting Auto Reload credit job...")

	AutoReloadCredits()

	log.Println("Auto Reload credit job completed.")
}

func AutoReloadCredits() {
	query := `
		SELECT id, organization_id, stripe_id, subscription_id, auto_reload_slug
		FROM credit_limits
		WHERE cancelled_at IS NULL
		  AND balance_credit+purchased_credit < 2000
		  AND auto_reload = true
		  AND auto_reload_slug IS NOT NULL
		  AND slug != 'free_ente_mont'
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Println("❌ Failed to query active subscriptions with less credits:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var orgID string
		var stripe_id string
		var subscription_id string
		var auto_credit_slug string

		if err := rows.Scan(&id, &orgID, &stripe_id, &subscription_id, &auto_credit_slug); err != nil {
			log.Println("❌ Failed to scan row:", err)
			continue
		}

		// fetch existing license
		var existingLicense License
		row := DB.QueryRow(`
        SELECT pl.id, pl.slug, pl.purchase_type, pl.subscription_id, pl.created_at, pp.billing_cycle,
               cl.balance_credit, cl.total_credit, pl.price, pl.start_date, pl.end_date, pp.name, pp.display_name
        FROM purchase_logs AS pl
		JOIN purchase_plans AS pp ON pp.slug = pl.slug
		JOIN credit_limits AS cl ON cl.organization_id = pl.organization_id
        WHERE pl.subscription_id = $1 
          AND pl.organization_id = $2 
          AND pl.purchase_type = 'subscription'
    `, subscription_id, orgID)

		if err := row.Scan(
			&existingLicense.ID,
			&existingLicense.SubscriptionPlanId,
			&existingLicense.PurchaseType,
			&existingLicense.SubscriptionId,
			&existingLicense.CreatedAt,
			&existingLicense.BillingCycle,
			&existingLicense.BalanceCredit,
			&existingLicense.TotalCredit,
			&existingLicense.Price,
			&existingLicense.StartDate,
			&existingLicense.EndDate,
			&existingLicense.Name,
			&existingLicense.DisplayName,
		); err != nil {
			return
		}

		// load existing plan
		var existingPlan SubscriptionPlan
		row = DB.QueryRow(`
        SELECT slug, name, price, user_type, billing_cycle, value_score, billing_type, credits,
               concurrent_sessions, max_session_duration, details
        FROM purchase_plans
        WHERE slug = $1
    `, existingLicense.SubscriptionPlanId)

		if err := row.Scan(
			&existingPlan.ID, &existingPlan.Name, &existingPlan.Price,
			&existingPlan.CustomerType, &existingPlan.BillingCycle, &existingPlan.ValueScore,
			&existingPlan.BillingType, &existingPlan.Credits, &existingPlan.ConcurrentSessions,
			&existingPlan.MaxSessionDuration, &existingPlan.Details,
		); err != nil {
			return
		}

		// load new plan
		var newPlan SubscriptionPlan
		row = DB.QueryRow(`
        SELECT slug, name, price, user_type, billing_cycle, value_score, billing_type, credits,
               concurrent_sessions, max_session_duration, details
        FROM purchase_plans
        WHERE slug = $1
    `, auto_credit_slug)

		if err := row.Scan(
			&newPlan.ID, &newPlan.Name, &newPlan.Price, &newPlan.CustomerType, &newPlan.BillingCycle,
			&newPlan.ValueScore, &newPlan.BillingType, &newPlan.Credits, &newPlan.ConcurrentSessions,
			&newPlan.MaxSessionDuration, &newPlan.Details,
		); err != nil {
			return
		}
		var price int = 0
		// one time plans
		if newPlan.BillingType == "one-time" {
			PriceResult, err := CreditRateForEnterprise(existingPlan, newPlan)
			if err != nil {
				return
			}

			price = int(math.Round(PriceResult.TargetPlanPrice))
		}

		if price == 0 {
			continue
		}

		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

		inv, err := invoice.New(&stripe.InvoiceParams{
			Customer:         stripe.String(stripe_id),
			AutoAdvance:      stripe.Bool(false), // IMPORTANT
			CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically)),
			Metadata: map[string]string{
				"userId":             "admin-auto-reload",
				"userName":           "Trugen Auto Reload",
				"orgId":              orgID,
				"subscriptionPlanId": newPlan.ID,
				"subscriptionName":   newPlan.Name,
				"concurrentSessions": strconv.Itoa(newPlan.ConcurrentSessions),
				"maxSessionDuration": strconv.Itoa(newPlan.MaxSessionDuration),
				"type":               "one-time-payment",
				"userType":           "Enterprise",
				"billingCycle":       "one-time",
				"creditAmount":       strconv.Itoa(newPlan.Credits),
				"price":              strconv.Itoa(price),
			},
		})
		if err != nil {
			log.Printf("❌ Invoice creation failed: %v\n", err)
			continue
		}

		_, err = invoiceitem.New(&stripe.InvoiceItemParams{
			Customer: stripe.String(stripe_id),
			Invoice:  stripe.String(inv.ID), // 🔥 THIS CONNECTS THEM
			Amount:   stripe.Int64(int64(price * 100)),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
			Description: stripe.String(
				fmt.Sprintf("Trugen Auto Reload Credits (%d) — Amount Paid: $%.2f", newPlan.Credits, float64(price)),
			),
		})
		if err != nil {
			log.Printf("❌ InvoiceItem creation failed: %v\n", err)
			continue
		}

		_, err = invoice.FinalizeInvoice(inv.ID, nil)
		if err != nil {
			log.Printf("❌ Invoice finalize failed: %v\n", err)
			continue
		}

		_, err = invoice.Pay(inv.ID, &stripe.InvoicePayParams{
			OffSession: stripe.Bool(true),
		})
		if err != nil {
			log.Printf("❌ Invoice payment failed: %v\n", err)
			continue
		}
		log.Println("✔️ Auto Reload Credits purchased for org:", orgID)
	}
}

// ResetMonthlyCredits Cron job runs every day night
func ResetMonthlyCredits() {
	log.Println("Starting monthly credit reset job...")

	resetActiveSubscriptions()
	resetCancelledSubscriptions()

	log.Println("Monthly credit reset job completed.")
}

// -------------------------------------------------------------------
// Reset active subscriptions that have expired
// -------------------------------------------------------------------
func resetActiveSubscriptions() {
	query := `
		SELECT id, organization_id, total_credit
		FROM credit_limits
		WHERE cancelled_at IS NULL
		  AND current_period_end <= NOW();
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Println("❌ Failed to query active subscriptions:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var orgID string
		var totalCredit int

		if err := rows.Scan(&id, &orgID, &totalCredit); err != nil {
			log.Println("❌ Failed to scan row:", err)
			continue
		}

		err := updateActiveCredit(id, totalCredit)
		if err != nil {
			log.Printf("❌ Update failed for org=%s id=%s: %v\n", orgID, id, err)
			continue
		}

		log.Println("✔️ Credits reset for active org:", orgID)
	}
}

func updateActiveCredit(id string, totalCredit int) error {
	_, err := DB.Exec(`
		UPDATE credit_limits 
		SET 
			balance_credit = $1,
			current_period_end = NOW() + INTERVAL '1 month',
			updated_at = NOW()
		WHERE id = $2
	`, totalCredit, id)

	return err
}

// -------------------------------------------------------------------
// Reset CANCELLED subscriptions back to free tier
// -------------------------------------------------------------------
func resetCancelledSubscriptions() {
	query := `
		SELECT id, organization_id
		FROM credit_limits
		WHERE cancelled_at IS NOT NULL
		  AND subscription_id IS NOT NULL
		  AND current_period_end <= NOW();
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Println("❌ Failed to query cancelled subscriptions:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var orgID string

		if err := rows.Scan(&id, &orgID); err != nil {
			log.Println("❌ Failed to scan row:", err)
			continue
		}

		err := downgradeToFreePlan(id)
		if err != nil {
			log.Printf("❌ Downgrade failed for org=%s id=%s: %v\n", orgID, id, err)
			continue
		}

		log.Println("✔️ Downgraded to free plan:", orgID)
	}
}

func downgradeToFreePlan(id string) error {
	_, err := DB.Exec(`
		UPDATE credit_limits 
		SET 
			concurrent_sessions = 1,
			max_session_duration = 5,
			balance_credit = 1500,
			total_credit = 1500,
			slug = 'free_ente_mont',
			subscription_id = NULL,
			stripe_id = NULL,
			cancelled_at = NULL,
			current_period_end = NOW() + INTERVAL '1 month',
			updated_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

// billingCycle -> intervalCount mapping
func billingCycleToIntervalCount(billingCycle string) (int64, error) {
	switch strings.ToLower(billingCycle) {
	case "year":
		return 12, nil
	case "month":
		return 1, nil
	case "quarter":
		return 3, nil
	default:
		return 0, fmt.Errorf("invalid billing cycle: %s", billingCycle)
	}
}

type ProRateResult struct {
	TargetPlanPrice float64
	Credit          int
}

func CreditRateForEnterprise(
	existingPlan SubscriptionPlan,
	newPlan SubscriptionPlan,
) (ProRateResult, error) {
	return ProRateResult{
		TargetPlanPrice: float64(newPlan.Credits/50) * existingPlan.ValueScore,
		Credit:          newPlan.Credits,
	}, nil
}

func ProRateAmountForEnterprise(
	existingPlan SubscriptionPlan,
	newPlan SubscriptionPlan,
) (ProRateResult, error) {
	return ProRateResult{
		TargetPlanPrice: newPlan.Price,
		Credit:          newPlan.Credits,
	}, nil
}

func HandleCreateStripeCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stripe.Key = configs.GetEnv("STRIPE_SECRET_KEY")
	ctx := r.Context()

	// decode request
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.SubscriptionPlanId == "" {
		WriteError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// get user (auth)
	user, err := GetUserFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	log.Printf("OrgID: %v", user.OrgID)

	// load user profile
	var profile OrgProfile
	row := DB.QueryRowContext(ctx, `
		SELECT 
			COALESCE(stripe_id, '') AS stripe, COALESCE(subscription_id, '') AS sub, COALESCE(slug, '') AS slug
		FROM credit_limits
		where organization_id = $1
	`, user.OrgID)
	if err := row.Scan(&profile.StripeID, &profile.SubscriptionID, &profile.Slug); err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Org profile not found")
			return
		}
		log.Printf("DB error loading org profile: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to load org profile")
		return
	}

	tx, err := DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// create stripe customer if not exists
	if !profile.StripeID.Valid || profile.StripeID.String == "" {
		if user.OwnerEmail == "" {
			WriteError(w, http.StatusBadRequest, "Email required to create Stripe customer")
			return
		}

		custParams := &stripe.CustomerParams{
			Email: stripe.String(user.OwnerEmail),
			Name:  stripe.String(user.OrgName),
		}
		custParams.Metadata = map[string]string{"userId": user.UserID, "orgId": user.OrgID, "orgName": user.OrgName}

		cus, err := customer.New(custParams)
		if err != nil {
			log.Printf("Stripe customer create error: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to create Stripe customer")
			return
		}

		// persist stripe id
		_, err = tx.ExecContext(ctx, `
			UPDATE credit_limits SET stripe_id = $1, updated_at = CURRENT_TIMESTAMP WHERE organization_id = $2
		`, cus.ID, user.OrgID)
		if err != nil {
			log.Printf("DB error updating stripe id: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to save stripe customer")
			return
		}
		profile.StripeID.Valid = true
		profile.StripeID.String = cus.ID
	}

	var billingCycle string

	row = tx.QueryRowContext(ctx, `
			SELECT billing_cycle
			FROM purchase_plans
			WHERE slug = $1
		`, req.SubscriptionPlanId)
	if err := row.Scan(&billingCycle); err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Not a valid purchase plan")
			return
		}
		log.Printf("DB error loading new plan: %v", err)
		WriteError(w, http.StatusInternalServerError, "Not a valid purchase plan")
		return
	}

	// mapping billing cycle
	intervalCount, _ := billingCycleToIntervalCount(billingCycle)

	// If user has existing license -> handle proration / transition
	if profile.SubscriptionID.Valid && profile.SubscriptionID.String != "" {
		// fetch existing license
		var existingLicense License
		row = tx.QueryRowContext(ctx, `
			SELECT pl.id, pl.slug, pl.purchase_type, pl.subscription_id, pp.name, pp.display_name
			FROM purchase_logs AS pl
			JOIN purchase_plans AS pp ON pp.slug = pl.slug
			WHERE pl.subscription_id = $1 and pl.organization_id = $2 and pl.purchase_type = 'subscription'
		`, profile.SubscriptionID.String, user.OrgID)
		if err := row.Scan(&existingLicense.ID, &existingLicense.SubscriptionPlanId, &existingLicense.PurchaseType, &existingLicense.SubscriptionId, &existingLicense.Name, &existingLicense.DisplayName); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Existing license not found for id: %s", profile.SubscriptionID.String)
				WriteError(w, http.StatusNotFound, "Not a valid subscription")
				return
			}
			log.Printf("DB error loading existing license: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to load license")
			return
		}

		// load existing plan
		var existingPlan SubscriptionPlan
		row = tx.QueryRowContext(ctx, `
			SELECT slug, display_name, price, user_type, billing_cycle, value_score, billing_type, credits, concurrent_sessions, max_session_duration, details
			FROM purchase_plans
			WHERE slug = $1
		`, existingLicense.SubscriptionPlanId)
		if err := row.Scan(&existingPlan.ID, &existingPlan.Name, &existingPlan.Price, &existingPlan.CustomerType, &existingPlan.BillingCycle, &existingPlan.ValueScore, &existingPlan.BillingType, &existingPlan.Credits, &existingPlan.ConcurrentSessions, &existingPlan.MaxSessionDuration, &existingPlan.Details); err != nil {
			log.Printf("DB error loading existing plan: %v", err)
			WriteError(w, http.StatusInternalServerError, "Not a valid purchase plan")
			return
		}

		// load new plan
		var newPlan SubscriptionPlan
		row = tx.QueryRowContext(ctx, `
			SELECT slug, display_name, price, user_type, billing_cycle, value_score, billing_type, credits, concurrent_sessions, max_session_duration, details
			FROM purchase_plans
			WHERE slug = $1
		`, req.SubscriptionPlanId)
		if err := row.Scan(&newPlan.ID, &newPlan.Name, &newPlan.Price, &newPlan.CustomerType, &newPlan.BillingCycle, &newPlan.ValueScore, &newPlan.BillingType, &newPlan.Credits, &newPlan.ConcurrentSessions, &newPlan.MaxSessionDuration, &newPlan.Details); err != nil {
			if err == sql.ErrNoRows {
				WriteError(w, http.StatusNotFound, "Not a valid purchase plan")
				return
			}
			log.Printf("DB error loading new plan: %v", err)
			WriteError(w, http.StatusInternalServerError, "Not a valid purchase plan")
			return
		}

		// transition rules (month/quarter/year)
		canTransition := func(from, to string) bool {
			from = strings.ToLower(from)
			to = strings.ToLower(to)
			allowed := map[string][]string{
				"month":   {"month", "quarter", "year"},
				"quarter": {"quarter", "year"},
				"year":    {"year"},
			}
			for _, v := range allowed[from] {
				if v == to {
					return true
				}
			}
			return false
		}

		// transition rules (Free/Essential/Pro)
		canTransition_plan := func(from, to string) bool {
			from = strings.ToLower(from)
			to = strings.ToLower(to)
			allowed := map[string][]string{
				"free": {"free", "pro"},
				"pro":  {"pro"},
			}
			for _, v := range allowed[from] {
				if v == to {
					return true
				}
			}
			return false
		}

		if newPlan.BillingType == "one-time" {

			PriceResult, err := CreditRateForEnterprise(existingPlan, newPlan)
			if err != nil {
				log.Printf("Error prorating enterprise: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to calculate prorated amount")
				return
			}

			// create stripe product
			prodParams := &stripe.ProductParams{
				Name: stripe.String(fmt.Sprintf("%s Purchase", newPlan.Name)),
			}
			prodParams.AddMetadata("subscriptionPlanId", newPlan.ID)
			prodParams.AddMetadata("customerType", newPlan.CustomerType)

			prod, err := product.New(prodParams)
			if err != nil {
				log.Printf("Stripe product create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create product")
				return
			}

			priceParams := &stripe.PriceParams{
				Product:    stripe.String(prod.ID),
				UnitAmount: stripe.Int64(int64(PriceResult.TargetPlanPrice * 100)),
				Currency:   stripe.String("usd"),
			}

			p, err := price.New(priceParams)
			if err != nil {
				log.Printf("Stripe price create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create price")
				return
			}

			sessParams := &stripe.CheckoutSessionParams{
				Mode:     stripe.String(string(stripe.CheckoutSessionModePayment)),
				Customer: stripe.String(profile.StripeID.String),
				LineItems: []*stripe.CheckoutSessionLineItemParams{
					{
						Price:    stripe.String(p.ID),
						Quantity: stripe.Int64(1),
					},
				},
				SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", os.Getenv("STRIPE_PAYMENT_SUCCESS_URL"))),
				CancelURL:  stripe.String(os.Getenv("STRIPE_PAYMENT_CANCEL_URL")),
				InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
					Enabled: stripe.Bool(true),
				},
			}

			sessParams.AddMetadata("userId", user.UserID)
			sessParams.AddMetadata("userName", user.OwnerName)
			sessParams.AddMetadata("orgId", user.OrgID)
			sessParams.AddMetadata("subscriptionPlanId", newPlan.ID)
			sessParams.AddMetadata("subscriptionName", newPlan.Name)
			sessParams.AddMetadata("concurrentSessions", strconv.Itoa(newPlan.ConcurrentSessions))
			sessParams.AddMetadata("maxSessionDuration", strconv.Itoa(newPlan.MaxSessionDuration))
			sessParams.AddMetadata("type", "one-time-payment")
			sessParams.AddMetadata("userType", "Enterprise")
			sessParams.AddMetadata("billingCycle", "one-time")
			sessParams.AddMetadata("creditAmount", strconv.Itoa(PriceResult.Credit))

			sess, err := session.New(sessParams)
			if err != nil {
				log.Printf("Stripe session create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create checkout session")
				return
			}

			_ = tx.Commit()
			WriteJSON(w, http.StatusOK, map[string]string{"checkoutUrl": sess.URL})
			return
		}
		if existingPlan.ID == newPlan.ID {
			msg := fmt.Sprintf("Old and New plan (%s) are same.", existingPlan.Name)
			log.Print(msg)
			WriteError(w, http.StatusBadRequest, msg)
			return
		} else {
			if !canTransition(existingPlan.BillingCycle, newPlan.BillingCycle) {
				msg := fmt.Sprintf("Downgrade from %s to %s is not allowed", existingPlan.BillingCycle, newPlan.BillingCycle)
				log.Print(msg)
				WriteError(w, http.StatusBadRequest, msg)
				return
			}

			if !canTransition_plan(existingPlan.Name, newPlan.Name) {
				msg := fmt.Sprintf("Downgrade from %s to %s is not allowed", existingPlan.Name, newPlan.Name)
				log.Print(msg)
				WriteError(w, http.StatusBadRequest, msg)
				return
			}

			proRateResult, err := ProRateAmountForEnterprise(existingPlan, newPlan)
			if err != nil {
				log.Printf("Error prorating enterprise: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to calculate prorated amount")
				return
			}

			// create stripe product
			prodParams := &stripe.ProductParams{
				Name: stripe.String(fmt.Sprintf("%s Subscription", newPlan.Name)),
			}
			prodParams.AddMetadata("subscriptionPlanId", newPlan.ID)
			prodParams.AddMetadata("customerType", newPlan.CustomerType)

			prod, err := product.New(prodParams)
			if err != nil {
				log.Printf("Stripe product create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create product")
				return
			}

			priceParams := &stripe.PriceParams{
				Product:    stripe.String(prod.ID),
				UnitAmount: stripe.Int64(int64(proRateResult.TargetPlanPrice * 100)),
				Currency:   stripe.String("usd"),
				Recurring: &stripe.PriceRecurringParams{
					Interval:      stripe.String("month"),
					IntervalCount: stripe.Int64(intervalCount),
				},
			}
			p, err := price.New(priceParams)
			if err != nil {
				log.Printf("Stripe price create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create price")
				return
			}

			sessParams := &stripe.CheckoutSessionParams{
				Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
				Customer: stripe.String(profile.StripeID.String),
				LineItems: []*stripe.CheckoutSessionLineItemParams{
					{
						Price:    stripe.String(p.ID),
						Quantity: stripe.Int64(1),
					},
				},
				SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", os.Getenv("STRIPE_PAYMENT_SUCCESS_URL"))),
				CancelURL:  stripe.String(os.Getenv("STRIPE_PAYMENT_CANCEL_URL")),
			}
			sessParams.AddMetadata("userId", user.UserID)
			sessParams.AddMetadata("userName", user.OwnerName)
			sessParams.AddMetadata("orgId", user.OrgID)
			sessParams.AddMetadata("subscriptionPlanId", newPlan.ID)
			sessParams.AddMetadata("subscriptionName", newPlan.Name)
			sessParams.AddMetadata("concurrentSessions", strconv.Itoa(newPlan.ConcurrentSessions))
			sessParams.AddMetadata("maxSessionDuration", strconv.Itoa(newPlan.MaxSessionDuration))
			sessParams.AddMetadata("oldSubscriptionPlanId", existingPlan.ID)
			sessParams.AddMetadata("oldSubscriptionId", existingLicense.SubscriptionId)
			sessParams.AddMetadata("type", "subscription")
			sessParams.AddMetadata("userType", "Enterprise")
			sessParams.AddMetadata("billingCycle", billingCycle)
			sessParams.AddMetadata("creditAmount", strconv.Itoa(proRateResult.Credit))

			result, err := CalculatePlanPrice(ctx, user.OrgID, newPlan.ID)
			if err != nil {
				log.Println("calc plan err:", err)
			}

			discountAmount := result.Discount
			//need to call HandleGetPlanPrice() to get discount amount
			if discountAmount > 0 {
				// 👈 APPLY COUPON HERE
				couponParams := &stripe.CouponParams{
					AmountOff: stripe.Int64(int64(discountAmount * 100)),
					Currency:  stripe.String("usd"),
					Duration:  stripe.String("once"),
					Name:      stripe.String("Available credit from your current plan"),
				}

				coupon, err := coupon.New(couponParams)
				if err != nil {
					log.Printf("Stripe coupon error: %v", err)
				}
				if coupon.ID != "" {
					sessParams.Discounts = []*stripe.CheckoutSessionDiscountParams{
						{
							Coupon: stripe.String(coupon.ID),
						},
					}
				}
			}

			sess, err := session.New(sessParams)
			if err != nil {
				log.Printf("Stripe session create error: %v", err)
				WriteError(w, http.StatusInternalServerError, "Failed to create checkout session")
				return
			}

			_ = tx.Commit()
			WriteJSON(w, http.StatusOK, map[string]string{"checkoutUrl": sess.URL})
			return
		}
	}

	// No existing license or fallback: create a new subscription / one-time payment
	var plan SubscriptionPlan
	row = tx.QueryRowContext(ctx, `
		SELECT slug, display_name, price, user_type, billing_cycle, value_score, billing_type, credits, concurrent_sessions, max_session_duration, details
			FROM purchase_plans
			WHERE slug = $1
	`, req.SubscriptionPlanId)
	if err := row.Scan(&plan.ID, &plan.Name, &plan.Price, &plan.CustomerType, &plan.BillingCycle, &plan.ValueScore, &plan.BillingType, &plan.Credits, &plan.ConcurrentSessions, &plan.MaxSessionDuration, &plan.Details); err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Subscription plan not found")
			return
		}
		log.Printf("DB error loading plan: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to load subscription plan")
		return
	}

	// compute total price in cents
	basePrice := plan.Price
	totalPriceCents := float32(basePrice * 100)

	if totalPriceCents == 0 {
		WriteError(w, http.StatusBadRequest, "Total price cannot be zero")
		return
	}

	planName := fmt.Sprintf("%s Subscription", plan.Name)

	// create product
	prodParams := &stripe.ProductParams{
		Name: stripe.String(planName),
	}
	prodParams.AddMetadata("subscriptionPlanId", plan.ID)
	prodParams.AddMetadata("customerType", plan.CustomerType)

	prod, err := product.New(prodParams)
	if err != nil {
		log.Printf("Stripe product create error: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	// recurring price
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(prod.ID),
		UnitAmount: stripe.Int64(int64(totalPriceCents)),
		Currency:   stripe.String("usd"),
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String("month"),
			IntervalCount: stripe.Int64(intervalCount),
		},
	}
	p, err := price.New(priceParams)
	if err != nil {
		log.Printf("Stripe price create error: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to create price")
		return
	}

	// create checkout session
	sessParams := &stripe.CheckoutSessionParams{
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		Customer: stripe.String(profile.StripeID.String),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(p.ID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", os.Getenv("STRIPE_PAYMENT_SUCCESS_URL"))),
		CancelURL:  stripe.String(os.Getenv("STRIPE_PAYMENT_CANCEL_URL")),
	}
	sessParams.AddMetadata("userId", user.UserID)
	sessParams.AddMetadata("userName", user.OwnerName)
	sessParams.AddMetadata("orgId", user.OrgID)
	sessParams.AddMetadata("subscriptionPlanId", plan.ID)
	sessParams.AddMetadata("subscriptionName", plan.Name)
	sessParams.AddMetadata("concurrentSessions", strconv.Itoa(plan.ConcurrentSessions))
	sessParams.AddMetadata("maxSessionDuration", strconv.Itoa(plan.MaxSessionDuration))
	sessParams.AddMetadata("type", "subscription")
	sessParams.AddMetadata("userType", "Enterprise")
	sessParams.AddMetadata("billingCycle", billingCycle)
	sessParams.AddMetadata("creditAmount", strconv.Itoa(plan.Credits))

	sess, err := session.New(sessParams)
	if err != nil {
		log.Printf("Stripe session create error: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"checkoutUrl": sess.URL})
}

func HandleGetPlanPrice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.SubscriptionPlanId == "" {
		WriteError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	user, err := GetUserFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := CalculatePlanPrice(r.Context(), user.OrgID, req.SubscriptionPlanId)
	if err != nil {
		log.Printf("calc price error: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to calculate plan price")
		return
	}

	WriteJSON(w, http.StatusOK, res)
}

func CalculatePlanPrice(
	ctx context.Context,
	orgID string,
	newPlanSlug string,
) (*PlanPriceResult, error) {

	var result PlanPriceResult

	// load user profile
	var profile OrgProfile
	row := DB.QueryRowContext(ctx, `
		SELECT COALESCE(stripe_id,''), COALESCE(subscription_id,''), COALESCE(slug,'')
		FROM credit_limits
		WHERE organization_id = $1
	`, orgID)
	if err := row.Scan(&profile.StripeID, &profile.SubscriptionID, &profile.Slug); err != nil {
		return nil, fmt.Errorf("load org profile: %w", err)
	}

	// load new plan billing cycle
	var billingCycle string
	row = DB.QueryRowContext(ctx, `
		SELECT billing_cycle
		FROM purchase_plans
		WHERE slug = $1
	`, newPlanSlug)
	if err := row.Scan(&billingCycle); err != nil {
		return nil, fmt.Errorf("load billing cycle: %w", err)
	}

	// If NO active subscription → return new plan price and details
	if !profile.SubscriptionID.Valid || profile.SubscriptionID.String == "" {
		// load new plan
		var newPlan SubscriptionPlan
		row = DB.QueryRowContext(ctx, `
        SELECT slug, name, price, user_type, billing_cycle, value_score, billing_type, credits,
               concurrent_sessions, max_session_duration, details
        FROM purchase_plans
        WHERE slug = $1
    `, newPlanSlug)

		if err := row.Scan(
			&newPlan.ID, &newPlan.Name, &newPlan.Price, &newPlan.CustomerType, &newPlan.BillingCycle,
			&newPlan.ValueScore, &newPlan.BillingType, &newPlan.Credits, &newPlan.ConcurrentSessions,
			&newPlan.MaxSessionDuration, &newPlan.Details,
		); err != nil {
			return nil, fmt.Errorf("load new plan: %w", err)
		}

		// one time plans
		if newPlan.BillingType == "one-time" {
			return &PlanPriceResult{}, nil
		} else {
			result.Price = int(math.Round(newPlan.Price))
			result.Credit = int(math.Round(float64(newPlan.Credits)))
			result.BalanceCredit = 0
			result.Discount = 0
			result.DiscountedPrice = 0
			return &result, nil
		}
	}

	// fetch existing license
	var existingLicense License
	row = DB.QueryRowContext(ctx, `
        SELECT pl.id, pl.slug, pl.purchase_type, pl.subscription_id, pl.created_at, pp.billing_cycle,
               cl.balance_credit, cl.total_credit, pl.price, pl.start_date, pl.end_date, pp.name, pp.display_name
        FROM purchase_logs AS pl
		JOIN purchase_plans AS pp ON pp.slug = pl.slug
		JOIN credit_limits AS cl ON cl.organization_id = pl.organization_id
        WHERE pl.subscription_id = $1 
          AND pl.organization_id = $2 
          AND pl.purchase_type = 'subscription'
    `, profile.SubscriptionID.String, orgID)

	if err := row.Scan(
		&existingLicense.ID,
		&existingLicense.SubscriptionPlanId,
		&existingLicense.PurchaseType,
		&existingLicense.SubscriptionId,
		&existingLicense.CreatedAt,
		&existingLicense.BillingCycle,
		&existingLicense.BalanceCredit,
		&existingLicense.TotalCredit,
		&existingLicense.Price,
		&existingLicense.StartDate,
		&existingLicense.EndDate,
		&existingLicense.Name,
		&existingLicense.DisplayName,
	); err != nil {
		return nil, fmt.Errorf("load existing license: %w", err)
	}

	// load existing plan
	var existingPlan SubscriptionPlan
	row = DB.QueryRowContext(ctx, `
        SELECT slug, name, price, user_type, billing_cycle, value_score, billing_type, credits,
               concurrent_sessions, max_session_duration, details
        FROM purchase_plans
        WHERE slug = $1
    `, existingLicense.SubscriptionPlanId)

	if err := row.Scan(
		&existingPlan.ID, &existingPlan.Name, &existingPlan.Price,
		&existingPlan.CustomerType, &existingPlan.BillingCycle, &existingPlan.ValueScore,
		&existingPlan.BillingType, &existingPlan.Credits, &existingPlan.ConcurrentSessions,
		&existingPlan.MaxSessionDuration, &existingPlan.Details,
	); err != nil {
		return nil, fmt.Errorf("load existing plan: %w", err)
	}

	// load new plan
	var newPlan SubscriptionPlan
	row = DB.QueryRowContext(ctx, `
        SELECT slug, name, price, user_type, billing_cycle, value_score, billing_type, credits,
               concurrent_sessions, max_session_duration, details
        FROM purchase_plans
        WHERE slug = $1
    `, newPlanSlug)

	if err := row.Scan(
		&newPlan.ID, &newPlan.Name, &newPlan.Price, &newPlan.CustomerType, &newPlan.BillingCycle,
		&newPlan.ValueScore, &newPlan.BillingType, &newPlan.Credits, &newPlan.ConcurrentSessions,
		&newPlan.MaxSessionDuration, &newPlan.Details,
	); err != nil {
		return nil, fmt.Errorf("load new plan: %w", err)
	}

	// one time plans
	if newPlan.BillingType == "one-time" {
		PriceResult, err := CreditRateForEnterprise(existingPlan, newPlan)
		if err != nil {
			return nil, fmt.Errorf("prorate enterprise: %w", err)
		}

		result.Price = int(math.Round(PriceResult.TargetPlanPrice))
		result.Credit = int(math.Round(float64(PriceResult.Credit)))
		result.BalanceCredit = 0
		result.Discount = 0
		result.DiscountedPrice = 0
		return &result, nil
	}

	// ------------------------------------------------------
	// Subscription → Pro-rate discount / credit calculation
	// ------------------------------------------------------

	start := existingLicense.StartDate
	end := existingLicense.EndDate

	totalDays := end.Sub(start).Hours() / 24
	if totalDays <= 0 {
		totalDays = 1
	}

	usedDays := time.Since(start).Hours() / 24
	remainingDays := totalDays - usedDays
	if remainingDays < 0 {
		remainingDays = 0
	}

	// Total credit
	totalCredit := float64(existingLicense.TotalCredit)
	if existingLicense.BillingCycle == "year" {
		totalCredit = totalCredit * 12
	}

	perDayCredit := totalCredit / totalDays
	earnedRemaining := remainingDays * perDayCredit
	actualRemaining := float64(existingLicense.BalanceCredit)

	remainingCredit := math.Min(actualRemaining, earnedRemaining)

	// discount value based on remaining credit
	discount := (remainingCredit / totalCredit) * existingLicense.Price

	if discount < 0 {
		discount = 0
	}
	if discount > existingLicense.Price {
		discount = math.Round(existingLicense.Price)
	}

	// assign result
	result.Price = int(math.Round(newPlan.Price))
	result.Credit = int(math.Round(float64(newPlan.Credits)))
	result.BalanceCredit = int(math.Round(remainingCredit*100) / 100)
	result.Discount = int(math.Round(discount*100) / 100)
	result.DiscountedPrice = int(math.Round(float64(result.Price - result.Discount)))
	return &result, nil
}

// GetAllPurchaseLogs retrieves all entries from purchase_logs
func GetAllPurchaseLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	keyid := r.Header.Get("X-API-Key")

	query := `
        SELECT DISTINCT
			spl.id,
			spl.organization_id,
			spl.created_at,
			credit_value,
			COALESCE(meta_data, '{}') as meta_data,
			COALESCE(payment_id, '') as payment_id,
			COALESCE(spl.user_id, '') as user_id,
			spl.price,
			subscription_id,
			spl.slug,
			pp.name,
			pp.display_name,
			COALESCE(spl.invoice_url, '') as invoice_url
		FROM purchase_logs spl
		JOIN purchase_plans pp ON pp.slug = spl.slug
		JOIN workspaces w ON spl.organization_id = w.organization_id           
		JOIN api_keys ak ON ak.workspace_id = w.id 
		WHERE ak.key_hash = $1
		ORDER BY spl.created_at DESC;
    `

	rows, err := DB.Query(query, keyid)
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch purchase logs: "+err.Error())
		return
	}
	defer rows.Close()

	var logs []PurchaseLog

	for rows.Next() {
		var (
			id, orgID, paymentID, userID, subscriptionID, slug, name, displayName, invoiceURL string
			createdAt                                                                         time.Time
			creditValue                                                                       int
			price                                                                             float64
			metaData                                                                          json.RawMessage
		)

		// Scan the DB row
		err := rows.Scan(
			&id,
			&orgID,
			&createdAt,
			&creditValue,
			&metaData,
			&paymentID,
			&userID,
			&price,
			&subscriptionID,
			&slug,
			&name,
			&displayName,
			&invoiceURL,
		)
		if err != nil {
			WriteInternalServerError(w, "Error scanning purchase log: "+err.Error())
			return
		}

		logs = append(logs, PurchaseLog{
			ID:             id,
			OrganizationID: orgID,
			CreatedAt:      createdAt,
			CreditValue:    creditValue,
			MetaData:       metaData,
			PaymentID:      paymentID,
			UserID:         userID,
			Price:          price,
			SubscriptionID: subscriptionID,
			Slug:           slug,
			Name:           name,
			DisplayName:    displayName,
			InvoiceURL:     invoiceURL,
		})
	}

	json.NewEncoder(w).Encode(logs)
}

func GetSubscriptionDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apiKeyHash := r.Header.Get("X-API-Key")

	if apiKeyHash == "" {
		WriteInternalServerError(w, "Missing X-API-Key header")
		return
	}

	query := `
        SELECT
            spl.id,
            spl.organization_id,
            spl.concurrent_sessions,
            spl.total_credit,
            spl.balance_credit,
            spl.purchased_credit,
            spl.created_at,
            spl.updated_at,
            spl.max_session_duration,
            COALESCE(spl.subscription_id, '') as subscription_id,
            COALESCE(spl.stripe_id, '') as stripe_id,
            spl.slug,
			spl.cancelled_at,
			pp.billing_cycle,
			pl.start_date,
			pl.end_date,
			spl.auto_reload,
			COALESCE(spl.auto_reload_slug, '') as auto_reload_slug
        FROM credit_limits spl
        JOIN workspaces w ON spl.organization_id = w.organization_id
        JOIN api_keys ak ON ak.workspace_id = w.id
		JOIN purchase_plans pp ON pp.slug = spl.slug
		LEFT JOIN purchase_logs pl ON pl.subscription_id = spl.subscription_id
        WHERE ak.key_hash = $1
        ORDER BY spl.created_at DESC
        LIMIT 1;
    `

	var scl StripeCreditLimit

	err := DB.QueryRow(query, apiKeyHash).Scan(
		&scl.ID,
		&scl.OrganizationID,
		&scl.ConcurrentSessions,
		&scl.TotalCredit,
		&scl.BalanceCredit,
		&scl.PurchasedCredit,
		&scl.CreatedAt,
		&scl.UpdatedAt,
		&scl.MaxSessionDuration,
		&scl.SubscriptionID,
		&scl.StripeID,
		&scl.Slug,
		&scl.CancelledAt,
		&scl.BillingCycle,
		&scl.StartsAt,
		&scl.ExpiresAt,
		&scl.AutoReload,
		&scl.AutoReloadSlug,
	)

	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(nil) // or {} if you prefer
		return
	}

	if err != nil {
		WriteInternalServerError(w, "Failed to fetch subscription details: "+err.Error())
		return
	}

	json.NewEncoder(w).Encode(scl)
}

// GetAllPurchasePlans retrieves all subscription plans
func GetAllPurchasePlans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `
        SELECT id, name, display_name, description, user_type, billing_cycle, 
               price, credits, billing_type, slug, concurrent_sessions, max_session_duration, details
        FROM public.purchase_plans
        ORDER BY value_score DESC
    `

	rows, err := DB.Query(query)
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch stripe plans"+err.Error())
		return
	}
	defer rows.Close()

	var plans []StripePlan

	for rows.Next() {
		var (
			id, name, displayName, description, userType, billingCycle string
			billingType, slug                                          string
			price                                                      float64
			credits, concurrentSessions, maxSessionDuration            int
			details                                                    json.RawMessage
		)

		err := rows.Scan(
			&id,
			&name,
			&displayName,
			&description,
			&userType,
			&billingCycle,
			&price,
			&credits,
			&billingType,
			&slug,
			&concurrentSessions,
			&maxSessionDuration,
			&details,
		)
		if err != nil {

			WriteInternalServerError(w, "error scanning subscription plan: "+err.Error())
			return
		}

		plans = append(plans, StripePlan{
			ID:                 id,
			Name:               name,
			DisplayName:        displayName,
			Description:        description,
			UserType:           userType,
			BillingCycle:       billingCycle,
			Price:              price,
			Credits:            credits,
			BillingType:        billingType,
			Slug:               slug,
			ConcurrentSessions: concurrentSessions,
			MaxSessionDuration: maxSessionDuration,
			Details:            details,
		})
	}
	json.NewEncoder(w).Encode(plans)
}

func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Verify signature
	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(
		body,
		sig,
		webhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Signature verification failed: %v", err), http.StatusBadRequest)
		return
	}

	// Handle different event types
	switch event.Type {

	case "checkout.session.completed":
		handleCheckoutCompleted(event)

	case "customer.subscription.updated":
		handleSubscriptionUpdated(event)

	case "invoice.paid":
		handleInvoicePaid(event)

	case "invoice.payment_succeeded":
		handlePaymentSuccess(event)

	case "invoice.payment_failed":
		handlePaymentFailure(event)

	}
	w.WriteHeader(http.StatusOK)
}

func handlePaymentSuccess(event stripe.Event) {
	stripe.Key = configs.GetEnv("STRIPE_SECRET_KEY")
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Println("unmarshal invoice:", err)
		return
	}

	orgID := extractOrgID(invoice)
	if orgID == "" {
		log.Println("orgId missing in payment success")
		return
	}

	const q = `
		UPDATE credit_limits
		SET payment_failed = FALSE,
		    error_message = NULL,
		    updated_at = NOW()
		WHERE organization_id = $1;
	`

	if _, err := DB.Exec(q, orgID); err != nil {
		log.Println("DB update error:", err)
		return
	}

	log.Println("✅ Payment success → cleared failure for org", orgID)
}

func handlePaymentFailure(event stripe.Event) {
	stripe.Key = configs.GetEnv("STRIPE_SECRET_KEY")
	var invEvent stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invEvent); err != nil {
		log.Println("unmarshal invoice:", err)
		return
	}

	// ✅ Use invoice package correctly
	inv, err := invoice.Get(
		invEvent.ID,
		&stripe.InvoiceParams{
			Expand: []*string{
				stripe.String("payment_intent"),
			},
		},
	)
	if err != nil {
		log.Println("invoice refetch failed:", err)
		return
	}

	errMsg := "Payment failed"
	if inv.PaymentIntent != nil {
		if lpe := inv.PaymentIntent.LastPaymentError; lpe != nil {

			switch {
			case lpe.Msg != "":
				errMsg = lpe.Msg

			case lpe.DeclineCode != "":
				errMsg = "Payment declined: " + string(lpe.DeclineCode)

			case lpe.Code != "":
				errMsg = "Payment error: " + string(lpe.Code)

			default:
				errMsg = "Payment failed"
			}
		}
	}

	orgID := extractOrgID(invEvent)
	if orgID == "" {
		log.Println("orgId missing in payment failure")
		return
	}

	const q = `
		UPDATE credit_limits
		SET payment_failed = TRUE,
		    error_message = $2,
		    updated_at = NOW()
		WHERE organization_id = $1;
	`

	if _, err := DB.Exec(q, orgID, errMsg); err != nil {
		log.Println("DB update error:", err)
		return
	}

	log.Println("❌ Payment failed → marked org", orgID)
}

func extractOrgID(inv stripe.Invoice) string {

	// 1️⃣ Invoice metadata
	if inv.Metadata != nil {
		if v := inv.Metadata["orgId"]; v != "" {
			return v
		}
	}

	// 2️⃣ Subscription metadata
	if inv.Subscription != nil {
		sub, err := subscription.Get(inv.Subscription.ID, nil)
		if err == nil && sub.Metadata != nil {
			if v := sub.Metadata["orgId"]; v != "" {
				return v
			}
		}
	}

	// 3️⃣ Customer metadata
	if inv.Customer != nil {
		cus, err := customer.Get(inv.Customer.ID, nil)
		if err == nil && cus.Metadata != nil {
			if v := cus.Metadata["orgId"]; v != "" {
				return v
			}
		}
	}

	return ""
}

func handleCheckoutCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Println("Failed to unmarshal checkout.session.completed:", err)
		return
	}

	if session.Subscription != nil {
		//🔑 Expand invoice properly
		if session.Invoice == nil {
			log.Println("⚠️ No invoice on checkout session")
			return
		}

		inv, err := invoice.Get(session.Invoice.ID, &stripe.InvoiceParams{
			Expand: []*string{
				stripe.String("payment_intent"),
				stripe.String("payment_intent.payment_method"),
			},
		})
		if err != nil {
			log.Println("❌ Failed to fetch invoice:", err)
			return
		}

		if inv.PaymentIntent == nil || inv.PaymentIntent.PaymentMethod == nil {
			log.Println("⚠️ No payment method on invoice payment intent")
			return
		}

		pmID := inv.PaymentIntent.PaymentMethod.ID
		customerID := session.Customer.ID
		subsID := session.Subscription.ID

		// 1️⃣ Attach PM to customer (REQUIRED)
		_, err = paymentmethod.Attach(pmID, &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(customerID),
		})
		if err != nil {
			log.Println("❌ Failed to attach payment method:", err)
			return
		}

		// 2️⃣ Set default PM on subscription
		_, err = subscription.Update(subsID, &stripe.SubscriptionParams{
			DefaultPaymentMethod: stripe.String(pmID),
		})
		if err != nil {
			log.Println("❌ Failed to update subscription default PM:", err)
			return
		}

		// 3️⃣ Set default PM on customer (future invoices)
		_, err = customer.Update(customerID, &stripe.CustomerParams{
			InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(pmID),
			},
		})
		if err != nil {
			log.Println("❌ Failed to update customer default PM:", err)
			return
		}

		log.Printf(
			"✅ Default PM set | pm=%s | sub=%s | customer=%s",
			pmID, subsID, customerID,
		)
	}

	log.Printf("Checkout Completed → Session: %s", session.ID)

	userId := session.Metadata["userId"]
	userName := session.Metadata["userName"]
	orgIdStr := session.Metadata["orgId"]
	orgId, err := uuid.Parse(orgIdStr)
	if err != nil {
		log.Println("Invalid orgId UUID:", err)
		return
	}

	price := float64(session.AmountTotal) / 100
	paymentID := ""
	subscriptionID := ""
	var start time.Time
	var end time.Time
	creditStr := session.Metadata["creditAmount"]
	slug := session.Metadata["subscriptionPlanId"]
	name := session.Metadata["subscriptionName"]
	purchaseType := session.Metadata["type"]
	concurrentSessions := session.Metadata["concurrentSessions"]
	maxSessionDuration := session.Metadata["maxSessionDuration"]
	invoiceURL := ""
	if session.PaymentIntent != nil {
		paymentID = session.PaymentIntent.ID
	}
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID

		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			log.Println("Failed to fetch subscription:", err)
			return
		}

		start = time.Unix(sub.CurrentPeriodStart, 0)
		end = time.Unix(sub.CurrentPeriodEnd, 0)

		log.Println("Period Start:", start)
		log.Println("Period End:", end)
	}

	invoiceURL, err = GetInvoicePDF(paymentID, subscriptionID)
	if err != nil {
		log.Println("Failed to get invoice PDF:", err)
	} else {
		log.Println("Invoice PDF URL:", invoiceURL)
	}

	// Parse credit
	credit := 0
	if creditStr != "" {
		credit, _ = strconv.Atoi(creditStr)
	}

	// Re-encode metadata
	metaJSON, _ := json.Marshal(session.Metadata)

	//------------------------------------------------------
	// BEGIN TRANSACTION
	//------------------------------------------------------
	tx, err := DB.Begin()
	if err != nil {
		log.Println("❌ Failed to begin transaction:", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	//------------------------------------------------------
	// 1️⃣ INSERT PURCHASE LOG
	//------------------------------------------------------
	insertPurchase := `
		INSERT INTO purchase_logs (
			id, organization_id, created_at, credit_value, meta_data,
			payment_id, user_id, price, subscription_id, slug, invoice_url, purchase_type, start_date, end_date
		)
		VALUES (
			gen_random_uuid(), $1, NOW(), $2,
			$3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		);
	`

	_, err = tx.Exec(
		insertPurchase,
		orgId,
		credit,
		metaJSON,
		paymentID,
		userId,
		price,
		subscriptionID,
		slug,
		invoiceURL,
		purchaseType,
		start,
		end,
	)
	if err != nil {
		log.Println("❌ Failed to insert purchase log:", err)
		return
	}

	//------------------------------------------------------
	// 2️⃣ UPDATE credit_limits
	//------------------------------------------------------

	// ---- Handle One-time or Subscription ----
	if purchaseType == "one-time-payment" {
		// 🔹 ONE-TIME → increase purchased_credit only
		updateOneTime := `
			UPDATE credit_limits
			SET 
				purchased_credit = purchased_credit + $1,
				updated_at = NOW()
			WHERE organization_id = $2;
		`

		_, err = tx.Exec(updateOneTime, credit, orgId)
		if err != nil {
			log.Println("❌ Failed to update one-time credit:", err)
			return
		}

		log.Println("✅ One-time credit updated successfully")

	} else if purchaseType == "subscription" {
		//for new subscriptions, enable auto reload by default
		enableAutoReload(orgId)
		updateSubscription := `
			UPDATE credit_limits
			SET 
				concurrent_sessions = $1,
				total_credit = $2,
				balance_credit = $2,
				max_session_duration = $3,
				subscription_id = $4,
				slug = $6,
				updated_at = NOW(),
				current_period_end = NOW() + INTERVAL '1 month'
			WHERE organization_id = $5;
		`

		_, err = tx.Exec(
			updateSubscription,
			concurrentSessions,
			credit,
			maxSessionDuration,
			subscriptionID,
			orgId,
			slug,
		)

		if err != nil {
			log.Println("❌ Failed to update subscription credit limits:", err)
			return
		}

		log.Println("✅ Subscription credit limits updated")
	}

	//Cancel old subscription for plan upgrade
	oldSubscriptionId, ok := session.Metadata["oldSubscriptionId"]

	if ok && oldSubscriptionId != "" {
		log.Println("Cancelling existing subscription:", oldSubscriptionId)
		err := cancelStripeSubscription(oldSubscriptionId)
		if err != nil {
			log.Println("❌ Failed to cancel old subscription:", err)
		}
	} else {
		log.Println("ℹ️ No previous subscription to cancel")
	}

	if purchaseType == "one-time-payment" {
		//Send Email
		emailHtml := GetEmailTemplateHTML("credits")

		//Replace the required details to the templates
		emailHtml = strings.ReplaceAll(emailHtml, "Krishna", userName)
		emailHtml = strings.ReplaceAll(emailHtml, "$34", "$"+strconv.Itoa(int(price)))
		emailHtml = strings.ReplaceAll(emailHtml, "10,000", FormatWithCommas(int64(credit)))
		emailHtml = strings.ReplaceAll(emailHtml, "ReplaceInvoiceURL", invoiceURL)
		err = SendEmail(session.CustomerDetails.Email, "New credits added to your account!", emailHtml)

		if err != nil {
			log.Println("Email error:", err)
		}
	} else {
		//Send Email
		emailHtml := GetEmailTemplateHTML("subscription")

		//Replace the required details to the templates
		emailHtml = strings.ReplaceAll(emailHtml, "Growth", name)
		emailHtml = strings.ReplaceAll(emailHtml, "Krishna", userName)
		emailHtml = strings.ReplaceAll(emailHtml, "$34", "$"+strconv.Itoa(int(price)))
		emailHtml = strings.ReplaceAll(emailHtml, "7,777", FormatWithCommas(int64(credit)))
		emailHtml = strings.ReplaceAll(emailHtml, "ReplaceInvoiceURL", invoiceURL)
		err = SendEmail(session.CustomerDetails.Email, "Subscription activated successfully!", emailHtml)

		if err != nil {
			log.Println("Email error:", err)
		}
	}

	log.Println("✅ Transaction completed successfully")
}

func cancelStripeSubscription(subscriptionID string) error {
	if subscriptionID == "" {
		return nil
	}

	canceled, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		return err
	}

	log.Printf("✅ Subscription cancelled: %s (status=%s)", canceled.ID, canceled.Status)
	return nil
}

func handleInvoicePaid(event stripe.Event) {
	var session stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Println("Failed to unmarshal Invoice.completed:", err)
		return
	}

	log.Printf("Invoice Paid → Session: %s", session.ID)

	userId := session.Metadata["userId"]
	orgIdStr := session.Metadata["orgId"]
	orgId, err := uuid.Parse(orgIdStr)
	if err != nil {
		log.Println("Invalid orgId UUID:", err)
		return
	}

	price := float64(session.AmountPaid) / 100
	paymentID := ""
	subscriptionID := ""
	var start time.Time
	var end time.Time
	creditStr := session.Metadata["creditAmount"]
	slug := session.Metadata["subscriptionPlanId"]
	purchaseType := session.Metadata["type"]
	invoiceURL := session.InvoicePDF

	// Parse credit
	credit := 0
	if creditStr != "" {
		credit, _ = strconv.Atoi(creditStr)
	}

	// Re-encode metadata
	metaJSON, _ := json.Marshal(session.Metadata)

	//------------------------------------------------------
	// BEGIN TRANSACTION
	//------------------------------------------------------
	tx, err := DB.Begin()
	if err != nil {
		log.Println("❌ Failed to begin transaction:", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	//------------------------------------------------------
	// 1️⃣ INSERT PURCHASE LOG
	//------------------------------------------------------
	insertPurchase := `
		INSERT INTO purchase_logs (
			id, organization_id, created_at, credit_value, meta_data,
			payment_id, user_id, price, subscription_id, slug, invoice_url, purchase_type, start_date, end_date
		)
		VALUES (
			gen_random_uuid(), $1, NOW(), $2,
			$3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		);
	`
	_, err = tx.Exec(
		insertPurchase,
		orgId,
		credit,
		metaJSON,
		paymentID,
		userId,
		price,
		subscriptionID,
		slug,
		invoiceURL,
		purchaseType,
		start,
		end,
	)
	if err != nil {
		log.Println("❌ Failed to insert purchase log:", err)
		return
	}

	//------------------------------------------------------
	// 2️⃣ UPDATE credit_limits
	//------------------------------------------------------

	// ---- Handle One-time or Subscription ----
	if purchaseType == "one-time-payment" {
		// 🔹 ONE-TIME → increase purchased_credit only
		updateOneTime := `
			UPDATE credit_limits
			SET 
				purchased_credit = purchased_credit + $1,
				updated_at = NOW()
			WHERE organization_id = $2;
		`

		_, err = tx.Exec(updateOneTime, credit, orgId)
		if err != nil {
			log.Println("❌ Failed to update one-time credit:", err)
			return
		}

		log.Println("✅ One-time credit updated successfully")
	}

	if purchaseType == "one-time-payment" {
		//Send Email
		emailHtml := GetEmailTemplateHTML("credits")

		//Replace the required details to the templates
		emailHtml = strings.ReplaceAll(emailHtml, "Krishna", "")
		emailHtml = strings.ReplaceAll(emailHtml, "$34", "$"+strconv.Itoa(int(price)))
		emailHtml = strings.ReplaceAll(emailHtml, "10,000", FormatWithCommas(int64(credit)))
		emailHtml = strings.ReplaceAll(emailHtml, "ReplaceInvoiceURL", invoiceURL)
		err = SendEmail(session.CustomerEmail, "New credits added to your account!", emailHtml)

		if err != nil {
			log.Println("Email error:", err)
		}
	}

	log.Println("✅ Transaction completed successfully")
}

func enableAutoReload(orgId uuid.UUID) {

	update := `
	UPDATE credit_limits
	SET
		auto_reload = true,
		auto_reload_slug = $1,
		updated_at = NOW()
	WHERE
		organization_id = $2
		AND (subscription_id IS NULL OR subscription_id = '');
`

	res, err := DB.Exec(update, "10k_one_time", orgId)
	if err != nil {
		log.Println("❌ Failed to update auto reload:", err)

	}

	log.Println("✅ Transaction completed successfully" + fmt.Sprint(res.RowsAffected()))
}

// GetInvoicePDF returns the invoice PDF URL from paymentIntentID or subscriptionID.
func GetInvoicePDF(paymentIntentID, subscriptionID string) (string, error) {

	//---------------------------------------------------
	// 1️⃣ If PaymentIntent ID is provided → get invoice
	//---------------------------------------------------
	if paymentIntentID != "" {
		pi, err := paymentintent.Get(paymentIntentID, nil)
		if err != nil {
			return "", fmt.Errorf("failed to fetch payment_intent: %w", err)
		}

		if pi.Invoice == nil {
			return "", fmt.Errorf("no invoice linked with payment_intent %s", paymentIntentID)
		}

		inv, err := invoice.Get(pi.Invoice.ID, nil)
		if err != nil {
			return "", fmt.Errorf("failed to fetch invoice: %w", err)
		}

		return inv.InvoicePDF, nil
	}

	//---------------------------------------------------
	// 2️⃣ If Subscription ID is provided → get latest invoice
	//---------------------------------------------------
	if subscriptionID != "" {
		params := &stripe.InvoiceListParams{}
		params.Filters.AddFilter("subscription", "", subscriptionID)
		params.Limit = stripe.Int64(20)

		iter := invoice.List(params)
		if !iter.Next() {
			return "", fmt.Errorf("no invoices found for subscription %s", subscriptionID)
		}

		inv := iter.Invoice()
		return inv.InvoicePDF, nil
	}

	//---------------------------------------------------
	// 3️⃣ No identifier supplied
	//---------------------------------------------------
	return "", fmt.Errorf("either paymentIntentID or subscriptionID must be provided")
}

func handleSubscriptionUpdated(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("❌ Failed to unmarshal subscription.updated: %v", err)
		return
	}

	log.Printf("Subscription Updated → %s", sub.ID)

	// Find license record
	query := `
        SELECT organization_id 
        FROM credit_limits 
        WHERE subscription_id = $1
        LIMIT 1;
    `
	var organizationID string
	err := DB.QueryRow(query, sub.ID).Scan(&organizationID)
	if err == sql.ErrNoRows {
		log.Println("❌ No license found for subscription:", sub.ID)
		return
	} else if err != nil {
		log.Println("❌ DB error:", err)
		return
	}

	// Determine cancellation state
	var cancelledAt interface{}
	if sub.CancelAtPeriodEnd {
		cancelledAt = time.Now()
		log.Println("Subscription marked as canceled_at =", cancelledAt)
	} else {
		cancelledAt = nil
		log.Println("Subscription reactivated → canceled_at = NULL")
	}

	// Update DB
	update := `
        UPDATE credit_limits
        SET cancelled_at = $1
        WHERE organization_id = $2;
    `

	_, err = DB.Exec(update, cancelledAt, organizationID)
	if err != nil {
		log.Println("❌ Failed to update cancellation:", err)
		return
	}

	log.Printf("✅ Updated license %s cancellation state", organizationID)
}

func CreateBillingPortalSession(userStripeId string) (string, error) {
	// Set API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(userStripeId),
		ReturnURL: stripe.String(os.Getenv("STRIPE_PAYMENT_CANCEL_URL")),
	}

	// Create session
	s, err := portalSession.New(params)
	if err != nil {
		return "", err
	}

	// Return billing portal URL
	return s.URL, nil
}

func ManageSubscriptionPortal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apiKeyHash := r.Header.Get("X-API-Key")

	if apiKeyHash == "" {
		WriteInternalServerError(w, "Missing X-API-Key header")
		return
	}

	query := `
        SELECT
            COALESCE(spl.stripe_id, '') as stripe_id		
        FROM credit_limits spl
        JOIN workspaces w ON spl.organization_id = w.organization_id
        JOIN api_keys ak ON ak.workspace_id = w.id
		JOIN purchase_plans pp ON pp.slug = spl.slug
        WHERE ak.key_hash = $1
        ORDER BY spl.created_at DESC
        LIMIT 1;
    `

	var stripeId string

	err := DB.QueryRow(query, apiKeyHash).Scan(&stripeId)

	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(nil)
		return
	}

	if err != nil {
		WriteInternalServerError(w, "Failed to fetch subscription details: "+err.Error())
		return
	}

	if stripeId == "" {
		WriteNotFoundError(w, "No valid stripe details found for this user")
		return
	}

	link, err1 := CreateBillingPortalSession(stripeId)

	if err1 != nil {
		WriteInternalServerError(w, "Failed to fetch subscription details: "+err1.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"manageUrl": link})
}

func HandleAutoReloadUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apiKeyHash := r.Header.Get("X-API-Key")

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if apiKeyHash == "" {
		WriteInternalServerError(w, "Missing X-API-Key header")
		return
	}

	var req AutoReloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := `
        SELECT
            spl.organization_id
        FROM credit_limits spl
        JOIN workspaces w ON spl.organization_id = w.organization_id
        JOIN api_keys ak ON ak.workspace_id = w.id
		JOIN purchase_plans pp ON pp.slug = spl.slug
		LEFT JOIN purchase_logs pl ON pl.subscription_id = spl.subscription_id
        WHERE ak.key_hash = $1
        LIMIT 1;
    `

	var organization_id string

	err := DB.QueryRow(query, apiKeyHash).Scan(&organization_id)
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch organization ID: "+err.Error())
		return
	}

	// Update DB
	update := `
        UPDATE credit_limits
		SET
			auto_reload = $1,
			auto_reload_slug = $2,
			updated_at = NOW()
		WHERE
			organization_id = $3;
    `

	_, err = DB.Exec(update, req.AutoReload, req.AutoReloadSlug, organization_id)
	if err != nil {
		log.Println("Failed to update auto reload:", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "AutoReload Updated successfully"})
}
