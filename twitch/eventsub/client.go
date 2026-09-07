package eventsub

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/docker/go-units"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/nicklaw5/helix/v2"
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
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get eventsub subscriptions: status code %d", response.StatusCode)
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
	if sub.Transport.Method != "webhook" {
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
	if sub.Transport.Callback != c.evenSubURL.String() {
		return false
	}
	return true
}

func (c *Client) createEventSubSubscription(eventType string) error {
	response, err := c.helixClient.CreateEventSubSubscription(&helix.EventSubSubscription{
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
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("failed to create eventsub subscription for %s: status code %d", eventType, response.StatusCode)
	}
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
		response, err := c.helixClient.RemoveEventSubSubscription(sub.ID)
		if err != nil {
			return fmt.Errorf("failed to remove existing subscription: %w", err)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return fmt.Errorf("failed to remove existing subscription: status code %d", response.StatusCode)
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
