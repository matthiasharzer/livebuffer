package twitch

import (
	"fmt"
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/util/marshalutil"
	esb "github.com/matthiasharzer/twitch-eventsub-bindings"
)

type StreamOnlineState struct {
	IsOnline  bool
	StartedAt *time.Time
}

// Client provides an interface to interact with the twitch API. It allows listening to events and retrieving information about streams, users, and more.
type Client struct {
	apiClient     *APIClient
	userID        string
	onlineChannel observer.ReadWriteChannel[StreamOnlineState]
}

func NewClient(apiClient *APIClient, userID string) (*Client, error) {
	client := &Client{
		apiClient:     apiClient,
		userID:        userID,
		onlineChannel: observer.NewChannel[StreamOnlineState](),
	}

	apiClient.EventSubHandler.HandleStreamOnline = client.handleStreamOnline
	apiClient.EventSubHandler.HandleStreamOffline = client.handleStreamOffline

	return client, nil
}

func (c *Client) handleStreamOnline(_ *esb.ResponseHeaders, event *esb.EventStreamOnline) {
	logging.Info("stream is online:", "username", event.BroadcasterUserName)
	startedAt, err := time.Parse(time.RFC3339, event.StartedAt)
	if err != nil {
		logging.Error("failed to parse started at time:", "error", err)
		return
	}
	c.onlineChannel.Publish(StreamOnlineState{
		IsOnline:  true,
		StartedAt: &startedAt,
	})
}

func (c *Client) handleStreamOffline(_ *esb.ResponseHeaders, event *esb.EventStreamOffline) {
	logging.Info("stream is offline:", "username", event.BroadcasterUserName)
	c.onlineChannel.Publish(StreamOnlineState{
		IsOnline: false,
	})
}

func (c *Client) OnlineChannel() observer.ReadonlyChannel[StreamOnlineState] {
	return c.onlineChannel
}

func (c *Client) StartEventSub() error {
	subscriptions, err := c.apiClient.EventSubGetSubscriptions()
	if err != nil {
		return fmt.Errorf("failed to get eventsub subscriptions: %w", err)
	}

	var existingStreamOnlineSub, existingStreamOfflineSub bool
	for _, sub := range subscriptions {
		if sub.Type != "stream.online" && sub.Type != "stream.offline" {
			continue
		}

		// stream.online and stream.offline have the same condition structure, so we can unmarshal it into a common struct
		condition := struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		}{}

		err = marshalutil.UnmarshalAny(sub.Condition, &condition)
		if err != nil {
			logging.Warn("failed to unmarshal condition for subscription", "error", err)
			continue
		}
		isWebhook := sub.Transport.Method == "webhook"
		isSameBroadcaster := condition.BroadcasterUserID == c.userID
		isSameCallback := sub.Transport.Callback == c.apiClient.eventSubURL

		subscriptionExists := isWebhook && isSameBroadcaster && isSameCallback
		if !subscriptionExists {
			continue
		}

		if sub.Type == "stream.online" {
			existingStreamOnlineSub = true
		}
		if sub.Type == "stream.offline" {
			existingStreamOfflineSub = true
		}
	}

	if !existingStreamOnlineSub {
		err = c.apiClient.EventSubSubscribe("stream.online", esb.ConditionStreamOnline{
			BroadcasterUserID: c.userID,
		})
		if err != nil {
			return fmt.Errorf("failed to subscribe to stream.online event: %w", err)
		}
		logging.Info("subscribed to stream.online event")
	} else {
		logging.Info("already subscribed to stream.online event")
	}

	if !existingStreamOfflineSub {
		err = c.apiClient.EventSubSubscribe("stream.offline", esb.ConditionStreamOffline{
			BroadcasterUserID: c.userID,
		})
		if err != nil {
			return fmt.Errorf("failed to subscribe to stream.offline event: %w", err)
		}
		logging.Info("subscribed to stream.offline event")
	} else {
		logging.Info("already subscribed to stream.offline event")
	}
	return nil
}
