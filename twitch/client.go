package twitch

import (
	"fmt"
	"time"

	esb "github.com/dnsge/twitch-eventsub-bindings"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
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
	err := c.apiClient.EventSubSubscribe("stream.online", esb.ConditionStreamOnline{
		BroadcasterUserID: c.userID,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to stream.online event: %w", err)
	}

	err = c.apiClient.EventSubSubscribe("stream.offline", esb.ConditionStreamOffline{
		BroadcasterUserID: c.userID,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to stream.offline event: %w", err)
	}
	return nil
}
