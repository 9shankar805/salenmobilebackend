# Salen Mobile Guru — Complete API Reference

**Backend:** Go + Gin · **Base URL:** `http://209.50.241.46:8085` · **Auth:** `Authorization: Bearer <jwt>`

---

## Legend

| Symbol | Meaning |
|--------|---------|
| 🔓 | Public — no auth required |
| 🔐 | JWT required (`Authorization: Bearer <token>`) |
| 👑 | Admin JWT required (role = ADMIN) |
| 📱 | Requires MOBILE subscription |
| 💬 | Requires CHAT subscription |
| ✅ | Requires ANY active subscription |

---

## Request / Response Format

All endpoints return JSON:
```json
{ "success": true, "data": { ... } }
{ "success": false, "error": "message" }
```

## Auth Header
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Base URLs
```
Production:  http://209.50.241.46:8085
Local:       http://localhost:8085
```

---

## 1. Health & System

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | 🔓 | Server health check — returns status, version, db, storage |
| HEAD | `/health` | 🔓 | Health ping (no body) |
| GET | `/` | 🔓 | Redirects to `/admin` |

---

## 2. Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/auth/sign-up` | 🔓 | Register new account |
| POST | `/api/auth/register` | 🔓 | Register (alias) |
| POST | `/api/auth/sign-in` | 🔓 | Login with email + password |
| POST | `/api/auth/sign-in/email` | 🔓 | Login via email |
| POST | `/api/auth/sign-in/gmail` | 🔓 | Login via Gmail |
| POST | `/api/auth/sign-in/otp` | 🔓 | Login via OTP code |
| POST | `/api/auth/login` | 🔓 | Login (alias) |
| POST | `/api/auth/login/email` | 🔓 | Login via email (alias) |
| POST | `/api/auth/login/gmail` | 🔓 | Login via Gmail (alias) |
| POST | `/api/auth/login/otp` | 🔓 | Login via OTP (alias) |
| POST | `/api/auth/send-otp` | 🔓 | Send 6-digit OTP to Gmail |
| POST | `/api/auth/request-otp` | 🔓 | Send OTP (alias) |
| POST | `/api/auth/otp` | 🔓 | Send OTP (alias) |
| POST | `/api/auth/verify-otp` | 🔓 | Verify OTP code |
| POST | `/api/auth/reset-password` | 🔓 | Reset password using OTP |
| POST | `/api/auth/forgot-password/reset` | 🔓 | Reset password (alias) |
| POST | `/api/auth/sign-out` | 🔓 | Sign out |
| GET | `/api/auth/session` | 🔓 | Get current session |
| GET | `/api/auth/get-session` | 🔓 | Get session (alias) |
| GET | `/session` | 🔓 | Get session (alias) |
| GET | `/account` | 🔐 | Get current user account details |
| POST | `/api/auth/google` | 🔓 | ❌ Disabled — Google OAuth removed |
| POST | `/api/auth/sign-in/google` | 🔓 | ❌ Disabled |
| POST | `/api/auth/login/google` | 🔓 | ❌ Disabled |
| POST | `/api/auth/phone` | 🔓 | ❌ Disabled — Phone auth removed |
| POST | `/api/auth/sign-in/phone` | 🔓 | ❌ Disabled |
| POST | `/api/auth/send-phone-otp` | 🔓 | ❌ Disabled |

**Sign Up Request Body:**
```json
{
  "name": "Rohit Sharma",
  "email": "rohit@gmail.com",
  "password": "YourPassword123!"
}
```

**Login Request Body:**
```json
{
  "email": "rohit@gmail.com",
  "password": "YourPassword123!"
}
```

**Login Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { "id": "...", "name": "Rohit Sharma", "email": "rohit@gmail.com", "role": "student" }
}
```

---

## 3. User Profile

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/user/profile` | 🔐 | Get user profile |
| PUT | `/user/profile` | 🔐 | Update user profile |
| PATCH | `/user/profile` | 🔐 | Update user profile (partial) |

---

## 4. Courses — Public Catalog

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/courses` | 🔓 | List all published courses |
| GET | `/api/courses` | 🔓 | List published courses (alias) |
| GET | `/organization/courses/public` | 🔓 | Public course catalog |
| GET | `/course/:id` | 🔓 | Get course by ID |
| GET | `/course/slug/:slug` | 🔓 | Get course by slug |
| GET | `/course/:id/tags` | 🔓 | Get course tags |
| GET | `/course/:id/exercise` | 🔓 | List exercises for course |
| GET | `/course/:id/exercise/:exerciseId` | 🔓 | Get single exercise detail |
| GET | `/org-site/course/:courseSlug/exercises` | 🔓 | Exercises via org-site path |

**Query params for `/courses`:**
```
?search=android     — full-text search
?tags=Hardware      — filter by tag/category
?pricing=free       — free or paid
?page=1&limit=20    — pagination
```

---

## 5. Courses — CRUD

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/course` | 🔓 | Create new course |
| PUT | `/course/:id` | 🔓 | Update course |
| DELETE | `/course/:id` | 🔓 | Delete course |
| POST | `/course/:id/clone` | 🔓 | Clone a course |
| POST | `/course/:id/enroll` | 🔓 | Enroll student in course |
| GET | `/organization/courses` | 🔓 | List org courses |
| POST | `/organization/courses` | 🔓 | Create org course |

---

## 6. Course Modules & Lessons

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/course/:id/section` | 🔓 | Create module/section |
| PUT | `/course/:id/section/:sectionId` | 🔓 | Update section |
| DELETE | `/course/:id/section/:sectionId` | 🔓 | Delete section |
| POST | `/course/:id/section/reorder` | 🔓 | Reorder sections |
| POST | `/course/:id/section/promote-ungrouped` | 🔓 | Promote ungrouped lessons |
| GET | `/course/:id/lesson` | 🔓 | List lessons in course |
| POST | `/course/:id/lesson` | 🔓 | Create lesson |
| GET | `/course/:id/lesson/:lessonId` | 🔓 | Get lesson detail |
| PUT | `/course/:id/lesson/:lessonId` | 🔓 | Update lesson |
| DELETE | `/course/:id/lesson/:lessonId` | 🔓 | Delete lesson |
| POST | `/course/:id/lesson/reorder` | 🔓 | Reorder lessons |
| GET | `/course/:id/lesson/:lessonId/comment` | 🔓 | Get lesson comments |
| POST | `/course/:id/lesson/:lessonId/comment` | 🔓 | Post comment on lesson |
| PUT | `/course/:id/lesson/comment/:commentId` | 🔓 | Update comment |
| DELETE | `/course/:id/lesson/comment/:commentId` | 🔓 | Delete comment |
| GET | `/course/:id/curriculum` | 🔓 | Get full curriculum tree |
| GET | `/course/:id/members` | 🔓 | Get enrolled members |
| GET | `/course/:id/analytics` | 🔓 | Course analytics |
| GET | `/course/:id/ai-tutor` | 🔓 | Get AI tutor config |
| PUT | `/course/:id/ai-tutor` | 🔓 | Update AI tutor config |
| PUT | `/course/:id/tags` | 🔓 | Update course tags |

---

## 7. Course — Mobile App Player

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/org-site/course/:courseSlug` | 🔓 | Full course tree for Flutter player |
| GET | `/org-site/course/:courseSlug/item/:itemSlug` | 🔓 | Get specific lesson for player |
| GET | `/org-site/course/:courseSlug/item/:itemSlug/stream` | 🔓 | Get video stream URL + type |
| POST | `/org-site/course/:courseSlug/item/:itemSlug/hls-cookie` | 🔓 | Mint HLS auth token |

---

## 8. Course — API v1 (Flutter Recommended)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/courses/:courseSlug/curriculum` | 🔓 | Full module + lesson tree with quizIds |
| GET | `/api/v1/courses/:courseSlug/resources` | 🔓 | Downloadable schematics & PDFs |
| GET | `/api/v1/users/:userId/courses/:courseSlug/progress` | 🔓 | Student progress + quiz grades |

---

## 9. Watch Progress & Streaks

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/course/:id/lesson/:lessonId/watch-progress` | 🔓 | Get watch progress for lesson |
| PUT | `/course/:id/lesson/:lessonId/watch-progress` | 🔓 | Update watch position |
| PUT | `/course/:id/lesson/:lessonId/completion` | 🔓 | Mark lesson as complete |
| GET | `/user/learning-progress` | 🔓 | Overall learning progress |
| GET | `/user/learning-streak` | 🔓 | Daily streak data |
| GET | `/user/streak` | 🔓 | Streak (alias) |
| GET | `/user/continue-watching` | 🔓 | Continue watching list |
| GET | `/user/downloads` | 🔓 | Downloaded lessons list |
| POST | `/user/downloads` | 🔓 | Save downloaded lesson |
| DELETE | `/user/downloads/:id` | 🔓 | Remove download record |
| POST | `/user/lesson-progress` | ✅ | Update lesson progress (subscription required) |
| POST | `/api/v1/courses/:courseSlug/lessons/:lessonSlug/progress` | ✅ | Update progress v1 |

---

## 10. Quizzes

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/quizzes/:courseSlug` | 🔓 | Get quiz for a course |
| POST | `/quizzes/submit` | ✅ | Submit quiz answers |
| GET | `/quizzes/results` | ✅ | Get user's quiz results |
| GET | `/quizzes/leaderboard` | 🔓 | Quiz leaderboard |
| GET | `/quizzes/badges` | 🔓 | Badges earned from quizzes |

**Submit Quiz Body:**
```json
{
  "quizId": "qz1",
  "userId": "user_123",
  "courseSlug": "android-hardware-repairing-level-1",
  "answers": [
    { "questionId": "qq_c1_1", "selectedOption": 2 }
  ]
}
```

---

## 11. Subscriptions & Payments

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/subscriptions/status` | 🔓 | Get user subscription status |
| GET | `/user/subscription-status` | 🔓 | Subscription status (alias) |
| POST | `/payments/verify` | 🔓 | Submit payment + activate |
| POST | `/subscriptions/enroll` | 🔓 | Enroll (alias) |

**Verify Payment Body:**
```json
{
  "userId": "user_123",
  "planId": "mp_1month",
  "paymentMethod": "eSewa",
  "transactionId": "TXN-ABC123",
  "amount": 999,
  "currency": "INR",
  "status": "active"
}
```

**Payment Methods:** `eSewa` · `Khalti` · `IME Pay` · `Stripe` · `MANUAL`

**Plan Types:** `VIP` · `ANNUAL` · `LIFETIME` · `MOBILE` · `CHAT` · `SINGLE_COURSE`

---

## 12. Plans

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/plans` | 🔓 | Get all plans (mobile + chat) |
| GET | `/plans/mobile` | 🔓 | Mobile subscription plans |
| GET | `/plans/chat` | 🔓 | Expert chat plans |

---

## 13. Live Sessions

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/live/sessions` | 🔓 | List all sessions (no join URL exposed) |
| GET | `/live/schedule` | 🔓 | Schedule split: upcoming / live / completed |
| GET | `/live/sessions/:id/join` | 🔐 | Get join URL + meeting ID + passcode |

**Join Response:**
```json
{
  "success": true,
  "data": {
    "id": "ls_001",
    "platform": "GOOGLE_MEET",
    "platformName": "Google Meet",
    "joinUrl": "https://meet.google.com/abc-defg-hij",
    "meetingId": "abc-defg-hij",
    "passcode": ""
  }
}
```

**Platforms:** `ZOOM` · `GOOGLE_MEET` · `TEAMS` · `YOUTUBE_LIVE`

---

## 14. Certificates

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/user/certificates` | 🔓 | Get certificates for a user |
| GET | `/certificates/verify/:certNum` | 🔓 | Verify certificate (public) |

**Query:** `/user/certificates?userId=user_123`

---

## 15. Community / Q&A

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/community` | 🔓 | List community posts |
| POST | `/community` | 🔓 | Create post |
| POST | `/community/:id/upvote` | 🔓 | Upvote a post |
| POST | `/community/:id/reply` | 🔓 | Reply to post |
| GET | `/org-site/community/:orgId` | 🔓 | Org community feed |
| GET | `/api/community/posts/:slug/comments` | 🔓 | Get post comments |

---

## 16. Notifications (Flutter App)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/notifications/token` | 🔓 | Register FCM device token |
| GET | `/notifications` | 🔓 | Get notifications for user |
| POST | `/notifications/:id/read` | 🔓 | Mark notification as read |
| PATCH | `/notifications/:id/read` | 🔓 | Mark as read (Flutter prefers PATCH) |
| POST | `/notifications/mark-all-read` | 🔓 | Mark all as read |

**Register Token Body:**
```json
{
  "userId": "user_123",
  "token": "fcm_token_from_firebase_sdk",
  "platform": "android"
}
```

**Platforms:** `android` · `ios` · `web`

---

## 17. Messaging / Chat

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/messages/threads` | 🔓 | List chat threads |
| GET | `/messages/threads/:threadId/messages` | 🔓 | Get messages in a thread |
| POST | `/messages/send` | 💬 | Send message (CHAT subscription required) |

---

## 18. Hardware Tools (📱 Mobile Subscription)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/hardware/models` | 📱 | List hardware models (Xiaomi, Apple, Samsung) |
| POST | `/hardware/components` | 📱 | Get hardware components |
| GET | `/hardware/diode-values` | 📱 | List diode mode readings |
| GET | `/hardware/jumpers` | 📱 | List jumper traces |
| GET | `/api/hardware/components` | 🔓 | Hardware components (public, no sub) |

---

## 19. Calculator / Repair POS (📱 Mobile Subscription)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/calculator/pricing-matrix` | 📱 | Get repair pricing matrix |
| GET | `/calculator/work-orders` | 📱 | List work orders |
| POST | `/calculator/work-orders` | 📱 | Create work order |
| PATCH | `/calculator/work-orders/:id/status` | 📱 | Update work order status |
| GET | `/pricing-matrix` | 🔓 | Public pricing matrix |

**Work Order Statuses:** `PENDING` · `IN_PROGRESS` · `REPAIRED` · `DELIVERED`

---

## 20. Content & Search

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/search` | 🔓 | Global search across courses, guides, hardware |
| GET | `/api/guides` | 🔓 | List repair guides / manuals |
| GET | `/guides` | 🔓 | Guides (alias) |
| GET | `/api/testimonials` | 🔓 | List student testimonials |
| GET | `/api/tools` | 🔓 | List app tools |
| GET | `/api/categories` | 🔓 | List course categories |
| GET | `/tags` | 🔓 | Get all tags |
| GET | `/tags/public` | 🔓 | Public tags |

**Search Query:** `/api/search?q=reballing&type=course`

---

## 21. Home Screen

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/home` | 🔓 | Full home screen data block |
| GET | `/home/highlights` | 🔓 | Highlights / offer banners |
| GET | `/home/pdf-categories` | 🔓 | PDF category grid |
| GET | `/home/toolkit` | 🔓 | Toolkit items list |
| GET | `/home/intro-video` | 🔓 | App intro YouTube video |
| GET | `/api/app/intro-video` | 🔓 | Intro video (alias) |

---

## 22. Schematics & File Storage

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/schematics` | 🔓 | List all schematics |
| GET | `/storage/*key` | 🔓 | Stream file from Cloudflare R2 or local |
| HEAD | `/storage/*key` | 🔓 | File head check |
| GET | `/r2/*key` | 🔓 | Direct R2 stream |
| HEAD | `/r2/*key` | 🔓 | R2 head check |

**Example:** `GET /storage/schematics/pdf/backlight.pdf`

---

## 23. Support Categories

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/support/categories` | 🔓 | Top-level categories (SOFTWARE / HARDWARE) |
| GET | `/support/categories/:slug` | 🔓 | Children of a category |
| GET | `/support/categories/:slug/full` | 🔓 | Category + all nested children |

---

## 24. Widgets

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/widgets` | 🔓 | List all app widgets |
| GET | `/widgets/:id` | 🔓 | Get single widget |

---

## 25. App Screens / CMS

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/app-screens` | 🔓 | List all screen configs |
| GET | `/app-screens/:siteName/:screenKey` | 🔓 | Get specific screen config |
| GET | `/cms/banners` | 🔓 | Get CMS banners |

**Screen Keys:** `guest_home` · `member_home` · `live_sessions` · `tools`

---

## 26. Media Uploads

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/media` | 🔓 | List media assets |
| POST | `/media` | 🔓 | Create media record |
| DELETE | `/media/:id` | 🔓 | Delete media record |
| POST | `/media/upload` | 🔓 | Upload file to Cloudflare R2 |
| POST | `/api/upload` | 🔓 | Upload (alias) |

**Upload:** `multipart/form-data` with field `file` + optional `folder`

---

## 27. Cohorts

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/cohort` | 🔓 | List cohorts (`?courseSlug=` filter) |
| GET | `/cohort/:id` | 🔓 | Get cohort by ID |

---

## 28. SSO

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/sso/discover` | 🔓 | SSO discovery — returns `ssoAvailable: false` |
| GET | `/sso/org/:orgId` | 🔓 | Org SSO config — returns `enabled: false` |

---

## 29. Onboarding / Invite Links

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/invite/:token/preview` | 🔓 | Preview invite details |
| POST | `/invite/:token/accept` | 🔓 | Accept invite |
| GET | `/r/:token/preview` | 🔓 | Short invite preview |
| POST | `/r/:token/accept` | 🔓 | Short invite accept |
| GET | `/link/:token/preview` | 🔓 | Link invite preview |
| POST | `/link/:token/accept` | 🔓 | Link invite accept |
| GET | `/student/:token` | 🔓 | Student token lookup |

---

## 30. Organization

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/organization/dash/stats` | 🔓 | Org dashboard stats |
| GET | `/organization/pending` | 🔓 | Pending org members |
| POST | `/organization/:token/accept` | 🔓 | Accept org invite |
| GET | `/organization/:token/preview` | 🔓 | Preview org invite |
| POST | `/create-org` | 🔓 | Create organization |

---

## 31. Dashboard

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/dash/stats` | 🔓 | Platform-wide stats |
| GET | `/dash/login-activity` | 🔓 | Login activity log |

---

---

# ADMIN PANEL ENDPOINTS

> **All admin endpoints require:**
> ```
> Authorization: Bearer <admin-jwt>
> ```
> Get the token from `POST /api/admin/login`

---

## A1. Admin Login

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/admin/login` | 🔓 | Admin login — returns JWT |

**Body:**
```json
{
  "email": "admin@salenmobileguru.com",
  "password": "Admin@SMG2026!"
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "admin": { "id": "...", "name": "Salen Master Admin", "email": "admin@salenmobileguru.com" }
}
```

---

## A2. Admin Web Portal (HTML Pages)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin` | Main admin dashboard |
| GET | `/admin/*any` | Admin SPA catch-all routing |
| GET | `/classroom` | Classroom video player |
| GET | `/classroom/*any` | Classroom catch-all |
| GET | `/learn/*any` | Learn alias |
| GET | `/courses.html` | Courses admin page |
| GET | `/curriculum.html` | Curriculum builder |
| GET | `/quizzes.html` | Quizzes admin |
| GET | `/subscriptions.html` | Subscriptions admin |
| GET | `/users.html` | Users admin |
| GET | `/live.html` | Live sessions admin |
| GET | `/templates.html` | Templates admin |
| GET | `/smtp.html` | SMTP settings page |
| GET | `/course-workspace.html` | Course workspace builder |
| GET | `/internal/routes/admin/courses/*any` | Course submodule pages |

---

## A3. Admin Stats & Analytics

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/stats` | 👑 | Platform-wide stats (users, revenue, courses) |
| GET | `/api/admin/analytics` | 👑 | Detailed analytics data |
| GET | `/api/admin/system/status` | 👑 | System health — DB, R2, FCM, SMTP |

---

## A4. Admin — Courses

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/courses` | 👑 | List all courses (published + drafts) |
| POST | `/api/admin/courses` | 👑 | Create course |
| PUT | `/api/admin/courses/:id` | 👑 | Update course |
| DELETE | `/api/admin/courses/:id` | 👑 | Delete course |
| GET | `/api/admin/courses/:id/curriculum` | 👑 | Get full curriculum tree |
| POST | `/api/admin/courses/:id/modules` | 👑 | Add module to course |
| PUT | `/api/admin/modules/:id` | 👑 | Update module |
| DELETE | `/api/admin/modules/:id` | 👑 | Delete module |
| POST | `/api/admin/modules/:id/lessons` | 👑 | Add lesson to module |
| PUT | `/api/admin/lessons/:id` | 👑 | Update lesson |
| DELETE | `/api/admin/lessons/:id` | 👑 | Delete lesson |

---

## A5. Admin — Subscriptions

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/subscriptions` | 👑 | List all subscriptions |
| POST | `/api/admin/subscriptions/:id/approve` | 👑 | Approve pending subscription |
| POST | `/api/admin/subscriptions/:id/reject` | 👑 | Reject subscription |
| POST | `/api/admin/subscriptions/manual-enroll` | 👑 | Manually enroll a user |

---

## A6. Admin — Users

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/users` | 👑 | List all users |
| POST | `/api/admin/users` | 👑 | Create user |
| PUT | `/api/admin/users/:id` | 👑 | Update user |
| DELETE | `/api/admin/users/:id` | 👑 | Delete user |
| POST | `/api/admin/users/:id/toggle-vip` | 👑 | Toggle VIP status on/off |
| POST | `/api/admin/users/:id/activate-courses` | 👑 | Grant course access to user |

---

## A7. Admin — Live Sessions

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/live-sessions` | 👑 | List all live sessions |
| POST | `/api/admin/live-sessions` | 👑 | Create live session |
| PUT | `/api/admin/live-sessions/:id` | 👑 | Update session details |
| PATCH | `/api/admin/live-sessions/:id/status` | 👑 | Change status |
| DELETE | `/api/admin/live-sessions/:id` | 👑 | Delete session |
| POST | `/api/admin/live/sessions` | 👑 | Create session (live group) |
| PUT | `/api/admin/live/sessions/:id` | 👑 | Update (live group) |
| PATCH | `/api/admin/live/sessions/:id/status` | 👑 | Patch status (live group) |
| DELETE | `/api/admin/live/sessions/:id` | 👑 | Delete (live group) |

**Create Session Body:**
```json
{
  "title": "CPU Reballing Masterclass",
  "subtitle": "Complete BGA reballing process",
  "instructor": "Mr. Rohit Sharma",
  "date": "8 Nov (Sat), 07:00 PM",
  "time": "07:00 PM",
  "platform": "GOOGLE_MEET",
  "joinUrl": "https://meet.google.com/abc-defg-hij",
  "meetingId": "abc-defg-hij",
  "passcode": "",
  "durationMins": 90,
  "imageAsset": "assets/banner_bg_img/banner_1.webp"
}
```

**Platforms:** `ZOOM` · `GOOGLE_MEET` · `TEAMS` · `YOUTUBE_LIVE`
**Statuses:** `UPCOMING` · `LIVE` · `COMPLETED`

---

## A8. Admin — Community Moderation

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/community` | 👑 | List all community posts |
| DELETE | `/api/admin/community/:id` | 👑 | Delete post |
| POST | `/api/admin/community/:id/solve` | 👑 | Mark post as solved |
| POST | `/api/admin/community/:id/reply` | 👑 | Add admin reply to post |

---

## A9. Admin — Notifications & FCM Push

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/notifications` | 👑 | List all notifications |
| POST | `/api/admin/notifications` | 👑 | Broadcast push to all users |
| POST | `/api/admin/notifications/broadcast` | 👑 | Broadcast (alias) |
| POST | `/api/admin/notifications/send-user` | 👑 | Push to specific user's devices |
| POST | `/api/admin/notifications/send-topic` | 👑 | Push to FCM topic |
| DELETE | `/api/admin/notifications/:id` | 👑 | Delete notification |
| GET | `/api/admin/notifications/fcm-status` | 👑 | FCM health — ready/unavailable |
| POST | `/api/admin/notifications/fcm-config` | 👑 | Save Firebase service account JSON |
| POST | `/api/admin/notifications/fcm-validate` | 👑 | Validate Firebase credentials JSON |

**Broadcast Body:**
```json
{
  "title": "🔴 Live Class Starting Now!",
  "body": "CPU Reballing masterclass starts in 5 minutes.",
  "type": "live_reminder",
  "deepLink": "/live/sessions",
  "imageUrl": "https://...",
  "useTopic": true
}
```

**Notification Types:** `announcement` · `course_update` · `live_reminder` · `payment` · `personal`

**Send to User Body:**
```json
{
  "userId": "user_123",
  "title": "✅ Your Payment is Approved",
  "body": "You now have full access to all courses.",
  "type": "payment",
  "deepLink": "/member/home"
}
```

---

## A10. Admin — Quizzes

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/quizzes` | 👑 | List all quizzes |
| POST | `/api/admin/quizzes` | 👑 | Create / update quiz |
| DELETE | `/api/admin/quizzes/:id` | 👑 | Delete quiz |
| GET | `/api/admin/quizzes/results` | 👑 | List all student quiz results |
| DELETE | `/api/admin/quizzes/results/:id` | 👑 | Delete a quiz result |

---

## A11. Admin — Media & File Uploads

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/media` | 👑 | List all media assets |
| POST | `/api/admin/media` | 👑 | Create media record |
| POST | `/api/admin/media/upload` | 👑 | Upload file to R2 |
| DELETE | `/api/admin/media/:id` | 👑 | Delete media record |
| POST | `/api/admin/upload` | 👑 | Upload file (root admin alias) |

**Upload Response:**
```json
{
  "success": true,
  "publicUrl": "https://pub-xxx.r2.dev/uploads/file.pdf",
  "r2Key": "uploads/file.pdf",
  "size": 1048576
}
```

---

## A12. Admin — Schematics

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/schematics` | 👑 | List all schematics |
| POST | `/api/admin/schematics` | 👑 | Create schematic record |
| DELETE | `/api/admin/schematics/:id` | 👑 | Delete schematic |

---

## A13. Admin — Repair Guides

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/guides` | 👑 | List all repair guides |
| POST | `/api/admin/guides` | 👑 | Create guide |
| PUT | `/api/admin/guides/:id` | 👑 | Update guide |
| DELETE | `/api/admin/guides/:id` | 👑 | Delete guide |

---

## A14. Admin — Templates

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/templates` | 👑 | List all templates |
| GET | `/api/templates` | 🔓 | Public template list |
| POST | `/api/admin/templates` | 👑 | Create template |
| POST | `/api/admin/templates/upload` | 👑 | Upload template file |
| PATCH | `/api/admin/templates/:id/default` | 👑 | Toggle default template |
| DELETE | `/api/admin/templates/:id` | 👑 | Delete template |

**Template Categories:** `certificate` · `schematic` · `worksheet` · `quiz` · `curriculum` · `notification` · `banner` · `diagnostic`

---

## A15. Admin — SMTP / Email Settings

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/smtp` | 👑 | Get current SMTP settings |
| POST | `/api/admin/smtp` | 👑 | Save SMTP settings |
| POST | `/api/admin/smtp/test` | 👑 | Send a test email |

**SMTP Body:**
```json
{
  "enabled": true,
  "senderEmail": "support@salenmobileguru.com",
  "senderName": "Salen Mobile Guru",
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your@gmail.com",
  "password": "app-password",
  "securityMode": "STARTTLS"
}
```

**Security Modes:** `STARTTLS` · `SSL` · `NONE`

---

## A16. Admin — Hardware (Diode Mode / Jumpers)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/hardware/models` | 👑 | List hardware device models |
| POST | `/api/admin/hardware/models` | 👑 | Create hardware model |
| DELETE | `/api/admin/hardware/models/:id` | 👑 | Delete model |
| GET | `/api/admin/hardware/diode-values` | 👑 | List diode mode readings |
| POST | `/api/admin/hardware/diode-values` | 👑 | Create diode value entry |
| DELETE | `/api/admin/hardware/diode-values/:id` | 👑 | Delete diode value |
| GET | `/api/admin/hardware/jumpers` | 👑 | List jumper traces |
| POST | `/api/admin/hardware/jumpers` | 👑 | Create jumper trace |
| DELETE | `/api/admin/hardware/jumpers/:id` | 👑 | Delete jumper trace |

---

## A17. Admin — Calculator / Work Orders

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/calculator/pricing-matrix` | 👑 | List all pricing entries |
| POST | `/api/admin/calculator/pricing-matrix` | 👑 | Create pricing entry |
| DELETE | `/api/admin/calculator/pricing-matrix/:id` | 👑 | Delete pricing entry |
| GET | `/api/admin/calculator/work-orders` | 👑 | List all work orders |
| POST | `/api/admin/calculator/work-orders` | 👑 | Create work order |
| PUT | `/api/admin/calculator/work-orders/:id/status` | 👑 | Update work order status |

---

## A18. Admin — Cohorts

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/cohorts` | 👑 | List all cohorts |
| POST | `/api/admin/cohorts` | 👑 | Create cohort |
| PUT | `/api/admin/cohorts/:id` | 👑 | Update cohort |
| DELETE | `/api/admin/cohorts/:id` | 👑 | Delete cohort |

---

## A19. Admin — CMS & Home Screen

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/cms/banners` | 👑 | Get home screen banners |
| POST | `/api/admin/cms/banners` | 👑 | Update banners |
| GET | `/api/admin/home/intro-video` | 👑 | Get intro YouTube video |
| POST | `/api/admin/home/intro-video` | 👑 | Save intro video |
| PUT | `/api/admin/home/intro-video` | 👑 | Update intro video |

---

## A20. Admin — Certificates

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/certificates` | 👑 | List all issued certificates |
| POST | `/api/admin/certificates/issue` | 👑 | Issue certificate to student |
| DELETE | `/api/admin/certificates/:id` | 👑 | Revoke certificate |

**Issue Body:**
```json
{
  "userId": "user_123",
  "studentName": "Rohit Sharma",
  "courseTitle": "Android Hardware Repairing Level 1",
  "courseSlug": "android-hardware-repairing-level-1",
  "grade": "A+ (Master Technician)",
  "instructor": "Salen Dangol"
}
```

---

## A21. Admin — Progress Tracking

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/progress` | 👑 | All student lesson progress |
| GET | `/api/admin/progress/courses` | 👑 | Progress grouped by course |

---

## A22. Admin — Messages

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/messages/threads` | 👑 | List all chat threads |
| GET | `/api/admin/messages/threads/:id` | 👑 | Get messages in thread |
| POST | `/api/admin/messages/reply` | 👑 | Admin reply to a thread |
| POST | `/api/admin/messages/broadcast` | 👑 | Broadcast message to all |

---

## A23. Admin — Content Management

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/admin/content/tags` | 👑 | List content tags |
| POST | `/api/admin/content/categories` | 👑 | Create course category |
| POST | `/api/admin/content/testimonials` | 👑 | Create testimonial |
| POST | `/api/admin/content/update-metadata` | 👑 | Update metadata |
| GET | `/api/admin/content/assets` | 👑 | List content assets |

---

## Summary

| Category | Endpoints |
|----------|-----------|
| Public 🔓 | ~130 |
| JWT required 🔐 | ~15 |
| Admin only 👑 | ~90 |
| Subscription gated 📱💬✅ | ~12 |
| **Total** | **~257** |

---

## Firebase FCM — Project Details

| Setting | Value |
|---------|-------|
| Project ID | `salenmobile-8ec54` |
| Credentials File | `salenmobile-8ec54-firebase-adminsdk-fbsvc-4d33600c37.json` |
| FCM Status | ✅ Ready |
| VAPID Public Key | `BHlSAu0R1dXINv1BUw1wScq9HsDPeE8hHxJxZBJ7d9wnruuldSh5s1dpeDv5OdNz9wqtSNcOQFrFSM0CnqU48Bc` |

---

## Infrastructure

| Service | Port | Purpose |
|---------|------|---------|
| Go Backend | `8085` | Main API server |
| MediaMTX RTMP | `1935` | Live stream ingest (OBS/phone) |
| MediaMTX HLS | `8888` | Student HLS playback |
| MediaMTX WebRTC | `8889` | Low-latency WebRTC |
| MediaMTX RTSP | `8554` | RTSP protocol |
| coTURN STUN/TURN | `3478` | WebRTC NAT traversal |
| MySQL | `3306` | Database |
| Redis | `6379` | Cache (installed, not wired) |
| Nginx | `80/443` | Reverse proxy |

---

*Generated: Salen Mobile Guru Backend — smg-backend v1.0.0*
