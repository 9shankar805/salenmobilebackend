package appscreens

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smg-backend/internal/models"
)

type AppScreensRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	sr := &AppScreensRouter{DB: db}

	// App Screens Configuration (Guest Home, Member Home, Live, Tools)
	r.GET("/app-screens/:siteName/:screenKey", sr.GetScreen)
}

// GetScreen returns the screen config as stored in DB.
// ConfigJSON may be a JSON array (block-slot format) or a JSON object (legacy flat map).
// Both are returned as-is so the Flutter client receives the exact stored structure.
func (sr *AppScreensRouter) GetScreen(c *gin.Context) {
	siteName := c.Param("siteName")
	screenKey := c.Param("screenKey")

	var config models.AppScreenConfig
	err := sr.DB.Where("site_name = ? AND screen_key = ?", siteName, screenKey).First(&config).Error
	if err != nil {
		// Return block-typed default if DB record not found
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    sr.getDefaultScreenSlots(screenKey),
		})
		return
	}

	// Try parsing as a generic JSON value (supports both arrays and objects)
	var data interface{}
	if err := json.Unmarshal([]byte(config.ConfigJSON), &data); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    sr.getDefaultScreenSlots(screenKey),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// getDefaultScreenSlots returns block-typed slot arrays matching what the Flutter app
// expects when it loops over data as List<dynamic> and reads slot['blockType'].
func (sr *AppScreensRouter) getDefaultScreenSlots(screenKey string) interface{} {
	switch screenKey {
	case "guest_home":
		return []map[string]interface{}{
			{
				"blockType": "hero_banner",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_1.webp",
					"banners": []string{
						"assets/banner_bg_img/banner_1.webp",
						"assets/banner_bg_img/banner_2.webp",
						"assets/banner_bg_img/banner_3.webp",
						"assets/banner_bg_img/banner_4.webp",
						"assets/banner_bg_img/banner_5.webp",
					},
				},
			},
			{
				"blockType": "promo_banner",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_2.webp",
					"title":       "38% OFF — 1 Month Mobile Plan",
					"subtitle":    "Full access to schematics, IC finder & more at just ₹249",
					"actionUrl":   "/plans/mobile",
					"badge":       "OFFER",
				},
			},
		}

	case "member_home":
		return []map[string]interface{}{
			{
				"blockType": "video_card",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_5.webp",
					"title":       "Continue Learning",
					"subtitle":    "Pick up where you left off",
				},
			},
			{
				"blockType": "continue_learning",
				"config": map[string]interface{}{
					"bannerImage":     "assets/banner_bg_img/banner_1.webp",
					"continueCourse":  "Android Hardware Repairing Level 1",
					"continueLesson":  "PMIC & Power Section Diagnosis",
					"progressPercent": 0.65,
					"courseSlug":      "android-hardware-repairing-level-1",
				},
			},
			{
				"blockType": "live_countdown",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_2.webp",
					"title":       "Live This Saturday",
					"subtitle":    "CPU Double-Decker Reballing Masterclass",
					"date":        "This Saturday",
					"time":        "7:00 PM IST",
					"joinUrl":     "https://zoom.us/j/8421092839",
				},
			},
			{
				"blockType": "promo_banner",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_3.webp",
					"title":       "New: UFS 4.0 Pinout Guide Added",
					"subtitle":    "Samsung S24 & Pixel 8 Pro BGA-297 ISP pinouts in PDF Container",
					"actionUrl":   "/pdf-container/schematics",
					"badge":       "NEW",
				},
			},
		}

	case "live_sessions":
		return []map[string]interface{}{
			{
				"blockType": "live_countdown",
				"config": map[string]interface{}{
					"bannerImage": "assets/banner_bg_img/banner_2.webp",
					"title":       "Next Live Session",
					"subtitle":    "CPU Reballing & Thermal Camera Short Detection",
					"date":        "This Saturday",
					"time":        "7:00 PM IST",
					"joinUrl":     "https://zoom.us/j/8421092839",
					"passcode":    "SALEN2026",
				},
			},
		}

	case "tools":
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{"id": "t1", "title": "Circuit Schematic Viewer", "category": "Schematics", "route": "/tools/schematics"},
				{"id": "t2", "title": "PMIC Pinout & Voltage Tester", "category": "Hardware", "route": "/tools/pmic"},
				{"id": "t3", "title": "Resistor Diode Code Calculator", "category": "Reference", "route": "/tools/calculator"},
				{"id": "t4", "title": "IC Finder", "category": "Repair Toolkit", "route": "/tools/ic-finder"},
				{"id": "t5", "title": "PIN Abbreviations", "category": "Repair Toolkit", "route": "/tools/pin-abbreviations"},
				{"id": "t6", "title": "Parts Finder", "category": "Repair Toolkit", "route": "/tools/parts-finder"},
			},
		}

	default:
		return map[string]interface{}{}
	}
}
