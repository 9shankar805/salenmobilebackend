package main

import (
"log"
"net/http"

"github.com/gin-gonic/gin"

"smg-backend/internal/config"
"smg-backend/internal/database"
"smg-backend/internal/fcm"
"smg-backend/internal/middleware"
"smg-backend/internal/routes/account"
"smg-backend/internal/routes/admin"
"smg-backend/internal/routes/appscreens"
"smg-backend/internal/routes/calculator"
"smg-backend/internal/routes/certificate"
"smg-backend/internal/routes/classroom"
"smg-backend/internal/routes/cohort"
"smg-backend/internal/routes/community"
"smg-backend/internal/routes/content"
"smg-backend/internal/routes/course"
"smg-backend/internal/routes/dash"
"smg-backend/internal/routes/hardware"
"smg-backend/internal/routes/home"
"smg-backend/internal/routes/invite"
"smg-backend/internal/routes/live"
"smg-backend/internal/routes/media"
"smg-backend/internal/routes/messaging"
"smg-backend/internal/routes/notifications"
"smg-backend/internal/routes/onboarding"
"smg-backend/internal/routes/organization"
"smg-backend/internal/routes/plans"
"smg-backend/internal/routes/progress"
"smg-backend/internal/routes/quiz"
"smg-backend/internal/routes/search"
"smg-backend/internal/routes/sso"
"smg-backend/internal/routes/subscription"
"smg-backend/internal/routes/support"
"smg-backend/internal/routes/widgets"
"smg-backend/internal/storage"
)

func main() {
cfg := config.LoadConfig()

// Initialize Cloudflare R2 Cloud Storage
storage.InitStorage()

// Initialize Firebase FCM (non-fatal — works without credentials in dev)
if err := fcm.Init(); err != nil {
log.Printf("⚠️  FCM: %v", err)
} else {
log.Println("✅ FCM: push notifications ready")
}

// Initialize Database with Models & Seed Data
db := database.InitDB(cfg.DatabaseURL)

// Setup Gin Router with CORS
r := gin.Default()
r.Use(middleware.CorsMiddleware())

// Health Check
r.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"status":   "healthy",
"service":  "Salen Mobile Guru Backend (Go)",
"database": "MySQL (salenmobile_db)",
"storage":  "Cloudflare R2 (salenmobileguru)",
"version":  "1.0.0",
})
})
r.HEAD("/health", func(c *gin.Context) {
c.Status(http.StatusOK)
})

// Root redirect to Admin Web Portal
r.HEAD("/", func(c *gin.Context) {
c.Redirect(http.StatusTemporaryRedirect, "/admin")
})
r.GET("/", func(c *gin.Context) {
c.Redirect(http.StatusTemporaryRedirect, "/admin")
})

// Register all route packages
account.RegisterRoutes(r, db)
admin.RegisterRoutes(r, db)
appscreens.RegisterRoutes(r, db)
calculator.RegisterRoutes(r, db)
certificate.RegisterRoutes(r, db)
classroom.RegisterRoutes(r, db)
cohort.RegisterRoutes(r, db)
community.RegisterRoutes(r, db)
content.RegisterRoutes(r, db)
course.RegisterRoutes(r, db)
dash.RegisterRoutes(r, db)
hardware.RegisterRoutes(r, db)
home.RegisterRoutes(r, db)
invite.RegisterRoutes(r, db)
live.RegisterRoutes(r, db)
media.RegisterRoutes(r, db)
messaging.RegisterRoutes(r, db)
notifications.RegisterRoutes(r, db)
onboarding.RegisterRoutes(r, db)
organization.RegisterRoutes(r, db)
plans.RegisterRoutes(r, db)
progress.RegisterRoutes(r, db)
quiz.RegisterRoutes(r, db)
search.RegisterRoutes(r, db)
sso.RegisterRoutes(r, db)
subscription.RegisterRoutes(r, db)
support.RegisterRoutes(r, db)
widgets.RegisterRoutes(r, db)

log.Printf("🚀 Salen Mobile Guru Go Backend running on port %s", cfg.Port)
log.Printf("☁️ Storage: Pure Cloudflare R2 (bucket: salenmobileguru)")
log.Printf("🌐 Admin Portal: http://localhost:%s/admin", cfg.Port)
if err := r.Run(":" + cfg.Port); err != nil {
log.Fatalf("Server failed to start: %v", err)
}
}
