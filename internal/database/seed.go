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
