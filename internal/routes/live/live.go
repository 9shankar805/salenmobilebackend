// Package live manages live class sessions for Salen Mobile Guru.
//
// Supported platforms:
//   HLS_SELF_HOSTED  — Self-hosted HLS via MediaMTX on this VPS (FREE, unlimited)
//                      Instructor streams RTMP from OBS/phone → students watch .m3u8
//   ZOOM             — External Zoom meeting link
//   GOOGLE_MEET      — External Google Meet link
//   TEAMS            — External Microsoft Teams link
//   YOUTUBE_LIVE     — External YouTube Live link
//
// MediaMTX endpoints on this server (209.50.241.46):
//   RTMP ingest:  rtmp://209.50.241.46:1935/live/<streamKey>
//   HLS playback: http://209.50.241.46:8888/live/<streamKey>/index.m3u8
//   WebRTC:       http://209.50.241.46:8889/live/<streamKey>/
//   MediaMTX API: http://127.0.0.1:9997/v3/paths/list  (localhost only)
//
// LiveKit WebRTC Audio/Video:
//   Server:  ws://209.50.241.46:7880
//   Token:   GET /live/call-token?room=&userId=&name=&role=
package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

// ─── MediaMTX / HLS constants ─────────────────────────────────────────────

const (
	mediaMTXAPI   = "http://127.0.0.1:9997"    // internal API — Go backend only
	hlsBaseURL    = "http://209.50.241.46:8888" // students watch on this
	rtmpBaseURL   = "rtmp://209.50.241.46:1935" // instructors stream to this
	webrtcBaseURL = "http://209.50.241.46:8889" // ultra-low latency browser/app
	smgUser       = "smg_broadcaster"
	smgPass       = "SalenLive@SMG2026!"
)

// ─── LiveKit constants ─────────────────────────────────────────────────────

const (
	LiveKitAPIKey    = "APIKeySalenGuru2026"
	LiveKitAPISecret = "SecretKeySalenGuruAudioCallLive2026Master"
	LiveKitServerURL = "ws://209.50.241.46:7880"
)

// ─── HLS URL helpers ──────────────────────────────────────────────────────

func hlsURL(streamKey string) string {
	return fmt.Sprintf("%s/live/%s/index.m3u8", hlsBaseURL, streamKey)
}
func rtmpURL(streamKey string) string {
	return fmt.Sprintf("%s/live/%s", rtmpBaseURL, streamKey)
}
func webrtcURL(streamKey string) string {
	return fmt.Sprintf("%s/live/%s/", webrtcBaseURL, streamKey)
}

// isStreamLive queries MediaMTX to check if a stream key is currently publishing.
func isStreamLive(streamKey string) bool {
	if streamKey == "" {
		return false
	}
	resp, err := http.Get(fmt.Sprintf("%s/v3/paths/get/live%%2F%s", mediaMTXAPI, streamKey))
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	defer resp.Body.Close()
	var result struct {
		Ready bool `json:"ready"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &result)
	return result.Ready
}

// listActivePaths returns all currently live stream paths from MediaMTX.
func listActivePaths() []string {
	resp, err := http.Get(mediaMTXAPI + "/v3/paths/list")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Items []struct {
			Name  string `json:"name"`
			Ready bool   `json:"ready"`
		} `json:"items"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &result)
	keys := []string{}
	for _, item := range result.Items {
		if item.Ready {
			keys = append(keys, item.Name)
		}
	}
	return keys
}

// ─── Platform helpers ─────────────────────────────────────────────────────

func platformLabel(p string) string {
	switch strings.ToUpper(p) {
	case "HLS_SELF_HOSTED":
		return "Live Stream (SMG Server)"
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

func platformIcon(p string) string {
	switch strings.ToUpper(p) {
	case "HLS_SELF_HOSTED":
		return "hls"
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

// ─── Router ───────────────────────────────────────────────────────────────

type LiveRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	lr := &LiveRouter{DB: db}
	lr.seedSampleSessions()

	// ── Public (no auth) ─────────────────────────────────────────────────
	r.GET("/live/sessions", lr.GetLiveSessions)
	r.GET("/live/schedule", lr.GetSchedule)

	// HLS stream status — Flutter polls this to show LIVE badge / enable play button
	r.GET("/live/stream/:streamKey/status", lr.GetStreamStatus)

	// LiveKit WebRTC token — public so unauthenticated students can join
	r.GET("/live/call-token", lr.GetLiveKitCallToken)

	// ── Authenticated (JWT required) ─────────────────────────────────────
	auth := r.Group("", middleware.AuthMiddleware())
	{
		// Returns full join details: HLS → hlsUrl+webrtcUrl, Zoom/Meet → joinUrl+passcode
		auth.GET("/live/sessions/:id/join", lr.GetJoinDetails)
	}

	// ── Admin (JWT + admin role) ──────────────────────────────────────────
	adm := r.Group("/api/admin/live", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.POST("/sessions", lr.CreateSession)
		adm.PUT("/sessions/:id", lr.UpdateSession)
		adm.PATCH("/sessions/:id/status", lr.PatchSessionStatus)
		adm.DELETE("/sessions/:id", lr.DeleteSession)
		adm.GET("/streams/active", lr.GetActiveStreams)
		adm.POST("/sessions/:id/go-live", lr.GoLive)
		adm.POST("/sessions/:id/end-stream", lr.EndStream)
	}
}

// ─── Public shape (no sensitive fields) ──────────────────────────────────

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
	IsHls        bool   `json:"isHls"`
	ImageAsset   string `json:"imageAsset"`
	DurationMins int    `json:"durationMins"`
	Viewers      int    `json:"viewersCount"`
	IsStreamLive bool   `json:"isStreamLive"` // real-time from MediaMTX API
	RecordingURL string `json:"recordingUrl,omitempty"`
	// External meeting fields (safe to expose — no passcode)
	JoinURL   string `json:"joinUrl,omitempty"`
	MeetingID string `json:"meetingId,omitempty"`
}

func toPublic(s models.LiveSession) sessionPublic {
	platform := s.Platform
	if platform == "" {
		if s.StreamKey != "" {
			platform = "HLS_SELF_HOSTED"
		} else if s.ZoomID != "" {
			platform = "ZOOM"
		} else {
			platform = "ZOOM"
		}
	}
	img := s.ImageAsset
	if img == "" {
		img = "assets/banner_bg_img/banner_1.webp"
	}

	streamLive := false
	if strings.ToUpper(platform) == "HLS_SELF_HOSTED" && s.StreamKey != "" {
		streamLive = isStreamLive(s.StreamKey)
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
		IsHls:        strings.ToUpper(platform) == "HLS_SELF_HOSTED",
		ImageAsset:   img,
		DurationMins: s.DurationMins,
		Viewers:      s.ViewersCount,
		IsStreamLive: streamLive,
		JoinURL:      s.JoinURL,
		MeetingID:    s.MeetingID,
	}
	if strings.ToUpper(s.Status) == "COMPLETED" {
		out.RecordingURL = s.RecordingURL
	}
	return out
}

// ─── Public endpoints ─────────────────────────────────────────────────────

// GET /live/sessions
func (lr *LiveRouter) GetLiveSessions(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch sessions"})
		return
	}
	out := make([]sessionPublic, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, toPublic(s))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// GET /live/schedule
func (lr *LiveRouter) GetSchedule(c *gin.Context) {
	var sessions []models.LiveSession
	if err := lr.DB.Order("created_at asc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch schedule"})
		return
	}
	upcoming, live, completed := []sessionPublic{}, []sessionPublic{}, []sessionPublic{}
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
		"data":    gin.H{"upcoming": upcoming, "live": live, "completed": completed},
	})
}

// GET /live/stream/:streamKey/status
// Flutter polls every 10–30s to show/hide the LIVE badge and enable the play button.
func (lr *LiveRouter) GetStreamStatus(c *gin.Context) {
	key := c.Param("streamKey")
	live := isStreamLive(key)
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"streamKey": key,
		"isLive":    live,
		"hlsUrl":    hlsURL(key),
		"webrtcUrl": webrtcURL(key),
	})
}

// GET /live/sessions/:id/join  (JWT required)
// HLS sessions  → returns hlsUrl, webrtcUrl, streamKey, isStreamLive
// Zoom/Meet     → returns joinUrl, passcode, meetingId
func (lr *LiveRouter) GetJoinDetails(c *gin.Context) {
	id := c.Param("id")
	var s models.LiveSession
	if err := lr.DB.Where("id = ?", id).First(&s).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Session not found"})
		return
	}

	platform := s.Platform
	if platform == "" {
		if s.StreamKey != "" {
			platform = "HLS_SELF_HOSTED"
		} else {
			platform = "ZOOM"
		}
	}

	base := gin.H{
		"id":           s.ID,
		"title":        s.Title,
		"status":       s.Status,
		"platform":     platform,
		"platformName": platformLabel(platform),
		"durationMins": s.DurationMins,
	}

	if strings.ToUpper(platform) == "HLS_SELF_HOSTED" {
		live := isStreamLive(s.StreamKey)
		base["isHls"] = true
		base["streamKey"] = s.StreamKey
		base["hlsUrl"] = hlsURL(s.StreamKey)
		base["webrtcUrl"] = webrtcURL(s.StreamKey)
		base["isStreamLive"] = live
		base["rtmpUrl"] = rtmpURL(s.StreamKey)
		base["obsServer"] = rtmpBaseURL + "/live"
		base["obsStreamKey"] = s.StreamKey
	} else {
		base["isHls"] = false
		base["joinUrl"] = s.JoinURL
		base["meetingId"] = s.MeetingID
		base["zoomId"] = s.ZoomID
		base["passcode"] = s.Passcode
	}

	if strings.ToUpper(s.Status) == "COMPLETED" {
		base["recordingUrl"] = s.RecordingURL
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": base})
}

// ─── Admin endpoints ──────────────────────────────────────────────────────

// POST /api/admin/live/sessions
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
		if s.StreamKey != "" {
			s.Platform = "HLS_SELF_HOSTED"
		} else {
			s.Platform = "ZOOM"
		}
	}
	// Auto-generate stream key if HLS and none provided
	if strings.ToUpper(s.Platform) == "HLS_SELF_HOSTED" {
		if s.StreamKey == "" {
			s.StreamKey = "smg-" + uuid.New().String()[:8]
		}
		s.HlsUrl = hlsURL(s.StreamKey)
		s.RtmpUrl = rtmpURL(s.StreamKey)
		s.IsHls = true
	}
	if s.MeetingID != "" && s.ZoomID == "" && strings.ToUpper(s.Platform) == "ZOOM" {
		s.ZoomID = s.MeetingID
	}
	if err := lr.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := gin.H{"success": true, "data": s}
	if strings.ToUpper(s.Platform) == "HLS_SELF_HOSTED" {
		resp["obsInstructions"] = gin.H{
			"server":    rtmpBaseURL + "/live",
			"streamKey": s.StreamKey,
			"username":  smgUser,
			"password":  smgPass,
			"hlsUrl":    s.HlsUrl,
		}
	}
	c.JSON(http.StatusCreated, resp)
}

// PUT /api/admin/live/sessions/:id
func (lr *LiveRouter) UpdateSession(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if sk, ok := updates["streamKey"]; ok && sk != "" {
		key := sk.(string)
		updates["hls_url"] = hlsURL(key)
		updates["rtmp_url"] = rtmpURL(key)
		updates["is_hls"] = true
		updates["platform"] = "HLS_SELF_HOSTED"
	}
	if mid, ok := updates["meetingId"]; ok {
		if p, pp := updates["platform"]; pp && strings.ToUpper(fmt.Sprintf("%v", p)) == "ZOOM" {
			updates["zoom_id"] = mid
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
	lr.DB.Model(&models.LiveSession{}).Where("id = ?", id).Update("status", st)
	c.JSON(http.StatusOK, gin.H{"success": true, "status": st})
}

// POST /api/admin/live/sessions/:id/go-live
func (lr *LiveRouter) GoLive(c *gin.Context) {
	id := c.Param("id")
	var s models.LiveSession
	if err := lr.DB.Where("id = ?", id).First(&s).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Session not found"})
		return
	}
	lr.DB.Model(&s).Update("status", "LIVE")
	resp := gin.H{"success": true, "status": "LIVE", "message": "Session is now LIVE"}
	if strings.ToUpper(s.Platform) == "HLS_SELF_HOSTED" && s.StreamKey != "" {
		resp["streaming"] = gin.H{
			"platform":     "HLS_SELF_HOSTED",
			"obsServer":    rtmpBaseURL + "/live",
			"obsStreamKey": s.StreamKey,
			"obsUsername":  smgUser,
			"obsPassword":  smgPass,
			"hlsUrl":       hlsURL(s.StreamKey),
			"webrtcUrl":    webrtcURL(s.StreamKey),
			"isStreamLive": isStreamLive(s.StreamKey),
			"instructions": "OBS: Settings → Stream → Server: " + rtmpBaseURL + "/live  Stream Key: " + s.StreamKey,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/admin/live/sessions/:id/end-stream
func (lr *LiveRouter) EndStream(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		RecordingURL string `json:"recordingUrl"`
	}
	c.ShouldBindJSON(&body)
	updates := map[string]interface{}{"status": "COMPLETED"}
	if body.RecordingURL != "" {
		updates["recording_url"] = body.RecordingURL
	}
	lr.DB.Model(&models.LiveSession{}).Where("id = ?", id).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"success": true, "status": "COMPLETED"})
}

// GET /api/admin/live/streams/active
func (lr *LiveRouter) GetActiveStreams(c *gin.Context) {
	paths := listActivePaths()
	streams := make([]gin.H, 0, len(paths))
	for _, p := range paths {
		key := strings.TrimPrefix(p, "live/")
		streams = append(streams, gin.H{
			"path":      p,
			"streamKey": key,
			"hlsUrl":    hlsURL(key),
			"webrtcUrl": webrtcURL(key),
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "count": len(streams), "data": streams})
}

// DELETE /api/admin/live/sessions/:id
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
			ID: "ls_001", Title: "CPU Reballing Full Process",
			Subtitle:      "Learn complete CPU reballing from preparation to re-installation.",
			Instructor:    "Mr. Rohit Sharma",
			DateFormatted: now.Add(24 * time.Hour).Format("2 Jan (Mon)") + ", 07:00 PM",
			TimeFormatted: "07:00 PM", Status: "UPCOMING",
			Platform: "HLS_SELF_HOSTED", StreamKey: "smg-reballing-class",
			HlsUrl: hlsURL("smg-reballing-class"), RtmpUrl: rtmpURL("smg-reballing-class"),
			IsHls: true, DurationMins: 90, ImageAsset: "assets/banner_bg_img/banner_1.webp",
			CreatedAt: now,
		},
		{
			ID: "ls_002", Title: "EMMC Programming Complete Guide",
			Subtitle:      "Qualcomm & MediaTek EMMC flashing using EasyJTAG.",
			Instructor:    "Mr. Amit Kumar",
			DateFormatted: now.Add(72 * time.Hour).Format("2 Jan (Mon)") + ", 07:00 PM",
			TimeFormatted: "07:00 PM", Status: "UPCOMING",
			Platform: "GOOGLE_MEET", JoinURL: "https://meet.google.com/abc-defg-hij",
			MeetingID: "abc-defg-hij", IsHls: false, DurationMins: 90,
			ImageAsset: "assets/banner_bg_img/banner_2.webp",
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID: "ls_003", Title: "iPhone Baseband Repairing",
			Subtitle:      "Apple A15 baseband fault diagnosis and repair techniques.",
			Instructor:    "Mr. Pankaj Verma",
			DateFormatted: now.Add(120 * time.Hour).Format("2 Jan (Mon)") + ", 11:00 AM",
			TimeFormatted: "11:00 AM", Status: "UPCOMING",
			Platform: "HLS_SELF_HOSTED", StreamKey: "smg-iphone-class",
			HlsUrl: hlsURL("smg-iphone-class"), RtmpUrl: rtmpURL("smg-iphone-class"),
			IsHls: true, DurationMins: 120, ImageAsset: "assets/banner_bg_img/banner_3.webp",
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: "ls_004", Title: "Display Light Problem Finding",
			Subtitle:      "Complete backlight circuit tracing on Android & iPhone.",
			Instructor:    "Mr. Pankaj Verma",
			DateFormatted: now.Add(-48 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "06:00 PM", Status: "COMPLETED",
			Platform: "HLS_SELF_HOSTED", StreamKey: "smg-display-class",
			HlsUrl: hlsURL("smg-display-class"), IsHls: true,
			DurationMins: 90, ImageAsset: "assets/banner_bg_img/banner_1.webp",
			ViewersCount: 312, CreatedAt: now.Add(-48 * time.Hour),
		},
		{
			ID: "ls_005", Title: "Short Circuit Tracing Technique",
			Subtitle:      "Using DC power supply and thermal camera to locate PCB shorts.",
			Instructor:    "Mr. Amit Kumar",
			DateFormatted: now.Add(-96 * time.Hour).Format("2 Jan 2006"),
			TimeFormatted: "07:00 PM", Status: "COMPLETED",
			Platform: "ZOOM", IsHls: false, DurationMins: 90,
			ImageAsset: "assets/banner_bg_img/banner_2.webp",
			ViewersCount: 275, CreatedAt: now.Add(-96 * time.Hour),
		},
	}
	lr.DB.Create(&sessions)
}

// ─── LiveKit WebRTC Token ─────────────────────────────────────────────────

type liveKitVideoGrant struct {
	Room           string `json:"room"`
	RoomJoin       bool   `json:"roomJoin"`
	CanPublish     bool   `json:"canPublish"`
	CanSubscribe   bool   `json:"canSubscribe"`
	CanPublishData bool   `json:"canPublishData"`
	RoomAdmin      bool   `json:"roomAdmin,omitempty"`
}

type liveKitClaims struct {
	Video liveKitVideoGrant `json:"video"`
	Name  string            `json:"name"`
	jwt.RegisteredClaims
}

// CreateLiveKitToken crafts an HMAC-SHA256 JWT accepted by LiveKit server.
func CreateLiveKitToken(apiKey, secret, room, identity, name string, isAdmin bool) (string, error) {
	now := time.Now()
	claims := liveKitClaims{
		Video: liveKitVideoGrant{
			Room: room, RoomJoin: true,
			CanPublish: true, CanSubscribe: true, CanPublishData: true,
			RoomAdmin: isAdmin,
		},
		Name: name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    apiKey,
			Subject:   identity,
			ExpiresAt: jwt.NewNumericDate(now.Add(6 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GET /live/call-token?room=&userId=&name=&role=
// Returns a LiveKit JWT for audio/video calls.
func (lr *LiveRouter) GetLiveKitCallToken(c *gin.Context) {
	room := c.DefaultQuery("room", "salen_live_masterclass")
	role := strings.ToUpper(c.DefaultQuery("role", "STUDENT"))
	userName := c.DefaultQuery("name", "")
	userID := c.DefaultQuery("userId", "")

	if u, exists := c.Get("user"); exists {
		if userObj, ok := u.(models.User); ok {
			if userName == "" {
				userName = userObj.Name
			}
			if userID == "" {
				userID = userObj.ID
			}
			if userObj.Role == "ADMIN" {
				role = "ADMIN"
			}
		}
	}
	if userName == "" {
		if role == "ADMIN" {
			userName = "Salen Dangol (Instructor)"
		} else {
			userName = "Student"
		}
	}
	if userID == "" {
		userID = "user_" + uuid.New().String()[:8]
	}

	tokenStr, err := CreateLiveKitToken(
		LiveKitAPIKey, LiveKitAPISecret,
		room, userID, userName, role == "ADMIN",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Token generation failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"serverUrl":       LiveKitServerURL,
		"token":           tokenStr,
		"room":            room,
		"role":            role,
		"participantName": userName,
		"participantId":   userID,
	})
}
