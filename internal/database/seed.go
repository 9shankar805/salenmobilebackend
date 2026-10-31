package database

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"smg-backend/internal/models"
)

// SeedDatabase orchestrates seeding for all models and relations
func SeedDatabase(db *gorm.DB) {
	log.Println("⚡ Running Salen Mobile Guru comprehensive database seeder...")

	// 1. Admin & instructor credentials
	SeedAdminCredentials(db)

	// 2. All 5 courses with full 23 modules and 101 lessons
	SeedCoursesAndCurriculum(db)

	// 3. Hardware circuit schematics
	SeedSchematics(db)

	// 4. Practical workbench exercises & lab tasks
	SeedExercises(db)

	// 5. Interactive quizzes and assessments
	SeedQuizzes(db)

	// 6. Community technical Q&A and upcoming live sessions
	SeedCommunityAndLive(db)

	// 7. App screen layout configurations (guest home, member home, live)
	SeedAppScreenConfigs(db)

	// 8. Demo students and VIP active subscriptions
	SeedUsersAndSubscriptions(db)

	// 9. Mobile & Chat Plans, PDF Container categories, Support tree, Highlights
	SeedPlansAndHome(db)

	// 10. Official template blueprints & worksheets
	SeedTemplates(db)

	log.Println("🎉 Database seeding completed successfully!")
}

// hashSeedPassword hashes a password with bcrypt for secure storage.
func hashSeedPassword(pwd string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash seed password: %v", err)
	}
	return string(hashed)
}

// ═══════════════════════════════════════════════════════════════════════════
// ADMIN PANEL CREDENTIALS
// ═══════════════════════════════════════════════════════════════════════════

func SeedAdminCredentials(db *gorm.DB) {
	admins := []models.User{
		{
			ID:               "u_admin_master",
			Email:            "admin@salenmobileguru.com",
			Name:             "Salen Master Admin",
			Role:             "ADMIN",
			PasswordHash:     hashSeedPassword("Admin@SMG2026!"),
			Phone:            "+977-9800000000",
			AvatarURL:        "https://images.unsplash.com/photo-1581092918056-0c4c3acd3789?w=200",
			IsSubscribed:     true,
			IsVIP:            true,
			SubscriptionPlan: "Super Admin All Access",
			CreatedAt:        time.Now().AddDate(0, -6, 0),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               "u_admin_instructor",
			Email:            "instructor@salenmobileguru.com",
			Name:             "Rohit Sharma (Senior Instructor)",
			Role:             "ADMIN",
			PasswordHash:     hashSeedPassword("Instructor@SMG2026!"),
			Phone:            "+977-9801111111",
			AvatarURL:        "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=200",
			IsSubscribed:     true,
			IsVIP:            true,
			SubscriptionPlan: "Master Trainer",
			CreatedAt:        time.Now().AddDate(0, -5, 0),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               "u_admin_support",
			Email:            "support@salenmobileguru.com",
			Name:             "SMG Support Manager",
			Role:             "ADMIN",
			PasswordHash:     hashSeedPassword("Support@SMG2026!"),
			Phone:            "+977-9802222222",
			AvatarURL:        "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200",
			IsSubscribed:     true,
			IsVIP:            false,
			SubscriptionPlan: "Admin Manager",
			CreatedAt:        time.Now().AddDate(0, -4, 0),
			UpdatedAt:        time.Now(),
		},
	}

	for _, admin := range admins {
		var existing models.User
		if err := db.Where("LOWER(email) = ?", admin.Email).First(&existing).Error; err != nil {
			if err := db.Create(&admin).Error; err != nil {
				log.Printf("⚠️ Failed to create admin %s: %v", admin.Email, err)
			} else {
				log.Printf("✅ Seeded admin credential: %s (%s)", admin.Email, admin.Role)
			}
		} else {
			// Always refresh password hash and role on restart so credentials stay valid
			existing.PasswordHash = admin.PasswordHash
			existing.Role = "ADMIN"
			existing.Name = admin.Name
			existing.AvatarURL = admin.AvatarURL
			existing.UpdatedAt = time.Now()
			db.Save(&existing)
			log.Printf("✅ Updated admin credential: %s", admin.Email)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// OFFICIAL TEMPLATES & BLUEPRINTS SEEDING
// ═══════════════════════════════════════════════════════════════════════════

func SeedTemplates(db *gorm.DB) {
	var count int64
	db.Model(&models.TemplateAsset{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	templates := []models.TemplateAsset{
		{
			ID:             "tmpl_cert_distinction",
			Title:          "Gold Distinction Certificate Background Template",
			Category:       "certificate",
			Description:    "High-resolution official certificate template with Guilloche security borders, gold foil seals, and distinction typography.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_cert_gold_distinction.svg",
			ThumbnailURL:   "https://images.unsplash.com/photo-1589330694653-dad6ef0140be?w=600",
			FileType:       "SVG",
			FileSize:       "1.2 MB",
			DownloadsCount: 342,
			IsDefault:      true,
			Author:         "Salen Mobile Guru Academy",
			Tags:           "certificate, gold, distinction, official, diploma",
			CreatedAt:      now.AddDate(0, -2, 0),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_cert_android",
			Title:          "Android Hardware Specialist Certificate Template",
			Category:       "certificate",
			Description:    "Landscape printable diploma with micro-soldering, CPU reballing, and schematic tracing credentials branding.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_cert_android_hardware.pdf",
			ThumbnailURL:   "https://images.unsplash.com/photo-1607344645866-009c320c5ab8?w=600",
			FileType:       "PDF",
			FileSize:       "3.4 MB",
			DownloadsCount: 189,
			IsDefault:      false,
			Author:         "Salen Dangol",
			Tags:           "certificate, android, hardware, diploma, pmic",
			CreatedAt:      now.AddDate(0, -1, -15),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_diag_diode",
			Title:          "Motherboard Diode Mode Probing Worksheet",
			Category:       "worksheet",
			Description:    "Standard technician logging sheet for red-probe-to-ground diode readings across Type-C pins, display FPC, and PMIC coils.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_diode_mode_worksheet.pdf",
			ThumbnailURL:   "https://images.unsplash.com/photo-1581092160607-ee22621dd758?w=600",
			FileType:       "PDF",
			FileSize:       "850 KB",
			DownloadsCount: 520,
			IsDefault:      true,
			Author:         "Master Salen",
			Tags:           "worksheet, multimeter, diode mode, diagnostic, testing",
			CreatedAt:      now.AddDate(0, -2, -5),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_diag_vph",
			Title:          "PMIC Power Sequencing & Buck Coil Flowchart",
			Category:       "schematic",
			Description:    "Decision tree for dead boot troubleshooting: VPH_PWR, S2_PMU, LDO rails, PS_HOLD, and crystal oscillator oscillations.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_power_sequence_flowchart.pdf",
			ThumbnailURL:   "https://images.unsplash.com/photo-1518770660439-4636190af475?w=600",
			FileType:       "PDF",
			FileSize:       "2.1 MB",
			DownloadsCount: 714,
			IsDefault:      true,
			Author:         "Master Salen",
			Tags:           "schematic, pmic, buck, vph, flowchart, boot sequence",
			CreatedAt:      now.AddDate(0, -3, 0),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_quiz_csv",
			Title:          "Standard 4-Choice Assessment Import Template (CSV)",
			Category:       "quiz",
			Description:    "Pre-formatted spreadsheet template for bulk importing technical multiple choice questions with explanations and pass grades.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_quiz_import_template.csv",
			ThumbnailURL:   "https://images.unsplash.com/photo-1434030216411-0b793f4b4173?w=600",
			FileType:       "CSV",
			FileSize:       "45 KB",
			DownloadsCount: 156,
			IsDefault:      true,
			Author:         "SMG Examination Board",
			Tags:           "quiz, assessment, csv, import, questions",
			CreatedAt:      now.AddDate(0, -1, -20),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_user_csv",
			Title:          "Technician & Student Batch Import Template (CSV)",
			Category:       "worksheet",
			Description:    "Standard CSV columns for onboarding bulk students: Full Name, Email, Phone, Role, Plan, VIP Status.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_student_import_template.csv",
			ThumbnailURL:   "https://images.unsplash.com/photo-1522071820081-009f0129c71c?w=600",
			FileType:       "CSV",
			FileSize:       "32 KB",
			DownloadsCount: 94,
			IsDefault:      false,
			Author:         "SMG Admissions",
			Tags:           "students, csv, batch, import, directory",
			CreatedAt:      now.AddDate(0, -1, 0),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_les_notes",
			Title:          "3-Stage Hardware Repair Lesson Notes Blueprint (Markdown)",
			Category:       "curriculum",
			Description:    "Standardized markdown structure: 1. Symptom & Fault Line, 2. Multimeter Readings, 3. Micro-soldering Execution Steps.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_lesson_notes_blueprint.md",
			ThumbnailURL:   "https://images.unsplash.com/photo-1455390582262-044cdead277a?w=600",
			FileType:       "JSON",
			FileSize:       "18 KB",
			DownloadsCount: 230,
			IsDefault:      true,
			Author:         "Rohit Sharma",
			Tags:           "lesson, markdown, curriculum, blueprint, notes",
			CreatedAt:      now.AddDate(0, -2, -10),
			UpdatedAt:      now,
		},
		{
			ID:             "tmpl_banner_webinar",
			Title:          "Live Masterclass Broadcast Banner Frame (1920x1080)",
			Category:       "banner",
			Description:    "High-res YouTube & Zoom broadcast banner template with instructor frame, brand badge, and session metadata layout.",
			FileURL:        "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/templates/smg_live_banner_frame.png",
			ThumbnailURL:   "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=600",
			FileType:       "PNG",
			FileSize:       "1.8 MB",
			DownloadsCount: 310,
			IsDefault:      true,
			Author:         "SMG Media Team",
			Tags:           "banner, webinar, zoom, live, 1080p",
			CreatedAt:      now.AddDate(0, -1, -5),
			UpdatedAt:      now,
		},
	}

	for _, t := range templates {
		if err := db.Create(&t).Error; err != nil {
			log.Printf("⚠️ Template %s seeding notice: %v", t.ID, err)
		}
	}
	log.Printf("✅ Seeded %d official template blueprints", len(templates))
}
