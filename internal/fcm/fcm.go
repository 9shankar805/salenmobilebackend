// Package fcm wraps the Firebase Admin SDK to provide push notification
// dispatch for the Salen Mobile Guru Flutter app.
//
// Initialization order:
//  1. On server start, call fcm.Init() — reads credentials from env.
//  2. Use fcm.SendToToken()    — single device push.
//  3. Use fcm.SendToTopic()    — topic-based multicast (e.g. "all-users").
//  4. Use fcm.SendMulticast()  — batch up to 500 FCM tokens at once.
//
// Environment variables (set in .env):
//
//	FIREBASE_CREDENTIALS_FILE   path to serviceAccountKey.json
//	FIREBASE_CREDENTIALS_JSON   raw JSON content (alternative to file path)
//	FIREBASE_PROJECT_ID         Firebase project ID (optional if in credentials)
package fcm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// ─── Client singleton ─────────────────────────────────────────────────────

var (
	client     *messaging.Client
	initOnce   sync.Once
	initErr    error
	isReady    bool
)

// Init initialises the Firebase Admin SDK.
// Safe to call multiple times — initialises only once.
func Init() error {
	initOnce.Do(func() {
		ctx := context.Background()

		var opt option.ClientOption

		// Option 1: raw JSON in env (ideal for containers / secrets managers)
		if raw := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_JSON")); raw != "" {
			opt = option.WithCredentialsJSON([]byte(raw))
			log.Println("🔥 FCM: using credentials from FIREBASE_CREDENTIALS_JSON env var")
		} else if path := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE")); path != "" {
			// Option 2: path to serviceAccountKey.json file
			opt = option.WithCredentialsFile(path)
			log.Printf("🔥 FCM: using credentials file: %s", path)
		} else {
			// No credentials configured — FCM will run in no-op mode
			initErr = fmt.Errorf("FCM not configured: set FIREBASE_CREDENTIALS_FILE or FIREBASE_CREDENTIALS_JSON in .env")
			log.Printf("⚠️  FCM: %v — push notifications will be skipped", initErr)
			return
		}

		cfg := &firebase.Config{}
		if pid := os.Getenv("FIREBASE_PROJECT_ID"); pid != "" {
			cfg.ProjectID = pid
		}

		app, err := firebase.NewApp(ctx, cfg, opt)
		if err != nil {
			initErr = fmt.Errorf("firebase.NewApp: %w", err)
			log.Printf("❌ FCM init error: %v", initErr)
			return
		}

		mc, err := app.Messaging(ctx)
		if err != nil {
			initErr = fmt.Errorf("app.Messaging: %w", err)
			log.Printf("❌ FCM messaging client error: %v", initErr)
			return
		}

		client = mc
		isReady = true
		log.Println("✅ FCM: Firebase Messaging client ready")
	})
	return initErr
}

// IsReady reports whether FCM was successfully initialised.
func IsReady() bool { return isReady }

// Reset clears the singleton so Init() can be called again with new credentials.
// Used by the admin panel hot-reload flow.
func Reset() {
	initOnce = sync.Once{}
	client = nil
	isReady = false
	initErr = nil
}

// ─── Payload types ────────────────────────────────────────────────────────

// Notification is the human-readable notification payload shown in the system tray.
type Notification struct {
	Title    string
	Body     string
	ImageURL string // optional banner image (Android / iOS rich push)
}

// Data is the key-value map delivered silently to the app (even when foregrounded).
type Data map[string]string

// ─── Send to a single FCM token ───────────────────────────────────────────

// SendToToken dispatches a push to one specific device token.
// Returns the FCM message ID on success.
func SendToToken(ctx context.Context, token string, notif Notification, data Data) (string, error) {
	if !isReady {
		return "", fmt.Errorf("FCM not initialised")
	}
	if token == "" {
		return "", fmt.Errorf("empty FCM token")
	}

	msg := buildMessage(notif, data)
	msg.Token = token

	msgID, err := client.Send(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("FCM send to token: %w", err)
	}
	return msgID, nil
}

// ─── Send to a topic ─────────────────────────────────────────────────────

// SendToTopic dispatches to all devices subscribed to a topic (e.g. "all-users").
// Firebase handles fan-out automatically.
func SendToTopic(ctx context.Context, topic string, notif Notification, data Data) (string, error) {
	if !isReady {
		return "", fmt.Errorf("FCM not initialised")
	}

	msg := buildMessage(notif, data)
	msg.Topic = topic

	msgID, err := client.Send(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("FCM send to topic %q: %w", topic, err)
	}
	return msgID, nil
}

// ─── Multicast (batch up to 500 tokens) ──────────────────────────────────

// MulticastResult summarises a batch push.
type MulticastResult struct {
	SuccessCount int
	FailureCount int
	InvalidTokens []string // tokens that should be removed from DB
}

// SendMulticast sends to up to 500 FCM tokens in a single batch request.
// Tokens that are invalid / unregistered are returned in MulticastResult.InvalidTokens
// so the caller can clean them from the database.
func SendMulticast(ctx context.Context, tokens []string, notif Notification, data Data) (*MulticastResult, error) {
	if !isReady {
		return nil, fmt.Errorf("FCM not initialised")
	}
	if len(tokens) == 0 {
		return &MulticastResult{}, nil
	}
	if len(tokens) > 500 {
		tokens = tokens[:500] // FCM hard limit per batch
	}

	msg := &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: notif.Title, Body: notif.Body, ImageURL: notif.ImageURL},
		Data:         dataToStringMap(data, notif),
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title:    notif.Title,
				Body:     notif.Body,
				ImageURL: notif.ImageURL,
				Sound:    "default",
				ChannelID: "smg_notifications",
			},
		},
		APNS: buildAPNS(notif),
	}

	resp, err := client.SendEachForMulticast(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("FCM multicast: %w", err)
	}

	result := &MulticastResult{
		SuccessCount: resp.SuccessCount,
		FailureCount: resp.FailureCount,
	}

	for i, r := range resp.Responses {
		if !r.Success && i < len(tokens) {
			errCode := ""
			if r.Error != nil {
				errCode = r.Error.Error()
			}
			// Mark unregistered / invalid tokens for cleanup
			if strings.Contains(errCode, "registration-token-not-registered") ||
				strings.Contains(errCode, "invalid-argument") ||
				strings.Contains(errCode, "UNREGISTERED") {
				result.InvalidTokens = append(result.InvalidTokens, tokens[i])
			}
		}
	}

	return result, nil
}

// ─── Subscribe / Unsubscribe from topics ─────────────────────────────────

// SubscribeToTopic subscribes a batch of FCM tokens to a topic.
// Use topic "all-users" so SendToTopic("all-users", ...) reaches everyone.
func SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	if !isReady || len(tokens) == 0 {
		return nil
	}
	_, err := client.SubscribeToTopic(ctx, tokens, topic)
	return err
}

// UnsubscribeFromTopic removes tokens from a topic.
func UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	if !isReady || len(tokens) == 0 {
		return nil
	}
	_, err := client.UnsubscribeFromTopic(ctx, tokens, topic)
	return err
}

// ─── Status endpoint helper ───────────────────────────────────────────────

// StatusMap returns a map suitable for the admin /system/status endpoint.
func StatusMap() map[string]interface{} {
	if !isReady {
		reason := "not configured"
		if initErr != nil {
			reason = initErr.Error()
		}
		return map[string]interface{}{
			"enabled": false,
			"status":  "unavailable",
			"reason":  reason,
		}
	}
	return map[string]interface{}{
		"enabled": true,
		"status":  "ready",
		"reason":  "Firebase Admin SDK initialised",
	}
}

// ─── Credential validator (called from admin panel) ───────────────────────

// ValidateCredentialsJSON checks whether a raw service-account JSON string is
// structurally valid (has project_id, client_email, private_key).
func ValidateCredentialsJSON(raw string) error {
	var sa struct {
		Type        string `json:"type"`
		ProjectID   string `json:"project_id"`
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if sa.Type != "service_account" {
		return fmt.Errorf("JSON type must be 'service_account', got '%s'", sa.Type)
	}
	if sa.ProjectID == "" {
		return fmt.Errorf("missing project_id")
	}
	if sa.ClientEmail == "" {
		return fmt.Errorf("missing client_email")
	}
	if sa.PrivateKey == "" {
		return fmt.Errorf("missing private_key")
	}
	return nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────

func buildMessage(notif Notification, data Data) *messaging.Message {
	return &messaging.Message{
		Notification: &messaging.Notification{
			Title:    notif.Title,
			Body:     notif.Body,
			ImageURL: notif.ImageURL,
		},
		Data: dataToStringMap(data, notif),
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title:       notif.Title,
				Body:        notif.Body,
				ImageURL:    notif.ImageURL,
				Sound:       "default",
				ChannelID:   "smg_notifications",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: buildAPNS(notif),
	}
}

func buildAPNS(notif Notification) *messaging.APNSConfig {
	return &messaging.APNSConfig{
		Headers: map[string]string{
			"apns-priority": "10",
		},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{
					Title: notif.Title,
					Body:  notif.Body,
				},
				Sound:            "default",
				MutableContent:   true, // allows iOS notification extension to attach images
				ContentAvailable: true, // silent background wake
			},
		},
	}
}

// dataToStringMap merges custom data + standard SMG fields into one string map.
// FCM data payloads must be map[string]string.
func dataToStringMap(data Data, notif Notification) map[string]string {
	m := map[string]string{
		"title":     notif.Title,
		"body":      notif.Body,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"app":       "salenmobileguru",
	}
	for k, v := range data {
		m[k] = v
	}
	return m
}
