package subscription

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/models"
)

type SubscriptionRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	sr := &SubscriptionRouter{DB: db}

	r.GET("/user/subscription-status", sr.GetStatus)
	r.GET("/subscriptions/status", sr.GetStatus)
	r.POST("/payments/verify", sr.VerifyPayment)
	r.POST("/subscriptions/enroll", sr.VerifyPayment)
	r.GET("/api/v1/payment-settings", sr.GetPaymentSettings)
	r.GET("/payment-settings", sr.GetPaymentSettings)
	r.POST("/api/v1/subscriptions/submit-proof", sr.SubmitPaymentProof)
	r.POST("/subscriptions/submit-proof", sr.SubmitPaymentProof)
}

// GetStatus returns the user's active subscription details.
// Also returns which features are unlocked so the app can gate UI correctly.
func (sr *SubscriptionRouter) GetStatus(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.Query("id")
	}
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    notSubscribedResponse(),
		})
		return
	}

	// Admin / instructor / VIP always have full access
	var user models.User
	if err := sr.DB.Where("id = ? OR email = ?", userID, userID).First(&user).Error; err == nil {
		if strings.EqualFold(user.Role, "ADMIN") || strings.EqualFold(user.Role, "INSTRUCTOR") || user.IsVIP || user.IsSubscribed {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": map[string]interface{}{
					"isSubscribed":         true,
					"isVip":                true,
					"planType":             "VIP",
					"planCategory":         "FULL_ACCESS",
					"status":               "active",
					"expiresAt":            time.Now().AddDate(10, 0, 0).Format("2006-01-02"),
					"enrolledCourseSlugs":  allCourseSlugs(),
					"totalCertificates":    0,
					"features":             allFeatures(),
					"canAccessMobileTools": true,
					"canAccessChat":        true,
					"canAccessCourses":     true,
				},
			})
			return
		}
	}

	var sub models.Subscription
	err := sr.DB.Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at desc").First(&sub).Error

	if err != nil {
		// Check if user has a pending payment submitted awaiting admin review
		var pendingSub models.Subscription
		if pErr := sr.DB.Where("user_id = ? AND status = ?", userID, "pending").
			Order("created_at desc").First(&pendingSub).Error; pErr == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": map[string]interface{}{
					"isSubscribed":      false,
					"hasPendingPayment": true,
					"status":            "pending",
					"pendingSubscription": map[string]interface{}{
						"id":            pendingSub.ID,
						"planType":      pendingSub.PlanType,
						"amount":        pendingSub.Amount,
						"currency":      pendingSub.Currency,
						"paymentMethod": pendingSub.PaymentMethod,
						"transactionId": pendingSub.TransactionID,
						"courseSlug":    pendingSub.CourseSlug,
						"createdAt":     pendingSub.CreatedAt.Format(time.RFC3339),
					},
					"expiresAt":            nil,
					"enrolledCourseSlugs":  []string{},
					"totalCertificates":    0,
					"features":             []string{},
					"canAccessMobileTools": false,
					"canAccessChat":        false,
					"canAccessCourses":     false,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": notSubscribedResponse()})
		return
	}

	// Check expiry
	if time.Now().After(sub.EndDate) {
		sr.DB.Model(&sub).Update("status", "expired")
		c.JSON(http.StatusOK, gin.H{"success": true, "data": notSubscribedResponse()})
		return
	}

	enrolledList := []string{}
	if sub.EnrolledSlugs != "" {
		enrolledList = strings.Split(sub.EnrolledSlugs, ",")
	}

	var certCount int64
	sr.DB.Model(&models.Certificate{}).Where("user_id = ?", userID).Count(&certCount)

	planCategory := planTypeToCategory(sub.PlanType)
	features := featuresForPlan(sub.PlanType)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"isSubscribed":         true,
			"planType":             sub.PlanType,
			"planCategory":         planCategory,
			"status":               sub.Status,
			"expiresAt":            sub.EndDate.Format("2006-01-02"),
			"enrolledCourseSlugs":  enrolledList,
			"totalCertificates":    certCount,
			"features":             features,
			"canAccessMobileTools": canMobileTools(sub.PlanType),
			"canAccessChat":        canChat(sub.PlanType),
			"canAccessCourses":     true,
		},
	})
}

// VerifyPayment activates a subscription after payment.
// It looks up the real Plan from the plans table to set correct duration and features.
func (sr *SubscriptionRouter) VerifyPayment(c *gin.Context) {
	var req struct {
		UserID        string  `json:"userId" binding:"required"`
		PlanID        string  `json:"planId"`        // plan ID from plans table (e.g. "mp_1month")
		PlanType      string  `json:"planType"`      // legacy: VIP/ANNUAL etc or plan table ID
		CourseID      string  `json:"courseId"`
		CourseSlug    string  `json:"courseSlug"`
		PaymentMethod string  `json:"paymentMethod"`
		TransactionID string  `json:"transactionId"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		Status        string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "INR"
	}
	if req.TransactionID == "" {
		req.TransactionID = "TXN-" + uuid.New().String()[:8]
	}

	// ── Resolve plan from plans table ──────────────────────────────────────
	// Priority: planId field → planType if it looks like a plan table ID → fallback VIP
	planID := req.PlanID
	if planID == "" {
		planID = req.PlanType // app may send planId in planType field
	}

	var plan models.Plan
	planFound := false
	if planID != "" {
		if err := sr.DB.Where("id = ? OR UPPER(id) = UPPER(?)", planID, planID).First(&plan).Error; err == nil {
			planFound = true
		}
	}

	// Determine subscription type and duration from the plan
	var (
		subscriptionPlanType string
		durationMonths       int
		enrolledSlugs        string
	)

	if planFound {
		switch plan.PlanType {
		case "MOBILE":
			subscriptionPlanType = "MOBILE"
		case "CHAT":
			subscriptionPlanType = "CHAT"
		default:
			subscriptionPlanType = "VIP"
		}
		durationMonths = plan.DurationMonths
		if durationMonths == 0 {
			durationMonths = 1
		}
	} else {
		// Legacy path — planType sent directly (VIP/ANNUAL/LIFETIME/SINGLE_COURSE)
		subscriptionPlanType = req.PlanType
		if subscriptionPlanType == "" {
			subscriptionPlanType = "VIP"
		}
		switch strings.ToUpper(subscriptionPlanType) {
		case "ANNUAL":
			durationMonths = 12
		case "LIFETIME":
			durationMonths = 120 // 10 years
		default:
			durationMonths = 1
		}
	}

	// Enrolled course slugs: MOBILE/CHAT plans don't grant course access by themselves
	// VIP / ANNUAL / LIFETIME grants all courses
	pt := strings.ToUpper(subscriptionPlanType)
	if pt == "VIP" || pt == "ANNUAL" || pt == "LIFETIME" {
		enrolledSlugs = strings.Join(allCourseSlugs(), ",")
	} else if req.CourseSlug != "" {
		enrolledSlugs = req.CourseSlug
	}

	finalStatus := "active"
	if strings.EqualFold(req.Status, "pending") {
		finalStatus = "pending"
	}

	sub := models.Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		PlanType:      subscriptionPlanType,
		Status:        finalStatus,
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(0, durationMonths, 0),
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CourseID:      req.CourseID,
		CourseSlug:    req.CourseSlug,
		EnrolledSlugs: enrolledSlugs,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := sr.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create subscription"})
		return
	}

	if finalStatus == "pending" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Payment submitted successfully! Course access will activate once admin verifies your payment screenshot.",
			"data": gin.H{
				"subscriptionId": sub.ID,
				"planType":       sub.PlanType,
				"status":         "pending",
				"transactionId":  sub.TransactionID,
				"amount":         sub.Amount,
				"currency":       sub.Currency,
			},
		})
		return
	}

	// Update user's isSubscribed flag
	sr.DB.Model(&models.User{}).Where("id = ?", req.UserID).Updates(map[string]interface{}{
		"is_subscribed":     true,
		"is_vip":            pt == "VIP" || pt == "LIFETIME",
		"subscription_plan": sub.PlanType,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment verified and subscription activated successfully!",
		"data": gin.H{
			"subscriptionId":      sub.ID,
			"planType":            sub.PlanType,
			"planCategory":        planTypeToCategory(sub.PlanType),
			"status":              sub.Status,
			"expiresAt":           sub.EndDate.Format("2006-01-02"),
			"transactionId":       sub.TransactionID,
			"enrolledCourseSlugs": strings.Split(enrolledSlugs, ","),
			"features":            featuresForPlan(sub.PlanType),
			"canAccessMobileTools": canMobileTools(sub.PlanType),
			"canAccessChat":        canChat(sub.PlanType),
		},
	})
}

// ─── helpers ────────────────────────────────────────────────────────────────

func allCourseSlugs() []string {
	return []string{
		"android-hardware-repairing-level-1",
		"iphone-display-repairing",
		"emmc-ufs-programming",
		"borneo-zxw-schematic-reading",
		"pmic-power-section-diagnosis",
	}
}

func notSubscribedResponse() map[string]interface{} {
	return map[string]interface{}{
		"isSubscribed":         false,
		"planType":             "",
		"planCategory":         "NONE",
		"status":               "none",
		"expiresAt":            nil,
		"enrolledCourseSlugs":  []string{},
		"totalCertificates":    0,
		"features":             []string{},
		"canAccessMobileTools": false,
		"canAccessChat":        false,
		"canAccessCourses":     false,
	}
}

func planTypeToCategory(pt string) string {
	switch strings.ToUpper(pt) {
	case "MOBILE":
		return "MOBILE_TOOLS"
	case "CHAT":
		return "EXPERT_CHAT"
	case "VIP", "LIFETIME":
		return "FULL_ACCESS"
	case "ANNUAL":
		return "FULL_ACCESS"
	case "SINGLE_COURSE":
		return "SINGLE_COURSE"
	default:
		return "NONE"
	}
}

func canMobileTools(pt string) bool {
	p := strings.ToUpper(pt)
	return p == "VIP" || p == "ANNUAL" || p == "LIFETIME" || p == "MOBILE" || p == "SINGLE_COURSE"
}

func canChat(pt string) bool {
	p := strings.ToUpper(pt)
	return p == "VIP" || p == "ANNUAL" || p == "LIFETIME" || p == "CHAT"
}

func allFeatures() []string {
	return []string{
		"Full HD Video Lessons Access", "Motherboard Schematic Diagrams",
		"Multimeter & Diode Mode Procedures", "Hardware Diagnostic Guides",
		"PMIC & Power Sequence Schematics", "Live Masterclasses with Master Salen",
		"Priority 1-on-1 Hardware Support", "Official Academy Completion Certificate",
		"Expert Technician Chat Support",
	}
}

func featuresForPlan(pt string) []string {
	switch strings.ToUpper(pt) {
	case "MOBILE":
		return []string{
			"Full HD Video Lessons Access",
			"Motherboard Schematic Diagrams",
			"Multimeter & Diode Mode Procedures",
			"Hardware Diagnostic Guides",
			"Student Discussion Community",
		}
	case "CHAT":
		return []string{
			"Expert Chat Support", "Priority Response",
			"Unlimited Chats", "Hardware Diagnosis Help",
		}
	case "VIP", "ANNUAL", "LIFETIME":
		return allFeatures()
	case "SINGLE_COURSE":
		return []string{"Single Course Access", "Course Certificates", "Quiz Access"}
	default:
		return []string{}
	}
}

func (sr *SubscriptionRouter) GetPaymentSettings(c *gin.Context) {
	var ps models.PaymentSetting
	if err := sr.DB.Where("is_active = ?", true).First(&ps).Error; err != nil {
		if err := sr.DB.First(&ps).Error; err != nil {
			ps = models.PaymentSetting{
				ID:                 "ps_default",
				EsewaQrUrl:         "https://api.qrserver.com/v1/create-qr-code/?size=350x350&data=esewa://payment?id=9805912345&name=SalenMobileGuru",
				EsewaID:            "9805912345",
				EsewaMerchantName:  "Salen Mobile Guru Academy (Salen Dangol)",
				KhaltiQrUrl:        "https://api.qrserver.com/v1/create-qr-code/?size=350x350&data=khalti://payment?id=9805912345&name=SalenMobileGuru",
				KhaltiID:           "9805912345",
				KhaltiMerchantName: "Salen Mobile Guru Academy",
				BankName:           "Nabil Bank Ltd.",
				BankAccountName:    "SALEN DANGOL / SALEN MOBILE GURU",
				BankAccountNumber:  "01201017500123",
				Instructions:       "1. Open eSewa / Khalti / Mobile Banking and scan the QR code above.\n2. In remarks, enter your Student Name and Mobile Number.\n3. Complete the payment and save the confirmation screenshot.\n4. Enter your Transaction Ref / UTR ID below and upload your screenshot voucher.\n5. Tap 'Submit Payment for Verification'. Your plan will be verified and activated within 10-15 minutes.",
				IsActive:           true,
				UpdatedAt:          time.Now(),
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ps,
	})
}

func (sr *SubscriptionRouter) SubmitPaymentProof(c *gin.Context) {
	var req struct {
		UserID        string  `json:"userId" binding:"required"`
		PlanID        string  `json:"planId"`
		PlanType      string  `json:"planType"`
		CourseID      string  `json:"courseId"`
		CourseSlug    string  `json:"courseSlug"`
		PaymentMethod string  `json:"paymentMethod"`
		TransactionID string  `json:"transactionId"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		ProofImageUrl string  `json:"proofImageUrl"`
		Notes         string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.TransactionID == "" {
		req.TransactionID = "TXN-" + uuid.New().String()[:8]
	}
	if req.Currency == "" {
		req.Currency = "NPR"
	}
	if req.PlanType == "" {
		req.PlanType = "VIP All-Access Pass"
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "eSewa"
	}

	sub := models.Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		PlanType:      req.PlanType,
		Status:        "pending",
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(1, 0, 0),
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CourseID:      req.CourseID,
		CourseSlug:    req.CourseSlug,
		ProofImageUrl: req.ProofImageUrl,
		Notes:         req.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := sr.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to submit payment proof"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment proof submitted successfully! Admin will verify and activate your plan.",
		"data": gin.H{
			"subscriptionId": sub.ID,
			"planType":       sub.PlanType,
			"status":         "pending",
			"transactionId":  sub.TransactionID,
			"amount":         sub.Amount,
			"currency":       sub.Currency,
			"proofImageUrl":  sub.ProofImageUrl,
		},
	})
}
