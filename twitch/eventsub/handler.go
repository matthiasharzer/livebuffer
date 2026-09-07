package eventsub

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	"github.com/nicklaw5/helix/v2"
)

type Notification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

func (c *Client) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		limitedReader := io.LimitReader(r.Body, eventSubMaxPayload)
		body, err := io.ReadAll(limitedReader)
		if err != nil {
			logging.Error("failed to read request body", "error", err)
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			return
		}
		defer funcutils.LogError(r.Body.Close, "failed to close request body")

		if !helix.VerifyEventSubNotification(c.eventSubSecret, r.Header, string(body)) {
			logging.Error("failed to verify eventsub notification")
			http.Error(w, "failed to verify eventsub notification", http.StatusBadRequest)
			return
		}

		var notification Notification
		err = json.Unmarshal(body, &notification)
		if err != nil {
			logging.Error("failed to unmarshal eventsub notification", "error", err)
			http.Error(w, "failed to unmarshal eventsub notification", http.StatusBadRequest)
			return
		}

		if notification.Challenge != "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(notification.Challenge))
			return
		}

		go c.events.Publish(notification)
		w.WriteHeader(http.StatusOK)
	}
}
