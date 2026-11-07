package live

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

type LiveRouter struct {
	DB *gorm.DB
}

// platformLabel returns a human-readable platform name.
func platformLabel(p string) string {
	switch strings.ToUpper(p) {
	case "GOOGLE_MEET":
		return "Google Meet"
	case "TEAMS":
		return "Microsoft Teams"
	case "YOUTUBE_LIVE":
		return "YouTube Live"
	default:
		return "Zoom"
	}
}

// platformIcon returns an icon key the mobile app can resolve.
func platformIcon(p string) string {
	switch strings.ToUpper(p) {
	case "GOOGLE_MEET":
		return "google_meet"
	case "TEAMS":
		return "teams"
	case "YOUTUBE_LIVE":
		return "youtube"
	default:
		return "zoom"
	}
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	lr := &LiveRouter{DB: db}
	lr.seedSampleSessions()

	// ── Public endpoints (no auth) ─────────────────────────────────────────
	// Returns session list without join URLs / passcodes (safe for public)
	r.GET("/live/sessions", lr.GetLiveSessions)

	// Schedule endpoint: splits sessions into upcoming / live / past
	r.GET("/live/schedule", lr.GetSchedule)

	// ── Authenticated endpoints (JWT required) ────────────────────────────
	// Returns full session details including joinUrl and passcode
	auth := r.Group("", middleware.AuthMiddleware())
	{
		auth.GET("/live/sessions/:id/join", lr.GetJoinDetails)
	}

	// ── Admin management (JWT + admin role) ──────────────────────────────
	adm := r.Group("/api/admin/live", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.POST("/sessions", lr.CreateSession)
		adm.PUT("/sessions/:id", lr.UpdateSession)
		adm.PATCH("/sessions/:id/status", lr.PatchSessionStatus)
		adm.DELETE("/sessions/:id", lr.DeleteSession)
	}
}

// ─── sessionPublic is the safe public-facing shape (no joinUrl / passcode) ──

type sessionPublic struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Instructor   string `json:"instructor"`
	Date         string `json:"date"`
	Time         string `json:"time"`
	Status       string `json:"status"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platformName"`
	PlatformIcon string `json:"platformIcon"`
	ImageAsset   string `json:"imageAsset"`
	DurationMins int    `json:"durationMins"`
	Viewers      int    `json:"viewersCount"`
	RecordingURL string `json:"recordingUrl,omitempty"`
	// joinUrl / meetingId / passcode intentionally omitted here
}

func toPublic(s models.LiveSession) sessionPublic {
	platform := s.Platform
	if platform == "" {
		if s.ZoomID != "" {
			platform = "ZOOM"
		} else {
			platform = "ZOOM"
		}
	}
	img := s.ImageAsset
	if img == "" {
		img = "assets/banner_bg_img/banner_1.webp"
	}
	out := sessionPublic{
		ID:           s.ID,
		Title:        s.Title,
		Subtitle:     s.Subtitle,
		Instructor:   s.Instructor,
		Date:         s.DateFormatted,
		Time:         s.TimeFormatted,
		Status:       s.Status,
		Platform:     platform,
		PlatformName: platformLabel(platform),
		PlatformIcon: platformIcon(platform),
		ImageAsset:   img,
		DurationMins: s.DurationMins,
		Viewers:      s.ViewersCount,
	}
	// Only expose recording URL for completed sessions
	if strings.ToUpper(s.Status) == "COMPLETED" {
		out.RecordingURL = s.RecordingURL
	}
	return out
}

// GetLiveSessions returns all sessions (public, no sensitive join details).
// GET /live/sessions
func (lr *LiveRouter) GetLiveSessions(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch live sessions"})
		return
	}
	out := make([]sessionPublic, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, toPublic(s))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// GetSchedule returns sessions split by status (public).
// GET /live/schedule
func (lr *LiveRouter) GetSchedule(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at asc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch schedule"})
		return
	}

	upcoming := []sessionPublic{}
	live := []sessionPublic{}
	completed := []sessionPublic{}

	for _, s := range sessions {
		p := toPublic(s)
		switch strings.ToUpper(s.Status) {
		case "UPCOMING":
			upcoming = append(upcoming, p)
		case "LIVE":
			live = append(live, p)
		case "COMPLETED":
			completed = append(completed, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"upcoming":  upcoming,
			"live":      live,
			"completed": completed,
		},
	})
}

// GetJoinDetails returns the full join URL and passcode for a session.
// Requires a valid JWT. Students should call this just before joining.
// GET /live/sessions/:id/join
func (lr *LiveRouter) GetJoinDetails(c *gin.Context) {
	id := c.Param("id")
	var s models.LiveSession
	if err := lr.DB.Where("id = ?", id).First(&s).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Session not found"})
		return
	}

	st := strings.ToUpper(s.Status)
	if st == "UPCOMING" {
		// Reveal join link only 30 minutes before — for simplicity we allow UPCOMING too
		// (the mobile app can show a countdown and disable the join button)
	}

	platform := s.Platform
	if platform == "" {
		platform = "ZOOM"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           s.ID,
			"title":        s.Title,
			"status":       s.Status,
			"platform":     platform,
			"platformName": platformLabel(platform),
			"joinUrl":      s.JoinURL,
			"meetingId":    s.MeetingID,
			"zoomId":       s.ZoomID, // backward compat
			"passcode":     s.Passcode,
			"recordingUrl": s.RecordingURL,
		},
	})
}

// ─── Admin handlers ───────────────────────────────────────────────────────

func (lr *LiveRouter) CreateSession(c *gin.Context) {
	var s models.LiveSession
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if s.ID == "" {
		s.ID = "ls_" + uuid.New().String()[:8]
	}
	s.CreatedAt = time.Now()
	if s.Status == "" {
		s.Status = "UPCOMING"
	}
	if s.Platform == "" {
		s.Platform = "ZOOM"
	}
	// Backfill legacy ZoomID from MeetingID
	if s.MeetingID != "" && s.ZoomID == "" && strings.ToUpper(s.Platform) == "ZOOM" {
		s.ZoomID = s.MeetingID
	}
	if err := lr.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": s})
}

func (lr *LiveRouter) UpdateSession(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	// Sync meetingId → zoomId for backward compat
	if mid, ok := updates["meetingId"]; ok {
		if platform, p := updates["platform"]; p {
			if strings.ToUpper(platform.(string)) == "ZOOM" {
				updates["zoom_id"] = mid
			}
		}
	}
	if err := lr.DB.Model(&models.LiveSession{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	var updated models.LiveSession
	lr.DB.Where("id = ?", id).First(&updated)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}

// PatchSessionStatus quickly updates only the status field.
// PATCH /api/admin/live/sessions/:id/status
func (lr *LiveRouter) PatchSessionStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "status field required"})
		return
	}
	st := strings.ToUpper(body.Status)
	if st != "UPCOMING" && st != "LIVE" && st != "COMPLETED" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "status must be UPCOMING, LIVE, or COMPLETED"})
		return
	}
	if err := lr.DB.Model(&models.LiveSession{}).Where("id = ?", id).Update("status", st).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "status": st})
}

func (lr *LiveRouter) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	lr.DB.Delete(&models.LiveSession{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Seed ─────────────────────────────────────────────────────────────────

func (lr *LiveRouter) seedSampleSessions() {
	var count int64
	lr.DB.Model(&models.LiveSession{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	sessions := []models.LiveSession{
		{
			ID:            "ls_001",
			Title:         "CPU Reballing Full Process",
			Subtitle:      "Learn complete CPU reballing from preparation to re-installation.",
			Instructor:    "Mr. Rohit Sharma",
			DateFormatted: now.Add(24 * time.Hour).Format("2 Jan (Mon)") + ", 07:00 PM",
			TimeFormatted: "07:00 PM",
			Status:        "UPCOMING",
			Platform:      "ZOOM",
			JoinURL:       "https://zoom.us/j/123456789",
			ZoomID:        "123 456 789",
			MeetingID:     "123 456 789",
			Passcode:      "salen2026",
			DurationMins:  90,
			ImageAsset:    "assets/banner_bg_img/banner_1.webp",
			ViewersCount:  0,
			CreatedAt:     now,
		},
		{
			ID:            "ls_002",
			Title:         "EMMC Programming Complete Guide",
			Subtitle:      "Qualcomm & MediaTek EMMC flashing using EasyJTAG.",
			Instructor:    "Mr. Amit Kumar",
			DateFormatted: now.Add(72 * time.Hour).Format("2 Jan (Mon)") + ", 07:00 PM",
			TimeFormatted: "07:00 PM",
			Status:        "UPCOMING",
			Platform:      "GOOGLE_MEET",
			JoinURL:       "https://meet.google.com/abc-defg-hij",
			MeetingID:     "abc-defg-hij",
			Passcode:      "",
			DurationMins:  90,
			ImageAsset:    "assets/banner_bg_img/banner_2.webp",
			ViewersCount:  0,
			CreatedAt:     now.Add(-1 * time.Hour),
		},
		{
			ID:            "ls_003",
			Title:         "iPhone Baseband Repairing",
			Subtitle:      "Apple A15 baseband fault diagnosis and repair techniques.",
			Instructor:    "Mr. Pankaj Verma",
			DateFormatted: now.Add(120 * time.Hour).Format("2 Jan (Mon)") + ", 11:00 AM",
			TimeFormatted: "11:00 AM",
			Status:        "UPCOMING",
			Platform:      "ZOOM",
			JoinURL:       "https://zoom.us/j/111222333",
			ZoomID:        "111 222 333",
			MeetingID:     "111 222 333",
			Passcode:      "salen2026",
			DurationMins:  120,
			ImageAsset:    "assets/banner_bg_img/banner_3.webp",
			ViewersCount:  0,
			CreatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            "ls_004",
			Title:         "Display Light Problem Finding",
			Subtitle:      "Complete backlight circuit tracing on Android & iPhone.",
			Instructor:    "Mr. Pankaj Verma",
			DateFormatted: now.Add(-48 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "06:00 PM",
			Status:        "COMPLETED",
			Platform:      "ZOOM",
			RecordingURL:  "",
			DurationMins:  90,
			ImageAsset:    "assets/banner_bg_img/banner_1.webp",
			ViewersCount:  312,
			CreatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            "ls_005",
			Title:         "Short Circuit Tracing Technique",
			Subtitle:      "Using DC power supply and thermal camera to locate PCB shorts.",
			Instructor:    "Mr. Amit Kumar",
			DateFormatted: now.Add(-96 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "07:00 PM",
			Status:        "COMPLETED",
			Platform:      "GOOGLE_MEET",
			RecordingURL:  "",
			DurationMins:  90,
			ImageAsset:    "assets/banner_bg_img/banner_2.webp",
			ViewersCount:  275,
			CreatedAt:     now.Add(-96 * time.Hour),
		},
		{
			ID:            "ls_006",
			Title:         "Audio IC Replacement Step by Step",
			Subtitle:      "Micro-soldering audio codec IC on Android motherboards.",
			Instructor:    "Mr. Rohit Sharma",
			DateFormatted: now.Add(-168 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "07:00 PM",
			Status:        "COMPLETED",
			Platform:      "ZOOM",
			RecordingURL:  "",
			DurationMins:  90,
			ImageAsset:    "assets/banner_bg_img/banner_3.webp",
			ViewersCount:  198,
			CreatedAt:     now.Add(-168 * time.Hour),
		},
	}
	lr.DB.Create(&sessions)
}
