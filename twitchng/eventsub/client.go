package eventsub

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/docker/go-units"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/nicklaw5/helix"
)

const eventSubMaxPayload = 1 * units.MiB

type Client struct {
	userID            string
	evenSubURL        url.URL
	eventSubSecret    string
	subscriptionTypes []string

	helixClient *helix.Client
	events      observer.ReadWriteChannel[Notification]
}

func NewClient(helixClient *helix.Client, userID string, subscriptionTypes []string, evenSubURL url.URL, eventSubSecret string) *Client {
	return &Client{
		userID:            userID,
		evenSubURL:        evenSubURL,
		eventSubSecret:    eventSubSecret,
		subscriptionTypes: subscriptionTypes,
		helixClient:       helixClient,
		events:            observer.NewChannel[Notification](),
	}
}

func (c *Client) Events() observer.ReadonlyChannel[Notification] {
	return c.events
}

func (c *Client) getExistingEventSubSubscriptions() ([]helix.EventSubSubscription, error) {
	var after string
	var subscriptions []helix.EventSubSubscription

	for {
		params := &helix.EventSubSubscriptionsParams{
			After: after,
		}
		response, err := c.helixClient.GetEventSubSubscriptions(params)
		if err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, response.Data.EventSubSubscriptions...)

		if response.Data.Pagination.Cursor == "" {
			break
		}
		after = response.Data.Pagination.Cursor
	}

	return subscriptions, nil
}

func (c *Client) isMatchingEventSubSubscription(sub helix.EventSubSubscription, subTypes []string) bool {
	isWebsocket := sub.Transport.Method == "websocket"
	if !isWebsocket {
		return false
	}
	if sub.Transport.Callback != c.evenSubURL.String() {
		return false
	}
	if !slices.Contains(subTypes, sub.Type) {
		return false
	}
	if sub.Condition.BroadcasterUserID != c.userID {
		return false
	}
	//TODO: check if secret!!
	if sub.Transport.Callback != c.evenSubURL.String() {
		return false
	}
	return true
}

func (c *Client) createEventSubSubscription(eventType string) error {
	_, err := c.helixClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: c.evenSubURL.String(),
			Secret:   c.eventSubSecret,
		},
		Type:    eventType,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: c.userID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create eventsub subscription for %s: %w", eventType, err)
	}
	logging.Info("created eventsub subscription for", "type", eventType, "callback", c.evenSubURL.String())
	return nil
}

func (c *Client) Start() error {
	existingSubscriptions, err := c.getExistingEventSubSubscriptions()
	if err != nil {
		return fmt.Errorf("failed to get existing eventsub subscriptions: %w", err)
	}

	for _, sub := range existingSubscriptions {
		if !c.isMatchingEventSubSubscription(sub, c.subscriptionTypes) {
			continue
		}

		logging.Info("removing existing subscription", "type", sub.Type, "id", sub.ID)
		_, err = c.helixClient.RemoveEventSubSubscription(sub.ID)
		if err != nil {
			return fmt.Errorf("failed to remove existing subscription: %w", err)
		}
	}

	for _, subType := range c.subscriptionTypes {
		logging.Info("creating subscription for", "type", subType)
		err = c.createEventSubSubscription(subType)
		if err != nil {
			return fmt.Errorf("failed to create subscription for %s: %w", subType, err)
		}
	}

	return nil
}

func (c *Client) Close() error {
	c.events.Clear()
	return nil
}
