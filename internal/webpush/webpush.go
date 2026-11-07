// Package webpush implements VAPID-based Web Push notifications.
// These are sent to browsers and PWA (Progressive Web Apps) via the
// Web Push Protocol (RFC 8030) — no Firebase project needed.
//
// Your VAPID keys are already configured in .env:
//
//	VAPID_PUBLIC_KEY   BHlSAu0R1...  (share with Flutter/Web client)
//	VAPID_PRIVATE_KEY  UJY4fRJP...   (keep private on server)
//	VAPID_SUBJECT      mailto:support@salenmobileguru.com
//
// Flow:
//  1. Web/Flutter-web client calls the browser Push API to get a PushSubscription JSON.
//  2. Client POSTs that subscription to POST /webpush/subscribe.
//  3. Admin broadcasts via POST /api/admin/notifications/broadcast — hits both
//     FCM (Android/iOS) and Web Push (browser) in parallel.
package webpush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	webpushlib "github.com/SherClockHolmes/webpush-go"
)

// ─── State ────────────────────────────────────────────────────────────────

var (
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
	ready           bool
	initOnce        sync.Once
)

// Init reads VAPID keys from environment and validates them.
// Safe to call multiple times.
func Init() error {
	var initErr error
	initOnce.Do(func() {
		vapidPublicKey = strings.TrimSpace(os.Getenv("VAPID_PUBLIC_KEY"))
		vapidPrivateKey = strings.TrimSpace(os.Getenv("VAPID_PRIVATE_KEY"))
		vapidSubject = strings.TrimSpace(os.Getenv("VAPID_SUBJECT"))

		if vapidPublicKey == "" || vapidPrivateKey == "" {
			initErr = fmt.Errorf("VAPID_PUBLIC_KEY and VAPID_PRIVATE_KEY must be set in .env")
			log.Printf("⚠️  Web Push: %v", initErr)
			return
		}
		if vapidSubject == "" {
			vapidSubject = "mailto:support@salenmobileguru.com"
		}

		ready = true
		log.Printf("✅ Web Push: VAPID keys loaded (public: %s...)", vapidPublicKey[:16])
	})
	return initErr
}

// IsReady reports whether VAPID keys were loaded successfully.
func IsReady() bool { return ready }

// PublicKey returns the VAPID public key (safe to expose to clients).
func PublicKey() string { return vapidPublicKey }

// ─── Subscription type ────────────────────────────────────────────────────

// Subscription mirrors the browser PushSubscription JSON structure.
// The Flutter web / browser client sends this after calling
// PushManager.subscribe({ userVisibleOnly: true, applicationServerKey: VAPID_PUBLIC_KEY })
type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// Payload is the JSON body delivered inside the push envelope.
type Payload struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Icon     string `json:"icon,omitempty"`
	Badge    string `json:"badge,omitempty"`
	Image    string `json:"image,omitempty"`
	DeepLink string `json:"deepLink,omitempty"`
	Type     string `json:"type,omitempty"`
	Tag      string `json:"tag,omitempty"`     // deduplication key
	URL      string `json:"url,omitempty"`     // click destination
}

// ─── Send ─────────────────────────────────────────────────────────────────

// SendResult summarises the outcome of a push attempt.
type SendResult struct {
	Endpoint   string
	StatusCode int
	Err        error
	Expired    bool // true when the subscription is gone and should be deleted
}

// Send delivers a push notification to a single subscription.
func Send(sub Subscription, payload Payload) SendResult {
	if !ready {
		return SendResult{Endpoint: sub.Endpoint, Err: fmt.Errorf("web push not initialised")}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{Endpoint: sub.Endpoint, Err: fmt.Errorf("marshal payload: %w", err)}
	}

	wps := &webpushlib.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpushlib.Keys{
			P256dh: sub.Keys.P256dh,
			Auth:   sub.Keys.Auth,
		},
	}

	resp, err := webpushlib.SendNotification(body, wps, &webpushlib.Options{
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		Subscriber:      vapidSubject,
		TTL:             86400, // 24 hours
		Urgency:         webpushlib.UrgencyHigh,
	})
	if err != nil {
		return SendResult{Endpoint: sub.Endpoint, Err: fmt.Errorf("send: %w", err)}
	}
	defer resp.Body.Close()

	expired := resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound
	return SendResult{
		Endpoint:   sub.Endpoint,
		StatusCode: resp.StatusCode,
		Expired:    expired,
	}
}

// BroadcastResult summarises a bulk push operation.
type BroadcastResult struct {
	Total          int
	SuccessCount   int
	FailureCount   int
	ExpiredCount   int
	ExpiredEndpoints []string // subscriptions to delete from DB
	DurationMs     int64
}

// Broadcast sends to a slice of subscriptions concurrently (up to 50 goroutines).
func Broadcast(subs []Subscription, payload Payload) BroadcastResult {
	start := time.Now()
	result := BroadcastResult{Total: len(subs)}
	if len(subs) == 0 {
		return result
	}

	type job struct{ sub Subscription }
	jobs := make(chan job, len(subs))
	for _, s := range subs {
		jobs <- job{s}
	}
	close(jobs)

	type outcome struct{ r SendResult }
	outcomes := make(chan outcome, len(subs))

	// Fan-out with max 50 concurrent goroutines
	concurrency := 50
	if len(subs) < concurrency {
		concurrency = len(subs)
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				outcomes <- outcome{Send(j.sub, payload)}
			}
		}()
	}
	wg.Wait()
	close(outcomes)

	for o := range outcomes {
		if o.r.Err != nil || (o.r.StatusCode != 0 && o.r.StatusCode >= 400 && !o.r.Expired) {
			result.FailureCount++
			if o.r.Err != nil {
				log.Printf("⚠️  Web Push send error [%s]: %v", shortEndpoint(o.r.Endpoint), o.r.Err)
			}
		} else if o.r.Expired {
			result.ExpiredCount++
			result.ExpiredEndpoints = append(result.ExpiredEndpoints, o.r.Endpoint)
		} else {
			result.SuccessCount++
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	log.Printf("✅ Web Push broadcast — success:%d fail:%d expired:%d in %dms",
		result.SuccessCount, result.FailureCount, result.ExpiredCount, result.DurationMs)
	return result
}

// StatusMap returns a map suitable for the admin /system/status endpoint.
func StatusMap() map[string]interface{} {
	if !ready {
		return map[string]interface{}{
			"enabled":   false,
			"status":    "unavailable",
			"reason":    "VAPID keys not configured",
		}
	}
	return map[string]interface{}{
		"enabled":         true,
		"status":          "ready",
		"vapidPublicKey":  vapidPublicKey, // safe to expose
		"subject":         vapidSubject,
	}
}

// ValidateKeys checks that the supplied VAPID key pair looks structurally valid
// (correct base64url length for P-256 uncompressed public key = 87 chars,
// private key scalar = ~43 chars).
func ValidateKeys(pub, priv string) error {
	pub = strings.TrimSpace(pub)
	priv = strings.TrimSpace(priv)
	if len(pub) < 80 || len(pub) > 92 {
		return fmt.Errorf("VAPID public key looks wrong (got %d chars, expected ~87)", len(pub))
	}
	if len(priv) < 38 || len(priv) > 48 {
		return fmt.Errorf("VAPID private key looks wrong (got %d chars, expected ~43)", len(priv))
	}
	// Check base64url characters only
	for _, ch := range pub + priv {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return fmt.Errorf("VAPID keys must be base64url encoded (found invalid char '%c')", ch)
		}
	}
	return nil
}

// shortEndpoint trims a push endpoint URL for log brevity.
func shortEndpoint(e string) string {
	if len(e) > 60 {
		return e[:40] + "..." + e[len(e)-10:]
	}
	return e
}

// GenerateServiceWorkerSnippet returns a ready-to-paste JS snippet for the
// Flutter web / PWA service worker that registers push with this server's VAPID key.
func GenerateServiceWorkerSnippet(serverBaseURL string) string {
	return fmt.Sprintf(`// ── Salen Mobile Guru — Web Push Registration ──
// Paste this in your Flutter web main.dart or service-worker.js

const VAPID_PUBLIC_KEY = '%s';

async function registerWebPush() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    console.warn('Web Push not supported in this browser');
    return;
  }
  const reg = await navigator.serviceWorker.ready;
  const sub = await reg.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
  });
  // Send subscription to backend
  await fetch('%s/webpush/subscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ subscription: sub, userId: getCurrentUserId() }),
  });
  console.log('✅ Web Push subscription registered');
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length %% 4) %% 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
}`, vapidPublicKey, serverBaseURL)
}

// ─── Service Worker push handler snippet ──────────────────────────────────

// PushHandlerSnippet returns the service-worker.js push event listener code.
func PushHandlerSnippet() string {
	return `// ── service-worker.js — push event handler ──
self.addEventListener('push', event => {
  const data = event.data ? event.data.json() : {};
  const title = data.title || 'Salen Mobile Guru';
  const options = {
    body:    data.body  || '',
    icon:    data.icon  || '/icons/Icon-192.png',
    badge:   data.badge || '/icons/Icon-192.png',
    image:   data.image || undefined,
    tag:     data.tag   || 'smg-notification',
    data:    { deepLink: data.deepLink, url: data.url },
    actions: [{ action: 'open', title: 'Open App' }],
    vibrate: [200, 100, 200],
    requireInteraction: false,
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', event => {
  event.notification.close();
  const deepLink = event.notification.data && event.notification.data.deepLink;
  const url = deepLink ? ('/' + deepLink.replace(/^\//, '')) : '/';
  event.waitUntil(clients.openWindow(url));
});`
}

// Unused import guard
var _ = bytes.NewBuffer
