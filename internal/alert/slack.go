package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/vaultwatch/internal/lease"
)

// SlackHandler sends alerts to a Slack webhook.
type SlackHandler struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackHandler creates a SlackHandler with the given webhook URL.
func NewSlackHandler(webhookURL string) *SlackHandler {
	return &SlackHandler{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// OnAlert implements the Handler interface.
func (s *SlackHandler) OnAlert(l *lease.Lease) error {
	msg := fmt.Sprintf(":warning: *VaultWatch Alert*\nLease: `%s`\nStatus: `%s`\nExpires: `%s`",
		l.ID, l.Status(), l.ExpiresAt.Format(time.RFC3339))

	payload, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("slack: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
