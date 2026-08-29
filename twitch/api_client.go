package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	esf "github.com/dnsge/twitch-eventsub-framework"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

const twitchAPIBaseURL = "https://api.twitch.tv/helix"

type APIClient struct {
	io.Closer
	EventSubHandler *esf.SubHandler

	subClient      *esf.SubClient
	eventSubURL    string
	callbackSecret string
	clientID       string
	clientSecret   string

	subContext    context.Context
	contextCancel context.CancelFunc
}

func NewAPIClient(eventSubURL, eventSubSecret, clientID, clientSecret string) (*APIClient, error) {
	subHandler := esf.NewSubHandler(true, []byte(eventSubSecret))
	subClient := esf.NewSubClient(esf.NewStaticCredentials(clientID, clientSecret))

	subContext, cancel := context.WithCancel(context.Background())

	client := &APIClient{
		EventSubHandler: subHandler,

		clientID:       clientID,
		clientSecret:   clientSecret,
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

func (c *APIClient) EventSubHTTPHandler() http.Handler {
	return c.EventSubHandler
}

func (c *APIClient) getAccessToken() (string, error) {
	url := fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials", c.clientID, c.clientSecret)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer funcutils.LogError(resp.Body.Close, "failed to close response body")

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.AccessToken, nil
}

func (c *APIClient) GetUserID(username string) (string, error) {
	// TODO: If more twitch API calls are needed, consider using a proper Twitch API client library instead of making raw HTTP requests.

	url := fmt.Sprintf("%s/users?login=%s", twitchAPIBaseURL, username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	accessToken, err := c.getAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Client-Id", c.clientID)
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
