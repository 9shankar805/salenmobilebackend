package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// ─── JWT Auth Middleware ───────────────────────────────────────────────────

// AuthMiddleware validates the JWT token from the Authorization header.
// Expects: Authorization: Bearer <token>
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header format must be: Bearer <token>",
			})
			return
		}

		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "salen-mobile-guru-secret-key-2026"
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token claims",
			})
			return
		}

		c.Set("userId", claims["userId"])
		c.Set("userEmail", claims["email"])
		c.Set("userRole", claims["role"])
		c.Next()
	}
}

// AdminMiddleware ensures the authenticated user has an ADMIN role.
// Must be chained after AuthMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}
		roleStr, _ := role.(string)
		if !strings.EqualFold(roleStr, "ADMIN") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin access required",
			})
			return
		}
		c.Next()
	}
}

// ─── Subscription Middleware ───────────────────────────────────────────────
//
// RequireSubscription gates a route by plan category.
//
//	"ANY"    — any active subscription
//	"MOBILE" — MOBILE, VIP, ANNUAL, or LIFETIME
//	"CHAT"   — CHAT, VIP, ANNUAL, or LIFETIME
//	"VIP"    — VIP or LIFETIME only
//
// userId is read from JWT context first, then ?userId= query param.
// Admins and instructors always bypass the check.
func RequireSubscription(db *gorm.DB, planCategory string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := ""
		if id, exists := c.Get("userId"); exists {
			userID, _ = id.(string)
		}
		if userID == "" {
			userID = c.GetHeader("X-User-Id")
		}
		if userID == "" {
			userID = c.GetHeader("X-User-Email")
		}
		if userID == "" {
			userID = c.Param("userId")
		}
		if userID == "" {
			userID = c.Query("userId")
		}
		if userID == "" {
			userID = c.Query("email")
		}
		if userID == "" {
			userID = c.Query("id")
		}

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success":              false,
				"error":                "Sign in required to access this feature.",
				"subscriptionRequired": true,
				"requiredPlan":         planCategory,
				"upgradeUrl":           "/plans",
			})
			return
		}

		if !hasActiveSubscription(db, userID, planCategory) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success":              false,
				"error":                "Active " + planCategory + " subscription required.",
				"subscriptionRequired": true,
				"requiredPlan":         planCategory,
				"upgradeUrl":           "/plans",
			})
			return
		}

		c.Next()
	}
}

// hasActiveSubscription returns true if userID holds an active subscription
// satisfying the given planCategory.
func hasActiveSubscription(db *gorm.DB, userID, planCategory string) bool {
	// Admins, instructors, VIP, and subscribed users always pass
	var u struct {
		Role         string
		IsVIP        bool
		IsSubscribed bool
	}
	if db.Table("users").Select("role, is_vip, is_subscribed").Where("id = ? OR LOWER(email) = LOWER(?)", userID, userID).Scan(&u).Error == nil {
		r := strings.ToUpper(u.Role)
		if r == "ADMIN" || r == "INSTRUCTOR" || u.IsVIP || u.IsSubscribed {
			return true
		}
	}

	var sub struct {
		PlanType string
		Status   string
	}
	err := db.Table("subscriptions").
		Select("plan_type, status").
		Where("(user_id = ? OR user_id IN (SELECT id FROM users WHERE LOWER(email) = LOWER(?))) AND status IN ('active', 'approved')", userID, userID).
		Order("created_at desc").
		Limit(1).
		Scan(&sub).Error

	if err != nil || (sub.Status != "active" && sub.Status != "approved") {
		return false
	}

	pt := strings.ToUpper(sub.PlanType)

	switch strings.ToUpper(planCategory) {
	case "ANY":
		return true
	case "MOBILE":
		return pt == "VIP" || pt == "ANNUAL" || pt == "LIFETIME" || pt == "MOBILE" || pt == "SINGLE_COURSE"
	case "CHAT":
		return pt == "VIP" || pt == "ANNUAL" || pt == "LIFETIME" || pt == "CHAT"
	case "VIP":
		return pt == "VIP" || pt == "LIFETIME"
	default:
		return pt == strings.ToUpper(planCategory)
	}
}
