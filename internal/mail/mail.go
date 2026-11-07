// Package mail handles all outbound email for Salen Mobile Guru.
//
// Email types:
//   - OTP verification    (signup / login / password reset)
//   - Welcome             (after successful registration)
//   - Subscription active (payment approved by admin)
//   - Payment pending     (payment received, under review)
//   - Certificate issued  (course completion)
//   - Live class reminder (before session starts)
//   - Admin SMTP test     (connectivity probe)
//
// Gmail setup in .env:
//   SMTP_ENABLED=true
//   SMTP_HOST=smtp.gmail.com
//   SMTP_PORT=587
//   SMTP_SECURITY=STARTTLS
//   SMTP_USERNAME=you@gmail.com
//   SMTP_PASSWORD=xxxx xxxx xxxx xxxx   ← 16-char Google App Password
//   SMTP_FROM_EMAIL=you@gmail.com
//   SMTP_FROM_NAME=Salen Mobile Guru
package mail

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"smg-backend/internal/models"
)

// ─── Brand ────────────────────────────────────────────────────────────────

const (
	brandName     = "Salen Mobile Guru"
	brandTagline  = "Elite Hardware &amp; Micro-Soldering Academy"
	brandWebsite  = "https://salenmobileguru.com"
	brandSupport  = "support@salenmobileguru.com"
	brandWhatsApp = "+977-9800000000"
	brandLocation = "Kathmandu, Nepal"
)

// ─── SMTP Settings ────────────────────────────────────────────────────────

// GetSMTPSettings reads config from DB with .env fallback.
func GetSMTPSettings(db *gorm.DB) *models.SMTPSettings {
	var s models.SMTPSettings
	if db != nil {
		if err := db.Where("id = ?", "default").First(&s).Error; err == nil && s.Host != "" {
			return &s
		}
	}

	enabled := strings.ToLower(os.Getenv("SMTP_ENABLED")) == "true" || os.Getenv("SMTP_ENABLED") == "1"
	port := 587
	if p, err := strconv.Atoi(os.Getenv("SMTP_PORT")); err == nil && p > 0 {
		port = p
	}
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = brandSupport
	}
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = brandName
	}
	sec := strings.ToUpper(os.Getenv("SMTP_SECURITY"))
	if sec == "" {
		sec = "STARTTLS"
	}

	s = models.SMTPSettings{
		ID: "default", Enabled: enabled,
		SenderEmail: fromEmail, SenderName: fromName,
		Host: host, Port: port,
		Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"),
		SecurityMode: sec, UpdatedAt: time.Now(),
	}
	if db != nil && s.Host != "" {
		_ = db.Save(&s)
	}
	return &s
}

// SaveSMTPSettings persists updated SMTP config.
func SaveSMTPSettings(db *gorm.DB, updated *models.SMTPSettings) error {
	updated.ID = "default"
	updated.UpdatedAt = time.Now()
	return db.Save(updated).Error
}

// ─── OTP generator ────────────────────────────────────────────────────────

// GenerateOTP returns a cryptographically secure 6-digit code.
func GenerateOTP() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1_000_000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// ─── Core send ────────────────────────────────────────────────────────────

// SendMail dispatches an HTML email with up to 2 retries.
// Falls back to simulation log when SMTP is disabled or misconfigured.
func SendMail(db *gorm.DB, to, subject, htmlBody string) error {
	cfg := GetSMTPSettings(db)
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient email is empty")
	}

	if !cfg.Enabled || cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		log.Printf("📧 [SMTP DISABLED — SIMULATED] To: %s | Subject: %s", to, subject)
		return nil
	}

	msg := buildRFC2822(cfg, to, subject, htmlBody)

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		lastErr = smtpDeliver(cfg, to, msg)
		if lastErr == nil {
			log.Printf("✅ Email sent [attempt %d] → %s | %s", attempt, to, subject)
			return nil
		}
		log.Printf("⚠️  Email attempt %d failed [%s]: %v", attempt, to, lastErr)
		if attempt < 2 {
			time.Sleep(2 * time.Second)
		}
	}
	return fmt.Errorf("email delivery failed after 2 attempts: %w", lastErr)
}

// buildRFC2822 assembles an RFC 2822 email with base64-encoded HTML body.
func buildRFC2822(cfg *models.SMTPSettings, to, subject, htmlBody string) string {
	encodedSubject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	encodedBody := base64.StdEncoding.EncodeToString([]byte(htmlBody))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s <%s>\r\n", cfg.SenderName, cfg.SenderEmail))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	b.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	b.WriteString("X-Mailer: SMG-Backend/1.0\r\n")
	b.WriteString("\r\n")

	// Split base64 into 76-char lines (RFC 2045)
	for i := 0; i < len(encodedBody); i += 76 {
		end := i + 76
		if end > len(encodedBody) {
			end = len(encodedBody)
		}
		b.WriteString(encodedBody[i:end] + "\r\n")
	}
	return b.String()
}

// smtpDeliver handles the actual SMTP connection and message delivery.
func smtpDeliver(cfg *models.SMTPSettings, to, msg string) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	secMode := strings.ToUpper(cfg.SecurityMode)

	var client *smtp.Client
	var err error

	// SSL/TLS direct connection (port 465)
	if secMode == "SSL" || secMode == "TLS" || cfg.Port == 465 {
		conn, tlsErr := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host})
		if tlsErr != nil {
			return fmt.Errorf("TLS dial %s: %w", addr, tlsErr)
		}
		client, err = smtp.NewClient(conn, cfg.Host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("SMTP client init: %w", err)
		}
	} else {
		// STARTTLS (port 587 — Gmail standard)
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("SMTP dial %s: %w", addr, err)
		}
		if secMode != "NONE" {
			if err = client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
				client.Close()
				return fmt.Errorf("STARTTLS on %s: %w", addr, err)
			}
		}
	}
	defer client.Close()

	// Authenticate
	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth for %s: %w", cfg.Username, err)
		}
	}

	// Envelope
	if err = client.Mail(cfg.SenderEmail); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO %s: %w", to, err)
	}

	// Body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command: %w", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close DATA: %w", err)
	}
	_ = client.Quit()
	return nil
}

// ─── HTML layout ──────────────────────────────────────────────────────────

// layout wraps email content in the standard SMG dark-theme shell.
func layout(preheader, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1.0"/>
<meta name="color-scheme" content="dark"/>
<style>
body,table,td,a{-webkit-text-size-adjust:100%%;-ms-text-size-adjust:100%%}
img{border:0;outline:none;text-decoration:none}
body{margin:0;padding:0;background:#05070f}
a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important}
@media only screen and (max-width:600px){
  .email-container{width:100%%!important}
  .mobile-pad{padding:24px 16px!important}
  .otp-code{font-size:34px!important;letter-spacing:6px!important}
  .hide-mobile{display:none!important}
}
</style>
</head>
<body style="margin:0;padding:0;background:#05070f;">
<!-- preheader -->
<div style="display:none;max-height:0;overflow:hidden;font-size:1px;color:#05070f;">%s&zwnj;&nbsp;&zwnj;&nbsp;&zwnj;&nbsp;</div>

<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#05070f;">
<tr><td align="center" style="padding:32px 16px 0;">
<table class="email-container" role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="margin:0 auto;">

  <!-- HEADER -->
  <tr>
    <td style="background:linear-gradient(135deg,#1d4ed8,#7c3aed);border-radius:16px 16px 0 0;padding:28px 32px;text-align:center;">
      <div style="display:inline-block;background:rgba(255,255,255,0.15);border-radius:10px;padding:6px 14px;margin-bottom:10px;">
        <span style="font-size:13px;font-weight:700;color:#fff;letter-spacing:2px;font-family:Arial,sans-serif;text-transform:uppercase;">%s</span>
      </div>
      <div style="font-size:12px;color:rgba(255,255,255,0.7);font-family:Arial,sans-serif;">%s</div>
    </td>
  </tr>

  <!-- BODY -->
  <tr>
    <td style="background:#0d1117;border:1px solid #1e2d45;border-top:0;border-radius:0 0 16px 16px;">
      <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
        <tr><td class="mobile-pad" style="padding:36px 40px 32px;">
          %s
        </td></tr>
      </table>
    </td>
  </tr>

  <!-- FOOTER -->
  <tr>
    <td style="padding:24px 0 40px;text-align:center;">
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#3d5170;margin:0 0 6px;">&copy; %d %s · %s</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#3d5170;margin:0 0 6px;">
        <a href="%s" style="color:#4f6a9a;text-decoration:none;">Website</a> &nbsp;·&nbsp;
        <a href="mailto:%s" style="color:#4f6a9a;text-decoration:none;">Support</a> &nbsp;·&nbsp;
        WhatsApp %s
      </p>
      <p style="font-family:Arial,sans-serif;font-size:11px;color:#2a3a52;margin:0;">You received this because you have an account on Salen Mobile Guru.</p>
    </td>
  </tr>

</table>
</td></tr>
</table>
</body>
</html>`,
		preheader,
		brandName, brandTagline,
		content,
		time.Now().Year(), brandName, brandLocation,
		brandWebsite, brandSupport, brandWhatsApp,
	)
}

// renderTmpl renders a Go html/template string with data.
func renderTmpl(name, tmpl string, data interface{}) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ─── 1. OTP Email ─────────────────────────────────────────────────────────

// SendOTPEmail sends a professional verification code email.
// purpose: "signup" | "login" | "reset_password"
func SendOTPEmail(db *gorm.DB, toEmail, recipientName, code, purpose string) error {
	if recipientName == "" {
		recipientName = "Technician"
	}

	headings := map[string]string{
		"signup":         "Verify Your Email Address",
		"login":          "Your Login Code",
		"reset_password": "Reset Your Password",
	}
	bodyTexts := map[string]string{
		"signup":         "Thank you for joining <strong>Salen Mobile Guru</strong>. Use the code below to verify your email and activate your account.",
		"login":          "A sign-in request was made for your account. Use the code below to complete login.",
		"reset_password": "We received a password reset request. Use this code to set a new password. If this wasn't you, ignore this email.",
	}
	heading := headings[purpose]
	if heading == "" {
		heading = "Your Verification Code"
	}
	bodyText := bodyTexts[purpose]
	if bodyText == "" {
		bodyText = "Use the code below to proceed."
	}

	// Format as "123 456" for readability
	display := code
	if len(code) == 6 {
		display = code[:3] + "&nbsp;" + code[3:]
	}

	subject := fmt.Sprintf("[%s] Verification code: %s", brandName, code)

	content := fmt.Sprintf(`
<p style="font-family:Arial,sans-serif;font-size:16px;font-weight:700;color:#e8f0fe;margin:0 0 6px;">Hello, %s 👋</p>
<p style="font-family:Arial,sans-serif;font-size:15px;font-weight:600;color:#c9d8f0;margin:0 0 16px;">%s</p>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 28px;">%s</p>

<!-- OTP Box -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:28px;">
  <tr>
    <td style="background:#080d1a;border:2px dashed #2563eb;border-radius:14px;padding:28px;text-align:center;">
      <p style="font-family:Arial,sans-serif;font-size:11px;font-weight:700;color:#3b82f6;text-transform:uppercase;letter-spacing:2px;margin:0 0 14px;">One-Time Password</p>
      <p class="otp-code" style="font-family:'Courier New',Courier,monospace;font-size:44px;font-weight:900;letter-spacing:12px;color:#ffffff;margin:0 0 12px;">%s</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#4a6080;margin:0;">⏱ Expires in <strong style="color:#94a3b8;">10 minutes</strong></p>
    </td>
  </tr>
</table>

<!-- Steps -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#080d1a;border-radius:10px;margin-bottom:24px;">
  <tr>
    <td style="padding:18px 20px;">
      <p style="font-family:Arial,sans-serif;font-size:12px;font-weight:700;color:#4f6a8a;text-transform:uppercase;letter-spacing:1px;margin:0 0 10px;">How to use</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#5a7a9a;margin:0 0 6px;line-height:1.6;">1. Return to the Salen Mobile Guru app.</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#5a7a9a;margin:0 0 6px;line-height:1.6;">2. Enter the 6-digit code in the verification field.</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#5a7a9a;margin:0;line-height:1.6;">3. Tap <strong style="color:#7aa0c4;">Verify</strong> to continue.</p>
    </td>
  </tr>
</table>

<!-- Security -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
  <tr>
    <td style="background:#1a0a0a;border-left:3px solid #ef4444;border-radius:0 8px 8px 0;padding:14px 16px;">
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#c87070;margin:0;line-height:1.6;">
        🔒 <strong>Security:</strong> Salen Mobile Guru staff will <strong>never</strong> ask for your OTP by call, WhatsApp, or any other channel.
      </p>
    </td>
  </tr>
</table>`,
		recipientName, heading, bodyText, display,
	)

	preheader := fmt.Sprintf("Your SMG verification code: %s — expires in 10 minutes.", code)
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 2. Welcome Email ─────────────────────────────────────────────────────

// SendWelcomeEmail sends an onboarding email after successful registration.
func SendWelcomeEmail(db *gorm.DB, toEmail, recipientName string) error {
	if recipientName == "" {
		recipientName = "Technician"
	}
	subject := fmt.Sprintf("Welcome to %s, %s! 🎉 Let's Master Hardware Repair", brandName, recipientName)

	content := fmt.Sprintf(`
<!-- Hero -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:28px;">
  <tr>
    <td style="background:linear-gradient(135deg,#0f1f3d,#1a0f3d);border-radius:12px;padding:28px;text-align:center;">
      <p style="font-family:Arial,sans-serif;font-size:32px;margin:0 0 8px;">🎓</p>
      <p style="font-family:Arial,sans-serif;font-size:20px;font-weight:800;color:#e8f0fe;margin:0 0 6px;">You're in, %s!</p>
      <p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;margin:0;">Your Salen Mobile Guru account is now active.</p>
    </td>
  </tr>
</table>

<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  Welcome to the most advanced mobile hardware repair academy in Nepal &amp; India. You now have access to professional-grade repair education trusted by thousands of technicians.
</p>

<!-- Feature grid -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:28px;">
  <tr>
    <td width="48%%" valign="top" style="background:#080d1a;border-radius:10px;padding:16px;">
      <p style="font-size:20px;margin:0 0 8px;">📱</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;margin:0 0 4px;">HD Video Courses</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#4f6a8a;margin:0;line-height:1.5;">Reballing, EMMC, PMIC, schematic reading &amp; more</p>
    </td>
    <td width="4%%"></td>
    <td width="48%%" valign="top" style="background:#080d1a;border-radius:10px;padding:16px;">
      <p style="font-size:20px;margin:0 0 8px;">🔴</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;margin:0 0 4px;">Live Masterclasses</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#4f6a8a;margin:0;line-height:1.5;">Weekly live sessions with Master Salen &amp; instructors</p>
    </td>
  </tr>
  <tr><td colspan="3" style="height:12px;"></td></tr>
  <tr>
    <td width="48%%" valign="top" style="background:#080d1a;border-radius:10px;padding:16px;">
      <p style="font-size:20px;margin:0 0 8px;">📐</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;margin:0 0 4px;">Hardware Schematics</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#4f6a8a;margin:0;line-height:1.5;">Borneo, ZXW, diode mode &amp; jumper guides</p>
    </td>
    <td width="4%%"></td>
    <td width="48%%" valign="top" style="background:#080d1a;border-radius:10px;padding:16px;">
      <p style="font-size:20px;margin:0 0 8px;">🏆</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;margin:0 0 4px;">Certificates</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#4f6a8a;margin:0;line-height:1.5;">Industry-recognised upon course completion</p>
    </td>
  </tr>
</table>

<!-- CTA -->
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:20px;">
  <tr>
    <td align="center">
      <a href="%s" style="display:inline-block;background:linear-gradient(135deg,#1d4ed8,#7c3aed);color:#fff;font-family:Arial,sans-serif;font-size:15px;font-weight:700;text-decoration:none;padding:14px 36px;border-radius:10px;">
        Open Your Dashboard →
      </a>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;text-align:center;margin:0;">
  Questions? WhatsApp us at <strong style="color:#7aa0c4;">%s</strong>
</p>`,
		recipientName, brandWebsite, brandWhatsApp,
	)

	preheader := fmt.Sprintf("Your %s account is active. Start learning hardware repair today.", brandName)
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 3. Subscription Confirmed ────────────────────────────────────────────

// SubscriptionEmailData holds the data for a subscription confirmation email.
type SubscriptionEmailData struct {
	RecipientName string
	PlanName      string
	Amount        string
	Currency      string
	ExpiresAt     string
	TransactionID string
	PaymentMethod string
}

// SendSubscriptionConfirmedEmail notifies the student their payment was approved.
func SendSubscriptionConfirmedEmail(db *gorm.DB, toEmail string, d SubscriptionEmailData) error {
	if d.RecipientName == "" {
		d.RecipientName = "Technician"
	}
	subject := fmt.Sprintf("✅ Subscription Activated — %s | %s", d.PlanName, brandName)

	content := fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:24px;">
  <tr>
    <td style="background:linear-gradient(135deg,#052e16,#064e3b);border:1px solid #065f46;border-radius:12px;padding:24px;text-align:center;">
      <p style="font-size:36px;margin:0 0 8px;">✅</p>
      <p style="font-family:Arial,sans-serif;font-size:18px;font-weight:800;color:#d1fae5;margin:0 0 4px;">Payment Approved!</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#6ee7b7;margin:0;">Your subscription is now active.</p>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  Hi <strong style="color:#c9d8f0;">%s</strong>, your <strong>%s</strong> subscription is now active. You have full access to all plan features.
</p>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#080d1a;border-radius:12px;margin-bottom:24px;">
  <tr><td style="padding:16px 20px;border-bottom:1px solid #1e2d45;">
    <p style="font-family:Arial,sans-serif;font-size:11px;font-weight:700;color:#3b5270;text-transform:uppercase;letter-spacing:1px;margin:0;">Subscription Details</p>
  </td></tr>
  <tr><td style="padding:0 20px;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Plan</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Amount Paid</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s %s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Payment Method</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Transaction ID</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#7aa0c4;">%s</code></td></tr>
      <tr><td style="padding:12px 0;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Valid Until</td><td style="padding:12px 0;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#6ee7b7;">%s</td></tr>
    </table>
  </td></tr>
</table>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
  <tr><td align="center">
    <a href="%s" style="display:inline-block;background:linear-gradient(135deg,#1d4ed8,#7c3aed);color:#fff;font-family:Arial,sans-serif;font-size:15px;font-weight:700;text-decoration:none;padding:14px 36px;border-radius:10px;">
      Start Learning Now →
    </a>
  </td></tr>
</table>`,
		d.RecipientName, d.PlanName,
		d.PlanName, d.Amount, d.Currency,
		d.PaymentMethod, d.TransactionID, d.ExpiresAt,
		brandWebsite,
	)

	preheader := fmt.Sprintf("Your %s subscription is now active. Full access unlocked!", d.PlanName)
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 4. Payment Pending ───────────────────────────────────────────────────

// PaymentPendingEmailData holds data for payment-received-pending email.
type PaymentPendingEmailData struct {
	RecipientName string
	PlanName      string
	Amount        string
	Currency      string
	TransactionID string
	PaymentMethod string
}

// SendPaymentReceivedEmail notifies the student that payment is under review.
func SendPaymentReceivedEmail(db *gorm.DB, toEmail string, d PaymentPendingEmailData) error {
	if d.RecipientName == "" {
		d.RecipientName = "Technician"
	}
	subject := fmt.Sprintf("⏳ Payment Received — Under Review | %s", brandName)

	content := fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:24px;">
  <tr>
    <td style="background:linear-gradient(135deg,#1c1400,#2d2000);border:1px solid #92400e;border-radius:12px;padding:24px;text-align:center;">
      <p style="font-size:36px;margin:0 0 8px;">⏳</p>
      <p style="font-family:Arial,sans-serif;font-size:18px;font-weight:800;color:#fde68a;margin:0 0 4px;">Payment Received!</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#fbbf24;margin:0;">Under admin review — approved within 2 hours.</p>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  Hi <strong style="color:#c9d8f0;">%s</strong>, we have received your payment for the <strong>%s</strong> plan. Our team will verify and activate your subscription shortly.
</p>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#080d1a;border-radius:12px;margin-bottom:24px;">
  <tr><td style="padding:0 20px;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Plan</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Amount</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s %s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Method</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Transaction ID</td><td style="padding:12px 0;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#7aa0c4;">%s</code></td></tr>
    </table>
  </td></tr>
</table>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
  <tr>
    <td style="background:#080d1a;border-left:3px solid #3b82f6;border-radius:0 8px 8px 0;padding:14px 16px;">
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#7a9cc4;margin:0;line-height:1.6;">
        📬 You will receive another email once your subscription is activated. For faster support, WhatsApp us at <strong style="color:#93c5fd;">%s</strong>
      </p>
    </td>
  </tr>
</table>`,
		d.RecipientName, d.PlanName,
		d.PlanName, d.Amount, d.Currency, d.PaymentMethod, d.TransactionID,
		brandWhatsApp,
	)

	preheader := "Your payment was received and is being reviewed. Activation coming soon!"
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 5. Certificate Issued ────────────────────────────────────────────────

// CertificateEmailData holds the data for a certificate issued email.
type CertificateEmailData struct {
	RecipientName     string
	CourseTitle       string
	Grade             string
	CertificateNumber string
	IssuedAt          string
	VerifyURL         string
	PdfURL            string
}

// SendCertificateIssuedEmail notifies a student their certificate is ready.
func SendCertificateIssuedEmail(db *gorm.DB, toEmail string, d CertificateEmailData) error {
	if d.RecipientName == "" {
		d.RecipientName = "Technician"
	}
	subject := fmt.Sprintf("🏆 Certificate Ready — %s | %s", d.CourseTitle, brandName)

	content := fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:24px;">
  <tr>
    <td style="background:linear-gradient(135deg,#1e1200,#2d1f00);border:1px solid #78350f;border-radius:12px;padding:28px;text-align:center;">
      <p style="font-size:40px;margin:0 0 8px;">🏆</p>
      <p style="font-family:Arial,sans-serif;font-size:19px;font-weight:800;color:#fef3c7;margin:0 0 4px;">Congratulations, %s!</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#fcd34d;margin:0;">You have successfully completed your course.</p>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  We are proud to certify your completion of <strong style="color:#c9d8f0;">%s</strong>. Your certificate is available for download and can be verified online.
</p>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:linear-gradient(135deg,#0f1a2e,#1a0f2e);border:1px solid #2d3a5a;border-radius:14px;margin-bottom:24px;">
  <tr>
    <td style="padding:24px;text-align:center;">
      <p style="font-family:Arial,sans-serif;font-size:11px;font-weight:700;color:#4f6a9a;text-transform:uppercase;letter-spacing:2px;margin:0 0 12px;">Certificate of Completion</p>
      <p style="font-family:Arial,sans-serif;font-size:17px;font-weight:800;color:#e8f0fe;margin:0 0 6px;">%s</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#7a9cc4;margin:0 0 12px;">%s</p>
      <p style="font-family:Arial,sans-serif;font-size:22px;font-weight:900;color:#fbbf24;margin:0 0 8px;">%s</p>
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#3b5270;margin:0 0 12px;">Issued: %s</p>
      <p style="font-family:Arial,sans-serif;font-size:11px;color:#3b5270;margin:0;">No: <code style="font-family:'Courier New',monospace;color:#5a7a9a;">%s</code></p>
    </td>
  </tr>
</table>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:16px;">
  <tr>
    <td width="48%%" align="center" style="padding:4px;">
      <a href="%s" style="display:block;background:linear-gradient(135deg,#1d4ed8,#7c3aed);color:#fff;font-family:Arial,sans-serif;font-size:13px;font-weight:700;text-decoration:none;padding:12px 16px;border-radius:8px;">Download PDF</a>
    </td>
    <td width="4%%"></td>
    <td width="48%%" align="center" style="padding:4px;">
      <a href="%s" style="display:block;background:#080d1a;border:1px solid #2d3a5a;color:#93c5fd;font-family:Arial,sans-serif;font-size:13px;font-weight:700;text-decoration:none;padding:12px 16px;border-radius:8px;">Verify Online</a>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:12px;color:#4f6a8a;text-align:center;margin:0;">Share your achievement on LinkedIn to showcase your hardware repair expertise! 🚀</p>`,
		d.RecipientName, d.CourseTitle,
		d.RecipientName, d.CourseTitle, d.Grade, d.IssuedAt, d.CertificateNumber,
		d.PdfURL, d.VerifyURL,
	)

	preheader := fmt.Sprintf("Your certificate for %s is ready to download!", d.CourseTitle)
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 6. Live Class Reminder ───────────────────────────────────────────────

// LiveClassEmailData holds data for a live class reminder email.
type LiveClassEmailData struct {
	RecipientName string
	SessionTitle  string
	Instructor    string
	Date          string
	Time          string
	Platform      string
	JoinURL       string
	DurationMins  int
}

// SendLiveClassReminderEmail notifies enrolled students before a class starts.
func SendLiveClassReminderEmail(db *gorm.DB, toEmail string, d LiveClassEmailData) error {
	if d.RecipientName == "" {
		d.RecipientName = "Technician"
	}
	platformName := d.Platform
	switch strings.ToUpper(d.Platform) {
	case "GOOGLE_MEET":
		platformName = "Google Meet"
	case "ZOOM":
		platformName = "Zoom"
	case "TEAMS":
		platformName = "Microsoft Teams"
	case "YOUTUBE_LIVE":
		platformName = "YouTube Live"
	}

	subject := fmt.Sprintf("🔴 Live Class Starting Soon — %s | %s", d.SessionTitle, brandName)

	content := fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:24px;">
  <tr>
    <td style="background:linear-gradient(135deg,#1a0000,#2d0a0a);border:1px solid #7f1d1d;border-radius:12px;padding:24px;text-align:center;">
      <p style="font-family:Arial,sans-serif;font-size:11px;font-weight:700;color:#fca5a5;text-transform:uppercase;letter-spacing:2px;margin:0 0 8px;">🔴 Live Class Reminder</p>
      <p style="font-family:Arial,sans-serif;font-size:18px;font-weight:800;color:#fee2e2;margin:0 0 4px;">%s</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#fca5a5;margin:0;">by %s</p>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  Hi <strong style="color:#c9d8f0;">%s</strong>, your live masterclass is starting soon. Make sure you're ready to join on time!
</p>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#080d1a;border-radius:12px;margin-bottom:24px;">
  <tr><td style="padding:0 20px;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">📅 Date &amp; Time</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s · %s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">👨‍🏫 Instructor</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">🎥 Platform</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%s</td></tr>
      <tr><td style="padding:12px 0;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">⏱ Duration</td><td style="padding:12px 0;text-align:right;font-family:Arial,sans-serif;font-size:13px;font-weight:700;color:#c9d8f0;">%d minutes</td></tr>
    </table>
  </td></tr>
</table>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:20px;">
  <tr><td align="center">
    <a href="%s" style="display:inline-block;background:linear-gradient(135deg,#dc2626,#b91c1c);color:#fff;font-family:Arial,sans-serif;font-size:16px;font-weight:800;text-decoration:none;padding:16px 40px;border-radius:10px;">
      🔴 Join Live Class
    </a>
  </td></tr>
</table>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
  <tr>
    <td style="background:#080d1a;border-left:3px solid #f59e0b;border-radius:0 8px 8px 0;padding:14px 16px;">
      <p style="font-family:Arial,sans-serif;font-size:12px;color:#d97706;margin:0;line-height:1.6;">
        ⚡ <strong>Tip:</strong> Join 5 minutes early to test your audio/video. Ask questions in the chat during the session.
      </p>
    </td>
  </tr>
</table>`,
		d.SessionTitle, d.Instructor,
		d.RecipientName,
		d.Date, d.Time, d.Instructor, platformName, d.DurationMins,
		d.JoinURL,
	)

	preheader := fmt.Sprintf("Live class '%s' is starting soon. Join now!", d.SessionTitle)
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── 7. Admin SMTP Test ───────────────────────────────────────────────────

// SendTestEmail validates SMTP connectivity by sending a probe email.
func SendTestEmail(db *gorm.DB, toEmail string) error {
	cfg := GetSMTPSettings(db)
	ts := time.Now().Format("02 Jan 2006, 15:04:05 UTC")
	subject := fmt.Sprintf("✅ SMTP Test Successful — %s [%s]", brandName, ts)

	content := fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom:24px;">
  <tr>
    <td style="background:linear-gradient(135deg,#052e16,#064e3b);border:1px solid #065f46;border-radius:12px;padding:24px;text-align:center;">
      <p style="font-size:36px;margin:0 0 8px;">✅</p>
      <p style="font-family:Arial,sans-serif;font-size:18px;font-weight:800;color:#d1fae5;margin:0 0 4px;">SMTP Connection Verified!</p>
      <p style="font-family:Arial,sans-serif;font-size:13px;color:#6ee7b7;margin:0;">Your email system is working correctly.</p>
    </td>
  </tr>
</table>
<p style="font-family:Arial,sans-serif;font-size:14px;color:#7a9cc4;line-height:1.7;margin:0 0 24px;">
  This confirms that your SMTP server is properly configured and actively delivering emails for <strong style="color:#c9d8f0;">%s</strong>.
</p>
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#080d1a;border-radius:12px;margin-bottom:20px;">
  <tr><td style="padding:16px 20px;border-bottom:1px solid #1e2d45;">
    <p style="font-family:Arial,sans-serif;font-size:11px;font-weight:700;color:#3b5270;text-transform:uppercase;letter-spacing:1px;margin:0;">SMTP Configuration</p>
  </td></tr>
  <tr><td style="padding:0 20px;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Host</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#93c5fd;">%s</code></td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Port</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#93c5fd;">%d</code></td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">From</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#93c5fd;">%s &lt;%s&gt;</code></td></tr>
      <tr><td style="padding:12px 0;border-bottom:1px solid #111827;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Security</td><td style="padding:12px 0;border-bottom:1px solid #111827;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#93c5fd;">%s</code></td></tr>
      <tr><td style="padding:12px 0;font-family:Arial,sans-serif;font-size:13px;color:#4f6a8a;">Tested At</td><td style="padding:12px 0;text-align:right;"><code style="font-family:'Courier New',monospace;font-size:12px;color:#93c5fd;">%s</code></td></tr>
    </table>
  </td></tr>
</table>`,
		brandName,
		cfg.Host, cfg.Port, cfg.SenderName, cfg.SenderEmail, cfg.SecurityMode, ts,
	)

	preheader := "SMTP configuration test passed. Email delivery is working correctly."
	return SendMail(db, toEmail, subject, layout(preheader, content))
}

// ─── Async helpers ────────────────────────────────────────────────────────

// SendWelcomeEmailAsync fires the welcome email in a background goroutine.
func SendWelcomeEmailAsync(db *gorm.DB, toEmail, name string) {
	go func() {
		if err := SendWelcomeEmail(db, toEmail, name); err != nil {
			log.Printf("⚠️  Welcome email failed [%s]: %v", toEmail, err)
		}
	}()
}

// SendCertificateEmailAsync fires the certificate email in background.
func SendCertificateEmailAsync(db *gorm.DB, toEmail string, data CertificateEmailData) {
	go func() {
		if err := SendCertificateIssuedEmail(db, toEmail, data); err != nil {
			log.Printf("⚠️  Certificate email failed [%s]: %v", toEmail, err)
		}
	}()
}

// SendLiveReminderAsync fires the live class reminder in background.
func SendLiveReminderAsync(db *gorm.DB, toEmail string, data LiveClassEmailData) {
	go func() {
		if err := SendLiveClassReminderEmail(db, toEmail, data); err != nil {
			log.Printf("⚠️  Live reminder email failed [%s]: %v", toEmail, err)
		}
	}()
}

// SendSubscriptionEmailAsync fires the subscription confirmation in background.
func SendSubscriptionEmailAsync(db *gorm.DB, toEmail string, data SubscriptionEmailData) {
	go func() {
		if err := SendSubscriptionConfirmedEmail(db, toEmail, data); err != nil {
			log.Printf("⚠️  Subscription email failed [%s]: %v", toEmail, err)
		}
	}()
}

// _ suppress unused import for template (used by renderTmpl above)
var _ = template.New