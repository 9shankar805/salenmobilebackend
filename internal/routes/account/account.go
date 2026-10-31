package account

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"smg-backend/internal/mail"
	"smg-backend/internal/models"
)

type AccountRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	ar := &AccountRouter{DB: db}

	r.GET("/account", ar.GetAccount)
	r.GET("/session", ar.GetSession)
	r.GET("/api/auth/get-session", ar.GetSession)
	r.GET("/api/auth/session", ar.GetSession)
	r.POST("/api/auth/sign-in/email", ar.SignInEmail)
	r.POST("/api/auth/sign-out", ar.SignOut)
	r.POST("/api/auth/sign-up", ar.SignUp)
	r.POST("/api/auth/send-otp", ar.SendOTP)
	r.POST("/api/auth/request-otp", ar.SendOTP)
	r.POST("/api/auth/otp", ar.SendOTP)
	r.POST("/api/auth/verify-otp", ar.VerifyOTP)

	// User profile update — Flutter edit profile screen calls PUT /user/profile
	r.PUT("/user/profile", ar.UpdateProfile)
	r.PATCH("/user/profile", ar.UpdateProfile) // also accept PATCH
	r.GET("/user/profile", ar.GetProfile)
}

// issueJWT creates a signed JWT for the given user.
func issueJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "salen-mobile-guru-secret-key-2026"
	}

	claims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(30 * 24 * time.Hour).Unix(), // 30-day token
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (ar *AccountRouter) GetAccount(c *gin.Context) {
	// Try to identify user from token if present
	userID, _ := c.Get("userId")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Not authenticated"})
		return
	}

	var user models.User
	if err := ar.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"name":          user.Name,
			"role":          user.Role,
			"emailVerified": true,
		},
	})
}

func (ar *AccountRouter) GetSession(c *gin.Context) {
	// Return a minimal guest session; real auth uses the JWT token directly
	c.JSON(http.StatusOK, gin.H{
		"session": nil,
		"user":    nil,
	})
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ar *AccountRouter) SignInEmail(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid login credentials"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		// Constant-time response to prevent user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	tokenStr, err := issueJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   tokenStr,
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"name":          user.Name,
			"role":          user.Role,
			"emailVerified": true,
		},
	})
}

type SendOTPRequest struct {
	Email   string `json:"email" binding:"required"`
	Name    string `json:"name"`
	Purpose string `json:"purpose"` // "signup", "login", "reset_password"
}

func (ar *AccountRouter) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Email is required"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid email address format"})
		return
	}

	purpose := strings.ToLower(req.Purpose)
	if purpose == "" {
		purpose = "signup"
	}

	if purpose == "signup" {
		var existing models.User
		if err := ar.DB.Where("LOWER(email) = ?", email).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "This email is already registered. Please log in or reset your password.",
			})
			return
		}
	}

	// Generate secure 6-digit OTP code
	code := mail.GenerateOTP()

	// Invalidate previous OTPs for this email and purpose
	ar.DB.Model(&models.EmailOTP{}).Where("LOWER(email) = ? AND purpose = ? AND verified = ?", email, purpose, false).
		Update("verified", true)

	otpRecord := models.EmailOTP{
		ID:        uuid.New().String(),
		Email:     email,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		CreatedAt: time.Now(),
	}

	if err := ar.DB.Create(&otpRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed generating authorization code"})
		return
	}

	// Dispatch email via configured SMTP
	recipientName := req.Name
	if recipientName == "" {
		recipientName = strings.Split(email, "@")[0]
	}

	err := mail.SendOTPEmail(ar.DB, email, recipientName, code, purpose)

	smtpCfg := mail.GetSMTPSettings(ar.DB)
	res := gin.H{
		"success":   true,
		"message":   "6-digit verification code sent to " + email,
		"email":     email,
		"expiresIn": 600,
	}

	if err != nil {
		res["warning"] = "SMTP delivery notice: " + err.Error()
	}

	if !smtpCfg.Enabled {
		res["debugOtp"] = code
		res["note"] = "Custom SMTP is currently disabled in Admin Panel; using simulation mode."
	}

	c.JSON(http.StatusOK, res)
}

type VerifyOTPRequest struct {
	Email   string `json:"email" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Purpose string `json:"purpose"`
}

func (ar *AccountRouter) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Email and code are required"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	code := strings.TrimSpace(req.Code)
	purpose := strings.ToLower(req.Purpose)
	if purpose == "" {
		purpose = "signup"
	}

	var otp models.EmailOTP
	err := ar.DB.Where("LOWER(email) = ? AND code = ? AND verified = ? AND expires_at > ?", email, code, false, time.Now()).
		Order("created_at desc").First(&otp).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid or expired verification code. Please check your inbox or request a new code.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully ✓",
		"email":   email,
	})
}

type SignUpRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
	OTP      string `json:"otp"`
	Code     string `json:"code"`
}

func (ar *AccountRouter) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	// Check existing
	var existing models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Email already registered"})
		return
	}

	otpCode := strings.TrimSpace(req.OTP)
	if otpCode == "" {
		otpCode = strings.TrimSpace(req.Code)
	}

	smtpCfg := mail.GetSMTPSettings(ar.DB)
	emailVerified := false

	// Validate OTP if provided or if custom SMTP is actively enforced
	if otpCode != "" {
		var otp models.EmailOTP
		err := ar.DB.Where("LOWER(email) = ? AND code = ? AND verified = ? AND expires_at > ?", email, otpCode, false, time.Now()).
			Order("created_at desc").First(&otp).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid or expired OTP verification code",
			})
			return
		}
		otp.Verified = true
		ar.DB.Save(&otp)
		emailVerified = true
	} else if smtpCfg.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Verification code (OTP) required. Please click 'Send Code' to receive your OTP in Gmail.",
		})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not process password"})
		return
	}

	user := models.User{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Email:         email,
		Phone:         req.Phone,
		PasswordHash:  string(hashed),
		Role:          "student",
		EmailVerified: emailVerified,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := ar.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not create account"})
		return
	}

	// Dispatch welcome email asynchronously
	go mail.SendWelcomeEmail(ar.DB, user.Email, user.Name)

	tokenStr, err := issueJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Account created but could not issue token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"token":   tokenStr,
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"name":          user.Name,
			"role":          user.Role,
			"emailVerified": user.EmailVerified,
		},
	})
}

func (ar *AccountRouter) SignOut(c *gin.Context) {
	// JWT is stateless — client simply discards the token
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Signed out successfully"})
}

// GetProfile returns a user's profile by userId query param or JWT context.
func (ar *AccountRouter) GetProfile(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		if id, exists := c.Get("userId"); exists {
			userID, _ = id.(string)
		}
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	var user models.User
	if err := ar.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":               user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"phone":            user.Phone,
			"avatarUrl":        user.AvatarURL,
			"role":             user.Role,
			"isSubscribed":     user.IsSubscribed,
			"isVip":            user.IsVIP,
			"subscriptionPlan": user.SubscriptionPlan,
			"createdAt":        user.CreatedAt,
		},
	})
}

type UpdateProfileRequest struct {
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatarUrl"`
}

// UpdateProfile updates name, phone, and avatar for a user.
// Flutter edit profile screen calls PUT /user/profile with {userId, name, phone, avatarUrl}.
func (ar *AccountRouter) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Prefer userId from body; fall back to JWT context
	userID := req.UserID
	if userID == "" {
		if id, exists := c.Get("userId"); exists {
			userID, _ = id.(string)
		}
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	var user models.User
	if err := ar.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// Only update fields that were provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	user.UpdatedAt = time.Now()

	if err := ar.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data": gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"phone":     user.Phone,
			"avatarUrl": user.AvatarURL,
			"role":      user.Role,
		},
	})
}
