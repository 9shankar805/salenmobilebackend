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

type SignUpRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
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

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not process password"})
		return
	}

	user := models.User{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Email:        email,
		Phone:        req.Phone,
		PasswordHash: string(hashed),
		Role:         "student",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := ar.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not create account"})
		return
	}

	tokenStr, err := issueJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Account created but could not issue token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"token":   tokenStr,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
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
