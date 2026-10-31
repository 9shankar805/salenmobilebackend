package database

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
	"smg-backend/internal/models"
)

func SeedCommunityAndLive(db *gorm.DB) {
	now := time.Now()

	posts := []models.CommunityPost{
		{
			ID:             "p1",
			OrganizationID: "smg",
			CourseID:       "c1",
			Title:          "Redmi Note 10 Pro stuck at 0.045A on DC power supply — how to trace?",
			Body:           "When I connect to DC power supply and press power key, current jumps to 0.045A and freezes. Multimeter diode mode shows normal readings on VPH_PWR.",
			AuthorName:     "Karthik Raja",
			AuthorRole:     "Hardware Technician · Chennai",
			AuthorAvatar:   "assets/banner_bg_img/banner_1.webp",
			Upvotes:        24,
			RepliesCount:   3,
			IsSolved:       true,
			Solution:       "0.045A freeze on Snapdragon 732G indicates missing S2C or S3C secondary BUCK voltages from PM6150. Check coil L1002 and replace capacitor C1045.",
			CreatedAt:      now.Add(-48 * time.Hour),
		},
		{
			ID:             "p2",
			OrganizationID: "smg",
			CourseID:       "c1",
			Title:          "Recommended hot air rework temperature for UFS 3.1 chip removal?",
			Body:           "I am using Quick 861DW rework station. What air flow and temperature profile prevents board warping on 12-layer boards?",
			AuthorName:     "Bikash Shrestha",
			AuthorRole:     "Shop Owner · Kathmandu",
			AuthorAvatar:   "assets/banner_bg_img/banner_2.webp",
			Upvotes:        38,
			RepliesCount:   5,
			IsSolved:       true,
			Solution:       "Set preheater to 120°C for 2 minutes. On Quick 861DW, use 340°C with 45 airflow with circular motion for 45-60 seconds max.",
			CreatedAt:      now.Add(-24 * time.Hour),
		},
		{
			ID:             "p3",
			OrganizationID: "smg",
			CourseID:       "c2",
			Title:          "iPhone 13 Pro Max restart every 3 minutes — panic log sensor panic?",
			Body:           "The phone works normally for 180 seconds then shuts down and restarts. The panic log says: 'Missing sensor(s): mic1, prs0'.",
			AuthorName:     "Subash Tamang",
			AuthorRole:     "iOS Specialist · Pokhara",
			AuthorAvatar:   "assets/banner_bg_img/banner_3.webp",
			Upvotes:        45,
			RepliesCount:   4,
			IsSolved:       true,
			Solution:       "Missing prs0 and mic1 means the charging port flex cable is either disconnected or defective. Replacing the OEM charging port flex solves the 3-minute reboot loop immediately.",
			CreatedAt:      now.Add(-12 * time.Hour),
		},
		{
			ID:             "p4",
			OrganizationID: "smg",
			CourseID:       "c3",
			Title:          "EasyJtag Plus error: 'EMMC OCR: 0xC0FF8080 - CMD Timeout' on Samsung A12",
			Body:           "Soldered ISP wires (CLK, CMD, DAT0, VCC, VCCQ, GND). Box detects device but fails with CMD timeout at 24MHz.",
			AuthorName:     "Deepak Joshi",
			AuthorRole:     "Chip-Off Technician · Butwal",
			AuthorAvatar:   "assets/banner_bg_img/banner_4.webp",
			Upvotes:        19,
			RepliesCount:   2,
			IsSolved:       true,
			Solution:       "CMD line signal reflection: shorten ISP wire lengths to under 7cm and add a 100-ohm resistor inline on CLK or lower frequency to 12MHz.",
			CreatedAt:      now.Add(-6 * time.Hour),
		},
		{
			ID:             "p5",
			OrganizationID: "smg",
			CourseID:       "c5",
			Title:          "POCO X3 Pro CPU reballing: AP soldered successfully but RAM not detected?",
			Body:           "After soldering Qualcomm SM8150 CPU, PC recognizes 9008 port. But after soldering Samsung LPDDR4X RAM, device draws 0.120A and doesn't boot.",
			AuthorName:     "Roshan Giri",
			AuthorRole:     "Master Technician · Biratnagar",
			AuthorAvatar:   "assets/banner_bg_img/banner_5.webp",
			Upvotes:        31,
			RepliesCount:   3,
			IsSolved:       true,
			Solution:       "Check middle interposer balls on top of AP CPU under microscope. If balls are uneven, use Amaoe middle-layer stencil with 183°C paste and wipe with dry tissue before heating.",
			CreatedAt:      now.Add(-2 * time.Hour),
		},
	}

	for _, p := range posts {
		var existing models.CommunityPost
		if err := db.Where("id = ?", p.ID).First(&existing).Error; err != nil {
			db.Create(&p)
		} else {
			existing.Title = p.Title
			existing.Body = p.Body
			existing.Solution = p.Solution
			existing.Upvotes = p.Upvotes
			existing.RepliesCount = p.RepliesCount
			existing.IsSolved = p.IsSolved
			db.Save(&existing)
		}
	}

	liveSessions := []models.LiveSession{
		{
			ID:            "ls1",
			Title:         "CPU Double-Decker Reballing Live Masterclass",
			Subtitle:      "Hands-on microscopic demo with live Q&A",
			Instructor:    "Rohit Sharma",
			DateFormatted: "This Saturday",
			TimeFormatted: "7:00 PM IST",
			Status:        "UPCOMING",
			JoinURL:       "https://zoom.us/j/8421092839",
			ZoomID:        "842 109 2839",
			Passcode:      "SALEN2026",
			ViewersCount:  420,
			CreatedAt:     now,
		},
		{
			ID:            "ls2",
			Title:         "Thermal Camera Short Detection Live Case Studies",
			Subtitle:      "Clearing 5 dead phones sent in by students live",
			Instructor:    "Salen Dangol",
			DateFormatted: "This Sunday",
			TimeFormatted: "8:00 PM IST",
			Status:        "UPCOMING",
			JoinURL:       "https://zoom.us/j/9128301928",
			ZoomID:        "912 830 1928",
			Passcode:      "CHIPLEVEL",
			ViewersCount:  380,
			CreatedAt:     now,
		},
	}

	for _, ls := range liveSessions {
		var existing models.LiveSession
		if err := db.Where("id = ?", ls.ID).First(&existing).Error; err != nil {
			db.Create(&ls)
		} else {
			existing.Title = ls.Title
			existing.Subtitle = ls.Subtitle
			existing.Instructor = ls.Instructor
			existing.DateFormatted = ls.DateFormatted
			existing.TimeFormatted = ls.TimeFormatted
			existing.Status = ls.Status
			existing.ViewersCount = ls.ViewersCount
			db.Save(&existing)
		}
	}
	log.Printf("✅ Seeded community posts & live sessions!")
}

func SeedAppScreenConfigs(db *gorm.DB) {
	now := time.Now()

	// ─── Guest Home Config ───────────────────────────────────────────────
	// Flutter app loops over data as a List of {blockType, config} slots.
	// It handles: hero_banner, promo_banner
	guestHomeConfig := []map[string]interface{}{
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
	rawGuest, _ := json.Marshal(guestHomeConfig)

	var guestScreen models.AppScreenConfig
	if err := db.Where("screen_key = ?", "guest_home").First(&guestScreen).Error; err != nil {
		db.Create(&models.AppScreenConfig{
			ID:         "sc_guest",
			SiteName:   "smg",
			ScreenKey:  "guest_home",
			ConfigJSON: string(rawGuest),
			UpdatedAt:  now,
		})
	} else {
		guestScreen.ConfigJSON = string(rawGuest)
		guestScreen.UpdatedAt = now
		db.Save(&guestScreen)
	}

	// ─── Member Home Config ──────────────────────────────────────────────
	// Flutter app handles: video_card, continue_learning, live_countdown, promo_banner
	memberHomeConfig := []map[string]interface{}{
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
				"subtitle":    "Samsung S24 & Pixel 8 Pro BGA-297 ISP pinouts now in PDF Container",
				"actionUrl":   "/pdf-container/schematics",
				"badge":       "NEW",
			},
		},
	}
	rawMember, _ := json.Marshal(memberHomeConfig)

	var memberScreen models.AppScreenConfig
	if err := db.Where("screen_key = ?", "member_home").First(&memberScreen).Error; err != nil {
		db.Create(&models.AppScreenConfig{
			ID:         "sc_member",
			SiteName:   "smg",
			ScreenKey:  "member_home",
			ConfigJSON: string(rawMember),
			UpdatedAt:  now,
		})
	} else {
		memberScreen.ConfigJSON = string(rawMember)
		memberScreen.UpdatedAt = now
		db.Save(&memberScreen)
	}

	// ─── Live Sessions Config ────────────────────────────────────────────
	// Flutter app handles: live_countdown block
	liveConfig := []map[string]interface{}{
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
	rawLive, _ := json.Marshal(liveConfig)

	var liveScreen models.AppScreenConfig
	if err := db.Where("screen_key = ?", "live_sessions").First(&liveScreen).Error; err != nil {
		db.Create(&models.AppScreenConfig{
			ID:         "sc_live",
			SiteName:   "smg",
			ScreenKey:  "live_sessions",
			ConfigJSON: string(rawLive),
			UpdatedAt:  now,
		})
	} else {
		liveScreen.ConfigJSON = string(rawLive)
		liveScreen.UpdatedAt = now
		db.Save(&liveScreen)
	}

	log.Printf("✅ Seeded app screen configurations with blockType slot format!")
}

func SeedUsersAndSubscriptions(db *gorm.DB) {
	now := time.Now()
	allSlugs := "android-hardware-repairing-level-1,iphone-display-repairing,emmc-ufs-programming,borneo-zxw-schematic-reading,pmic-power-section-diagnosis"

	users := []models.User{
		{
			ID:               "u_admin",
			Email:            "admin@salenmobileguru.com",
			Name:             "Salen Master Admin",
			Role:             "ADMIN",
			IsSubscribed:     true,
			IsVIP:            true,
			SubscriptionPlan: "Super Admin All Access",
			CreatedAt:        now.AddDate(0, -6, 0),
		},
		{
			ID:               "u_ramesh",
			Email:            "ramesh.repair@gmail.com",
			Name:             "Ramesh KC",
			Role:             "STUDENT",
			IsSubscribed:     true,
			IsVIP:            true,
			SubscriptionPlan: "VIP All Courses Access",
			CreatedAt:        now.AddDate(0, -2, 0),
		},
		{
			ID:               "u_suresh",
			Email:            "suresh.tech@outlook.com",
			Name:             "Suresh Shrestha",
			Role:             "STUDENT",
			IsSubscribed:     true,
			IsVIP:            false,
			SubscriptionPlan: "Android Hardware Level 1",
			CreatedAt:        now.AddDate(0, -1, 0),
		},
		{
			ID:               "u_anil",
			Email:            "anil.chip@yahoo.com",
			Name:             "Anil Kumar",
			Role:             "STUDENT",
			IsSubscribed:     false,
			IsVIP:            false,
			SubscriptionPlan: "Free Member",
			CreatedAt:        now.Add(-2 * time.Hour),
		},
	}

	for _, u := range users {
		var existing models.User
		if err := db.Where("id = ? OR email = ?", u.ID, u.Email).First(&existing).Error; err != nil {
			db.Create(&u)
		} else {
			existing.IsSubscribed = u.IsSubscribed
			existing.IsVIP = u.IsVIP
			existing.SubscriptionPlan = u.SubscriptionPlan
			db.Save(&existing)
		}
	}

	subs := []models.Subscription{
		{
			ID:            "sub_1",
			UserID:        "u_ramesh",
			PlanType:      "VIP",
			Status:        "active",
			Amount:        9999.0,
			Currency:      "NPR",
			PaymentMethod: "ESEWA",
			TransactionID: "ESEWA-9281903",
			CourseSlug:    "android-hardware-repairing-level-1",
			EnrolledSlugs: allSlugs,
			StartDate:     now.AddDate(0, -2, 0),
			EndDate:       now.AddDate(1, 0, 0),
			CreatedAt:     now.AddDate(0, -2, 0),
		},
		{
			ID:            "sub_2",
			UserID:        "u_suresh",
			PlanType:      "SINGLE_COURSE",
			Status:        "active",
			Amount:        4999.0,
			Currency:      "NPR",
			PaymentMethod: "KHALTI",
			TransactionID: "KH-7381928",
			CourseSlug:    "android-hardware-repairing-level-1",
			EnrolledSlugs: "android-hardware-repairing-level-1",
			StartDate:     now.AddDate(0, -1, 0),
			EndDate:       now.AddDate(1, 0, 0),
			CreatedAt:     now.AddDate(0, -1, 0),
		},
	}

	for _, s := range subs {
		var existing models.Subscription
		if err := db.Where("id = ?", s.ID).First(&existing).Error; err != nil {
			db.Create(&s)
		} else {
			existing.Status = s.Status
			existing.EnrolledSlugs = s.EnrolledSlugs
			db.Save(&existing)
		}
	}

	log.Printf("✅ Seeded users and active subscriptions!")
}
