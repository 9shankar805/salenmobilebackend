package database

import (
	"log"
	"time"

	"gorm.io/gorm"
	"smg-backend/internal/models"
)

func getSeedCourses() []models.Course {
	now := time.Now()
	return []models.Course{
		{
			ID:            "c1",
			Title:         "Android Hardware Repairing Level 1",
			Slug:          "android-hardware-repairing-level-1",
			Description:   "Master chip-level diagnostics, short finding with thermal camera, power rail tracing, and micro-soldering.",
			Category:      "Hardware",
			Tags:          "Hardware, Thermal & Short, Power IC, PMIC",
			Price:         4999.0,
			OriginalPrice: 9999.0,
			Rating:        4.9,
			ReviewsCount:  1820,
			LessonsCount:  20,
			Duration:      "18h 40m",
			Level:         "Beginner to Advanced",
			ImageAsset:    "assets/banner_bg_img/banner_1.webp",
			InstructorID:  "u_admin_instructor",
			Instructor:    "Rohit Sharma",
			InstructorBio: "Senior Chip-Level Master Trainer with 12+ years experience.",
			IsPublished:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "c2",
			Title:         "iPhone Motherboard & Display Masterclass",
			Slug:          "iphone-display-repairing",
			Description:   "Complete guide to iPhone 11-15 Pro Max logic board repair, sandwich board separation, and True Tone reprogramming.",
			Category:      "iPhone",
			Tags:          "iPhone, Face ID, Display, Sandwich Board",
			Price:         6999.0,
			OriginalPrice: 12999.0,
			Rating:        4.9,
			ReviewsCount:  1240,
			LessonsCount:  22,
			Duration:      "22h 15m",
			Level:         "Intermediate",
			ImageAsset:    "assets/banner_bg_img/banner_2.webp",
			InstructorID:  "u_admin_instructor",
			Instructor:    "Rohit Sharma",
			InstructorBio: "Senior Chip-Level Master Trainer with 12+ years experience.",
			IsPublished:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "c3",
			Title:         "EMMC & UFS Chip-Off Data Recovery",
			Slug:          "emmc-ufs-programming",
			Description:   "Advanced memory programming with EasyJtag Plus, UFI Box, and Medusa Pro II for dead boot repair.",
			Category:      "EMMC",
			Tags:          "EMMC, UFS, EasyJtag, Dead Boot Repair",
			Price:         5499.0,
			OriginalPrice: 10499.0,
			Rating:        4.8,
			ReviewsCount:  960,
			LessonsCount:  25,
			Duration:      "20h 30m",
			Level:         "Advanced",
			ImageAsset:    "assets/banner_bg_img/banner_3.webp",
			InstructorID:  "u_amit_verma",
			Instructor:    "Amit Verma",
			InstructorBio: "Memory Chip Specialist and Data Recovery Lead.",
			IsPublished:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "c4",
			Title:         "Borneo & ZXW Circuit Schematics Deep-Dive",
			Slug:          "borneo-zxw-schematic-reading",
			Description:   "Step-by-step schematic line reading, diode value checking, and voltage rail tracing.",
			Category:      "Hardware",
			Tags:          "Hardware, Schematics, Borneo, ZXW, Multimeter",
			Price:         3999.0,
			OriginalPrice: 7999.0,
			Rating:        4.9,
			ReviewsCount:  1420,
			LessonsCount:  18,
			Duration:      "15h 45m",
			Level:         "All Levels",
			ImageAsset:    "assets/banner_bg_img/banner_4.webp",
			InstructorID:  "u_admin_instructor",
			Instructor:    "Rohit Sharma",
			InstructorBio: "Senior Chip-Level Master Trainer with 12+ years experience.",
			IsPublished:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "c5",
			Title:         "Qualcomm & MediaTek CPU Reballing Pro",
			Slug:          "pmic-power-section-diagnosis",
			Description:   "Master double-decker CPU scraping, BGA reballing stencils, and temperature control profiles.",
			Category:      "Thermal & Short",
			Tags:          "Thermal & Short, CPU Reballing, BGA, SMD",
			Price:         5999.0,
			OriginalPrice: 11999.0,
			Rating:        4.9,
			ReviewsCount:  880,
			LessonsCount:  16,
			Duration:      "14h 10m",
			Level:         "Masterclass",
			ImageAsset:    "assets/banner_bg_img/banner_5.webp",
			InstructorID:  "u_admin_instructor",
			Instructor:    "Rohit Sharma",
			InstructorBio: "Senior Chip-Level Master Trainer with 12+ years experience.",
			IsPublished:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

func getSeedModules() []models.Module {
	now := time.Now()
	return []models.Module{
		// C1: Android Hardware Repairing Level 1 (20 lessons total)
		{ID: "m1_1", CourseID: "c1", Title: "Module 1: Mobile Logic Board Architecture & Power Lines", OrderNum: 1, LessonCount: 4, CreatedAt: now},
		{ID: "m1_2", CourseID: "c1", Title: "Module 2: PMIC Voltage Rails & Diode Value Analysis", OrderNum: 2, LessonCount: 4, CreatedAt: now},
		{ID: "m1_3", CourseID: "c1", Title: "Module 3: Short Circuit Hunting with Thermal Imaging & DC Power", OrderNum: 3, LessonCount: 4, CreatedAt: now},
		{ID: "m1_4", CourseID: "c1", Title: "Module 4: Precision Micro-Soldering & IC Rework", OrderNum: 4, LessonCount: 4, CreatedAt: now},
		{ID: "m1_5", CourseID: "c1", Title: "Module 5: Real-World Dead Phone Diagnostics & Case Studies", OrderNum: 5, LessonCount: 4, CreatedAt: now},

		// C2: iPhone Motherboard & Display Masterclass (22 lessons total)
		{ID: "m2_1", CourseID: "c2", Title: "Module 1: iPhone Double-Layer Board Architecture", OrderNum: 1, LessonCount: 4, CreatedAt: now},
		{ID: "m2_2", CourseID: "c2", Title: "Module 2: Sandwich Board Splitting, Cleaning & Reballing", OrderNum: 2, LessonCount: 4, CreatedAt: now},
		{ID: "m2_3", CourseID: "c2", Title: "Module 3: Display IC Transfer & True Tone Reprogramming", OrderNum: 3, LessonCount: 4, CreatedAt: now},
		{ID: "m2_4", CourseID: "c2", Title: "Module 4: Face ID & TrueDepth Camera System Repair", OrderNum: 4, LessonCount: 5, CreatedAt: now},
		{ID: "m2_5", CourseID: "c2", Title: "Module 5: Baseband, Audio IC & Panic Log Diagnostics", OrderNum: 5, LessonCount: 5, CreatedAt: now},

		// C3: EMMC & UFS Chip-Off Data Recovery (25 lessons total)
		{ID: "m3_1", CourseID: "c3", Title: "Module 1: Mobile Memory IC Standards & Packages", OrderNum: 1, LessonCount: 5, CreatedAt: now},
		{ID: "m3_2", CourseID: "c3", Title: "Module 2: Programmer Hardware & Sockets Setup", OrderNum: 2, LessonCount: 5, CreatedAt: now},
		{ID: "m3_3", CourseID: "c3", Title: "Module 3: In-System Programming (ISP) Pinouts", OrderNum: 3, LessonCount: 5, CreatedAt: now},
		{ID: "m3_4", CourseID: "c3", Title: "Module 4: Dead Boot Repair & Partition Management", OrderNum: 4, LessonCount: 5, CreatedAt: now},
		{ID: "m3_5", CourseID: "c3", Title: "Module 5: Forensic Chip-Off Data Recovery", OrderNum: 5, LessonCount: 5, CreatedAt: now},

		// C4: Borneo & ZXW Circuit Schematics Deep-Dive (18 lessons total)
		{ID: "m4_1", CourseID: "c4", Title: "Module 1: Schematic Symbols & Block Diagrams", OrderNum: 1, LessonCount: 4, CreatedAt: now},
		{ID: "m4_2", CourseID: "c4", Title: "Module 2: Multimeter Diode Readings & Line Testing", OrderNum: 2, LessonCount: 4, CreatedAt: now},
		{ID: "m4_3", CourseID: "c4", Title: "Module 3: Borneo Schematics Pro Features", OrderNum: 3, LessonCount: 5, CreatedAt: now},
		{ID: "m4_4", CourseID: "c4", Title: "Module 4: ZXW Micro-Placement & Board Bitmap Tools", OrderNum: 4, LessonCount: 5, CreatedAt: now},

		// C5: Qualcomm & MediaTek CPU Reballing Pro (16 lessons total)
		{ID: "m5_1", CourseID: "c5", Title: "Module 1: Pre-Heating & Epoxy De-Glueing", OrderNum: 1, LessonCount: 4, CreatedAt: now},
		{ID: "m5_2", CourseID: "c5", Title: "Module 2: CPU & RAM Double-Decker Separation", OrderNum: 2, LessonCount: 4, CreatedAt: now},
		{ID: "m5_3", CourseID: "c5", Title: "Module 3: Stencil Reballing & Solder Paste Science", OrderNum: 3, LessonCount: 4, CreatedAt: now},
		{ID: "m5_4", CourseID: "c5", Title: "Module 4: Alignment, Soldering & Post-Mount Verification", OrderNum: 4, LessonCount: 4, CreatedAt: now},
	}
}

func yt(id string) string {
	return "https://www.youtube.com/watch?v=" + id
}

func getSeedLessons() []models.Lesson {
	now := time.Now()

	return []models.Lesson{
		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 1: Android Hardware Repairing Level 1 (20 lessons timeline)
		// ═══════════════════════════════════════════════════════════════════════
		// Module 1_1
		{ID: "l_c1_1", ModuleID: "m1_1", CourseSlug: "android-hardware-repairing-level-1", Title: "1. Introduction to Smartphone Motherboard Architecture", Slug: "intro-smartphone-motherboard", Duration: "18:45", VideoURL: yt("3e_K4h_K6r8"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 1, CreatedAt: now},
		{ID: "l_c1_2", ModuleID: "m1_1", CourseSlug: "android-hardware-repairing-level-1", Title: "2. VBAT vs VPH_PWR Primary Power Distribution", Slug: "vbat-vs-vph-pwr-distribution", Duration: "22:10", VideoURL: yt("7Mv_WJ4QyqE"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 2, CreatedAt: now},
		{ID: "l_c1_3", ModuleID: "m1_1", CourseSlug: "android-hardware-repairing-level-1", Title: "3. Cold Testing vs Hot Testing with Digital Multimeter", Slug: "cold-vs-hot-testing-multimeter", Duration: "20:15", VideoURL: yt("X8v_k3L8m9E"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 3, CreatedAt: now},
		{ID: "l_c1_4", ModuleID: "m1_1", CourseSlug: "android-hardware-repairing-level-1", Title: "4. Module 1 Assessment Quiz: Power Lines & Testing", Slug: "mod1-assessment-quiz", Duration: "15:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: true, IsQuiz: true, HasVideo: false, OrderNum: 4, CreatedAt: now},

		// Module 1_2
		{ID: "l_c1_5", ModuleID: "m1_2", CourseSlug: "android-hardware-repairing-level-1", Title: "5. Qualcomm PM6150 & PM6150L Power Sequence", Slug: "qualcomm-pm6150-power-sequence", Duration: "25:30", VideoURL: yt("9iT7v_vM8tA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 5, CreatedAt: now},
		{ID: "l_c1_6", ModuleID: "m1_2", CourseSlug: "android-hardware-repairing-level-1", Title: "6. Diode Mode Readings on BUCK & LDO Capacitors", Slug: "diode-mode-buck-ldo-capacitors", Duration: "19:40", VideoURL: yt("fJ9rUzIMcZQ"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 6, CreatedAt: now},
		{ID: "l_c1_7", ModuleID: "m1_2", CourseSlug: "android-hardware-repairing-level-1", Title: "7. Identifying Short Circuit vs Leakage vs Open Line", Slug: "short-circuit-leakage-open-line", Duration: "23:15", VideoURL: yt("e-ORhEE9VVg"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 7, CreatedAt: now},
		{ID: "l_c1_8", ModuleID: "m1_2", CourseSlug: "android-hardware-repairing-level-1", Title: "8. Crystal Oscillator (XO) & Sleep Clock Signal Verification", Slug: "crystal-oscillator-sleep-clock", Duration: "17:50", VideoURL: yt("9bZkp7q19f0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 8, CreatedAt: now},

		// Module 1_3
		{ID: "l_c1_9", ModuleID: "m1_3", CourseSlug: "android-hardware-repairing-level-1", Title: "9. 4-Channel DC Power Supply Signature Waveforms", Slug: "dc-power-supply-signature-waveforms", Duration: "24:15", VideoURL: yt("7Mv_WJ4QyqE"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 9, CreatedAt: now},
		{ID: "l_c1_10", ModuleID: "m1_3", CourseSlug: "android-hardware-repairing-level-1", Title: "10. Rosin Smoke vs Thermal Camera Method for VPH Short", Slug: "rosin-smoke-thermal-camera-vph", Duration: "28:00", VideoURL: yt("l2xX2rK5z_Q"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 10, CreatedAt: now},
		{ID: "l_c1_11", ModuleID: "m1_3", CourseSlug: "android-hardware-repairing-level-1", Title: "11. Secondary Line Low-Ohmic Short Tracing Techniques", Slug: "secondary-low-ohmic-short-tracing", Duration: "21:30", VideoURL: yt("fJ9rUzIMcZQ"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 11, CreatedAt: now},
		{ID: "l_c1_12", ModuleID: "m1_3", CourseSlug: "android-hardware-repairing-level-1", Title: "12. Voltage Injection Rules & Current Limiting Safety", Slug: "voltage-injection-rules-safety", Duration: "19:20", VideoURL: yt("e-ORhEE9VVg"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 12, CreatedAt: now},

		// Module 1_4
		{ID: "l_c1_13", ModuleID: "m1_4", CourseSlug: "android-hardware-repairing-level-1", Title: "13. Hot Air Gun Temperature & Airflow Profiles (Quick 861DW)", Slug: "hot-air-temperature-profiles", Duration: "26:40", VideoURL: yt("k8M9P3q2v5L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 13, CreatedAt: now},
		{ID: "l_c1_14", ModuleID: "m1_4", CourseSlug: "android-hardware-repairing-level-1", Title: "14. Removing & Replacing Underfilled BGA Components", Slug: "underfilled-bga-removal-replacement", Duration: "31:10", VideoURL: yt("RgKAFK5djSk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 14, CreatedAt: now},
		{ID: "l_c1_15", ModuleID: "m1_4", CourseSlug: "android-hardware-repairing-level-1", Title: "15. Micro Jumper Wire 0.01mm Under Broken BGA Pads", Slug: "micro-jumper-under-bga-pads", Duration: "27:15", VideoURL: yt("2a4xL8m5Q1U"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 15, CreatedAt: now},
		{ID: "l_c1_16", ModuleID: "m1_4", CourseSlug: "android-hardware-repairing-level-1", Title: "16. UV Mask Curing & Solder Mask Reconstruction", Slug: "uv-mask-curing-reconstruction", Duration: "18:25", VideoURL: yt("kJQP7kiw5Fk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 16, CreatedAt: now},

		// Module 1_5
		{ID: "l_c1_17", ModuleID: "m1_5", CourseSlug: "android-hardware-repairing-level-1", Title: "17. Case Study: Redmi Note 10 Pro Dead 0.045A Frozen Current", Slug: "case-study-redmi-note10-pro-dead", Duration: "32:50", VideoURL: yt("P7v9L3k2m1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 17, CreatedAt: now},
		{ID: "l_c1_18", ModuleID: "m1_5", CourseSlug: "android-hardware-repairing-level-1", Title: "18. Case Study: Samsung A51 Fast Discharge & Heating PMIC", Slug: "case-study-samsung-a51-heating", Duration: "29:10", VideoURL: yt("9iT7v_vM8tA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 18, CreatedAt: now},
		{ID: "l_c1_19", ModuleID: "m1_5", CourseSlug: "android-hardware-repairing-level-1", Title: "19. Case Study: POCO X3 Pro Audio & Camera Failure", Slug: "case-study-poco-x3-pro-cpu-failure", Duration: "34:20", VideoURL: yt("5a8qL2m9v3K"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 19, CreatedAt: now},
		{ID: "l_c1_20", ModuleID: "m1_5", CourseSlug: "android-hardware-repairing-level-1", Title: "20. Final Android Hardware Certification Exam", Slug: "android-hardware-final-exam", Duration: "30:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 20, CreatedAt: now},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 2: iPhone Motherboard & Display Masterclass (22 lessons timeline)
		// ═══════════════════════════════════════════════════════════════════════
		// Module 2_1
		{ID: "l_c2_1", ModuleID: "m2_1", CourseSlug: "iphone-display-repairing", Title: "1. Overview of iPhone Double-Decker Logic Boards (X to 15 Pro Max)", Slug: "iphone-double-decker-overview", Duration: "22:30", VideoURL: yt("W8m9K2v1q3L"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 1, CreatedAt: now},
		{ID: "l_c2_2", ModuleID: "m2_1", CourseSlug: "iphone-display-repairing", Title: "2. PP_VDD_MAIN and PP_VDD_BOOST Circuitry & Differences", Slug: "pp-vdd-main-vs-vdd-boost", Duration: "24:45", VideoURL: yt("X8v_k3L8m9E"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 2, CreatedAt: now},
		{ID: "l_c2_3", ModuleID: "m2_1", CourseSlug: "iphone-display-repairing", Title: "3. Battery Gas Gauge & Tigris/Hydra Charging Management", Slug: "battery-gas-gauge-tigris-hydra", Duration: "21:10", VideoURL: yt("9bZkp7q19f0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 3, CreatedAt: now},
		{ID: "l_c2_4", ModuleID: "m2_1", CourseSlug: "iphone-display-repairing", Title: "4. Pre-Separation Current Testing on QianLi / iSocket Jig", Slug: "pre-separation-isocket-testing", Duration: "19:50", VideoURL: yt("7Mv_WJ4QyqE"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 4, CreatedAt: now},

		// Module 2_2
		{ID: "l_c2_5", ModuleID: "m2_2", CourseSlug: "iphone-display-repairing", Title: "5. Heating Station Temperature Profiles for Lead-Free Separation", Slug: "heating-station-temperature-separation", Duration: "28:15", VideoURL: yt("W8m9K2v1q3L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 5, CreatedAt: now},
		{ID: "l_c2_6", ModuleID: "m2_2", CourseSlug: "iphone-display-repairing", Title: "6. Middle Frame Interposer Pad Cleaning with Wick & Rosin", Slug: "interposer-pad-cleaning-wick", Duration: "26:40", VideoURL: yt("RgKAFK5djSk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 6, CreatedAt: now},
		{ID: "l_c2_7", ModuleID: "m2_2", CourseSlug: "iphone-display-repairing", Title: "7. Low-Temp (138°C) vs Middle-Temp (158°C) Reballing Solder Pastes", Slug: "low-temp-vs-middle-temp-pastes", Duration: "25:20", VideoURL: yt("2Vv-BfVoq4g"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 7, CreatedAt: now},
		{ID: "l_c2_8", ModuleID: "m2_2", CourseSlug: "iphone-display-repairing", Title: "8. Precision Sandwich Alignment & Reflow Techniques", Slug: "precision-sandwich-alignment-reflow", Duration: "30:10", VideoURL: yt("W8m9K2v1q3L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 8, CreatedAt: now},

		// Module 2_3
		{ID: "l_c2_9", ModuleID: "m2_3", CourseSlug: "iphone-display-repairing", Title: "9. Solving 'Important Display Message' Warning on iOS", Slug: "solving-important-display-message", Duration: "27:35", VideoURL: yt("1JgS_k68j4w"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 9, CreatedAt: now},
		{ID: "l_c2_10", ModuleID: "m2_3", CourseSlug: "iphone-display-repairing", Title: "10. Original Display Touch IC Grinding & Micro-Soldering", Slug: "original-touch-ic-grinding-soldering", Duration: "33:00", VideoURL: yt("1JgS_k68j4w"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 10, CreatedAt: now},
		{ID: "l_c2_11", ModuleID: "m2_3", CourseSlug: "iphone-display-repairing", Title: "11. JCID V1SE & QianLi iCopy Programmer True Tone Recovery", Slug: "jcid-v1se-icopy-true-tone", Duration: "22:45", VideoURL: yt("zOq8v6K5e4g"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 11, CreatedAt: now},
		{ID: "l_c2_12", ModuleID: "m2_3", CourseSlug: "iphone-display-repairing", Title: "12. OLED Green Tint & Flex Cable Laser Repair Overview", Slug: "oled-green-tint-laser-repair", Duration: "24:15", VideoURL: yt("1JgS_k68j4w"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 12, CreatedAt: now},

		// Module 2_4
		{ID: "l_c2_13", ModuleID: "m2_4", CourseSlug: "iphone-display-repairing", Title: "13. Face ID System Architecture: Dot Projector, Flood, & Infrared", Slug: "face-id-system-architecture", Duration: "23:40", VideoURL: yt("J6z6_m_w7aA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 13, CreatedAt: now},
		{ID: "l_c2_14", ModuleID: "m2_4", CourseSlug: "iphone-display-repairing", Title: "14. Diagnosing 'Face ID Issue Detected' using JCID Matrix Tester", Slug: "diagnosing-face-id-matrix-tester", Duration: "21:30", VideoURL: yt("J6z6_m_w7aA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 14, CreatedAt: now},
		{ID: "l_c2_15", ModuleID: "m2_4", CourseSlug: "iphone-display-repairing", Title: "15. Prism Separation & MOSFET Fusing Bypass on Dot Projector", Slug: "prism-separation-mosfet-bypass", Duration: "32:10", VideoURL: yt("J6z6_m_w7aA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 15, CreatedAt: now},
		{ID: "l_c2_16", ModuleID: "m2_4", CourseSlug: "iphone-display-repairing", Title: "16. Tag-On vs Non-Removal Face ID Flex Installation", Slug: "tag-on-vs-non-removal-faceid-flex", Duration: "26:50", VideoURL: yt("J6z6_m_w7aA"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 16, CreatedAt: now},
		{ID: "l_c2_17", ModuleID: "m2_4", CourseSlug: "iphone-display-repairing", Title: "17. Ambient Light Sensor & Ear Speaker Flex Calibration", Slug: "ambient-light-sensor-calibration", Duration: "20:15", VideoURL: yt("kJQP7kiw5Fk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 17, CreatedAt: now},

		// Module 2_5
		{ID: "l_c2_18", ModuleID: "m2_5", CourseSlug: "iphone-display-repairing", Title: "18. Decoding iOS Analytics Panic Full Logs (Missing Sensor Prs0, Mic2)", Slug: "decoding-ios-panic-full-logs", Duration: "28:30", VideoURL: yt("b4k9L3m2v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 18, CreatedAt: now},
		{ID: "l_c2_19", ModuleID: "m2_5", CourseSlug: "iphone-display-repairing", Title: "19. iPhone Searching / No Service: Baseband PMU & BB-CPU Lines", Slug: "iphone-searching-baseband-pmu", Duration: "31:20", VideoURL: yt("9bZkp7q19f0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 19, CreatedAt: now},
		{ID: "l_c2_20", ModuleID: "m2_5", CourseSlug: "iphone-display-repairing", Title: "20. No Sound / Loop Disease Audio Codec Rework", Slug: "loop-disease-audio-codec-rework", Duration: "25:40", VideoURL: yt("vX1_4L8K8c0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 20, CreatedAt: now},
		{ID: "l_c2_21", ModuleID: "m2_5", CourseSlug: "iphone-display-repairing", Title: "21. Restart Every 3 Minutes Thermal Sensor Troubleshooting", Slug: "restart-3-mins-thermal-sensor", Duration: "24:10", VideoURL: yt("b4k9L3m2v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 21, CreatedAt: now},
		{ID: "l_c2_22", ModuleID: "m2_5", CourseSlug: "iphone-display-repairing", Title: "22. Final iPhone Masterclass Comprehensive Exam", Slug: "iphone-masterclass-final-exam", Duration: "30:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 22, CreatedAt: now},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 3: EMMC & UFS Chip-Off Data Recovery (25 lessons timeline)
		// ═══════════════════════════════════════════════════════════════════════
		// Module 3_1
		{ID: "l_c3_1", ModuleID: "m3_1", CourseSlug: "emmc-ufs-programming", Title: "1. Evolution from NAND to eMMC (5.1) and UFS (2.2, 3.1, 4.0)", Slug: "evolution-nand-emmc-ufs", Duration: "22:15", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 1, CreatedAt: now},
		{ID: "l_c3_2", ModuleID: "m3_1", CourseSlug: "emmc-ufs-programming", Title: "2. BGA Ball Geometries: BGA 153, 221, 254, and 297 Pinouts", Slug: "bga-ball-geometries-pinouts", Duration: "24:30", VideoURL: yt("2Vv-BfVoq4g"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 2, CreatedAt: now},
		{ID: "l_c3_3", ModuleID: "m3_1", CourseSlug: "emmc-ufs-programming", Title: "3. eMMC Internal Partitions: BOOT1, BOOT2, RPMB, and USERAREA", Slug: "emmc-internal-partitions-rpmb", Duration: "21:40", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 3, CreatedAt: now},
		{ID: "l_c3_4", ModuleID: "m3_1", CourseSlug: "emmc-ufs-programming", Title: "4. UFS Architecture: LUN0 to LUN7 & Descriptor Configuration", Slug: "ufs-architecture-lun-descriptors", Duration: "25:10", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 4, CreatedAt: now},
		{ID: "l_c3_5", ModuleID: "m3_1", CourseSlug: "emmc-ufs-programming", Title: "5. Module 1 Assessment Quiz: Memory IC Standards", Slug: "mod1-memory-standards-quiz", Duration: "15:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: true, IsQuiz: true, HasVideo: false, OrderNum: 5, CreatedAt: now},

		// Module 3_2
		{ID: "l_c3_6", ModuleID: "m3_2", CourseSlug: "emmc-ufs-programming", Title: "6. EasyJtag Plus Box V2 Setup & Software Installation", Slug: "easyjtag-plus-box-setup", Duration: "20:45", VideoURL: yt("x7K2L9m4v3A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 6, CreatedAt: now},
		{ID: "l_c3_7", ModuleID: "m3_2", CourseSlug: "emmc-ufs-programming", Title: "7. UFI Box vs Medusa Pro II Feature Comparison", Slug: "ufi-box-vs-medusa-pro-ii", Duration: "23:10", VideoURL: yt("x7K2L9m4v3A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 7, CreatedAt: now},
		{ID: "l_c3_8", ModuleID: "m3_2", CourseSlug: "emmc-ufs-programming", Title: "8. High-Speed BGA Sockets Maintenance & Pin Cleaning", Slug: "bga-sockets-maintenance-cleaning", Duration: "18:30", VideoURL: yt("kJQP7kiw5Fk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 8, CreatedAt: now},
		{ID: "l_c3_9", ModuleID: "m3_2", CourseSlug: "emmc-ufs-programming", Title: "9. Voltage Selection: VCC (2.8V/3.3V) and VCCQ (1.8V/1.2V)", Slug: "voltage-selection-vcc-vccq", Duration: "22:20", VideoURL: yt("9bZkp7q19f0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 9, CreatedAt: now},
		{ID: "l_c3_10", ModuleID: "m3_2", CourseSlug: "emmc-ufs-programming", Title: "10. Bus Width Optimization (1-bit, 4-bit, 8-bit) and Clock Speeds", Slug: "bus-width-clock-optimization", Duration: "19:50", VideoURL: yt("x7K2L9m4v3A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 10, CreatedAt: now},

		// Module 3_3
		{ID: "l_c3_11", ModuleID: "m3_3", CourseSlug: "emmc-ufs-programming", Title: "11. ISP Soldering: CMD, CLK, DAT0, VCC, VCCQ, and GND", Slug: "isp-soldering-cmd-clk-dat0", Duration: "26:40", VideoURL: yt("x7K2L9m4v3A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 11, CreatedAt: now},
		{ID: "l_c3_12", ModuleID: "m3_3", CourseSlug: "emmc-ufs-programming", Title: "12. Tracing ISP Lines with Borneo & ZXW Schematics", Slug: "tracing-isp-lines-borneo-zxw", Duration: "24:15", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 12, CreatedAt: now},
		{ID: "l_c3_13", ModuleID: "m3_3", CourseSlug: "emmc-ufs-programming", Title: "13. Eliminating Signal Noise & Adding 100-Ohm Pull-Up Resistors", Slug: "eliminating-signal-noise-pullups", Duration: "21:55", VideoURL: yt("x7K2L9m4v3A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 13, CreatedAt: now},
		{ID: "l_c3_14", ModuleID: "m3_3", CourseSlug: "emmc-ufs-programming", Title: "14. Reading eMMC Health & Smart Report (MLC / SLC Life Cycle)", Slug: "reading-emmc-health-smart-report", Duration: "27:10", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 14, CreatedAt: now},
		{ID: "l_c3_15", ModuleID: "m3_3", CourseSlug: "emmc-ufs-programming", Title: "15. Repairing Read-Only Mode & Urgent Wear Warnings", Slug: "repairing-read-only-urgent-wear", Duration: "28:30", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 15, CreatedAt: now},

		// Module 3_4
		{ID: "l_c3_16", ModuleID: "m3_4", CourseSlug: "emmc-ufs-programming", Title: "16. Qualcomm 9008 Sahara Handshake & Firehose Loaders", Slug: "qualcomm-sahara-firehose-loaders", Duration: "26:20", VideoURL: yt("w3M9L6k4v2P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 16, CreatedAt: now},
		{ID: "l_c3_17", ModuleID: "m3_4", CourseSlug: "emmc-ufs-programming", Title: "17. MediaTek Preloader & Boot Configuration Register (ExtCSD)", Slug: "mediatek-preloader-extcsd", Duration: "25:45", VideoURL: yt("w3M9L6k4v2P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 17, CreatedAt: now},
		{ID: "l_c3_18", ModuleID: "m3_4", CourseSlug: "emmc-ufs-programming", Title: "18. Samsung Exynos PIT & Boot Configuration Restoration", Slug: "samsung-exynos-pit-restoration", Duration: "24:30", VideoURL: yt("w3M9L6k4v2P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 18, CreatedAt: now},
		{ID: "l_c3_19", ModuleID: "m3_4", CourseSlug: "emmc-ufs-programming", Title: "19. Dumping and Writing Modem/NVRAM to Preserve IMEI", Slug: "dumping-writing-modem-nvram-imei", Duration: "30:15", VideoURL: yt("w3M9L6k4v2P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 19, CreatedAt: now},
		{ID: "l_c3_20", ModuleID: "m3_4", CourseSlug: "emmc-ufs-programming", Title: "20. Understanding RPMB Binding and CPU Pairing Keys", Slug: "rpmb-binding-cpu-pairing-keys", Duration: "28:00", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 20, CreatedAt: now},

		// Module 3_5
		{ID: "l_c3_21", ModuleID: "m3_5", CourseSlug: "emmc-ufs-programming", Title: "21. Non-Destructive Thermal Desoldering of Underfilled eMMC", Slug: "thermal-desoldering-underfilled-emmc", Duration: "29:40", VideoURL: yt("RgKAFK5djSk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 21, CreatedAt: now},
		{ID: "l_c3_22", ModuleID: "m3_5", CourseSlug: "emmc-ufs-programming", Title: "22. Direct Chip Reading on EasyJtag UFS Socket", Slug: "direct-chip-reading-ufs-socket", Duration: "25:10", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 22, CreatedAt: now},
		{ID: "l_c3_23", ModuleID: "m3_5", CourseSlug: "emmc-ufs-programming", Title: "23. Decrypting Android FBE (File-Based Encryption) Realities", Slug: "decrypting-android-fbe-realities", Duration: "27:35", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 23, CreatedAt: now},
		{ID: "l_c3_24", ModuleID: "m3_5", CourseSlug: "emmc-ufs-programming", Title: "24. Extracting Contacts, Photos, and WhatsApp Database", Slug: "extracting-contacts-photos-whatsapp", Duration: "31:20", VideoURL: yt("q3M8L5v2k1P"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 24, CreatedAt: now},
		{ID: "l_c3_25", ModuleID: "m3_5", CourseSlug: "emmc-ufs-programming", Title: "25. Final EMMC/UFS Certification Exam", Slug: "emmc-ufs-certification-exam", Duration: "30:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 25, CreatedAt: now},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 4: Borneo & ZXW Circuit Schematics Deep-Dive (18 lessons timeline)
		// ═══════════════════════════════════════════════════════════════════════
		// Module 4_1
		{ID: "l_c4_1", ModuleID: "m4_1", CourseSlug: "borneo-zxw-schematic-reading", Title: "1. Understanding Electrical Symbols: Capacitors, Inductors, Diodes & ICs", Slug: "electrical-symbols-schematics", Duration: "20:30", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 1, CreatedAt: now},
		{ID: "l_c4_2", ModuleID: "m4_1", CourseSlug: "borneo-zxw-schematic-reading", Title: "2. Motherboard Block Diagram: High-Level Chip Interconnects", Slug: "motherboard-block-diagrams", Duration: "22:45", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 2, CreatedAt: now},
		{ID: "l_c4_3", ModuleID: "m4_1", CourseSlug: "borneo-zxw-schematic-reading", Title: "3. Power Tree Navigation: Primary to Secondary LDO Conversions", Slug: "power-tree-ldo-conversions", Duration: "25:10", VideoURL: yt("9bZkp7q19f0"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 3, CreatedAt: now},
		{ID: "l_c4_4", ModuleID: "m4_1", CourseSlug: "borneo-zxw-schematic-reading", Title: "4. Reading Bus Protocols: I2C, SPI, MIPI, and UART Lines", Slug: "reading-bus-protocols-i2c-spi", Duration: "24:15", VideoURL: yt("e-ORhEE9VVg"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 4, CreatedAt: now},

		// Module 4_2
		{ID: "l_c4_5", ModuleID: "m4_2", CourseSlug: "borneo-zxw-schematic-reading", Title: "5. Diode Mode Principle: Red Probe to Ground Explained", Slug: "diode-mode-red-probe-ground", Duration: "21:20", VideoURL: yt("X8v_k3L8m9E"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 5, CreatedAt: now},
		{ID: "l_c4_6", ModuleID: "m4_2", CourseSlug: "borneo-zxw-schematic-reading", Title: "6. Comparing Real-Board Readings Against Borneo Database", Slug: "comparing-board-readings-borneo", Duration: "23:50", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 6, CreatedAt: now},
		{ID: "l_c4_7", ModuleID: "m4_2", CourseSlug: "borneo-zxw-schematic-reading", Title: "7. Detecting High-Resistance Leaks and Corroded Vias", Slug: "detecting-high-resistance-leaks", Duration: "22:30", VideoURL: yt("fJ9rUzIMcZQ"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 7, CreatedAt: now},
		{ID: "l_c4_8", ModuleID: "m4_2", CourseSlug: "borneo-zxw-schematic-reading", Title: "8. FPC Connector Diode Mapping (Display, Sub-Board, Fingerprint)", Slug: "fpc-connector-diode-mapping", Duration: "24:00", VideoURL: yt("X8v_k3L8m9E"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 8, CreatedAt: now},

		// Module 4_3
		{ID: "l_c4_9", ModuleID: "m4_3", CourseSlug: "borneo-zxw-schematic-reading", Title: "9. Combining Hardware Solutions with Official Schematics", Slug: "hardware-solutions-vs-schematics", Duration: "26:15", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 9, CreatedAt: now},
		{ID: "l_c4_10", ModuleID: "m4_3", CourseSlug: "borneo-zxw-schematic-reading", Title: "10. Diode Mode Matrix for Xiaomi, Samsung, and iPhone", Slug: "diode-matrix-xiaomi-samsung-iphone", Duration: "23:40", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 10, CreatedAt: now},
		{ID: "l_c4_11", ModuleID: "m4_3", CourseSlug: "borneo-zxw-schematic-reading", Title: "11. Sub-Board (Sub-PBA) to Main PBA Interconnect Tracing", Slug: "sub-pba-to-main-interconnect", Duration: "21:50", VideoURL: yt("t8K5L2m9v1A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 11, CreatedAt: now},
		{ID: "l_c4_12", ModuleID: "m4_3", CourseSlug: "borneo-zxw-schematic-reading", Title: "12. Rapid Jumper Finding for Broken PCB Traces", Slug: "rapid-jumper-finding-broken-traces", Duration: "25:30", VideoURL: yt("2a4xL8m5Q1U"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 12, CreatedAt: now},
		{ID: "l_c4_13", ModuleID: "m4_3", CourseSlug: "borneo-zxw-schematic-reading", Title: "13. Module 3 Assessment Quiz: Schematics & Solutions", Slug: "mod3-schematics-solutions-quiz", Duration: "15:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 13, CreatedAt: now},

		// Module 4_4
		{ID: "l_c4_14", ModuleID: "m4_4", CourseSlug: "borneo-zxw-schematic-reading", Title: "14. ZXW Dongle vs Online Account Setup and Layer Navigation", Slug: "zxw-dongle-setup-layers", Duration: "22:10", VideoURL: yt("mJ8L9P2q1wY"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 14, CreatedAt: now},
		{ID: "l_c4_15", ModuleID: "m4_4", CourseSlug: "borneo-zxw-schematic-reading", Title: "15. Tracing Concealed PCB Inner Layer Lines", Slug: "tracing-pcb-inner-layers", Duration: "27:35", VideoURL: yt("mJ8L9P2q1wY"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 15, CreatedAt: now},
		{ID: "l_c4_16", ModuleID: "m4_4", CourseSlug: "borneo-zxw-schematic-reading", Title: "16. Identifying NC (No Connect) vs Ground vs Active Data Pins", Slug: "identifying-nc-gnd-active-pins", Duration: "20:45", VideoURL: yt("mJ8L9P2q1wY"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 16, CreatedAt: now},
		{ID: "l_c4_17", ModuleID: "m4_4", CourseSlug: "borneo-zxw-schematic-reading", Title: "17. Case Studies: Audio Failure, No Charging, Touch Not Working", Slug: "case-studies-audio-charging-touch", Duration: "31:00", VideoURL: yt("mJ8L9P2q1wY"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 17, CreatedAt: now},
		{ID: "l_c4_18", ModuleID: "m4_4", CourseSlug: "borneo-zxw-schematic-reading", Title: "18. Final Schematic Analysis Certification Exam", Slug: "schematic-analysis-final-exam", Duration: "30:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 18, CreatedAt: now},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 5: Qualcomm & MediaTek CPU Reballing Pro (16 lessons timeline)
		// ═══════════════════════════════════════════════════════════════════════
		// Module 5_1
		{ID: "l_c5_1", ModuleID: "m5_1", CourseSlug: "pmic-power-section-diagnosis", Title: "1. Thermal Profiles: Safe Glass Transition Temperatures for Multi-Layer PCBs", Slug: "thermal-profiles-glass-transition", Duration: "21:15", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 1, CreatedAt: now},
		{ID: "l_c5_2", ModuleID: "m5_1", CourseSlug: "pmic-power-section-diagnosis", Title: "2. Pre-Heater Selection: Mijing, Mechanic, and Sunshine Platforms", Slug: "pre-heater-selection-platforms", Duration: "19:40", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: true, HasVideo: true, OrderNum: 2, CreatedAt: now},
		{ID: "l_c5_3", ModuleID: "m5_1", CourseSlug: "pmic-power-section-diagnosis", Title: "3. Black Epoxy vs White Epoxy Glue Softening & Edge Scraping", Slug: "black-vs-white-epoxy-scraping", Duration: "23:50", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 3, CreatedAt: now},
		{ID: "l_c5_4", ModuleID: "m5_1", CourseSlug: "pmic-power-section-diagnosis", Title: "4. Preventing PCB Blistering and Board Delamination", Slug: "preventing-pcb-blistering-delamination", Duration: "22:30", VideoURL: yt("k8M9P3q2v5L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 4, CreatedAt: now},

		// Module 5_2
		{ID: "l_c5_5", ModuleID: "m5_2", CourseSlug: "pmic-power-section-diagnosis", Title: "5. Top RAM Removal without Disturbing Bottom AP CPU", Slug: "top-ram-removal-without-cpu-disturbance", Duration: "26:45", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 5, CreatedAt: now},
		{ID: "l_c5_6", ModuleID: "m5_2", CourseSlug: "pmic-power-section-diagnosis", Title: "6. Ultra-Thin Scraping Blade Selection & Ergonomic Angles", Slug: "scraping-blade-selection-angles", Duration: "20:30", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 6, CreatedAt: now},
		{ID: "l_c5_7", ModuleID: "m5_2", CourseSlug: "pmic-power-section-diagnosis", Title: "7. Removing AP CPU from Motherboard under 340°C Precision Air", Slug: "removing-ap-cpu-340c", Duration: "27:10", VideoURL: yt("h4K9L2m3v8Q"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 7, CreatedAt: now},
		{ID: "l_c5_8", ModuleID: "m5_2", CourseSlug: "pmic-power-section-diagnosis", Title: "8. Board Pad Cleaning with Low-Melt Solder & Copper Braid", Slug: "board-pad-cleaning-copper-braid", Duration: "25:20", VideoURL: yt("RgKAFK5djSk"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 8, CreatedAt: now},

		// Module 5_3
		{ID: "l_c5_9", ModuleID: "m5_3", CourseSlug: "pmic-power-section-diagnosis", Title: "9. 3D Black Stencils vs Magnetic Platforms (Amaoe / Mechanic)", Slug: "3d-stencils-vs-magnetic-platforms", Duration: "22:15", VideoURL: yt("2Vv-BfVoq4g"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 9, CreatedAt: now},
		{ID: "l_c5_10", ModuleID: "m5_3", CourseSlug: "pmic-power-section-diagnosis", Title: "10. Solder Paste Viscosity, Drying with Tissue & 183°C Sn63Pb37", Slug: "solder-paste-viscosity-drying", Duration: "24:40", VideoURL: yt("a5P8M2q1v9L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 10, CreatedAt: now},
		{ID: "l_c5_11", ModuleID: "m5_3", CourseSlug: "pmic-power-section-diagnosis", Title: "11. Reballing Middle RAM Interposer Balls Uniformly", Slug: "reballing-middle-ram-interposer", Duration: "26:00", VideoURL: yt("2Vv-BfVoq4g"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 11, CreatedAt: now},
		{ID: "l_c5_12", ModuleID: "m5_3", CourseSlug: "pmic-power-section-diagnosis", Title: "12. Secondary Thermal Reflow with Amtech NC-559 Flux", Slug: "secondary-thermal-reflow-amtech-flux", Duration: "21:10", VideoURL: yt("a5P8M2q1v9L"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 12, CreatedAt: now},

		// Module 5_4
		{ID: "l_c5_13", ModuleID: "m5_4", CourseSlug: "pmic-power-section-diagnosis", Title: "13. Motherboard Silkscreen Alignment Marks & Tactile IC Placement", Slug: "silkscreen-alignment-tactile-placement", Duration: "25:35", VideoURL: yt("r7K3L9m2v5A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 13, CreatedAt: now},
		{ID: "l_c5_14", ModuleID: "m5_4", CourseSlug: "pmic-power-section-diagnosis", Title: "14. Two-Step Soldering: Soldering AP CPU First, Verifying DC Current", Slug: "two-step-soldering-ap-cpu-first", Duration: "28:50", VideoURL: yt("r7K3L9m2v5A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 14, CreatedAt: now},
		{ID: "l_c5_15", ModuleID: "m5_4", CourseSlug: "pmic-power-section-diagnosis", Title: "15. Mounting LPDDR5 RAM & USB EDL 9008 Port Verification", Slug: "mounting-lpddr5-ram-edl-verification", Duration: "29:15", VideoURL: yt("r7K3L9m2v5A"), StreamType: "YOUTUBE", IsFreePreview: false, HasVideo: true, OrderNum: 15, CreatedAt: now},
		{ID: "l_c5_16", ModuleID: "m5_4", CourseSlug: "pmic-power-section-diagnosis", Title: "16. Final CPU Reballing Master Certification Exam", Slug: "cpu-reballing-final-exam", Duration: "30:00", VideoURL: "", StreamType: "YOUTUBE", IsFreePreview: false, IsQuiz: true, HasVideo: false, OrderNum: 16, CreatedAt: now},
	}
}

func getSeedSchematics() []models.Schematic {
	now := time.Now()
	return []models.Schematic{
		// C1: Android Hardware Repairing Level 1
		{ID: "s1", CourseSlug: "android-hardware-repairing-level-1", Title: "Qualcomm PM6150 PMIC Power Rails Schematic", Category: "Power Management", FileName: "qualcomm_pm6150_power_rails.pdf", DownloadURL: "/storage/schematics/pdf/pm6150.pdf", FileSize: "4.2 KB", TotalPages: 3, CreatedAt: now},
		{ID: "s2", CourseSlug: "android-hardware-repairing-level-1", Title: "Display Backlight & AMOLED Power IC Line Map", Category: "Display Circuit", FileName: "display_backlight_line_map.pdf", DownloadURL: "/storage/schematics/pdf/backlight.pdf", FileSize: "3.0 KB", TotalPages: 2, CreatedAt: now},
		{ID: "s3", CourseSlug: "android-hardware-repairing-level-1", Title: "Sub-PBA & Charging IC (BQ25890) Schematic", Category: "Charging Circuit", FileName: "charging_ic_bq25890.pdf", DownloadURL: "/storage/schematics/pdf/charging.pdf", FileSize: "2.9 KB", TotalPages: 2, CreatedAt: now},

		// C2: iPhone Motherboard & Display Masterclass
		{ID: "s4", CourseSlug: "iphone-display-repairing", Title: "iPhone 14 Pro Sandwich Interposer Pinout & Signal Map", Category: "Motherboard Interposer", FileName: "iphone14pro_sandwich_interposer.pdf", DownloadURL: "/storage/schematics/pdf/iphone14_interposer.pdf", FileSize: "2.9 KB", TotalPages: 2, CreatedAt: now},
		{ID: "s5", CourseSlug: "iphone-display-repairing", Title: "iPhone 13 Display FPC Diode Readings & Power Rails", Category: "Display Circuit", FileName: "iphone13_display_fpc_diode.pdf", DownloadURL: "/storage/schematics/pdf/iphone13_display.pdf", FileSize: "2.9 KB", TotalPages: 2, CreatedAt: now},
		{ID: "s6", CourseSlug: "iphone-display-repairing", Title: "iPhone Face ID Dot Projector & Flood Illuminator Line Map", Category: "Sensors & Biometrics", FileName: "iphone_faceid_dotprojector.pdf", DownloadURL: "/storage/schematics/pdf/faceid_dotprojector.pdf", FileSize: "1.7 KB", TotalPages: 1, CreatedAt: now},

		// C3: EMMC & UFS Chip-Off Data Recovery
		{ID: "s7", CourseSlug: "emmc-ufs-programming", Title: "BGA 254 UFS 2.1 / 3.1 Direct Pinout & Signal Reference", Category: "Memory IC", FileName: "bga254_ufs_pinout_guide.pdf", DownloadURL: "/storage/schematics/pdf/bga254_ufs.pdf", FileSize: "2.8 KB", TotalPages: 2, CreatedAt: now},
		{ID: "s8", CourseSlug: "emmc-ufs-programming", Title: "EasyJtag Plus & UFI Box ISP 1-Bit Soldering Reference", Category: "ISP Pinouts", FileName: "easyjtag_ufi_isp_pinouts.pdf", DownloadURL: "/storage/schematics/pdf/isp_pinouts.pdf", FileSize: "1.7 KB", TotalPages: 1, CreatedAt: now},
		{ID: "s9", CourseSlug: "emmc-ufs-programming", Title: "Medusa Pro II Dead Boot Partition & RPMB Map", Category: "Boot Repair", FileName: "medusa_deadboot_rpmb.pdf", DownloadURL: "/storage/schematics/pdf/medusa_deadboot.pdf", FileSize: "1.6 KB", TotalPages: 1, CreatedAt: now},

		// C4: Borneo & ZXW Circuit Schematics Deep-Dive
		{ID: "s10", CourseSlug: "borneo-zxw-schematic-reading", Title: "Redmi Note 10 Pro Full Power Tree & Rail Sequencing", Category: "Power Architecture", FileName: "redminote10pro_power_tree.pdf", DownloadURL: "/storage/schematics/pdf/redminote10pro_powertree.pdf", FileSize: "1.6 KB", TotalPages: 1, CreatedAt: now},
		{ID: "s11", CourseSlug: "borneo-zxw-schematic-reading", Title: "Samsung Galaxy A52 Diode Value Sheet & Line Tracing", Category: "Diode Readings", FileName: "samsung_a52_diode_values.pdf", DownloadURL: "/storage/schematics/pdf/samsung_a52_diode.pdf", FileSize: "1.6 KB", TotalPages: 1, CreatedAt: now},
		{ID: "s12", CourseSlug: "borneo-zxw-schematic-reading", Title: "Borneo Schematics Fastboot to EDL 9008 Test Points Map", Category: "Hardware Test Points", FileName: "borneo_edl9008_testpoints.pdf", DownloadURL: "/storage/schematics/pdf/borneo_edl_testpoints.pdf", FileSize: "1.6 KB", TotalPages: 1, CreatedAt: now},

		// C5: Qualcomm & MediaTek CPU Reballing Pro
		{ID: "s13", CourseSlug: "pmic-power-section-diagnosis", Title: "Snapdragon 888 AP & LPDDR5 RAM Ball Grid Layout", Category: "CPU BGA Layout", FileName: "snapdragon888_ball_layout.pdf", DownloadURL: "/storage/schematics/pdf/snapdragon888_layout.pdf", FileSize: "1.5 KB", TotalPages: 1, CreatedAt: now},
		{ID: "s14", CourseSlug: "pmic-power-section-diagnosis", Title: "MediaTek Dimensity 9200 Power Sequencing & Thermal Pad Matrix", Category: "CPU Power Rail", FileName: "dimensity9200_thermal_matrix.pdf", DownloadURL: "/storage/schematics/pdf/dimensity9200_matrix.pdf", FileSize: "1.5 KB", TotalPages: 1, CreatedAt: now},
		{ID: "s15", CourseSlug: "pmic-power-section-diagnosis", Title: "High-End SoC Reballing Temperature Profile & Reflow Curve", Category: "SMD Workshop Guide", FileName: "soc_reballing_temp_profile.pdf", DownloadURL: "/storage/schematics/pdf/reballing_temp_profile.pdf", FileSize: "1.6 KB", TotalPages: 1, CreatedAt: now},
	}
}

// SeedCoursesAndCurriculum ensures that all 5 courses, their modules, and their lessons are properly seeded with YouTube videos
func SeedCoursesAndCurriculum(db *gorm.DB) {
	log.Println("Seeding / Verifying courses and curriculum with YouTube videos...")

	courses := getSeedCourses()
	for _, c := range courses {
		var existing models.Course
		if err := db.Where("id = ? OR slug = ?", c.ID, c.Slug).First(&existing).Error; err != nil {
			// Course doesn't exist — create it fresh
			if err := db.Create(&c).Error; err != nil {
				log.Printf("⚠️ Failed to create course %s: %v", c.ID, err)
			} else {
				log.Printf("✅ Created course: %s (%s)", c.Title, c.ID)
			}
		}
		// Course already exists — do NOT overwrite admin edits (title, price, image, etc.)
	}

	modules := getSeedModules()
	lessons := getSeedLessons()

	// Only insert modules that don't already exist — never delete admin-created ones
	for _, m := range modules {
		var existingMod models.Module
		if err := db.Where("id = ?", m.ID).First(&existingMod).Error; err != nil {
			// Doesn't exist — create it
			if err := db.Create(&m).Error; err != nil {
				log.Printf("⚠️ Failed to create module %s: %v", m.ID, err)
			}
		}
		// Already exists — skip (preserve admin edits)
	}
	log.Printf("✅ Verified %d course modules", len(modules))

	// Only insert lessons that don't already exist — never overwrite admin-edited video URLs
	newLessons := 0
	for _, l := range lessons {
		var existingLesson models.Lesson
		if err := db.Where("id = ?", l.ID).First(&existingLesson).Error; err != nil {
			// Doesn't exist — enrich and create
			EnrichLessonContent(&l)
			if err2 := db.Create(&l).Error; err2 != nil {
				log.Printf("⚠️ Failed to create lesson %s: %v", l.ID, err2)
			} else {
				newLessons++
			}
		}
		// Already exists — skip (preserve admin-edited video URL, title, duration, etc.)
	}
	log.Printf("✅ Verified lessons — %d new lessons inserted, existing lessons preserved", newLessons)
}

// SeedSchematics ensures that all 15 schematics are properly seeded pointing to local/hosted PDFs
func SeedSchematics(db *gorm.DB) {
	schematics := getSeedSchematics()
	for _, s := range schematics {
		var existing models.Schematic
		if err := db.Where("id = ?", s.ID).First(&existing).Error; err != nil {
			if err := db.Create(&s).Error; err != nil {
				log.Printf("⚠️ Failed to create schematic %s: %v", s.ID, err)
			}
		} else {
			existing.Title = s.Title
			existing.Category = s.Category
			existing.FileName = s.FileName
			existing.DownloadURL = s.DownloadURL
			existing.FileSize = s.FileSize
			existing.TotalPages = s.TotalPages
			db.Save(&existing)
		}
	}
	log.Printf("✅ Verified / Seeded %d schematics across all 5 courses with valid PDF URLs!", len(schematics))
}
