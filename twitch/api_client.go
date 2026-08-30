package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	esb "github.com/matthiasharzer/twitch-eventsub-bindings"
	esf "github.com/matthiasharzer/twitch-eventsub-framework"
)

const twitchAPIBaseURL = "https://api.twitch.tv/helix"

type APIClient struct {
	io.Closer
	EventSubHandler *esf.SubHandler

	subClient      *esf.SubClient
	eventSubURL    string
	callbackSecret string
	credentials    Credentials

	subContext    context.Context
	contextCancel context.CancelFunc
}

func NewAPIClient(eventSubURL, eventSubSecret, clientID, clientSecret string) (*APIClient, error) {
	subHandler := esf.NewSubHandler(true, []byte(eventSubSecret))
	credentials := NewRollingCredentials(clientID, clientSecret)
	subClient := esf.NewSubClient(credentials)

	subContext, cancel := context.WithCancel(context.Background())

	client := &APIClient{
		EventSubHandler: subHandler,

		credentials:    credentials,
		subClient:      subClient,
		eventSubURL:    eventSubURL,
		subContext:     subContext,
		contextCancel:  cancel,
		callbackSecret: eventSubSecret,
	}

	return client, nil
}

func (c *APIClient) EventSubSubscribe(eventType string, condition any) error {
	response, err := c.subClient.Subscribe(c.subContext, &esf.SubRequest{
		Type:      eventType,
		Condition: condition,
		Callback:  c.eventSubURL,
		Secret:    c.callbackSecret,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s event: %w", eventType, err)
	}
	logging.Debug("Using %d/%d of webhook cost limit", response.TotalCost, response.MaxTotalCost)
	return nil
}

func (c *APIClient) EventSubGetSubscriptions() ([]esb.Subscription, error) {
	subscriptions, err := c.subClient.GetSubscriptions(c.subContext, esf.StatusAny)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions.Data, nil
}

func (c *APIClient) EventSubDeleteSubscription(subscriptionID string) error {
	err := c.subClient.Unsubscribe(c.subContext, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription %s: %w", subscriptionID, err)
	}
	return nil
}

func (c *APIClient) EventSubHTTPHandler() http.Handler {
	return c.EventSubHandler
}

func (c *APIClient) GetUserID(username string) (string, error) {
	// TODO: If more twitch API calls are needed, consider using a proper Twitch API client library instead of making raw HTTP requests.

	url := fmt.Sprintf("%s/users?login=%s", twitchAPIBaseURL, username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	accessToken, err := c.credentials.AppToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	clientID, err := c.credentials.ClientID()
	if err != nil {
		return "", fmt.Errorf("failed to get client ID: %w", err)
	}

	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer funcutils.LogError(resp.Body.Close, "failed to close response body")

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("user not found")
	}

	return result.Data[0].ID, nil
}

func (c *APIClient) Close() error {
	c.contextCancel()
	return nil
}
