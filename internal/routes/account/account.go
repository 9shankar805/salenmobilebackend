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

	// Gmail-only Authentication
	r.POST("/api/auth/sign-in/email", ar.SignInEmail)
	r.POST("/api/auth/sign-in/gmail", ar.SignInEmail)
	r.POST("/api/auth/sign-in", ar.SignInEmail)
	r.POST("/api/auth/login", ar.SignInEmail)
	r.POST("/api/auth/login/email", ar.SignInEmail)
	r.POST("/api/auth/login/gmail", ar.SignInEmail)
	r.POST("/api/auth/sign-in/otp", ar.SignInOTP)
	r.POST("/api/auth/login/otp", ar.SignInOTP)
	r.POST("/api/auth/sign-out", ar.SignOut)
	r.POST("/api/auth/sign-up", ar.SignUp)
	r.POST("/api/auth/register", ar.SignUp)
	r.POST("/api/auth/send-otp", ar.SendOTP)
	r.POST("/api/auth/request-otp", ar.SendOTP)
	r.POST("/api/auth/otp", ar.SendOTP)
	r.POST("/api/auth/verify-otp", ar.VerifyOTP)
	r.POST("/api/auth/reset-password", ar.ResetPassword)
	r.POST("/api/auth/forgot-password/reset", ar.ResetPassword)

	// Explicitly reject Google & Phone logins
	r.POST("/api/auth/google", ar.DisabledAuthMethod)
	r.POST("/api/auth/sign-in/google", ar.DisabledAuthMethod)
	r.POST("/api/auth/login/google", ar.DisabledAuthMethod)
	r.POST("/api/auth/phone", ar.DisabledAuthMethod)
	r.POST("/api/auth/sign-in/phone", ar.DisabledAuthMethod)
	r.POST("/api/auth/login/phone", ar.DisabledAuthMethod)
	r.POST("/api/auth/send-phone-otp", ar.DisabledAuthMethod)

	// User profile update — Flutter edit profile screen calls PUT /user/profile
	r.PUT("/user/profile", ar.UpdateProfile)
	r.PATCH("/user/profile", ar.UpdateProfile) // also accept PATCH
	r.GET("/user/profile", ar.GetProfile)
}

// isGmailAddress verifies that the email is a valid Google Gmail address
func isGmailAddress(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return strings.HasSuffix(email, "@gmail.com") || strings.HasSuffix(email, "@googlemail.com")
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

	// Enforce Gmail only
	if !isGmailAddress(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Only Gmail addresses (@gmail.com) are allowed. Google OAuth and phone number login have been removed.",
		})
		return
	}

	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		// Constant-time response to prevent user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid Gmail address or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid Gmail address or password"})
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

type SignInOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	OTP   string `json:"otp"`
}

// SignInOTP enables direct passwordless Gmail login using the 6-digit OTP code
func (ar *AccountRouter) SignInOTP(c *gin.Context) {
	var req SignInOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !isGmailAddress(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Only Gmail addresses (@gmail.com) are allowed. Google OAuth and phone number login have been removed.",
		})
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = strings.TrimSpace(req.OTP)
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Gmail OTP verification code is required"})
		return
	}

	var otp models.EmailOTP
	err := ar.DB.Where("LOWER(email) = ? AND code = ? AND verified = ? AND expires_at > ?", email, code, false, time.Now()).
		Order("created_at desc").First(&otp).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid or expired OTP verification code. Please check your Gmail or request a new code.",
		})
		return
	}

	otp.Verified = true
	ar.DB.Save(&otp)

	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		username := strings.Split(email, "@")[0]
		user = models.User{
			ID:            uuid.New().String(),
			Name:          username,
			Email:         email,
			Role:          "student",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := ar.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not create user session"})
			return
		}
	} else if !user.EmailVerified {
		user.EmailVerified = true
		ar.DB.Model(&user).Update("email_verified", true)
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

func (ar *AccountRouter) DisabledAuthMethod(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "Google OAuth and Phone number login have been removed. Only Gmail login is supported.",
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

	if !isGmailAddress(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Only Gmail addresses (@gmail.com) are supported for verification code.",
		})
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
	Email   string `json:"email"`
	Code    string `json:"code"`
	OTP     string `json:"otp"`
	Purpose string `json:"purpose"`
}

func (ar *AccountRouter) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = strings.TrimSpace(req.OTP)
	}
	if email == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Email and code (or otp) are required"})
		return
	}
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

type ResetPasswordRequest struct {
	Email           string `json:"email" binding:"required"`
	OTP             string `json:"otp"`
	Code            string `json:"code"`
	NewPassword     string `json:"newPassword"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (ar *AccountRouter) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !isGmailAddress(email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Only Gmail addresses (@gmail.com) are supported"})
		return
	}

	code := strings.TrimSpace(req.OTP)
	if code == "" {
		code = strings.TrimSpace(req.Code)
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Verification code (OTP) is required"})
		return
	}

	newPass := strings.TrimSpace(req.NewPassword)
	if newPass == "" {
		newPass = strings.TrimSpace(req.Password)
	}
	if len(newPass) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "New password must be at least 6 characters"})
		return
	}

	if req.ConfirmPassword != "" && req.ConfirmPassword != newPass {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Passwords do not match"})
		return
	}

	// Verify OTP for reset_password or general signup
	var otp models.EmailOTP
	err := ar.DB.Where("LOWER(email) = ? AND code = ? AND verified = ? AND expires_at > ?", email, code, false, time.Now()).
		Order("created_at desc").First(&otp).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid or expired verification code. Please check your Gmail or request a new OTP.",
		})
		return
	}

	// Mark OTP verified
	otp.Verified = true
	ar.DB.Save(&otp)

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not process new password"})
		return
	}

	// Update user password in DB
	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No account found registered with " + email})
		return
	}

	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()
	if err := ar.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update password in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully! You can now log in with your new password.",
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

	if !isGmailAddress(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Registration is restricted to Gmail accounts (@gmail.com). Google OAuth and phone login have been removed.",
		})
		return
	}

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

	var lessonsCompleted int64
	ar.DB.Model(&models.UserLessonProgress{}).Where("user_id = ? AND is_completed = ?", userID, true).Count(&lessonsCompleted)

	var certificatesEarned int64
	ar.DB.Model(&models.Certificate{}).Where("user_id = ?", userID).Count(&certificatesEarned)

	var totalSeconds int64
	ar.DB.Model(&models.UserLessonProgress{}).Where("user_id = ?", userID).Select("COALESCE(SUM(watched_seconds), 0)").Scan(&totalSeconds)

	// Enrollments calculation
	var subs []models.Subscription
	ar.DB.Where("user_id = ? AND (status = 'active' OR status = 'approved')", userID).Find(&subs)
	enrolledSet := make(map[string]bool)
	for _, s := range subs {
		if s.CourseSlug != "" {
			enrolledSet[s.CourseSlug] = true
		}
		if s.EnrolledSlugs != "" {
			for _, slug := range strings.Split(s.EnrolledSlugs, ",") {
				slug = strings.TrimSpace(slug)
				if slug != "" {
					enrolledSet[slug] = true
				}
			}
		}
	}
	if user.IsVIP || user.IsSubscribed {
		var allCourses []models.Course
		ar.DB.Where("is_published = ?", true).Find(&allCourses)
		for _, crs := range allCourses {
			enrolledSet[crs.Slug] = true
		}
	}

	enrolledSlugsList := make([]string, 0, len(enrolledSet))
	for s := range enrolledSet {
		enrolledSlugsList = append(enrolledSlugsList, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":                   user.ID,
			"name":                 user.Name,
			"email":                user.Email,
			"phone":                user.Phone,
			"avatarUrl":            user.AvatarURL,
			"role":                 user.Role,
			"isSubscribed":         user.IsSubscribed,
			"isVip":                user.IsVIP,
			"subscriptionPlan":     user.SubscriptionPlan,
			"coursesEnrolled":      len(enrolledSlugsList),
			"lessonsCompleted":     lessonsCompleted,
			"certificatesEarned":   certificatesEarned,
			"totalLearningMinutes": totalSeconds / 60,
			"enrolledCourseIds":    enrolledSlugsList,
			"createdAt":            user.CreatedAt,
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
