package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name"`
	Email            string    `json:"email" gorm:"size:191;uniqueIndex"`
	Phone            string    `json:"phone"`
	AvatarURL        string    `json:"avatarUrl"`
	Role             string    `json:"role"` // student, instructor, admin
	PasswordHash     string    `json:"-"`
	IsSubscribed     bool      `json:"isSubscribed"`
	IsVIP            bool      `json:"isVip"`
	EmailVerified    bool      `json:"emailVerified"`
	SubscriptionPlan string    `json:"subscriptionPlan"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type Subscription struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	UserID             string    `json:"userId" gorm:"index"`
	PlanType           string    `json:"planType"` // VIP, ANNUAL, LIFETIME, SINGLE_COURSE
	Status             string    `json:"status"`   // active, expired, pending
	Amount             float64   `json:"amount"`
	Currency           string    `json:"currency"`
	PaymentMethod      string    `json:"paymentMethod"` // eSewa, Khalti, IME Pay, Stripe
	TransactionID      string    `json:"transactionId"`
	CourseID           string    `json:"courseId"`
	CourseSlug         string    `json:"courseSlug"`
	EnrolledSlugs      string    `json:"enrolledSlugs"` // comma-separated course slugs
	StartDate          time.Time `json:"startDate"`
	EndDate            time.Time `json:"endDate"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type Course struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug" gorm:"size:191;uniqueIndex"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Tags          string    `json:"tags"` // comma-separated tags
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"originalPrice"`
	Rating        float64   `json:"rating"`
	ReviewsCount  int       `json:"reviews"`
	LessonsCount  int       `json:"lessons"`
	Duration      string    `json:"duration"`
	Level         string    `json:"level"`
	ImageAsset    string    `json:"imageAsset"`
	InstructorID  string    `json:"instructorId"`
	Instructor    string    `json:"instructorName"`
	InstructorBio string    `json:"instructorBio"`
	IsPublished   bool      `json:"isPublished"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	Modules []Module `json:"modules,omitempty" gorm:"foreignKey:CourseID"`
}

type Module struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	CourseID    string    `json:"courseId" gorm:"index"`
	Title       string    `json:"title"`
	LessonCount int       `json:"lessonCount"`
	OrderNum    int       `json:"order"`
	CreatedAt   time.Time `json:"createdAt"`

	Lessons []Lesson `json:"lessons,omitempty" gorm:"foreignKey:ModuleID"`
}

func (Module) TableName() string {
	return "course_modules"
}

type Lesson struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	ModuleID              string    `json:"moduleId" gorm:"index"`
	CourseSlug            string    `json:"courseSlug" gorm:"index"`
	Title                 string    `json:"title"`
	Slug                  string    `json:"slug" gorm:"index"`
	Duration              string    `json:"duration"`
	VideoURL              string    `json:"videoUrl"`
	StreamType            string    `json:"streamType"` // HLS, MP4, YOUTUBE
	IsFreePreview         bool      `json:"isFreePreview"`
	IsQuiz                bool      `json:"isQuiz"`
	HasVideo              bool      `json:"hasVideo"`
	OrderNum              int       `json:"order"`
	Description           string    `json:"description" gorm:"type:text"`
	Content               string    `json:"content" gorm:"type:longtext"`
	KeyPoints             string    `json:"keyPoints,omitempty" gorm:"type:text"`
	ToolsRequired         string    `json:"toolsRequired,omitempty" gorm:"type:text"`
	ExerciseTitle         string    `json:"exerciseTitle,omitempty" gorm:"type:varchar(255)"`
	ExerciseDescription   string    `json:"exerciseDescription,omitempty" gorm:"type:text"`
	ExerciseType          string    `json:"exerciseType,omitempty" gorm:"type:varchar(50)"`
	ExerciseEstimatedMins int       `json:"exerciseEstimatedMins,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

type Exercise struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	CourseID        string    `json:"courseId" gorm:"index"`
	CourseSlug      string    `json:"courseSlug" gorm:"index"`
	ModuleID        string    `json:"moduleId" gorm:"index"`
	LessonID        string    `json:"lessonId,omitempty" gorm:"index"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug" gorm:"index"`
	Type            string    `json:"type"`       // PRACTICAL_LAB, BOARD_DIAGNOSIS, REBALLING, COMPONENT_SWAP, TRACE_REPAIR
	Difficulty      string    `json:"difficulty"` // Beginner, Intermediate, Advanced, Master
	EstimatedMins   int       `json:"estimatedMins"`
	Objective       string    `json:"objective" gorm:"type:text"`
	Instructions    string    `json:"instructions" gorm:"type:longtext"`
	ToolsRequired   string    `json:"toolsRequired" gorm:"type:text"`
	ExpectedResults string    `json:"expectedResults" gorm:"type:text"`
	SafetyChecklist string    `json:"safetyChecklist" gorm:"type:text"`
	PassingCriteria string    `json:"passingCriteria" gorm:"type:text"`
	OrderNum        int       `json:"order"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Schematic struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	CourseSlug  string    `json:"courseSlug" gorm:"index"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	FileName    string    `json:"fileName"`
	DownloadURL string    `json:"downloadUrl"`
	FileSize    string    `json:"fileSize"`
	TotalPages  int       `json:"totalPages"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CommunityPost struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organizationId" gorm:"index"`
	CourseID       string    `json:"courseId"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	AuthorName     string    `json:"authorName"`
	AuthorRole     string    `json:"authorRole"`
	AuthorAvatar   string    `json:"authorAvatar"`
	Upvotes        int       `json:"upvotes"`
	RepliesCount   int       `json:"replies"`
	IsSolved       bool      `json:"isSolved"`
	Solution       string    `json:"solution,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`

	Replies []CommunityReply `json:"repliesList,omitempty" gorm:"foreignKey:PostID"`
}

type CommunityReply struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	PostID       string    `json:"postId" gorm:"index"`
	AuthorName   string    `json:"authorName"`
	AuthorRole   string    `json:"authorRole"`
	AuthorAvatar string    `json:"authorAvatar,omitempty"`
	Body         string    `json:"body"`
	Upvotes      int       `json:"upvotes"`
	IsSolution   bool      `json:"isSolution"`
	CreatedAt    time.Time `json:"createdAt"`
}

type LiveSession struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"`
	Instructor    string    `json:"instructor"`
	DateFormatted string    `json:"date"`
	TimeFormatted string    `json:"time"`
	Status        string    `json:"status"`   // UPCOMING, LIVE, COMPLETED
	Platform      string    `json:"platform"` // ZOOM, GOOGLE_MEET, TEAMS, YOUTUBE_LIVE
	JoinURL       string    `json:"joinUrl"`
	MeetingID     string    `json:"meetingId"` // Zoom ID or Meet code (e.g. abc-defg-hij)
	ZoomID        string    `json:"zoomId"`    // kept for backward compat
	Passcode      string    `json:"passcode"`
	RecordingURL  string    `json:"recordingUrl"` // post-session replay link
	ImageAsset    string    `json:"imageAsset"`   // thumbnail for the session card
	DurationMins  int       `json:"durationMins"` // scheduled duration
	ViewersCount  int       `json:"viewersCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ToolItem struct {
	ID          string `json:"id" gorm:"primaryKey"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Route       string `json:"route"`
	Icon        string `json:"icon"`
}

type AppScreenConfig struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SiteName   string    `json:"siteName" gorm:"index"`
	ScreenKey  string    `json:"screenKey" gorm:"index"`          // guest_home, member_home, live_sessions, tools
	ConfigJSON string    `json:"configJson" gorm:"type:mediumtext"` // Serialized JSON array of blockType slots
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Quiz struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	CourseSlug     string    `json:"courseSlug" gorm:"index"`
	ModuleID       string    `json:"moduleId"`
	Title          string    `json:"title"`
	DurationMins   int       `json:"durationMinutes"`
	TotalQuestions int       `json:"totalQuestions"`
	PassingScore   int       `json:"passingScorePercent"` // e.g. 80
	CreatedAt      time.Time `json:"createdAt"`

	Questions []QuizQuestion `json:"questions,omitempty" gorm:"foreignKey:QuizID"`
}

type QuizQuestion struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	QuizID        string       `json:"quizId" gorm:"index"`
	QuestionText  string       `json:"question"`
	Explanation   string       `json:"explanation"`
	OrderNum      int          `json:"order"`
	CorrectOption int          `json:"correctOptionIndex"` // 0-based
	Options       []QuizOption `json:"options,omitempty" gorm:"foreignKey:QuestionID"`
}

type QuizOption struct {
	ID         string `json:"id" gorm:"primaryKey"`
	QuestionID string `json:"questionId" gorm:"index"`
	Text       string `json:"text"`
	Index      int    `json:"index" gorm:"column:option_index"`
}

// QuizResult stores every attempt a student makes on a quiz.
type QuizResult struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	QuizID       string    `json:"quizId" gorm:"index"`
	CourseSlug   string    `json:"courseSlug" gorm:"index"`
	UserID       string    `json:"userId" gorm:"index"`
	ScorePercent int       `json:"scorePercent"`
	CorrectCount int       `json:"correctCount"`
	TotalQ       int       `json:"totalQuestions"`
	IsPassed     bool      `json:"isPassed"`
	XPEarned     int       `json:"xpEarned"`
	QuizTitle    string    `json:"quizTitle"`
	AttemptedAt  time.Time `json:"attemptedAt"`
}

type MediaAsset struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name"`
	Kind          string    `json:"kind"` // video, audio, image, document, schematic
	URL           string    `json:"url"`
	SizeFormatted string    `json:"size"`
	Category      string    `json:"category"`
	CreatedAt     time.Time `json:"createdAt"`
}

type TemplateAsset struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Title          string    `json:"title"`
	Category       string    `json:"category"` // certificate, schematic, worksheet, quiz, curriculum, notification, banner, diagnostic
	Description    string    `json:"description"`
	FileURL        string    `json:"fileUrl"`
	ThumbnailURL   string    `json:"thumbnailUrl"`
	FileType       string    `json:"fileType"` // PDF, PNG, SVG, JSON, CSV, ZIP
	FileSize       string    `json:"fileSize"`
	DownloadsCount int       `json:"downloadsCount"`
	IsDefault      bool      `json:"isDefault"`
	Author         string    `json:"author"`
	Tags           string    `json:"tags"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// HARDWARE SCHEMATICS & DIODE MODE
// ═══════════════════════════════════════════════════════════════════════════

type HardwareModel struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name"`              // e.g. "Redmi Note 11"
	Brand             string    `json:"brand"`             // Xiaomi, Apple, Samsung
	Processor         string    `json:"processor"`         // Snapdragon 680
	PMIC              string    `json:"pmic"`              // PM6150 / PM6150L
	BoardArchitecture string    `json:"boardArchitecture"` // Split Motherboard
	Description       string    `json:"description"`
	CommonFault       string    `json:"commonFault"`
	CreatedAt         time.Time `json:"createdAt"`
}

type DiodeValue struct {
	ID                  string `json:"id" gorm:"primaryKey"`
	ModelID             string `json:"modelId" gorm:"index"`
	ConnectorName       string `json:"connectorName"`       // Display FPC, Battery FPC
	PinNumber           int    `json:"pinNumber"`
	PinName             string `json:"pinName"`             // VDD_MAIN, RESET_N
	LineType            string `json:"lineType"`            // POWER, GROUND, DATA
	NormalReadingMv     int    `json:"normalReadingMv"`     // e.g. 420 (-1 for GND, 9999 for OL)
	TestVoltage         string `json:"testVoltage"`
	FunctionDescription string `json:"functionDescription"`
	FaultDiagnosis      string `json:"faultDiagnosis"`
}

type JumperTrace struct {
	ID           string  `json:"id" gorm:"primaryKey"`
	ModelID      string  `json:"modelId" gorm:"index"`
	Title        string  `json:"title"`        // e.g. "No Backlight Display Jumper"
	Category     string  `json:"category"`     // DISPLAY, CHARGING, AUDIO
	Difficulty   string  `json:"difficulty"`   // Level 2, Chip-Level
	Symptom      string  `json:"symptom"`
	RootCause    string  `json:"rootCause"`
	StartPoint   string  `json:"startPoint"`   // e.g. "Diode D2401 Cathode"
	EndPoint     string  `json:"endPoint"`     // e.g. "FPC Pin 1"
	WireThickness string `json:"wireThickness"`
	IronTemp     string  `json:"ironTemp"`
	UVMask       string  `json:"uvMask"`
	StepByStep   string  `json:"stepByStep"`   // JSON array string
	StartX       float64 `json:"startX"`
	StartY       float64 `json:"startY"`
	EndX         float64 `json:"endX"`
	EndY         float64 `json:"endY"`
	WireColorHex string  `json:"wireColorHex"` // e.g. "#EF4444"
}

// ═══════════════════════════════════════════════════════════════════════════
// MOBILE STORE CALCULATOR & REPAIR POS
// ═══════════════════════════════════════════════════════════════════════════

type RepairPriceGuide struct {
	ID          string  `json:"id" gorm:"primaryKey"`
	Brand       string  `json:"brand"`
	ModelName   string  `json:"modelName"`
	RepairType  string  `json:"repairType"`  // Screen, Battery, Charging Port, Camera Glass
	PartCost    float64 `json:"partCost"`    // Wholesale part cost
	LaborCharge float64 `json:"laborCharge"` // Recommended technician fee
	TotalPrice  float64 `json:"totalPrice"`  // Customer quote
}

type WorkOrder struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	TechnicianID  string    `json:"technicianId" gorm:"index"`
	CustomerName  string    `json:"customerName"`
	CustomerPhone string    `json:"customerPhone"`
	DeviceModel   string    `json:"deviceModel"`
	IssueReported string    `json:"issueReported"`
	EstimatedCost float64   `json:"estimatedCost"`
	AdvancePaid   float64   `json:"advancePaid"`
	Status        string    `json:"status"` // PENDING, IN_PROGRESS, REPAIRED, DELIVERED
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// STUDENT PROGRESS & STREAKS
// ═══════════════════════════════════════════════════════════════════════════

type CourseModule = Module

type UserLessonProgress struct {
	ID               string     `json:"id" gorm:"primaryKey"`
	UserID           string     `json:"userId" gorm:"index"`
	LessonID         string     `json:"lessonId"`
	CourseSlug       string     `json:"courseSlug" gorm:"index"`
	LessonSlug       string     `json:"lessonSlug"`
	WatchedSeconds   int        `json:"watchedSeconds"`
	WatchDurationSec int        `json:"watchDurationSec"`
	IsCompleted      bool       `json:"isCompleted"`
	LastWatchedAt    time.Time  `json:"lastWatchedAt"`
	CompletedAt      *time.Time `json:"completedAt"`
}

func (UserLessonProgress) TableName() string {
	return "user_lesson_progress"
}

type UserStreak struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId" gorm:"primaryKey"`
	CurrentStreak  int       `json:"currentStreak"`
	LongestStreak  int       `json:"longestStreak"`
	LastActiveDate time.Time `json:"lastActiveDate"`
}

// ═══════════════════════════════════════════════════════════════════════════
// CERTIFICATES & MESSAGES & NOTIFICATIONS
// ═══════════════════════════════════════════════════════════════════════════

type Certificate struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	CertificateNumber string    `json:"certificateNumber" gorm:"size:191;uniqueIndex"` // e.g. "SMG-2026-8849"
	UserID            string    `json:"userId" gorm:"index"`
	StudentName       string    `json:"studentName"`
	CourseID          string    `json:"courseId"`
	CourseTitle       string    `json:"courseTitle"`
	CourseSlug        string    `json:"courseSlug"`
	Grade             string    `json:"grade"` // Grade A+ (94%)
	Instructor        string    `json:"instructor"`
	IssuedAt          time.Time `json:"issuedAt"`
	IssueDate         time.Time `json:"issueDate"`
	PdfURL            string    `json:"pdfUrl"`
	VerificationURL   string    `json:"verificationUrl"`
}

type ChatThread struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	UserID          string        `json:"userId" gorm:"index"`
	InstructorID    string        `json:"instructorId"`
	InstructorName  string        `json:"instructorName"`
	ParticipantName string        `json:"name"`
	ParticipantRole string        `json:"role"`
	Subject         string        `json:"subject"`
	AvatarURL       string        `json:"avatarUrl"`
	LastMessage     string        `json:"lastMessage"`
	LastMessageAt   time.Time     `json:"lastMessageAt"`
	UnreadCount     int           `json:"unreadCount"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	Messages        []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ThreadID"`
}

type ChatMessage struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ThreadID      string    `json:"threadId" gorm:"index"`
	SenderID      string    `json:"senderId"`
	ReceiverID    string    `json:"receiverId"`
	SenderRole    string    `json:"senderRole"`
	Message       string    `json:"message"`
	MessageText   string    `json:"text"`
	AttachmentURL string    `json:"attachmentUrl,omitempty"`
	ImageURL      string    `json:"imageUrl,omitempty"`
	IsRead        bool      `json:"isRead"`
	CreatedAt     time.Time `json:"createdAt"`
}

type DeviceToken struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"userId" gorm:"index"`
	Token       string    `json:"token"`
	FCMToken    string    `json:"fcmToken" gorm:"size:191;uniqueIndex"`
	Platform    string    `json:"platform"` // android, ios, web
	DeviceModel string    `json:"deviceModel,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type InAppNotification struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"userId" gorm:"index"` // "ALL" or specific user ID
	Type      string    `json:"type"`                // LIVE, SCHEMATIC, QUIZ, SYSTEM
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	DeepLink  string    `json:"deepLink"`
	ActionURL string    `json:"actionUrl"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// PLANS (MOBILE PLANS & CHAT PLANS)
// ═══════════════════════════════════════════════════════════════════════════

type Plan struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	PlanType         string    `json:"planType" gorm:"index"` // MOBILE, CHAT
	Name             string    `json:"name"`
	DurationLabel    string    `json:"durationLabel"`    // "1 month mobile", "3 Month Plan"
	DurationMonths   int       `json:"durationMonths"`
	Price            float64   `json:"price"`
	OriginalPrice    float64   `json:"originalPrice"`
	DiscountPercent  int       `json:"discountPercent"`
	Currency         string    `json:"currency"`          // INR, NPR
	CurrencySymbol   string    `json:"currencySymbol"`    // ₹, Rs.
	Badge            string    `json:"badge"`             // BEST VALUE, POPULAR, MOST POPULAR
	Features         string    `json:"features" gorm:"type:text"` // JSON array of feature strings
	ResponseTimeMins int       `json:"responseTimeMins"`  // 0 for mobile plans
	IsActive         bool      `json:"isActive"`
	OrderNum         int       `json:"order"`
	CreatedAt        time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// SUPPORT CATEGORIES (SOFTWARE / HARDWARE EXPERT CHAT)
// ═══════════════════════════════════════════════════════════════════════════

type SupportCategory struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ParentID    string    `json:"parentId" gorm:"index"` // empty for top-level (SOFTWARE/HARDWARE)
	Name        string    `json:"name"`
	Slug        string    `json:"slug" gorm:"index"`
	IconAsset   string    `json:"iconAsset"`  // image path or icon name
	Description string    `json:"description"`
	OrderNum    int       `json:"order"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`

	Children []SupportCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// ═══════════════════════════════════════════════════════════════════════════
// HOME SCREEN — PDF CONTAINER CATEGORIES & HIGHLIGHTS
// ═══════════════════════════════════════════════════════════════════════════

type PdfCategory struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug" gorm:"index"`
	IconAsset   string    `json:"iconAsset"`
	Description string    `json:"description"`
	Route       string    `json:"route"`        // deep-link route in the app
	RequiresSub bool      `json:"requiresSub"`  // true = plan required
	OrderNum    int       `json:"order"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Highlight struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	ImageAsset  string    `json:"imageAsset"`
	ActionURL   string    `json:"actionUrl"`
	ActionLabel string    `json:"actionLabel"`
	Badge       string    `json:"badge"`    // OFFER, NEW, etc.
	OrderNum    int       `json:"order"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type HomeIntroVideo struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	YouTubeURL   string    `json:"youtubeUrl"`
	YouTubeID    string    `json:"youtubeId"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Badge        string    `json:"badge"`
	Duration     string    `json:"duration"`
	IsActive     bool      `json:"isActive"`
	UpdatedAt    time.Time `json:"updatedAt"`
}


// ═══════════════════════════════════════════════════════════════════════════
// GUIDES / REPAIR MANUALS
// ═══════════════════════════════════════════════════════════════════════════

type Guide struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Category    string    `json:"category"` // Circuit Diagrams, IC Pinouts, Short Tracing, Error Codes, Tool Usage
	Tag         string    `json:"tag"`      // POPULAR, NEW, FEATURED, MUST READ
	ReadTime    string    `json:"readTime"` // e.g. "5 min read"
	Downloads   int       `json:"downloads"`
	ImageAsset  string    `json:"imageAsset"`
	PdfURL      string    `json:"pdfUrl"`
	Description string    `json:"description" gorm:"type:text"`
	IsPublished bool      `json:"isPublished"`
	OrderNum    int       `json:"order"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// TESTIMONIALS (GUEST HOME SCREEN)
// ═══════════════════════════════════════════════════════════════════════════

type Testimonial struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Avatar     string    `json:"avatar"`
	Rating     int       `json:"rating"`
	ReviewText string    `json:"reviewText" gorm:"type:text"`
	Location   string    `json:"location"`
	IsActive   bool      `json:"isActive"`
	OrderNum   int       `json:"order"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// COURSE CATEGORIES (GUEST HOME SCREEN)
// ═══════════════════════════════════════════════════════════════════════════

type CourseCategory struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Label     string    `json:"label"`
	Icon      string    `json:"icon"`
	ColorHex  string    `json:"colorHex"`
	Count     int       `json:"count"`
	IsActive  bool      `json:"isActive"`
	OrderNum  int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// SIMULATOR COMPONENTS (MOBILE SIMULATOR SCREEN)
// ═══════════════════════════════════════════════════════════════════════════

type SimulatorComponent struct {
	ID           string  `json:"id" gorm:"primaryKey"`
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Category     string  `json:"category"` // Power Section, Processor & RAM, Display & Touch, Network & RF, Audio & Sensor
	LayerIndex   int     `json:"layerIndex"` // 0=Board Top, 1=Board Bottom, 2=Phone Assembly
	RelativePosX float64 `json:"relativePosX"`
	RelativePosY float64 `json:"relativePosY"`
	ColorHex     string  `json:"colorHex"`
	Icon         string  `json:"icon"`
	Summary      string  `json:"summary" gorm:"type:text"`
	Voltage      string  `json:"voltage"`
	DiodeReading string  `json:"diodeReading"`
	Symptoms     string  `json:"symptoms" gorm:"type:text"` // JSON array
	Steps        string  `json:"steps" gorm:"type:text"`    // JSON array
	EquivalentICs string `json:"equivalentICs" gorm:"type:text"` // JSON array
	IsActive     bool    `json:"isActive"`
	OrderNum     int     `json:"order"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// USER CONTINUE WATCHING & DOWNLOADS
// ═══════════════════════════════════════════════════════════════════════════

type UserDownload struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"userId" gorm:"index"`
	LessonID   string    `json:"lessonId" gorm:"index"`
	CourseSlug string    `json:"courseSlug"`
	Title      string    `json:"title"`
	Duration   string    `json:"duration"`
	SizeBytes  int64     `json:"sizeBytes"`
	SizeFmt    string    `json:"size"`
	ImageAsset string    `json:"imageAsset"`
	DownloadedAt time.Time `json:"downloadedAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// SMTP SETTINGS & EMAIL OTP VERIFICATION
// ═══════════════════════════════════════════════════════════════════════════

type SMTPSettings struct {
	ID           string    `json:"id" gorm:"primaryKey"` // "default"
	Enabled      bool      `json:"enabled"`
	SenderEmail  string    `json:"senderEmail"`  // support@yourdomain.com
	SenderName   string    `json:"senderName"`   // Salen Mobile Guru
	Host         string    `json:"host"`         // smtp.host.com (e.g. smtp.gmail.com)
	Port         int       `json:"port"`         // 587 or 465
	Username     string    `json:"username"`     // username or email
	Password     string    `json:"password"`     // app password
	SecurityMode string    `json:"securityMode"` // STARTTLS, SSL, NONE
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// COHORTS (Live Batch / Online Class Groups)
// ═══════════════════════════════════════════════════════════════════════════

type Cohort struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	CourseID    string    `json:"courseId" gorm:"index"`
	CourseSlug  string    `json:"courseSlug" gorm:"index"`
	TotalSeats  int       `json:"totalSeats"`
	Enrolled    int       `json:"enrolled"`
	StartDate   string    `json:"startDate"`
	EndDate     string    `json:"endDate"`
	Status      string    `json:"status"` // active, completed, upcoming
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type EmailOTP struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"size:191;index"`
	Code      string    `json:"code" gorm:"size:10"`
	Purpose   string    `json:"purpose" gorm:"size:50"` // "signup", "login", "reset_password"
	ExpiresAt time.Time `json:"expiresAt"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"createdAt"`
}

func (o *EmailOTP) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// SUPPORT THREADS & MESSAGES (Expert Chat System)
// ═══════════════════════════════════════════════════════════════════════════

// SupportThread is a full conversation ticket between a student and an instructor.
// It replaces/extends the basic ChatThread for support use cases.
type SupportThread struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	UserID             string    `json:"userId" gorm:"index"`
	UserName           string    `json:"userName"`
	UserAvatar         string    `json:"userAvatar"`
	CategoryID         string    `json:"categoryId" gorm:"index"`
	CategorySlug       string    `json:"categorySlug" gorm:"index"`
	CategoryName       string    `json:"categoryName"`
	InstructorID       string    `json:"instructorId"`
	InstructorName     string    `json:"instructorName"`
	InstructorAvatar   string    `json:"instructorAvatar"`
	Subject            string    `json:"subject"`
	DeviceModel        string    `json:"deviceModel"`
	Status             string    `json:"status"` // open, active, pending, resolved, closed
	Priority           string    `json:"priority"` // low, normal, high, urgent
	LastMessage        string    `json:"lastMessage"`
	LastMessageAt      time.Time `json:"lastMessageAt"`
	UnreadStudentCount int       `json:"unreadStudentCount"`
	UnreadAdminCount   int       `json:"unreadAdminCount"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	Messages []SupportMessage `json:"messages,omitempty" gorm:"foreignKey:ThreadID"`
}

// SupportMessage is a single message within a SupportThread.
type SupportMessage struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ThreadID       string    `json:"threadId" gorm:"index"`
	SenderID       string    `json:"senderId"`
	SenderName     string    `json:"senderName"`
	SenderRole     string    `json:"senderRole"` // student, instructor, admin
	SenderAvatar   string    `json:"senderAvatar"`
	Message        string    `json:"message" gorm:"type:text"`
	AttachmentURL  string    `json:"attachmentUrl"`
	AttachmentType string    `json:"attachmentType"` // image, pdf, diagram
	IsRead         bool      `json:"isRead"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// CLASSROOM — VIDEO NOTES
// ═══════════════════════════════════════════════════════════════════════════

// UserVideoNote is a personal timestamped note saved by a student during a lesson.
type UserVideoNote struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	UserID           string    `json:"userId" gorm:"index"`
	LessonID         string    `json:"lessonId" gorm:"index"`
	CourseSlug       string    `json:"courseSlug" gorm:"index"`
	TimestampSeconds float64   `json:"timestampSeconds"`
	Text             string    `json:"text" gorm:"type:text"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// CLASSROOM — LESSON SCHEMATICS
// ═══════════════════════════════════════════════════════════════════════════

// LessonSchematic is a PDF or diagram resource attached to a specific lesson.
type LessonSchematic struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	LessonID    string    `json:"lessonId" gorm:"index"`
	CourseSlug  string    `json:"courseSlug" gorm:"index"`
	Title       string    `json:"title"`
	Description string    `json:"description" gorm:"type:text"`
	FileURL     string    `json:"fileUrl"`
	FileType    string    `json:"fileType"` // pdf, image, svg
	FileSize    string    `json:"fileSize"`
	TotalPages  int       `json:"totalPages"`
	Category    string    `json:"category"` // Motherboard Schematic, ISP Pinout, Jumper Solution, etc.
	CreatedAt   time.Time `json:"createdAt"`
}

// ═══════════════════════════════════════════════════════════════════════════
// CLASSROOM — LESSON Q&A
// ═══════════════════════════════════════════════════════════════════════════

// LessonQAQuestion is a student doubt posted on a lesson (optionally timestamped).
type LessonQAQuestion struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	LessonID         string    `json:"lessonId" gorm:"index"`
	CourseSlug       string    `json:"courseSlug" gorm:"index"`
	UserID           string    `json:"userId" gorm:"index"`
	AuthorName       string    `json:"authorName"`
	AuthorRole       string    `json:"authorRole"` // VIP Technician, Hardware Student, etc.
	AuthorAvatar     string    `json:"authorAvatar"`
	TimestampSeconds float64   `json:"timestampSeconds"` // video seek position when question was asked
	Question         string    `json:"question" gorm:"type:text"`
	IsResolved       bool      `json:"isResolved"`
	Upvotes          int       `json:"upvotes"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	Answers []LessonQAAnswer `json:"answers,omitempty" gorm:"foreignKey:QuestionID"`
}

// LessonQAAnswer is a reply to a LessonQAQuestion.
type LessonQAAnswer struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	QuestionID   string    `json:"questionId" gorm:"index"`
	UserID       string    `json:"userId"`
	AuthorName   string    `json:"authorName"`
	AuthorRole   string    `json:"authorRole"`
	AuthorAvatar string    `json:"authorAvatar"`
	IsInstructor bool      `json:"isInstructor"`
	Answer       string    `json:"answer" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt"`
}
