package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

type Credentials interface {
	ClientID() (string, error)
	AppToken() (string, error)
}

type rollingCredentials struct {
	clientID     string
	clientSecret string
	appToken     string
	expiresAt    time.Time
}

func NewRollingCredentials(clientID, clientSecret string) Credentials {
	return &rollingCredentials{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (c *rollingCredentials) getAccessToken() (string, error) {
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

func (c *rollingCredentials) ClientID() (string, error) {
	return c.clientID, nil
}

func (c *rollingCredentials) AppToken() (string, error) {
	if c.appToken != "" && time.Now().Before(c.expiresAt) {
		return c.appToken, nil
	}

	token, err := c.getAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	c.appToken = token
	c.expiresAt = time.Now().Add(1 * time.Hour)

	return c.appToken, nil
}
