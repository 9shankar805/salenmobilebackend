package database

import (
"encoding/json"
"log"
"time"

"gorm.io/gorm"
"smg-backend/internal/models"
)

// SeedPlansAndHome seeds Mobile Plans, Chat Plans, PDF Container categories,
// Repair Toolkit tool items, Support Categories, and Home Highlights.
func SeedPlansAndHome(db *gorm.DB) {
now := time.Now()
seedMobilePlans(db, now)
seedChatPlans(db, now)
seedPdfCategories(db, now)
seedSupportCategories(db, now)
seedHighlights(db, now)
seedRepairToolkitItems(db)
seedHomeIntroVideo(db, now)
log.Println("✅ Seeded plans, PDF categories, support tree, highlights, toolkit items & intro video!")
}

// ═══════════════════════════════════════════════════════════════════════════
// 1. MOBILE PLANS
// ═══════════════════════════════════════════════════════════════════════════

func seedMobilePlans(db *gorm.DB, now time.Time) {
mobilePlans := []models.Plan{
{
ID: "mp_1month", PlanType: "MOBILE", Name: "1 Month Starter Pass",
DurationLabel: "1 Month Access", DurationMonths: 1,
Price: 249, OriginalPrice: 399, DiscountPercent: 38,
Currency: "INR", CurrencySymbol: "₹", Badge: "STARTER",
Features: mustJSON([]string{
"Full HD Video Lessons Access",
"Motherboard Schematic Diagrams",
"Multimeter & Diode Mode Procedures",
"Hardware Diagnostic Guides",
"Student Discussion Community",
}),
IsActive: true, OrderNum: 1, CreatedAt: now,
},
{
ID: "mp_3month", PlanType: "MOBILE", Name: "3 Months Pro Technician",
DurationLabel: "3 Months Access", DurationMonths: 3,
Price: 399, OriginalPrice: 599, DiscountPercent: 33,
Currency: "INR", CurrencySymbol: "₹", Badge: "POPULAR",
Features: mustJSON([]string{
"Full HD Video Lessons Access",
"Motherboard Schematic Diagrams",
"Multimeter & Diode Mode Procedures",
"PMIC & Power Sequence Schematics",
"Weekly Live Q&A Sessions",
"Priority Student Support",
}),
IsActive: true, OrderNum: 2, CreatedAt: now,
},
{
ID: "mp_6month", PlanType: "MOBILE", Name: "6 Months Master Technician",
DurationLabel: "6 Months Access", DurationMonths: 6,
Price: 699, OriginalPrice: 1199, DiscountPercent: 42,
Currency: "INR", CurrencySymbol: "₹", Badge: "RECOMMENDED",
Features: mustJSON([]string{
"Full HD Video Lessons Access",
"Motherboard Schematic Diagrams",
"Multimeter & Diode Mode Procedures",
"Advanced BUCK & LDO Tracing",
"Double-Decker CPU & EMMC Lab",
"Direct Instructor Guidance",
"Official Course Certificate",
}),
IsActive: true, OrderNum: 3, CreatedAt: now,
},
{
ID: "mp_1year", PlanType: "MOBILE", Name: "1 Year VIP All-Access Pass",
DurationLabel: "12 Months VIP", DurationMonths: 12,
Price: 999, OriginalPrice: 1999, DiscountPercent: 50,
Currency: "INR", CurrencySymbol: "₹", Badge: "MAX SAVINGS",
Features: mustJSON([]string{
"All Courses & Masterclasses Unlocked",
"Full Schematics & Circuit Diagrams Vault",
"Interactive Diode Mode & Jumper Diagnostics",
"Live Masterclasses with Master Salen",
"Priority 1-on-1 Hardware Support",
"Official Academy Completion Certificate",
"Lifetime Community Network",
}),
IsActive: true, OrderNum: 4, CreatedAt: now,
},
}
upsertPlans(db, mobilePlans)
}

// ═══════════════════════════════════════════════════════════════════════════
// 2. CHAT PLANS
// ═══════════════════════════════════════════════════════════════════════════

func seedChatPlans(db *gorm.DB, now time.Time) {
chatPlans := []models.Plan{
{
ID: "cp_monthly", PlanType: "CHAT", Name: "Monthly Plan",
DurationLabel: "Monthly", DurationMonths: 1,
Price: 699, OriginalPrice: 899, DiscountPercent: 22,
Currency: "INR", CurrencySymbol: "₹", Badge: "",
ResponseTimeMins: 20,
Features: mustJSON([]string{
"20 Mins Response Time", "Unlimited Chats", "Expert Support",
}),
IsActive: true, OrderNum: 1, CreatedAt: now,
},
{
ID: "cp_3month", PlanType: "CHAT", Name: "3 Month Plan",
DurationLabel: "3 Months", DurationMonths: 3,
Price: 1799, OriginalPrice: 2697, DiscountPercent: 33,
Currency: "INR", CurrencySymbol: "₹", Badge: "MOST POPULAR",
ResponseTimeMins: 10,
Features: mustJSON([]string{
"10 Mins Response Time", "Unlimited Chats", "Priority Support",
}),
IsActive: true, OrderNum: 2, CreatedAt: now,
},
{
ID: "cp_6month", PlanType: "CHAT", Name: "6 Month Plan",
DurationLabel: "6 Months", DurationMonths: 6,
Price: 2999, OriginalPrice: 5394, DiscountPercent: 44,
Currency: "INR", CurrencySymbol: "₹", Badge: "MOST POPULAR",
ResponseTimeMins: 5,
Features: mustJSON([]string{
"5 Mins Response Time", "Unlimited Chats", "Premium Support",
}),
IsActive: true, OrderNum: 3, CreatedAt: now,
},
{
ID: "cp_1year", PlanType: "CHAT", Name: "1 Year Plan",
DurationLabel: "1 Year", DurationMonths: 12,
Price: 4999, OriginalPrice: 10788, DiscountPercent: 54,
Currency: "INR", CurrencySymbol: "₹", Badge: "",
ResponseTimeMins: 5,
Features: mustJSON([]string{
"5 Mins Response Time", "Unlimited Everything", "VIP Support",
}),
IsActive: true, OrderNum: 4, CreatedAt: now,
},
}
upsertPlans(db, chatPlans)
}

func upsertPlans(db *gorm.DB, plans []models.Plan) {
for _, p := range plans {
var existing models.Plan
if err := db.Where("id = ?", p.ID).First(&existing).Error; err != nil {
db.Create(&p)
} else {
existing.Name = p.Name
existing.Price = p.Price
existing.OriginalPrice = p.OriginalPrice
existing.DiscountPercent = p.DiscountPercent
existing.Badge = p.Badge
existing.Features = p.Features
existing.ResponseTimeMins = p.ResponseTimeMins
existing.IsActive = p.IsActive
existing.OrderNum = p.OrderNum
db.Save(&existing)
}
}
}

// ═══════════════════════════════════════════════════════════════════════════
// 3. PDF CONTAINER CATEGORIES  (6 grid items on the Home screen)
// ═══════════════════════════════════════════════════════════════════════════

func seedPdfCategories(db *gorm.DB, now time.Time) {
cats := []models.PdfCategory{
{ID: "pdf_schematics", Title: "Schematics Diagram", Slug: "schematics-diagram", IconAsset: "assets/pdf_icons/schematics_diagram.png", Description: "Full circuit schematics for Android & iPhone motherboards", Route: "/pdf-container/schematics", RequiresSub: true, OrderNum: 1, IsActive: true, CreatedAt: now},
{ID: "pdf_unique", Title: "Unique Schematics", Slug: "unique-schematics", IconAsset: "assets/pdf_icons/unique_schematics.png", Description: "Exclusive SMG schematic solutions not found elsewhere", Route: "/pdf-container/unique-schematics", RequiresSub: true, OrderNum: 2, IsActive: true, CreatedAt: now},
{ID: "pdf_direct", Title: "Direct Solution", Slug: "direct-solution", IconAsset: "assets/pdf_icons/direct_solution.png", Description: "Step-by-step repair solutions with annotated board photos", Route: "/pdf-container/direct-solution", RequiresSub: true, OrderNum: 3, IsActive: true, CreatedAt: now},
{ID: "pdf_bitmap", Title: "Bitmap and Layers", Slug: "bitmap-and-layers", IconAsset: "assets/pdf_icons/bitmap_layers.png", Description: "PCB bitmap overlays with power & data layer highlighting", Route: "/pdf-container/bitmap-layers", RequiresSub: true, OrderNum: 4, IsActive: true, CreatedAt: now},
{ID: "pdf_cpu_bitmap", Title: "CPU Bitmap", Slug: "cpu-bitmap", IconAsset: "assets/pdf_icons/cpu_bitmap.png", Description: "Qualcomm & MediaTek CPU/RAM interposer bitmap maps", Route: "/pdf-container/cpu-bitmap", RequiresSub: true, OrderNum: 5, IsActive: true, CreatedAt: now},
{ID: "pdf_diode", Title: "Diode Value", Slug: "diode-value", IconAsset: "assets/pdf_icons/diode_value.png", Description: "Multimeter diode-mode reference readings for all connectors", Route: "/pdf-container/diode-value", RequiresSub: true, OrderNum: 6, IsActive: true, CreatedAt: now},
}
for _, cat := range cats {
var existing models.PdfCategory
if err := db.Where("id = ?", cat.ID).First(&existing).Error; err != nil {
db.Create(&cat)
} else {
existing.Title = cat.Title
existing.IconAsset = cat.IconAsset
existing.Description = cat.Description
existing.Route = cat.Route
existing.OrderNum = cat.OrderNum
existing.IsActive = cat.IsActive
db.Save(&existing)
}
}
}

// ═══════════════════════════════════════════════════════════════════════════
// 4. SUPPORT CATEGORIES  (SOFTWARE tree + HARDWARE tree)
// ═══════════════════════════════════════════════════════════════════════════

func seedSupportCategories(db *gorm.DB, now time.Time) {
all := []models.SupportCategory{
// Top-level
{ID: "sc_software", ParentID: "", Name: "SOFTWARE", Slug: "software", IconAsset: "assets/support/software_icon.png", Description: "Software flashing, unlocking, and firmware issues", OrderNum: 1, IsActive: true, CreatedAt: now},
{ID: "sc_hardware", ParentID: "", Name: "HARDWARE", Slug: "hardware", IconAsset: "assets/support/hardware_icon.png", Description: "Chip-level board repair and component diagnosis", OrderNum: 2, IsActive: true, CreatedAt: now},
// SOFTWARE children
{ID: "sc_sw_ufs", ParentID: "sc_software", Name: "UFS", Slug: "ufs", IconAsset: "assets/support/ufs_chip.png", Description: "UFS 2.1/3.1 programming, ISP pinout support", OrderNum: 1, IsActive: true, CreatedAt: now},
{ID: "sc_sw_others", ParentID: "sc_software", Name: "Others", Slug: "others", IconAsset: "assets/support/others_icon.png", Description: "FRP bypass, pattern unlock, miscellaneous", OrderNum: 2, IsActive: true, CreatedAt: now},
{ID: "sc_sw_emmc", ParentID: "sc_software", Name: "EMMC", Slug: "emmc", IconAsset: "assets/support/emmc_chip.png", Description: "eMMC 5.1 programming via ISP or programmer box", OrderNum: 3, IsActive: true, CreatedAt: now},
{ID: "sc_sw_usb_flash", ParentID: "sc_software", Name: "USB Flashing", Slug: "usb-flashing", IconAsset: "assets/support/usb_flash.png", Description: "Qualcomm 9008, MTK preloader, Samsung Odin USB flash", OrderNum: 4, IsActive: true, CreatedAt: now},
// HARDWARE children
{ID: "sc_hw_power", ParentID: "sc_hardware", Name: "Power Section", Slug: "power-section", IconAsset: "assets/support/power_ic.png", Description: "PMIC, charging IC, voltage rail diagnosis", OrderNum: 1, IsActive: true, CreatedAt: now},
{ID: "sc_hw_display", ParentID: "sc_hardware", Name: "Display & Touch", Slug: "display-touch", IconAsset: "assets/support/display_icon.png", Description: "Display IC, backlight, FPC connector repair", OrderNum: 2, IsActive: true, CreatedAt: now},
{ID: "sc_hw_cpu", ParentID: "sc_hardware", Name: "CPU / RAM", Slug: "cpu-ram", IconAsset: "assets/support/cpu_icon.png", Description: "CPU reballing, RAM replacement, interposer work", OrderNum: 3, IsActive: true, CreatedAt: now},
{ID: "sc_hw_charging", ParentID: "sc_hardware", Name: "Charging Port", Slug: "charging-port", IconAsset: "assets/support/charging_icon.png", Description: "USB-C, Micro-USB, lightning port and sub-board repair", OrderNum: 4, IsActive: true, CreatedAt: now},
}
for _, sc := range all {
var existing models.SupportCategory
if err := db.Where("id = ?", sc.ID).First(&existing).Error; err != nil {
db.Create(&sc)
} else {
existing.Name = sc.Name
existing.ParentID = sc.ParentID
existing.IconAsset = sc.IconAsset
existing.Description = sc.Description
existing.OrderNum = sc.OrderNum
existing.IsActive = sc.IsActive
db.Save(&existing)
}
}
}

// ═══════════════════════════════════════════════════════════════════════════
// 5. HOME HIGHLIGHTS  (Offers & announcements section)
// ═══════════════════════════════════════════════════════════════════════════

func seedHighlights(db *gorm.DB, now time.Time) {
highlights := []models.Highlight{
{ID: "hl_1", Title: "38% OFF — 1 Month Mobile Plan", Subtitle: "Get full access to schematics, bitmap diagrams, IC finder & more at just ₹249", ImageAsset: "assets/banner_bg_img/banner_1.webp", ActionURL: "/plans/mobile", ActionLabel: "Get Plan", Badge: "OFFER", OrderNum: 1, IsActive: true, CreatedAt: now},
{ID: "hl_2", Title: "New: UFS 4.0 Pinout Guide Added", Subtitle: "Samsung S24 & Pixel 8 Pro UFS 4.0 BGA-297 ISP pinouts now in PDF Container", ImageAsset: "assets/banner_bg_img/banner_2.webp", ActionURL: "/pdf-container/schematics", ActionLabel: "View Now", Badge: "NEW", OrderNum: 2, IsActive: true, CreatedAt: now.Add(-48 * time.Hour)},
{ID: "hl_3", Title: "Live Session Saturday — CPU Reballing", Subtitle: "Rohit Sharma demonstrates double-decker Qualcomm reballing live with Q&A", ImageAsset: "assets/banner_bg_img/banner_3.webp", ActionURL: "/live", ActionLabel: "Join Live", Badge: "LIVE", OrderNum: 3, IsActive: true, CreatedAt: now.Add(-24 * time.Hour)},
{ID: "hl_4", Title: "54% OFF — Annual Chat Plan ₹4999", Subtitle: "5 min response time, VIP support, unlimited expert chat for a full year", ImageAsset: "assets/banner_bg_img/banner_4.webp", ActionURL: "/plans/chat", ActionLabel: "Subscribe", Badge: "OFFER", OrderNum: 4, IsActive: true, CreatedAt: now.Add(-72 * time.Hour)},
}
for _, h := range highlights {
var existing models.Highlight
if err := db.Where("id = ?", h.ID).First(&existing).Error; err != nil {
db.Create(&h)
} else {
existing.Title = h.Title
existing.Subtitle = h.Subtitle
existing.ImageAsset = h.ImageAsset
existing.ActionURL = h.ActionURL
existing.ActionLabel = h.ActionLabel
existing.Badge = h.Badge
existing.OrderNum = h.OrderNum
existing.IsActive = h.IsActive
db.Save(&existing)
}
}
}

// ═══════════════════════════════════════════════════════════════════════════
// 6. REPAIR TOOLKIT ITEMS  (IC Finder, PIN Abbreviations, Parts Finder)
// ═══════════════════════════════════════════════════════════════════════════

func seedRepairToolkitItems(db *gorm.DB) {
items := []models.ToolItem{
{ID: "tool_ic_finder", Title: "IC Finder", Category: "Repair Toolkit", Description: "Identify any IC chip by markings, package type or function", Route: "/tools/ic-finder", Icon: "assets/toolkit/ic_finder.png"},
{ID: "tool_pin_abbr", Title: "PIN Abbreviations", Category: "Repair Toolkit", Description: "Full reference list of mobile PCB pin name abbreviations", Route: "/tools/pin-abbreviations", Icon: "assets/toolkit/pin_abbr.png"},
{ID: "tool_parts", Title: "Parts Finder", Category: "Repair Toolkit", Description: "Find compatible replacement parts by model and component", Route: "/tools/parts-finder", Icon: "assets/toolkit/parts_finder.png"},
}
for _, item := range items {
var existing models.ToolItem
if err := db.Where("id = ?", item.ID).First(&existing).Error; err != nil {
db.Create(&item)
} else {
existing.Title = item.Title
existing.Category = item.Category
existing.Description = item.Description
existing.Route = item.Route
existing.Icon = item.Icon
db.Save(&existing)
}
}
}

// ═══════════════════════════════════════════════════════════════════════════
// 7. HOME INTRO VIDEO
// ═══════════════════════════════════════════════════════════════════════════

func seedHomeIntroVideo(db *gorm.DB, now time.Time) {
	var existing models.HomeIntroVideo
	if err := db.Where("id = ?", "main_intro_video").First(&existing).Error; err != nil {
		video := models.HomeIntroVideo{
			ID:           "main_intro_video",
			Title:        "Welcome to Salen Mobile Guru | Masterclass & Academy Tour",
			Subtitle:     "Watch this introductory video to learn how to access chip-level courses, circuit schematics, and live training.",
			YouTubeURL:   "https://www.youtube.com/watch?v=3e_K4h_K6r8",
			YouTubeID:    "3e_K4h_K6r8",
			ThumbnailURL: "https://img.youtube.com/vi/3e_K4h_K6r8/hqdefault.jpg",
			Badge:        "APP TOUR & INTRODUCTION",
			Duration:     "18:45",
			IsActive:     true,
			UpdatedAt:    now,
		}
		db.Create(&video)
	}
}

// mustJSON marshals a value to a JSON string; panics only on seed data errors.
func mustJSON(v interface{}) string {
b, err := json.Marshal(v)
if err != nil {
panic(err)
}
return string(b)
}
