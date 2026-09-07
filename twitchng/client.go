package twitchng

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitchng/eventsub"
	"github.com/nicklaw5/helix"
)

type StreamOnlineState struct {
	IsOnline            bool
	BroadcasterUserName string
	StartedAt           *time.Time
}

type streamOnlineOfflineEventPayload struct {
	ID                  string    `json:"id"`
	BroadcasterUserName string    `json:"broadcaster_user_name"`
	StartedAt           time.Time `json:"started_at"`
}

type Client struct {
	userID string

	helixClient    *helix.Client
	eventSubClient *eventsub.Client
	onlineChannel  observer.ReadWriteChannel[StreamOnlineState]
}

func NewClient(clientID, clientSecret, userName string, evenSubURL url.URL, eventSubSecret string) (*Client, error) {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	response, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{userName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if len(response.Data.Users) == 0 {
		return nil, fmt.Errorf("user '%s' not found", userName)
	}
	userID := response.Data.Users[0].ID
	eventSubClient := eventsub.NewClient(
		helixClient,
		userID,
		[]string{"stream.online", "stream.offline"},
		evenSubURL,
		eventSubSecret,
	)

	return &Client{
		userID:         userID,
		helixClient:    helixClient,
		eventSubClient: eventSubClient,
		onlineChannel:  observer.NewChannel[StreamOnlineState](),
	}, nil
}

func (c *Client) handleEventSubNotification(notification eventsub.Notification) {
	switch notification.Subscription.Type {
	case "stream.online":
	case "stream.offline":
		var payload streamOnlineOfflineEventPayload
		err := json.Unmarshal(notification.Event, &payload)
		if err != nil {
			fmt.Printf("failed to unmarshal payload for event %s: %v\n", notification.Subscription.Type, err)
			return
		}
		c.onlineChannel.Publish(StreamOnlineState{
			IsOnline:            notification.Subscription.Type == "stream.online",
			BroadcasterUserName: payload.BroadcasterUserName,
			StartedAt:           &payload.StartedAt,
		})
	default:
		fmt.Printf("received unknown event type: %s\n", notification.Subscription.Type)
	}
}

func (c *Client) StartEventSub() error {
	err := c.eventSubClient.Start()
	if err != nil {
		return fmt.Errorf("failed to start eventsub client: %w", err)
	}
	return nil
}

func (c *Client) EventSubHTTPHandler() http.Handler {
	return c.eventSubClient.HTTPHandler()
}

func (c *Client) OnlineChannel() observer.ReadonlyChannel[StreamOnlineState] {
	return c.onlineChannel
}

func (c *Client) Close() error {
	return nil
}
