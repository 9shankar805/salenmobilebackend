package database

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"smg-backend/internal/models"
)

func parseMySQLURL(rawURL string) string {
	if strings.Contains(rawURL, "@tcp(") {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	if host == "" {
		host = "127.0.0.1:3306"
	} else if !strings.Contains(host, ":") {
		host = host + ":3306"
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	query := u.RawQuery
	if query == "" {
		query = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, pass, host, dbName, query)
}

func InitDB(databaseURL string) *gorm.DB {
	var dialector gorm.Dialector

	if strings.HasPrefix(databaseURL, "mysql://") || strings.Contains(databaseURL, "@tcp(") {
		log.Println("Connecting to MySQL database...")
		dsn := parseMySQLURL(databaseURL)
		dialector = mysql.New(mysql.Config{
			DSN:                       dsn,
			DefaultStringSize:         191,
			SkipInitializeWithVersion: false,
		})
	} else if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		log.Println("Connecting to PostgreSQL database...")
		dialector = postgres.Open(databaseURL)
	} else {
		log.Println("Using SQLite database:", databaseURL)
		dialector = sqlite.Open(databaseURL)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Subscription{},
		&models.Course{},
		&models.Module{},
		&models.Lesson{},
		&models.Exercise{},
		&models.Schematic{},
		&models.CommunityPost{},
		&models.CommunityReply{},
		&models.LiveSession{},
		&models.ToolItem{},
		&models.AppScreenConfig{},
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizOption{},
		&models.QuizResult{},
		&models.MediaAsset{},
		&models.TemplateAsset{},
		&models.HardwareModel{},
		&models.DiodeValue{},
		&models.JumperTrace{},
		&models.RepairPriceGuide{},
		&models.WorkOrder{},
		&models.UserLessonProgress{},
		&models.UserStreak{},
		&models.Certificate{},
		&models.ChatThread{},
		&models.ChatMessage{},
		&models.DeviceToken{},
		&models.InAppNotification{},
		// ── New tables for app screens ──────────────────────────────────────
		&models.Plan{},            // Mobile Plans & Chat Plans
		&models.SupportCategory{}, // SOFTWARE / HARDWARE support tree
		&models.PdfCategory{},     // Home screen PDF Container grid
		&models.Highlight{},       // Home screen Highlights / offers
		// ── Content & Search tables ─────────────────────────────────────────
		&models.Guide{},              // Repair guides & manuals
		&models.Testimonial{},        // Guest home testimonials
		&models.CourseCategory{},     // Guest home category chips
		&models.SimulatorComponent{}, // Mobile simulator IC components
		&models.UserDownload{},       // User downloaded lessons
		// ── Email & SMTP Settings ──────────────────────────────────────────
		&models.SMTPSettings{},
		&models.EmailOTP{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	log.Println("Database schema auto-migrated successfully!")

	// Seed initial data
	SeedDatabase(db)

	return db
}
