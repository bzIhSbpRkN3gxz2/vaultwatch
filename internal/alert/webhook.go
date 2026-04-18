package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// WebhookHandler sends alert payloads to a generic HTTP webhook endpoint.
type WebhookHandler struct {
	URL    string
	Client *http.Client
}

type webhookPayload struct {
	LeaseID   string    `json:"lease_id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// NewWebhookHandler creates a WebhookHandler that posts JSON alerts to url.
func NewWebhookHandler(url string) *WebhookHandler {
	return &WebhookHandler{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// OnAlert implements Handler. It serialises the lease alert and POSTs it.
func (w *WebhookHandler) OnAlert(l *lease.Lease) error {
	payload := webhookPayload{
		LeaseID:   l.ID,
		Status:    l.Status().String(),
		ExpiresAt: l.ExpiresAt,
		Message:   fmt.Sprintf("Lease %s is %s", l.ID, l.Status().String()),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
