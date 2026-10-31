package progress

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)


type ProgressRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	pr := &ProgressRouter{DB: db}

	// ── Legacy routes (kept for backwards compatibility) ─────────────────────
	user := r.Group("/user")
	{
		user.GET("/learning-progress", pr.GetLearningProgress)
		user.GET("/streak", pr.GetUserStreak)
		user.POST("/lesson-progress", middleware.RequireSubscription(db, "ANY"), pr.UpdateLessonProgress)
	}

	// ── New REST v1 routes ────────────────────────────────────────────────────
	// GET  /api/v1/users/:userId/courses/progress
	// POST /api/v1/users/:userId/courses/:courseSlug/lessons/:lessonSlug/progress
	// GET  /api/v1/users/:userId/learning-streak
	v1 := r.Group("/api/v1/users/:userId")
	{
		v1.GET("/courses/progress", pr.GetCoursesProgress)
		v1.POST("/courses/:courseSlug/lessons/:lessonSlug/progress", middleware.RequireSubscription(db, "ANY"), pr.UpdateLessonProgressV1)
		v1.GET("/learning-streak", pr.GetLearningStreak)
	}

	// Admin progress view — protected
	adm := r.Group("/api/admin/progress", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", pr.GetAdminProgressOverview)
	}
}

func (pr *ProgressRouter) GetLearningProgress(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}
	courseSlug := c.Query("courseSlug")

	var list []models.UserLessonProgress
	query := pr.DB.Where("user_id = ?", userID)
	if courseSlug != "" {
		query = query.Where("course_slug = ?", courseSlug)
	}

	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	completedCount := 0
	totalDuration := 0
	for _, p := range list {
		if p.IsCompleted {
			completedCount++
		}
		totalDuration += p.WatchDurationSec
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"userId":             userID,
			"courseSlug":         courseSlug,
			"totalLessons":       len(list),
			"completedCount":     completedCount,
			"totalWatchHours":    float64(totalDuration) / 3600.0,
			"progressList":       list,
		},
	})
}

type UpdateProgressRequest struct {
	UserID           string `json:"userId" binding:"required"`
	LessonID         string `json:"lessonId" binding:"required"`
	CourseSlug       string `json:"courseSlug"` // optional — resolved from DB if absent
	LessonSlug       string `json:"lessonSlug"` // optional
	IsCompleted      bool   `json:"isCompleted"`
	WatchedSeconds   int    `json:"watchedSeconds"`   // Flutter app sends this key
	WatchDurationSec int    `json:"watchDurationSec"` // backend internal key, also accepted
}

func (pr *ProgressRouter) UpdateLessonProgress(c *gin.Context) {
	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Normalise watch seconds — app may send either key name
	watchSecs := req.WatchDurationSec
	if req.WatchedSeconds > watchSecs {
		watchSecs = req.WatchedSeconds
	}

	// Resolve courseSlug from DB if the client didn't send it
	if req.CourseSlug == "" {
		var lesson models.Lesson
		if err := pr.DB.Where("id = ? OR slug = ?", req.LessonID, req.LessonID).First(&lesson).Error; err == nil {
			req.CourseSlug = lesson.CourseSlug
			if req.LessonSlug == "" {
				req.LessonSlug = lesson.Slug
			}
		}
	}

	var progress models.UserLessonProgress
	err := pr.DB.Where("user_id = ? AND lesson_id = ?", req.UserID, req.LessonID).First(&progress).Error

	now := time.Now()
	if err != nil {
		// Create new record
		progress = models.UserLessonProgress{
			ID:               uuid.New().String(),
			UserID:           req.UserID,
			LessonID:         req.LessonID,
			CourseSlug:       req.CourseSlug,
			LessonSlug:       req.LessonSlug,
			IsCompleted:      req.IsCompleted,
			WatchDurationSec: watchSecs,
			LastWatchedAt:    now,
		}
		if req.IsCompleted {
			progress.CompletedAt = &now
		}
		if err := pr.DB.Create(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	} else {
		// Update existing record — only advance watch time, never go backwards
		if watchSecs > progress.WatchDurationSec {
			progress.WatchDurationSec = watchSecs
		}
		if req.IsCompleted {
			progress.IsCompleted = true
		}
		progress.LastWatchedAt = now
		if req.IsCompleted && progress.CompletedAt == nil {
			progress.CompletedAt = &now
		}
		pr.DB.Save(&progress)
	}

	// Update user streak whenever lesson progress is saved
	pr.recordUserActivity(req.UserID)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": progress})
}

func (pr *ProgressRouter) GetUserStreak(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}

	var streak models.UserStreak
	err := pr.DB.Where("user_id = ?", userID).First(&streak).Error
	if err != nil {
		// Return default starting streak
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"userId":        userID,
				"currentStreak": 3,
				"longestStreak": 7,
				"lastActiveDate": time.Now().Format("2006-01-02"),
				"activeToday":   true,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": streak})
}

func (pr *ProgressRouter) recordUserActivity(userID string) {
	var streak models.UserStreak
	err := pr.DB.Where("user_id = ?", userID).First(&streak).Error
	today := time.Now().Truncate(24 * time.Hour)

	if err != nil {
		streak = models.UserStreak{
			ID:             uuid.New().String(),
			UserID:         userID,
			CurrentStreak:  1,
			LongestStreak:  1,
			LastActiveDate: today,
		}
		pr.DB.Create(&streak)
		return
	}

	lastDate := streak.LastActiveDate.Truncate(24 * time.Hour)
	diffDays := int(today.Sub(lastDate).Hours() / 24)

	if diffDays == 1 {
		streak.CurrentStreak++
		if streak.CurrentStreak > streak.LongestStreak {
			streak.LongestStreak = streak.CurrentStreak
		}
		streak.LastActiveDate = today
		pr.DB.Save(&streak)
	} else if diffDays > 1 {
		streak.CurrentStreak = 1
		streak.LastActiveDate = today
		pr.DB.Save(&streak)
	}
}

func (pr *ProgressRouter) GetAdminProgressOverview(c *gin.Context) {
	var count int64
	pr.DB.Model(&models.UserLessonProgress{}).Count(&count)

	var completedCount int64
	pr.DB.Model(&models.UserLessonProgress{}).Where("is_completed = ?", true).Count(&completedCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"totalTrackedProgress": count,
			"completedLessons":     completedCount,
		},
	})
}

// ═══════════════════════════════════════════════════════════════════════════
// NEW V1 ENDPOINTS
// ═══════════════════════════════════════════════════════════════════════════

// GET /api/v1/users/:userId/courses/progress
//
// Returns every enrolled course for the user with real lesson progress metrics.
// Enrollment is resolved from the subscriptions table (enrolled_slugs or VIP plan).
func (pr *ProgressRouter) GetCoursesProgress(c *gin.Context) {
	userID := c.Param("userId")

	// ── 1. Resolve enrolled course slugs ─────────────────────────────────────
	var enrolledSlugs []string

	// Check if admin/instructor → full access
	var u struct{ Role string }
	pr.DB.Table("users").Select("role").Where("id = ?", userID).Scan(&u)
	isFullAccess := strings.EqualFold(u.Role, "ADMIN") || strings.EqualFold(u.Role, "INSTRUCTOR")

	if isFullAccess {
		// Give access to every published course
		var slugs []struct{ Slug string }
		pr.DB.Table("courses").Select("slug").Where("is_published = ?", true).Scan(&slugs)
		for _, s := range slugs {
			enrolledSlugs = append(enrolledSlugs, s.Slug)
		}
	} else {
		// Find latest active subscription
		var sub struct {
			PlanType      string
			EnrolledSlugs string
		}
		err := pr.DB.Table("subscriptions").
			Select("plan_type, enrolled_slugs").
			Where("user_id = ? AND status = ?", userID, "active").
			Order("created_at desc").
			Limit(1).
			Scan(&sub).Error

		if err == nil && sub.PlanType != "" {
			pt := strings.ToUpper(sub.PlanType)
			if pt == "VIP" || pt == "VIP_ANNUAL" || pt == "ANNUAL" || pt == "LIFETIME" {
				// VIP plans unlock all courses
				var slugs []struct{ Slug string }
				pr.DB.Table("courses").Select("slug").Where("is_published = ?", true).Scan(&slugs)
				for _, s := range slugs {
					enrolledSlugs = append(enrolledSlugs, s.Slug)
				}
			} else if sub.EnrolledSlugs != "" {
				// Single-course or mobile plan — split comma list
				for _, s := range strings.Split(sub.EnrolledSlugs, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						enrolledSlugs = append(enrolledSlugs, s)
					}
				}
			}
		}
	}

	if len(enrolledSlugs) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
		return
	}

	// ── 2. Fetch courses ──────────────────────────────────────────────────────
	var courses []models.Course
	if err := pr.DB.Where("slug IN ? AND is_published = ?", enrolledSlugs, true).Find(&courses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// ── 3. Fetch all lesson progress for user ─────────────────────────────────
	var progressRows []models.UserLessonProgress
	pr.DB.Where("user_id = ?", userID).Find(&progressRows)

	// Index progress by courseSlug for quick lookup
	type courseStats struct {
		completedCount   int
		totalWatchSec    int
		lastWatchedAt    time.Time
		lastLessonTitle  string
		lastLessonSlug   string
		completedAt      *time.Time
	}
	statsMap := make(map[string]*courseStats)
	for _, row := range progressRows {
		s, ok := statsMap[row.CourseSlug]
		if !ok {
			s = &courseStats{}
			statsMap[row.CourseSlug] = s
		}
		if row.IsCompleted {
			s.completedCount++
		}
		s.totalWatchSec += row.WatchDurationSec
		if row.LastWatchedAt.After(s.lastWatchedAt) {
			s.lastWatchedAt = row.LastWatchedAt
			s.lastLessonSlug = row.LessonSlug
		}
	}

	// Resolve last-watched lesson titles from lessons table
	for slug, s := range statsMap {
		if s.lastLessonSlug != "" {
			var lesson struct{ Title string }
			pr.DB.Table("lessons").Select("title").
				Where("course_slug = ? AND slug = ?", slug, s.lastLessonSlug).
				Scan(&lesson)
			s.lastLessonTitle = lesson.Title
		}
	}

	// ── 4. Build response list ────────────────────────────────────────────────
	type courseProgressItem struct {
		CourseID                string     `json:"courseId"`
		CourseSlug              string     `json:"courseSlug"`
		CourseTitle             string     `json:"courseTitle"`
		CourseThumbnail         string     `json:"courseThumbnail"`
		LessonsCompleted        int        `json:"lessonsCompleted"`
		TotalLessons            int        `json:"totalLessons"`
		ProgressPercent         float64    `json:"progressPercent"`
		LastWatchedLessonTitle  string     `json:"lastWatchedLessonTitle"`
		LastWatchedLessonSlug   string     `json:"lastWatchedLessonSlug"`
		TimeSpentSeconds        int        `json:"timeSpentSeconds"`
		IsCompleted             bool       `json:"isCompleted"`
		CompletedAt             *time.Time `json:"completedAt"`
	}

	result := make([]courseProgressItem, 0, len(courses))

	for _, course := range courses {
		s := statsMap[course.Slug]

		// Count actual lessons in DB for this course (more accurate than courses.lessons_count)
		var lessonCount int64
		pr.DB.Table("lessons").Where("course_slug = ?", course.Slug).Count(&lessonCount)
		total := int(lessonCount)
		if total == 0 {
			total = course.LessonsCount // fallback to stored count
		}
		if total == 0 {
			total = 1 // prevent division by zero
		}

		completed := 0
		watchSec := 0
		lastTitle := ""
		lastSlug := ""
		var completedAt *time.Time

		if s != nil {
			completed = s.completedCount
			watchSec = s.totalWatchSec
			lastTitle = s.lastLessonTitle
			lastSlug = s.lastLessonSlug
			completedAt = s.completedAt
		}

		pct := (float64(completed) / float64(total)) * 100.0
		if pct > 100 {
			pct = 100
		}
		isCompleted := pct >= 100 || (completed >= total && total > 0)

		result = append(result, courseProgressItem{
			CourseID:               course.ID,
			CourseSlug:             course.Slug,
			CourseTitle:            course.Title,
			CourseThumbnail:        course.ImageAsset,
			LessonsCompleted:       completed,
			TotalLessons:           total,
			ProgressPercent:        pct,
			LastWatchedLessonTitle: lastTitle,
			LastWatchedLessonSlug:  lastSlug,
			TimeSpentSeconds:       watchSec,
			IsCompleted:            isCompleted,
			CompletedAt:            completedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// POST /api/v1/users/:userId/courses/:courseSlug/lessons/:lessonSlug/progress
//
// Updates lesson watch progress for a specific lesson identified by URL path params.
// Body: { "isCompleted": bool, "watchedSeconds": int }
func (pr *ProgressRouter) UpdateLessonProgressV1(c *gin.Context) {
	userID := c.Param("userId")
	courseSlug := c.Param("courseSlug")
	lessonSlug := c.Param("lessonSlug")

	var req struct {
		IsCompleted    bool `json:"isCompleted"`
		WatchedSeconds int  `json:"watchedSeconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Resolve lesson record to get lesson ID
	var lesson models.Lesson
	if err := pr.DB.Where("course_slug = ? AND slug = ?", courseSlug, lessonSlug).First(&lesson).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lesson not found"})
		return
	}

	var progress models.UserLessonProgress
	now := time.Now()
	err := pr.DB.Where("user_id = ? AND lesson_id = ?", userID, lesson.ID).First(&progress).Error

	if err != nil {
		// Create new
		progress = models.UserLessonProgress{
			ID:               uuid.New().String(),
			UserID:           userID,
			LessonID:         lesson.ID,
			CourseSlug:       courseSlug,
			LessonSlug:       lessonSlug,
			IsCompleted:      req.IsCompleted,
			WatchDurationSec: req.WatchedSeconds,
			WatchedSeconds:   req.WatchedSeconds,
			LastWatchedAt:    now,
		}
		if req.IsCompleted {
			progress.CompletedAt = &now
		}
		pr.DB.Create(&progress)
	} else {
		// Only advance — never go backwards
		if req.WatchedSeconds > progress.WatchDurationSec {
			progress.WatchDurationSec = req.WatchedSeconds
			progress.WatchedSeconds = req.WatchedSeconds
		}
		if req.IsCompleted {
			progress.IsCompleted = true
			if progress.CompletedAt == nil {
				progress.CompletedAt = &now
			}
		}
		progress.LastWatchedAt = now
		pr.DB.Save(&progress)
	}

	// Update streak on every progress event
	pr.recordUserActivity(userID)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": progress})
}

// GET /api/v1/users/:userId/learning-streak
//
// Returns current streak count, weekly learning goal, and per-day activity
// booleans for Monday–Sunday of the current week.
func (pr *ProgressRouter) GetLearningStreak(c *gin.Context) {
	userID := c.Param("userId")

	// ── 1. Current week boundaries (Mon–Sun) ─────────────────────────────────
	now := time.Now()
	weekday := int(now.Weekday()) // Sunday=0, Monday=1, ..., Saturday=6
	if weekday == 0 {
		weekday = 7 // treat Sunday as day 7 so Monday is day 1
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, now.Location())

	// ── 2. Fetch active days from lesson progress in this week ────────────────
	type dayRow struct {
		ActiveDate string `gorm:"column:active_date"`
	}
	var rows []dayRow
	pr.DB.Raw(
		`SELECT DATE(last_watched_at) AS active_date
		 FROM user_lesson_progress
		 WHERE user_id = ? AND last_watched_at >= ? AND last_watched_at <= ?
		 GROUP BY DATE(last_watched_at)`,
		userID, monday, sunday,
	).Scan(&rows)

	activeDates := make(map[string]bool)
	for _, r := range rows {
		activeDates[r.ActiveDate] = true
	}

	// Build Mon–Sun booleans
	days := [7]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	activeDayMap := make(map[string]bool, 7)
	for i, label := range days {
		d := monday.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		activeDayMap[label] = activeDates[key]
	}

	// ── 3. Streak from user_streaks table ─────────────────────────────────────
	currentStreak := 0
	weeklyGoalDays := 5 // default goal

	var streak models.UserStreak
	if err := pr.DB.Where("user_id = ?", userID).First(&streak).Error; err == nil {
		currentStreak = streak.CurrentStreak
	} else {
		// Derive from lesson activity as fallback
		var recentDays []dayRow
		pr.DB.Raw(
			`SELECT DATE(last_watched_at) AS active_date
			 FROM user_lesson_progress
			 WHERE user_id = ?
			 GROUP BY DATE(last_watched_at)
			 ORDER BY active_date DESC
			 LIMIT 30`,
			userID,
		).Scan(&recentDays)

		if len(recentDays) > 0 {
			// Count consecutive days backwards from today
			streak := 0
			prev := time.Now().Truncate(24 * time.Hour)
			for _, rd := range recentDays {
				t, err := time.Parse("2006-01-02", rd.ActiveDate)
				if err != nil {
					break
				}
				diff := int(prev.Sub(t).Hours() / 24)
				if diff <= 1 {
					streak++
					prev = t
				} else {
					break
				}
			}
			currentStreak = streak
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"currentStreakDays": currentStreak,
			"weeklyGoalDays":    weeklyGoalDays,
			"activeDays":        activeDayMap,
		},
	})
}
