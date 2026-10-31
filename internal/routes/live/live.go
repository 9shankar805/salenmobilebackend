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

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	lr := &LiveRouter{DB: db}
	lr.seedSampleSessions()

	// Live Masterclass Sessions — public
	r.GET("/live/sessions", lr.GetLiveSessions)

	// Schedule endpoint: splits sessions into upcoming / live / past
	r.GET("/live/schedule", lr.GetSchedule)

	// Admin management
	adm := r.Group("/api/admin/live", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.POST("/sessions", lr.CreateSession)
		adm.PUT("/sessions/:id", lr.UpdateSession)
		adm.DELETE("/sessions/:id", lr.DeleteSession)
	}
}

func (lr *LiveRouter) GetLiveSessions(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch live sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": sessions})
}

// GetSchedule returns sessions split by status: upcoming, live, completed.
// GET /live/schedule
func (lr *LiveRouter) GetSchedule(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at asc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch schedule"})
		return
	}

	type sessionOut struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Subtitle   string `json:"subtitle"`
		Instructor string `json:"instructor"`
		Date       string `json:"date"`
		Time       string `json:"time"`
		Status     string `json:"status"`
		JoinURL    string `json:"joinUrl"`
		ZoomID     string `json:"zoomId"`
		Passcode   string `json:"passcode"`
		Viewers    int    `json:"viewersCount"`
		ImageAsset string `json:"imageAsset"`
		Duration   string `json:"duration,omitempty"`
	}

	upcoming := []sessionOut{}
	live := []sessionOut{}
	completed := []sessionOut{}

	for _, s := range sessions {
		st := strings.ToUpper(s.Status)
		out := sessionOut{
			ID:         s.ID,
			Title:      s.Title,
			Subtitle:   s.Subtitle,
			Instructor: s.Instructor,
			Date:       s.DateFormatted,
			Time:       s.TimeFormatted,
			Status:     s.Status,
			JoinURL:    s.JoinURL,
			ZoomID:     s.ZoomID,
			Passcode:   s.Passcode,
			Viewers:    s.ViewersCount,
			ImageAsset: "assets/banner_bg_img/banner_1.webp",
		}
		switch st {
		case "UPCOMING":
			upcoming = append(upcoming, out)
		case "LIVE":
			live = append(live, out)
		case "COMPLETED":
			// Add duration placeholder for past classes
			out.Duration = "01:30:00"
			completed = append(completed, out)
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
	if err := lr.DB.Model(&models.LiveSession{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (lr *LiveRouter) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	lr.DB.Delete(&models.LiveSession{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

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
			JoinURL:       "https://zoom.us/j/123456789",
			ZoomID:        "123 456 789",
			Passcode:      "salen2026",
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
			JoinURL:       "https://zoom.us/j/987654321",
			ZoomID:        "987 654 321",
			Passcode:      "salen2026",
			ViewersCount:  0,
			CreatedAt:     now.Add(-1 * time.Hour),
		},
		{
			ID:            "ls_003",
			Title:         "iPhone Baseband Repairing",
			Subtitle:      "Apple A15 baseband fault diagnosis and repair techniques.",
			Instructor:    "By Mr. Pankaj Verma",
			DateFormatted: now.Add(120 * time.Hour).Format("2 Jan (Mon)") + ", 11:00 AM",
			TimeFormatted: "11:00 AM",
			Status:        "UPCOMING",
			JoinURL:       "https://zoom.us/j/111222333",
			ZoomID:        "111 222 333",
			Passcode:      "salen2026",
			ViewersCount:  0,
			CreatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            "ls_004",
			Title:         "Display Light Problem Finding",
			Subtitle:      "Complete backlight circuit tracing on Android & iPhone.",
			Instructor:    "By Mr. Pankaj Verma",
			DateFormatted: now.Add(-48 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "06:00 PM",
			Status:        "COMPLETED",
			JoinURL:       "",
			ZoomID:        "",
			Passcode:      "",
			ViewersCount:  312,
			CreatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            "ls_005",
			Title:         "Short Circuit Tracing Technique",
			Subtitle:      "Using DC power supply and thermal camera to locate PCB shorts.",
			Instructor:    "By Mr. Amit Kumar",
			DateFormatted: now.Add(-96 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "07:00 PM",
			Status:        "COMPLETED",
			JoinURL:       "",
			ZoomID:        "",
			Passcode:      "",
			ViewersCount:  275,
			CreatedAt:     now.Add(-96 * time.Hour),
		},
		{
			ID:            "ls_006",
			Title:         "Audio IC Replacement Step by Step",
			Subtitle:      "Micro-soldering audio codec IC on Android motherboards.",
			Instructor:    "By Mr. Rohit Sharma",
			DateFormatted: now.Add(-168 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "07:00 PM",
			Status:        "COMPLETED",
			JoinURL:       "",
			ZoomID:        "",
			Passcode:      "",
			ViewersCount:  198,
			CreatedAt:     now.Add(-168 * time.Hour),
		},
	}
	lr.DB.Create(&sessions)
}
