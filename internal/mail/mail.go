package mail

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
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

// GetSMTPSettings retrieves settings from DB with fallback to environment variables
func GetSMTPSettings(db *gorm.DB) *models.SMTPSettings {
	var s models.SMTPSettings
	if db != nil {
		if err := db.Where("id = ?", "default").First(&s).Error; err == nil {
			return &s
		}
	}

	// Environment variable defaults
	enabled := strings.ToLower(os.Getenv("SMTP_ENABLED")) == "true" || strings.ToLower(os.Getenv("SMTP_ENABLED")) == "1"
	portStr := os.Getenv("SMTP_PORT")
	port := 587
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		port = p
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "support@salenmobileguru.com"
	}
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "Salen Mobile Guru Academy"
	}
	sec := os.Getenv("SMTP_SECURITY")
	if sec == "" {
		sec = "STARTTLS"
	}

	s = models.SMTPSettings{
		ID:           "default",
		Enabled:      enabled,
		SenderEmail:  fromEmail,
		SenderName:   fromName,
		Host:         host,
		Port:         port,
		Username:     user,
		Password:     pass,
		SecurityMode: strings.ToUpper(sec),
		UpdatedAt:    time.Now(),
	}

	if db != nil {
		_ = db.Create(&s)
	}
	return &s
}

// SaveSMTPSettings stores updated SMTP configurations into the database
func SaveSMTPSettings(db *gorm.DB, updated *models.SMTPSettings) error {
	updated.ID = "default"
	updated.UpdatedAt = time.Now()
	return db.Save(updated).Error
}

// GenerateOTP produces a secure 6-digit numeric verification code
func GenerateOTP() string {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// SendMail dispatches an email using the configured SMTP server or logs a simulation
func SendMail(db *gorm.DB, to string, subject string, htmlBody string) error {
	cfg := GetSMTPSettings(db)

	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient email address cannot be empty")
	}

	// If custom SMTP is disabled or credentials missing, log simulation
	if !cfg.Enabled || cfg.Host == "" {
		log.Printf("📧 [SIMULATED EMAIL - SMTP Disabled] To: %s | Subject: %s", to, subject)
		return nil
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	fromHeader := fmt.Sprintf("%s <%s>", cfg.SenderName, cfg.SenderEmail)
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	headers := make(map[string]string)
	headers["From"] = fromHeader
	headers["To"] = to
	headers["Subject"] = encodedSubject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "8bit"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	secMode := strings.ToUpper(cfg.SecurityMode)

	// Mode 1: SSL / TLS direct connection (Port 465)
	if secMode == "SSL" || secMode == "TLS" || cfg.Port == 465 {
		tlsConfig := &tls.Config{
			ServerName: cfg.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to establish SSL/TLS connection to %s: %w", addr, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("SMTP client initialization error: %w", err)
		}
		defer client.Close()

		if cfg.Username != "" && cfg.Password != "" {
			auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed for %s: %w", cfg.Username, err)
			}
		}

		if err = client.Mail(cfg.SenderEmail); err != nil {
			return fmt.Errorf("MAIL FROM error: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("RCPT TO error: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("SMTP DATA stream error: %w", err)
		}
		if _, err = w.Write([]byte(message)); err != nil {
			return fmt.Errorf("failed writing email message: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("failed closing DATA stream: %w", err)
		}

		_ = client.Quit()
		log.Printf("✅ Email successfully sent via SSL to %s (Subject: %s)", to, subject)
		return nil
	}

	// Mode 2: STARTTLS (Port 587 - Gmail Standard)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP host %s: %w", addr, err)
	}
	defer client.Close()

	if secMode != "NONE" {
		tlsConfig := &tls.Config{
			ServerName: cfg.Host,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS negotiation failed on %s: %w", addr, err)
		}
	}

	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed for %s: %w", cfg.Username, err)
		}
	}

	if err = client.Mail(cfg.SenderEmail); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA stream error: %w", err)
	}
	if _, err = w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed writing email message: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("failed closing DATA stream: %w", err)
	}

	_ = client.Quit()
	log.Printf("✅ Email successfully sent via STARTTLS to %s (Subject: %s)", to, subject)
	return nil
}

// SendOTPEmail sends a high-contrast, responsive OTP verification email
func SendOTPEmail(db *gorm.DB, toEmail string, recipientName string, code string, purpose string) error {
	if recipientName == "" {
		recipientName = "Technician"
	}

	purposeText := "Complete Your Sign Up"
	if purpose == "login" {
		purposeText = "Log In to Your Account"
	} else if purpose == "reset_password" {
		purposeText = "Reset Your Account Password"
	}

	subject := fmt.Sprintf("%s — Your Verification Code: %s", code, purposeText)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Verification Code</title>
</head>
<body style="margin:0;padding:24px;background-color:#07090e;font-family:'Segoe UI',Roboto,-apple-system,sans-serif;color:#f0f4fc;">
  <div style="max-width:560px;margin:0 auto;background:#0d111a;border:1px solid #1a2336;border-radius:16px;overflow:hidden;box-shadow:0 8px 32px rgba(0,0,0,0.5);">
    <!-- Header -->
    <div style="background:linear-gradient(135deg,#2563eb,#8b5cf6);padding:28px 24px;text-align:center;">
      <div style="font-size:26px;font-weight:800;color:#ffffff;letter-spacing:-0.5px;">Salen Mobile Guru</div>
      <div style="font-size:12px;color:rgba(255,255,255,0.85);font-weight:600;margin-top:4px;text-transform:uppercase;letter-spacing:1px;">Elite Hardware &amp; Micro-Soldering Academy</div>
    </div>

    <!-- Body -->
    <div style="padding:32px 28px;">
      <h2 style="font-size:20px;font-weight:700;color:#f0f4fc;margin:0 0 12px 0;">Hello %s,</h2>
      <p style="font-size:14px;line-height:1.6;color:#9cb0d1;margin:0 0 20px 0;">
        You recently requested an authorization code to <b>%s</b> on the Salen Mobile Guru platform. Please use the verification code below to proceed:
      </p>

      <!-- OTP Box -->
      <div style="background:#111622;border:2px dashed #2563eb;border-radius:12px;padding:22px;text-align:center;margin:24px 0;">
        <div style="font-size:11px;font-weight:700;color:#3b82f6;text-transform:uppercase;letter-spacing:1.5px;margin-bottom:8px;">Your One-Time Password (OTP)</div>
        <div style="font-family:'Courier New',Courier,monospace;font-size:38px;font-weight:900;letter-spacing:10px;color:#ffffff;padding-left:10px;">%s</div>
        <div style="font-size:11px;color:#5e7399;margin-top:8px;">Expires in <b>10 minutes</b></div>
      </div>

      <p style="font-size:13px;line-height:1.6;color:#5e7399;margin:0 0 20px 0;">
        <b>Security Notice:</b> Never share this verification code with anyone. Salen Mobile Guru instructors and staff will never ask for your OTP.
      </p>

      <div style="padding-top:20px;border-top:1px solid #1a2336;font-size:12px;color:#3d4f6e;">
        If you did not request this verification, please safely ignore this email or contact support.
      </div>
    </div>

    <!-- Footer -->
    <div style="background:#0a0d14;padding:16px 28px;text-align:center;font-size:11px;color:#3d4f6e;">
      &copy; %d Salen Mobile Guru Academy. All rights reserved.<br/>
      Kathmandu, Nepal • WhatsApp Support: +977-9800000000
    </div>
  </div>
</body>
</html>`, recipientName, purposeText, code, time.Now().Year())

	return SendMail(db, toEmail, subject, html)
}

// SendWelcomeEmail sends a welcome message upon successful registration
func SendWelcomeEmail(db *gorm.DB, toEmail string, recipientName string) error {
	if recipientName == "" {
		recipientName = "Technician"
	}
	subject := "Welcome to Salen Mobile Guru Academy — Let's Master Hardware Repair!"
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:24px;background:#07090e;font-family:sans-serif;color:#f0f4fc;">
  <div style="max-width:540px;margin:0 auto;background:#0d111a;border:1px solid #1a2336;border-radius:14px;padding:32px;">
    <h2 style="color:#3b82f6;margin-top:0;">Welcome, %s!</h2>
    <p style="color:#9cb0d1;line-height:1.6;">Your student technician account at <b>Salen Mobile Guru</b> is now verified and active.</p>
    <p style="color:#9cb0d1;line-height:1.6;">You now have instant access to modular hardware repair courses, live multimeter schematics, diagnostic assessments, and webinars with Master Salen.</p>
    <a href="https://salenmobileguru.com" style="display:inline-block;padding:12px 24px;background:#2563eb;color:#fff;border-radius:8px;text-decoration:none;font-weight:700;margin-top:14px;">Open Student Dashboard</a>
  </div>
</body>
</html>`, recipientName)

	return SendMail(db, toEmail, subject, html)
}

// SendTestEmail validates the SMTP connectivity and credentials
func SendTestEmail(db *gorm.DB, toEmail string) error {
	cfg := GetSMTPSettings(db)
	subject := fmt.Sprintf("✅ SMG SMTP Probe Test Successful [%s]", time.Now().Format("15:04:05"))
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:24px;background:#07090e;font-family:sans-serif;color:#f0f4fc;">
  <div style="max-width:540px;margin:0 auto;background:#0d111a;border:1px solid #10b981;border-radius:14px;padding:28px;">
    <div style="color:#10b981;font-size:20px;font-weight:800;margin-bottom:8px;">✓ SMTP Connection Verified!</div>
    <p style="color:#9cb0d1;">This test email confirms your custom SMTP server is properly configured and actively delivering emails for Salen Mobile Guru.</p>
    <div style="background:#111622;padding:14px;border-radius:8px;font-family:monospace;font-size:12px;color:#5e7399;margin:16px 0;">
      <div>SMTP Host: %s</div>
      <div>SMTP Port: %d</div>
      <div>Sender: %s &lt;%s&gt;</div>
      <div>Security: %s</div>
      <div>Timestamp: %s</div>
    </div>
  </div>
</body>
</html>`, cfg.Host, cfg.Port, cfg.SenderName, cfg.SenderEmail, cfg.SecurityMode, time.Now().Format(time.RFC1123))

	return SendMail(db, toEmail, subject, html)
}
