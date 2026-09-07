package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch/eventsub"
	"github.com/nicklaw5/helix/v2"
)

type StreamOnlineState struct {
	IsOnline            bool
	BroadcasterUserName string
	Title               string
	StartedAt           *time.Time
}

type streamOnlineOfflineEventPayload struct {
	ID                  string    `json:"id"`
	BroadcasterUserName string    `json:"broadcaster_user_name"`
	StartedAt           time.Time `json:"started_at"`
}

type Client struct {
	userID string

	unsubscribeEventSub observer.UnsubscribeFunc
	helixClient         *helix.Client
	eventSubClient      *eventsub.Client
	onlineChannel       observer.ReadWriteChannel[StreamOnlineState]
}

func NewClient(clientID, clientSecret, userName string, evenSubURL url.URL, eventSubSecret string) (*Client, error) {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	accessTokenResponse, err := helixClient.RequestAppAccessToken([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to request app access token: %w", err)
	}
	helixClient.SetAppAccessToken(accessTokenResponse.Data.AccessToken)

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
	case "stream.online", "stream.offline":
		var payload streamOnlineOfflineEventPayload
		err := json.Unmarshal(notification.Event, &payload)
		if err != nil {
			logging.Error("failed to unmarshal payload for event", "type", notification.Subscription.Type, "error", err)
			return
		}
		stream, err := c.getCurrentUserStream()
		if err != nil {
			logging.Error("failed to get stream for event", "type", notification.Subscription.Type, "error", err)
			return
		}
		if stream == nil {
			logging.Info("stream not found for event", "type", notification.Subscription.Type)
			return
		}

		logging.Info("received event", "type", notification.Subscription.Type, "broadcaster", payload.BroadcasterUserName, "title", stream.Title, "started_at", payload.StartedAt)
		c.onlineChannel.Publish(StreamOnlineState{
			IsOnline:            notification.Subscription.Type == "stream.online",
			BroadcasterUserName: payload.BroadcasterUserName,
			Title:               stream.Title,
			StartedAt:           &payload.StartedAt,
		})
	default:
		logging.Warn("received unknown event", "type", notification.Subscription.Type)
	}
}

func (c *Client) getCurrentUserStream() (*helix.Stream, error) {
	response, err := c.helixClient.GetStreams(&helix.StreamsParams{
		UserIDs: []string{c.userID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}
	if len(response.Data.Streams) == 0 {
		return nil, nil
	}
	return &response.Data.Streams[0], nil
}

func (c *Client) StartEventSub() error {
	if c.unsubscribeEventSub != nil {
		c.unsubscribeEventSub = nil
		c.unsubscribeEventSub()
	}

	err := c.eventSubClient.Start()
	if err != nil {
		return fmt.Errorf("failed to start eventsub client: %w", err)
	}
	c.unsubscribeEventSub = c.eventSubClient.Events().Subscribe(c.handleEventSubNotification)
	return nil
}

func (c *Client) EventSubHTTPHandler() http.Handler {
	return c.eventSubClient.HTTPHandler()
}

func (c *Client) OnlineChannel() observer.ReadonlyChannel[StreamOnlineState] {
	return c.onlineChannel
}

func (c *Client) Close() error {
	if c.unsubscribeEventSub != nil {
		c.unsubscribeEventSub = nil
		c.unsubscribeEventSub()
	}
	return nil
}
