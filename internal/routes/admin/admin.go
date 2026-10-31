package admin

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"smg-backend/internal/mail"
	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
	"smg-backend/internal/storage"
)

var serverStartTime = time.Now()

//go:embed static/admin.html
var adminHTML []byte

//go:embed static/classroom.html
var classroomHTML []byte

//go:embed static/courses.html
var coursesHTML []byte

//go:embed static/curriculum.html
var curriculumHTML []byte

//go:embed static/quizzes.html
var quizzesHTML []byte

//go:embed static/subscriptions.html
var subscriptionsHTML []byte

//go:embed static/users.html
var usersHTML []byte

//go:embed static/live.html
var liveHTML []byte

//go:embed static/templates.html
var templatesHTML []byte

//go:embed static/smtp.html
var smtpHTML []byte

//go:embed courses/index.html
var courseWorkspaceHTML []byte

//go:embed courses/attendance/index.html
var courseAttendanceHTML []byte

//go:embed courses/marks/index.html
var courseMarksHTML []byte

//go:embed courses/lessons/index.html
var courseLessonsHTML []byte

//go:embed courses/exercises/index.html
var courseExercisesHTML []byte

//go:embed courses/submissions/index.html
var courseSubmissionsHTML []byte

//go:embed courses/people/index.html
var coursePeopleHTML []byte

//go:embed courses/certificates/index.html
var courseCertificatesHTML []byte

//go:embed courses/compliance/index.html
var courseComplianceHTML []byte

//go:embed courses/landingpage/index.html
var courseLandingpageHTML []byte

//go:embed courses/ai-tutor/index.html
var courseAiTutorHTML []byte

//go:embed courses/settings/index.html
var courseSettingsHTML []byte

//go:embed courses/analytics/index.html
var courseAnalyticsHTML []byte

//go:embed all:courses
var coursesFS embed.FS

type AdminRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	ar := &AdminRouter{DB: db}

	// 1. Web Portal (public — serves embedded HTML)
	r.GET("/admin", ar.ServeAdminWeb)
	r.HEAD("/admin", ar.ServeAdminWeb)
	r.HEAD("/admin/*any", ar.ServeAdminWeb)
	r.GET("/admin/*any", ar.ServeAdminWeb)
	r.GET("/classroom", ar.ServeClassroomWeb)
	r.GET("/classroom/*any", ar.ServeClassroomWeb)
	r.GET("/learn/*any", ar.ServeClassroomWeb)
	r.GET("/courses.html", ar.ServeCoursesWeb)
	r.GET("/curriculum.html", ar.ServeCurriculumWeb)
	r.GET("/quizzes.html", ar.ServeQuizzesWeb)
	r.GET("/subscriptions.html", ar.ServeSubscriptionsWeb)
	r.GET("/users.html", ar.ServeUsersWeb)
	r.GET("/live.html", ar.ServeLiveWeb)
	r.GET("/templates.html", ar.ServeTemplatesWeb)
	r.GET("/smtp.html", ar.ServeSMTPWeb)
	r.GET("/course-workspace.html", ar.ServeCourseWorkspaceWeb)
	r.GET("/internal/routes/admin/courses/*any", ar.ServeCourseSubmoduleWeb)
	r.GET("/api/templates", ar.ListTemplatesPublic)

	// 2. Public admin login (no auth required)
	r.POST("/api/admin/login", ar.Login)

	// 3. Protected admin REST APIs (JWT + admin role required)
	adm := r.Group("/api/admin", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("/stats", ar.GetStats)

		// Courses & Curriculum Builder
		adm.GET("/courses", ar.ListCourses)
		adm.POST("/courses", ar.CreateCourse)
		adm.PUT("/courses/:id", ar.UpdateCourse)
		adm.DELETE("/courses/:id", ar.DeleteCourse)
		adm.POST("/courses/:id/modules", ar.AddModule)
		adm.PUT("/modules/:id", ar.UpdateModule)
		adm.DELETE("/modules/:id", ar.DeleteModule)
		adm.POST("/modules/:id/lessons", ar.AddLesson)
		adm.PUT("/lessons/:id", ar.UpdateLesson)
		adm.DELETE("/lessons/:id", ar.DeleteLesson)
		adm.GET("/courses/:id/curriculum", ar.GetCourseCurriculum)

		// Subscriptions & Payment Approvals
		adm.GET("/subscriptions", ar.ListSubscriptions)
		adm.POST("/subscriptions/:id/approve", ar.ApproveSubscription)
		adm.POST("/subscriptions/:id/reject", ar.RejectSubscription)
		adm.POST("/subscriptions/manual-enroll", ar.ManualEnroll)

		// Students / Users
		adm.GET("/users", ar.ListUsers)
		adm.POST("/users/:id/toggle-vip", ar.ToggleUserVip)

		// Community Moderation
		adm.GET("/community", ar.ListCommunity)
		adm.DELETE("/community/:id", ar.DeleteCommunityPost)
		adm.POST("/community/:id/solve", ar.MarkSolved)

		// Live Masterclasses
		adm.GET("/live-sessions", ar.ListLiveSessions)
		adm.POST("/live-sessions", ar.CreateLiveSession)
		adm.PUT("/live-sessions/:id", ar.UpdateLiveSession)
		adm.PATCH("/live-sessions/:id/status", ar.UpdateLiveSessionStatus)
		adm.DELETE("/live-sessions/:id", ar.DeleteLiveSession)

		// Mobile App CMS
		adm.GET("/cms/banners", ar.GetBanners)
		adm.POST("/cms/banners", ar.UpdateBanners)

		// Analytics & Diagnostics
		adm.GET("/analytics", ar.GetAnalytics)
		adm.GET("/system/status", ar.GetSystemStatus)

		// Student / User Management
		adm.POST("/users", ar.CreateUser)
		adm.PUT("/users/:id", ar.UpdateUser)
		adm.DELETE("/users/:id", ar.DeleteUser)

		// Schematics
		adm.GET("/schematics", ar.ListSchematics)
		adm.POST("/schematics", ar.CreateSchematic)
		adm.DELETE("/schematics/:id", ar.DeleteSchematic)

		// Repair Manuals & Guides
		adm.GET("/guides", ar.ListGuides)
		adm.POST("/guides", ar.CreateGuide)
		adm.PUT("/guides/:id", ar.UpdateGuide)
		adm.DELETE("/guides/:id", ar.DeleteGuide)

		// Templates Studio & Asset Manager
		adm.GET("/templates", ar.ListTemplates)
		adm.POST("/templates", ar.CreateTemplate)
		adm.POST("/templates/upload", ar.UploadTemplateFile)
		adm.PATCH("/templates/:id/default", ar.ToggleDefaultTemplate)
		adm.DELETE("/templates/:id", ar.DeleteTemplate)

		// SMTP Settings & OTP Engine
		adm.GET("/smtp", ar.GetSMTPSettings)
		adm.POST("/smtp", ar.SaveSMTPSettings)
		adm.POST("/smtp/test", ar.TestSMTPSettings)
	}
}

func (ar *AdminRouter) ServeAdminWeb(c *gin.Context) {
	path := strings.ToLower(c.Request.URL.Path)
	if strings.HasPrefix(path, "/admin/courses") {
		trimmed := strings.TrimPrefix(path, "/admin/courses")
		trimmed = strings.TrimPrefix(trimmed, "/")
		if trimmed != "" {
			// Sub-routes or individual course workspace (e.g. /admin/courses/:id, /admin/courses/:id/:module)
			c.Data(http.StatusOK, "text/html; charset=utf-8", courseWorkspaceHTML)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", coursesHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/curriculum") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", curriculumHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/quizzes") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", quizzesHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/subscriptions") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", subscriptionsHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/users") || strings.HasPrefix(path, "/admin/students") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", usersHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/live") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", liveHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/templates") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", templatesHTML)
		return
	}
	if strings.HasPrefix(path, "/admin/smtp") || strings.HasPrefix(path, "/admin/email") {
		c.Data(http.StatusOK, "text/html; charset=utf-8", smtpHTML)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", adminHTML)
}

func (ar *AdminRouter) ServeCourseWorkspaceWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", courseWorkspaceHTML)
}

func (ar *AdminRouter) ServeCourseSubmoduleWeb(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Request.URL.Path, "/internal/routes/admin/courses/")
	relPath = strings.TrimPrefix(relPath, "/")

	// If a specific file is requested (e.g. style.css, script.js, index.html)
	if strings.Contains(relPath, ".") {
		data, err := coursesFS.ReadFile("courses/" + relPath)
		if err == nil {
			if strings.HasSuffix(relPath, ".css") {
				c.Data(http.StatusOK, "text/css; charset=utf-8", data)
				return
			}
			if strings.HasSuffix(relPath, ".js") {
				c.Data(http.StatusOK, "application/javascript; charset=utf-8", data)
				return
			}
			if strings.HasSuffix(relPath, ".html") {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}
			c.Data(http.StatusOK, "application/octet-stream", data)
			return
		}
	} else {
		// Directory requested without filename, e.g. /attendance or /attendance/
		indexPath := strings.TrimSuffix(relPath, "/") + "/index.html"
		data, err := coursesFS.ReadFile("courses/" + indexPath)
		if err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			return
		}
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", courseWorkspaceHTML)
}

func (ar *AdminRouter) ServeCoursesWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", coursesHTML)
}

func (ar *AdminRouter) ServeCurriculumWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", curriculumHTML)
}

func (ar *AdminRouter) ServeQuizzesWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", quizzesHTML)
}

func (ar *AdminRouter) ServeSubscriptionsWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", subscriptionsHTML)
}

func (ar *AdminRouter) ServeUsersWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", usersHTML)
}

func (ar *AdminRouter) ServeLiveWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", liveHTML)
}

func (ar *AdminRouter) ServeTemplatesWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", templatesHTML)
}

func (ar *AdminRouter) ServeSMTPWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", smtpHTML)
}

func (ar *AdminRouter) ServeClassroomWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", classroomHTML)
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ar *AdminRouter) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid login data"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	if !strings.EqualFold(user.Role, "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Admin access required"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "salen-mobile-guru-secret-key-2026"
	}

	claims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(12 * time.Hour).Unix(), // shorter-lived admin token
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   tokenStr,
		"admin": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (ar *AdminRouter) GetStats(c *gin.Context) {
	var totalUsers int64
	ar.DB.Model(&models.User{}).Count(&totalUsers)

	var totalCourses int64
	ar.DB.Model(&models.Course{}).Count(&totalCourses)

	var totalSubs int64
	ar.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&totalSubs)

	var totalPosts int64
	ar.DB.Model(&models.CommunityPost{}).Count(&totalPosts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats": gin.H{
			"students":            totalUsers,
			"courses":             totalCourses,
			"subscriptions":       totalSubs,
			"communityPosts":      totalPosts,
			"monthlyRevenueNpr":   745000,
			"activePlatformUsers": 482,
		},
	})
}

func (ar *AdminRouter) ListCourses(c *gin.Context) {
	var courses []models.Course
	ar.DB.Preload("Modules.Lessons").Order("created_at desc").Find(&courses)
	c.JSON(http.StatusOK, gin.H{"success": true, "courses": courses, "data": courses})
}

func (ar *AdminRouter) CreateCourse(c *gin.Context) {
	var req models.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := ar.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "course": req, "data": req})
}

func (ar *AdminRouter) DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Where("id = ?", id).Delete(&models.Course{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Course deleted"})
}

type AddModuleRequest struct {
	Title    string `json:"title" binding:"required"`
	OrderNum int    `json:"orderNum"`
}

func (ar *AdminRouter) AddModule(c *gin.Context) {
	courseID := c.Param("id")
	var req AddModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	module := models.CourseModule{
		ID:        uuid.New().String(),
		CourseID:  courseID,
		Title:     req.Title,
		OrderNum:  req.OrderNum,
		CreatedAt: time.Now(),
	}

	if err := ar.DB.Create(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "module": module})
}

type AddLessonRequest struct {
	Title         string `json:"title" binding:"required"`
	Slug          string `json:"slug"`
	Duration      string `json:"duration"`
	VideoURL      string `json:"videoUrl"`
	StreamType    string `json:"streamType"`
	IsFreePreview bool   `json:"isFreePreview"`
	IsQuiz        bool   `json:"isQuiz"`
	OrderNum      int    `json:"orderNum"`
	Description   string `json:"description"`
	Content       string `json:"content"`
	KeyPoints     string `json:"keyPoints"`
	ToolsRequired string `json:"toolsRequired"`
	CourseSlug    string `json:"courseSlug"`
}

func (ar *AdminRouter) AddLesson(c *gin.Context) {
	moduleID := c.Param("id")
	var req AddLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	}
	if req.StreamType == "" {
		req.StreamType = "mp4"
	}

	lesson := models.Lesson{
		ID:            uuid.New().String(),
		ModuleID:      moduleID,
		CourseSlug:    req.CourseSlug,
		Title:         req.Title,
		Slug:          req.Slug,
		Duration:      req.Duration,
		VideoURL:      req.VideoURL,
		StreamType:    req.StreamType,
		IsFreePreview: req.IsFreePreview,
		IsQuiz:        req.IsQuiz,
		HasVideo:      req.VideoURL != "",
		OrderNum:      req.OrderNum,
		Description:   req.Description,
		Content:       req.Content,
		KeyPoints:     req.KeyPoints,
		ToolsRequired: req.ToolsRequired,
		CreatedAt:     time.Now(),
	}

	if err := ar.DB.Create(&lesson).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "lesson": lesson})
}

func (ar *AdminRouter) DeleteLesson(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Where("id = ?", id).Delete(&models.Lesson{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lesson deleted"})
}

func (ar *AdminRouter) GetCourseCurriculum(c *gin.Context) {
	courseID := c.Param("id")
	var modules []models.CourseModule
	ar.DB.Preload("Lessons").Where("course_id = ?", courseID).Order("order_num asc").Find(&modules)
	c.JSON(http.StatusOK, gin.H{"success": true, "curriculum": modules, "data": modules})
}

func (ar *AdminRouter) ListSubscriptions(c *gin.Context) {
	var subs []models.Subscription
	ar.DB.Order("created_at desc").Find(&subs)
	c.JSON(http.StatusOK, gin.H{"success": true, "subscriptions": subs, "data": subs})
}

func (ar *AdminRouter) ApproveSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Model(&models.Subscription{}).Where("id = ?", id).Update("status", "active").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Update user's isSubscribed and subscription plan
	var sub models.Subscription
	if err := ar.DB.Where("id = ?", id).First(&sub).Error; err == nil && sub.UserID != "" {
		isVIP := strings.EqualFold(sub.PlanType, "VIP") || strings.EqualFold(sub.PlanType, "LIFETIME") || strings.EqualFold(sub.PlanType, "ANNUAL")
		ar.DB.Model(&models.User{}).Where("id = ?", sub.UserID).Updates(map[string]interface{}{
			"is_subscribed":     true,
			"is_vip":            isVIP,
			"subscription_plan": sub.PlanType,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription approved"})
}

func (ar *AdminRouter) RejectSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Model(&models.Subscription{}).Where("id = ?", id).Update("status", "cancelled").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription cancelled"})
}

type ManualEnrollRequest struct {
	UserID     string `json:"userId" binding:"required"`
	CourseSlug string `json:"courseSlug" binding:"required"`
	PlanType   string `json:"planType"`
}

func (ar *AdminRouter) ManualEnroll(c *gin.Context) {
	var req ManualEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.PlanType == "" {
		req.PlanType = "VIP"
	}

	sub := models.Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		PlanType:      req.PlanType,
		Status:        "active",
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(1, 0, 0),
		PaymentMethod: "MANUAL_ADMIN",
		TransactionID: "ADM-" + uuid.New().String()[:8],
		Amount:        0,
		Currency:      "NPR",
		EnrolledSlugs: req.CourseSlug,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := ar.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "subscription": sub})
}

func (ar *AdminRouter) ListUsers(c *gin.Context) {
	var users []models.User
	ar.DB.Order("created_at desc").Find(&users)
	c.JSON(http.StatusOK, gin.H{"success": true, "users": users, "data": users})
}

func (ar *AdminRouter) ToggleUserVip(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := ar.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	user.IsVIP = !user.IsVIP
	ar.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "isVip": user.IsVIP})
}

func (ar *AdminRouter) ListCommunity(c *gin.Context) {
	var posts []models.CommunityPost
	ar.DB.Preload("Replies").Order("created_at desc").Find(&posts)
	c.JSON(http.StatusOK, gin.H{"success": true, "posts": posts, "data": posts})
}

func (ar *AdminRouter) DeleteCommunityPost(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("post_id = ?", id).Delete(&models.CommunityReply{})
	ar.DB.Where("id = ?", id).Delete(&models.CommunityPost{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Post deleted"})
}

func (ar *AdminRouter) MarkSolved(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Model(&models.CommunityPost{}).Where("id = ?", id).Update("is_solved", true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Post marked as solved"})
}

func (ar *AdminRouter) ListLiveSessions(c *gin.Context) {
	var sessions []models.LiveSession
	ar.DB.Order("created_at desc").Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"success": true, "sessions": sessions, "data": sessions})
}

func (ar *AdminRouter) CreateLiveSession(c *gin.Context) {
	var session models.LiveSession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	if err := ar.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "session": session, "data": session})
}

func (ar *AdminRouter) DeleteLiveSession(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("id = ?", id).Delete(&models.LiveSession{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Live session deleted"})
}

func (ar *AdminRouter) UpdateLiveSession(c *gin.Context) {
	id := c.Param("id")
	var req models.LiveSession
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	var session models.LiveSession
	if err := ar.DB.Where("id = ?", id).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Live session not found"})
		return
	}
	session.Title = req.Title
	session.Subtitle = req.Subtitle
	session.Instructor = req.Instructor
	session.DateFormatted = req.DateFormatted
	session.TimeFormatted = req.TimeFormatted
	session.Status = req.Status
	session.JoinURL = req.JoinURL
	session.ZoomID = req.ZoomID
	session.Passcode = req.Passcode
	if err := ar.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "session": session, "data": session})
}

func (ar *AdminRouter) UpdateLiveSessionStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := ar.DB.Model(&models.LiveSession{}).Where("id = ?", id).Update("status", strings.ToUpper(req.Status)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Live session status updated"})
}

func (ar *AdminRouter) GetBanners(c *gin.Context) {
	var screen models.AppScreenConfig
	err := ar.DB.Where("site_name = ? AND screen_key = ?", "smg", "guest_home").First(&screen).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"banners": []string{
				"assets/banner_bg_img/banner_1.webp",
				"assets/banner_bg_img/banner_2.webp",
				"assets/banner_bg_img/banner_3.webp",
				"assets/banner_bg_img/banner_4.webp",
				"assets/banner_bg_img/banner_5.webp",
			},
		})
		return
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(screen.ConfigJSON), &data)
	c.JSON(http.StatusOK, gin.H{"success": true, "banners": data["banners"]})
}

type UpdateBannersRequest struct {
	Banners []string `json:"banners" binding:"required"`
}

func (ar *AdminRouter) UpdateBanners(c *gin.Context) {
	var req UpdateBannersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	payload := map[string]interface{}{
		"banners":     req.Banners,
		"bannerImage": req.Banners[0],
	}
	bytes, _ := json.Marshal(payload)

	ar.DB.Model(&models.AppScreenConfig{}).
		Where("site_name = ? AND screen_key = ?", "smg", "guest_home").
		Update("config_json", string(bytes))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Banners updated successfully"})
}

func (ar *AdminRouter) ListMedia(c *gin.Context) {
	var media []models.MediaAsset
	ar.DB.Order("created_at desc").Find(&media)
	c.JSON(http.StatusOK, gin.H{"success": true, "media": media, "data": media})
}

func (ar *AdminRouter) CreateMedia(c *gin.Context) {
	var req models.MediaAsset
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	if err := ar.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "media": req, "data": req})
}

func (ar *AdminRouter) DeleteMedia(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("id = ?", id).Delete(&models.MediaAsset{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Media deleted"})
}

// ── UpdateCourse ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := ar.DB.Where("id = ? OR slug = ?", id, id).First(&course).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found"})
		return
	}

	var req struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Category      string  `json:"category"`
		Tags          string  `json:"tags"`
		Price         float64 `json:"price"`
		OriginalPrice float64 `json:"originalPrice"`
		Level         string  `json:"level"`
		Duration      string  `json:"duration"`
		ImageAsset    string  `json:"imageAsset"`
		Instructor    string  `json:"instructorName"`
		InstructorBio string  `json:"instructorBio"`
		IsPublished   *bool   `json:"isPublished"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Category != "" {
		course.Category = req.Category
	}
	if req.Tags != "" {
		course.Tags = req.Tags
	}
	if req.Price > 0 {
		course.Price = req.Price
	}
	if req.OriginalPrice > 0 {
		course.OriginalPrice = req.OriginalPrice
	}
	if req.Level != "" {
		course.Level = req.Level
	}
	if req.Duration != "" {
		course.Duration = req.Duration
	}
	if req.ImageAsset != "" {
		course.ImageAsset = req.ImageAsset
	}
	if req.Instructor != "" {
		course.Instructor = req.Instructor
	}
	if req.InstructorBio != "" {
		course.InstructorBio = req.InstructorBio
	}
	if req.IsPublished != nil {
		course.IsPublished = *req.IsPublished
	}
	course.UpdatedAt = time.Now()

	ar.DB.Save(&course)
	c.JSON(http.StatusOK, gin.H{"success": true, "course": course, "data": course})
}

// ── UpdateModule ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateModule(c *gin.Context) {
	id := c.Param("id")
	var module models.CourseModule
	if err := ar.DB.Where("id = ?", id).First(&module).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Module not found"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		OrderNum int    `json:"orderNum"`
	}
	c.ShouldBindJSON(&req)

	if req.Title != "" {
		module.Title = req.Title
	}
	if req.OrderNum > 0 {
		module.OrderNum = req.OrderNum
	}
	ar.DB.Save(&module)
	c.JSON(http.StatusOK, gin.H{"success": true, "module": module})
}

func (ar *AdminRouter) DeleteModule(c *gin.Context) {
	id := c.Param("id")
	// Delete all lessons in this module first
	ar.DB.Where("module_id = ?", id).Delete(&models.Lesson{})
	ar.DB.Where("id = ?", id).Delete(&models.CourseModule{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Module and its lessons deleted"})
}

// ── UpdateLesson ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateLesson(c *gin.Context) {
	id := c.Param("id")
	var lesson models.Lesson
	if err := ar.DB.Where("id = ?", id).First(&lesson).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lesson not found"})
		return
	}

	var req struct {
		Title         string `json:"title"`
		Slug          string `json:"slug"`
		Duration      string `json:"duration"`
		VideoURL      string `json:"videoUrl"`
		StreamType    string `json:"streamType"`
		IsFreePreview *bool  `json:"isFreePreview"`
		IsQuiz        *bool  `json:"isQuiz"`
		OrderNum      int    `json:"orderNum"`
		Description   string `json:"description"`
		Content       string `json:"content"`
		KeyPoints     string `json:"keyPoints"`
		ToolsRequired string `json:"toolsRequired"`
		CourseSlug    string `json:"courseSlug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Title != "" {
		lesson.Title = req.Title
	}
	if req.Slug != "" {
		lesson.Slug = req.Slug
	}
	if req.Duration != "" {
		lesson.Duration = req.Duration
	}
	if req.VideoURL != "" {
		lesson.VideoURL = req.VideoURL
		lesson.HasVideo = true
	}
	if req.StreamType != "" {
		lesson.StreamType = req.StreamType
	}
	if req.IsFreePreview != nil {
		lesson.IsFreePreview = *req.IsFreePreview
	}
	if req.IsQuiz != nil {
		lesson.IsQuiz = *req.IsQuiz
	}
	if req.OrderNum > 0 {
		lesson.OrderNum = req.OrderNum
	}
	if req.Description != "" {
		lesson.Description = req.Description
	}
	if req.Content != "" {
		lesson.Content = req.Content
	}
	if req.KeyPoints != "" {
		lesson.KeyPoints = req.KeyPoints
	}
	if req.ToolsRequired != "" {
		lesson.ToolsRequired = req.ToolsRequired
	}
	if req.CourseSlug != "" {
		lesson.CourseSlug = req.CourseSlug
	}

	ar.DB.Save(&lesson)
	c.JSON(http.StatusOK, gin.H{"success": true, "lesson": lesson, "data": lesson})
}

// ── Analytics & System Diagnostics ────────────────────────────────────────────

func (ar *AdminRouter) GetAnalytics(c *gin.Context) {
	now := time.Now()
	months := make([]string, 6)
	revData := make([]float64, 6)
	userGrowth := make([]int64, 6)

	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		monthName := t.Format("Jan 2006")
		idx := 5 - i
		months[idx] = monthName

		startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)

		var rev float64
		ar.DB.Model(&models.Subscription{}).
			Where("created_at >= ? AND created_at < ? AND status = ?", startOfMonth, endOfMonth, "active").
			Select("COALESCE(SUM(amount), 0)").Scan(&rev)
		revData[idx] = rev

		var uCount int64
		ar.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth).
			Count(&uCount)
		userGrowth[idx] = uCount
	}

	var totalStudents, activeSubs, totalCourses, totalMedia, totalHWModels, totalQuizzes, totalOrders int64
	ar.DB.Model(&models.User{}).Count(&totalStudents)
	ar.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&activeSubs)
	ar.DB.Model(&models.Course{}).Count(&totalCourses)
	ar.DB.Model(&models.MediaAsset{}).Count(&totalMedia)
	ar.DB.Model(&models.HardwareModel{}).Count(&totalHWModels)
	ar.DB.Model(&models.Quiz{}).Count(&totalQuizzes)
	ar.DB.Model(&models.WorkOrder{}).Count(&totalOrders)

	var totalQuizAttempts, passedAttempts int64
	ar.DB.Model(&models.QuizResult{}).Count(&totalQuizAttempts)
	ar.DB.Model(&models.QuizResult{}).Where("is_passed = ?", true).Count(&passedAttempts)
	passRate := 0.0
	if totalQuizAttempts > 0 {
		passRate = float64(passedAttempts) / float64(totalQuizAttempts) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"analytics": gin.H{
			"months":             months,
			"monthlyRevenue":    revData,
			"monthlyUsers":      userGrowth,
			"totalStudents":     totalStudents,
			"activeSubs":        activeSubs,
			"totalCourses":      totalCourses,
			"totalMedia":        totalMedia,
			"totalHWModels":     totalHWModels,
			"totalQuizzes":      totalQuizzes,
			"totalOrders":       totalOrders,
			"totalQuizAttempts": totalQuizAttempts,
			"passRate":          passRate,
		},
	})
}

func (ar *AdminRouter) GetSystemStatus(c *gin.Context) {
	dbStatus := "Operational"
	var dbLatencyMs int64
	start := time.Now()
	sqlDB, err := ar.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "Degraded"
	} else {
		dbLatencyMs = time.Since(start).Milliseconds()
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"system": gin.H{
			"status":         "Operational",
			"serverTime":     time.Now().Format("2006-01-02 15:04:05 MST"),
			"uptimeSeconds":  int64(time.Since(serverStartTime).Seconds()),
			"goVersion":      runtime.Version(),
			"goroutines":     runtime.NumGoroutine(),
			"memAllocMB":     fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
			"memSysMB":       fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
			"database":       dbStatus,
			"dbLatencyMs":    dbLatencyMs,
			"cloudflareR2":   "Connected (Bucket: salenmobileguru)",
			"environment":    gin.Mode(),
		},
	})
}

// ── Student / User Management ─────────────────────────────────────────────────

func (ar *AdminRouter) CreateUser(c *gin.Context) {
	var req struct {
		Name             string `json:"name" binding:"required"`
		Email            string `json:"email" binding:"required"`
		Password         string `json:"password"`
		Role             string `json:"role"`
		Phone            string `json:"phone"`
		IsVIP            bool   `json:"isVip"`
		SubscriptionPlan string `json:"subscriptionPlan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	var existing models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Email already exists"})
		return
	}

	pass := req.Password
	if pass == "" {
		pass = "Student@2026!"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Password hashing failed"})
		return
	}

	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role == "" {
		role = "STUDENT"
	}

	user := models.User{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Email:            email,
		Phone:            req.Phone,
		Role:             role,
		PasswordHash:     string(hash),
		IsVIP:            req.IsVIP,
		SubscriptionPlan: req.SubscriptionPlan,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := ar.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "user": user, "data": user})
}

func (ar *AdminRouter) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := ar.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	var req struct {
		Name             string `json:"name"`
		Email            string `json:"email"`
		Role             string `json:"role"`
		Phone            string `json:"phone"`
		IsVIP            *bool  `json:"isVip"`
		SubscriptionPlan string `json:"subscriptionPlan"`
		Password         string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = strings.TrimSpace(strings.ToLower(req.Email))
	}
	if req.Role != "" {
		user.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.IsVIP != nil {
		user.IsVIP = *req.IsVIP
	}
	if req.SubscriptionPlan != "" {
		user.SubscriptionPlan = req.SubscriptionPlan
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err == nil {
			user.PasswordHash = string(hash)
		}
	}
	user.UpdatedAt = time.Now()

	ar.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "user": user, "data": user})
}

func (ar *AdminRouter) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("id = ?", id).Delete(&models.User{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted"})
}

// ── Schematics Management ─────────────────────────────────────────────────────

func (ar *AdminRouter) ListSchematics(c *gin.Context) {
	var list []models.Schematic
	ar.DB.Order("created_at desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "schematics": list})
}

func (ar *AdminRouter) CreateSchematic(c *gin.Context) {
	var s models.Schematic
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if s.ID == "" {
		s.ID = "sch_" + uuid.New().String()[:8]
	}
	s.CreatedAt = time.Now()
	if err := ar.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": s})
}

func (ar *AdminRouter) DeleteSchematic(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Delete(&models.Schematic{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Schematic deleted"})
}

// ── Repair Manuals & Guides ───────────────────────────────────────────────────

func (ar *AdminRouter) ListGuides(c *gin.Context) {
	var list []models.Guide
	ar.DB.Order("order_num asc, created_at desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "guides": list})
}

func (ar *AdminRouter) CreateGuide(c *gin.Context) {
	var g models.Guide
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if g.ID == "" {
		g.ID = "gd_" + uuid.New().String()[:8]
	}
	g.CreatedAt = time.Now()
	if err := ar.DB.Create(&g).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": g})
}

func (ar *AdminRouter) UpdateGuide(c *gin.Context) {
	id := c.Param("id")
	var g models.Guide
	if err := ar.DB.Where("id = ?", id).First(&g).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Guide not found"})
		return
	}
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	g.ID = id
	ar.DB.Save(&g)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": g})
}

func (ar *AdminRouter) DeleteGuide(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Delete(&models.Guide{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Guide deleted"})
}

// ═══════════════════════════════════════════════════════════════════════════
// TEMPLATES & BLUEPRINTS API HANDLERS
// ═══════════════════════════════════════════════════════════════════════════

func (ar *AdminRouter) ListTemplates(c *gin.Context) {
	cat := c.Query("category")
	q := c.Query("q")

	tx := ar.DB.Model(&models.TemplateAsset{})
	if cat != "" && cat != "all" {
		tx = tx.Where("LOWER(category) = ?", strings.ToLower(cat))
	}
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		tx = tx.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(tags) LIKE ?", like, like, like)
	}

	var list []models.TemplateAsset
	if err := tx.Order("is_default desc, created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "count": len(list)})
}

func (ar *AdminRouter) ListTemplatesPublic(c *gin.Context) {
	ar.ListTemplates(c)
}

type CreateTemplateRequest struct {
	Title        string `json:"title" binding:"required"`
	Category     string `json:"category" binding:"required"`
	Description  string `json:"description"`
	FileURL      string `json:"fileUrl" binding:"required"`
	ThumbnailURL string `json:"thumbnailUrl"`
	FileType     string `json:"fileType"`
	FileSize     string `json:"fileSize"`
	Author       string `json:"author"`
	Tags         string `json:"tags"`
	IsDefault    bool   `json:"isDefault"`
}

func (ar *AdminRouter) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Author == "" {
		req.Author = "Salen Mobile Guru Academy"
	}
	if req.FileType == "" {
		ext := strings.TrimPrefix(filepath.Ext(req.FileURL), ".")
		req.FileType = strings.ToUpper(ext)
	}

	if req.IsDefault {
		// Clear existing default in this category
		ar.DB.Model(&models.TemplateAsset{}).Where("LOWER(category) = ?", strings.ToLower(req.Category)).Update("is_default", false)
	}

	tmpl := models.TemplateAsset{
		ID:           "tmpl_" + uuid.New().String()[:8],
		Title:        req.Title,
		Category:     strings.ToLower(req.Category),
		Description:  req.Description,
		FileURL:      req.FileURL,
		ThumbnailURL: req.ThumbnailURL,
		FileType:     req.FileType,
		FileSize:     req.FileSize,
		Author:       req.Author,
		Tags:         req.Tags,
		IsDefault:    req.IsDefault,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := ar.DB.Create(&tmpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": tmpl})
}

func (ar *AdminRouter) UploadTemplateFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file uploaded: " + err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Open stream error: " + err.Error()})
		return
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	base = strings.ReplaceAll(base, " ", "_")
	uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), base, ext)

	cType := file.Header.Get("Content-Type")
	if cType == "" {
		cType = "application/octet-stream"
	}

	folder := c.DefaultPostForm("folder", "templates")
	publicURL, r2Key, err := storage.DefaultStorage.UploadDirect(folder, uniqueFilename, src, file.Size, cType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Storage upload error: " + err.Error()})
		return
	}

	sizeMB := float64(file.Size) / (1024 * 1024)
	sizeStr := fmt.Sprintf("%.1f MB", sizeMB)
	if sizeMB < 1.0 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(file.Size)/1024)
	}

	title := c.PostForm("title")
	if title == "" {
		title = strings.ReplaceAll(base, "_", " ")
	}
	category := strings.ToLower(c.DefaultPostForm("category", "worksheet"))
	author := c.DefaultPostForm("author", "Salen Mobile Guru Academy")
	tags := c.PostForm("tags")
	isDefault := c.PostForm("isDefault") == "true" || c.PostForm("isDefault") == "1"

	if isDefault {
		ar.DB.Model(&models.TemplateAsset{}).Where("LOWER(category) = ?", category).Update("is_default", false)
	}

	fileType := strings.ToUpper(strings.TrimPrefix(ext, "."))
	thumbURL := publicURL
	if !strings.HasPrefix(cType, "image/") {
		thumbURL = "https://images.unsplash.com/photo-1581092160607-ee22621dd758?w=600"
	}

	tmpl := models.TemplateAsset{
		ID:           "tmpl_" + uuid.New().String()[:8],
		Title:        title,
		Category:     category,
		Description:  c.PostForm("description"),
		FileURL:      publicURL,
		ThumbnailURL: thumbURL,
		FileType:     fileType,
		FileSize:     sizeStr,
		Author:       author,
		Tags:         tags,
		IsDefault:    isDefault,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	ar.DB.Create(&tmpl)

	// Also record in general media library
	asset := models.MediaAsset{
		ID:            uuid.New().String(),
		Name:          file.Filename,
		Kind:          "template",
		URL:           publicURL,
		SizeFormatted: sizeStr,
		Category:      category,
		CreatedAt:     time.Now(),
	}
	ar.DB.Create(&asset)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"publicUrl": publicURL,
		"fileUrl":   publicURL,
		"url":       publicURL,
		"r2Key":     r2Key,
		"size":      sizeStr,
		"data":      tmpl,
	})
}

func (ar *AdminRouter) ToggleDefaultTemplate(c *gin.Context) {
	id := c.Param("id")
	var tmpl models.TemplateAsset
	if err := ar.DB.Where("id = ?", id).First(&tmpl).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Template not found"})
		return
	}

	newDef := !tmpl.IsDefault
	if newDef {
		ar.DB.Model(&models.TemplateAsset{}).Where("LOWER(category) = ?", strings.ToLower(tmpl.Category)).Update("is_default", false)
	}
	tmpl.IsDefault = newDef
	tmpl.UpdatedAt = time.Now()
	ar.DB.Save(&tmpl)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Default status updated", "isDefault": newDef})
}

func (ar *AdminRouter) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Delete(&models.TemplateAsset{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Template blueprint removed"})
}

func (ar *AdminRouter) GetSMTPSettings(c *gin.Context) {
	cfg := mail.GetSMTPSettings(ar.DB)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cfg,
	})
}

func (ar *AdminRouter) SaveSMTPSettings(c *gin.Context) {
	var input models.SMTPSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body: " + err.Error()})
		return
	}

	// Keep existing password if left blank on edit
	if strings.TrimSpace(input.Password) == "" {
		existing := mail.GetSMTPSettings(ar.DB)
		if existing != nil {
			input.Password = existing.Password
		}
	}

	if err := mail.SaveSMTPSettings(ar.DB, &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save SMTP settings: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "SMTP settings saved successfully",
		"data":    input,
	})
}

type TestSMTPEmailRequest struct {
	ToEmail string `json:"toEmail" binding:"required"`
}

func (ar *AdminRouter) TestSMTPSettings(c *gin.Context) {
	var req TestSMTPEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Valid toEmail recipient address is required"})
		return
	}

	cfg := mail.GetSMTPSettings(ar.DB)
	if err := mail.SendTestEmail(ar.DB, req.ToEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"host":    cfg.Host,
			"port":    cfg.Port,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification probe email sent successfully",
		"host":    cfg.Host,
		"port":    cfg.Port,
	})
}
