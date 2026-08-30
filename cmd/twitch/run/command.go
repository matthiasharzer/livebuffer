package run

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	"github.com/matthiasharzer/livebuffer/util/stringutil"
	"github.com/spf13/cobra"
)

var httpPort = 4000
var httpHost string
var username string
var bufferDirectoryArg string
var liveBufferPublicURL string
var maxStreams = 2
var eventSubSecretArg string

func init() {
	Command.Flags().IntVarP(&httpPort, "port", "p", httpPort, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
	Command.Flags().StringVarP(&username, "username", "u", "", "Twitch username to buffer (required)")
	Command.Flags().StringVarP(&bufferDirectoryArg, "buffer-dir", "", bufferDirectoryArg, "Directory to store live buffer segments (default: temporary directory)")
	Command.Flags().StringVarP(&liveBufferPublicURL, "public-url", "", "", "Public URL for the live buffer (required)")
	Command.Flags().IntVarP(&maxStreams, "max-streams", "", maxStreams, "Maximum number of concurrent streams to buffer")
	Command.Flags().StringVarP(&eventSubSecretArg, "eventsub-secret", "", eventSubSecretArg, "Secret for Twitch EventSub (10-100 chars) (default: random string)")

	err := Command.MarkFlagRequired("username")
	if err != nil {
		panic(err)
	}
	err = Command.MarkFlagRequired("public-url")
	if err != nil {
		panic(err)
	}
}

func getTwitchAPIClient(secret, eventSubCallbackURL string) (*twitch.APIClient, error) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID is not set")
	}

	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET is not set")
	}

	return twitch.NewAPIClient(eventSubCallbackURL, secret, clientID, clientSecret)
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livebuffer server for twitch",
	Long:  "Run the livebuffer server for twitch, which buffers the live stream and provides a public URL for it. Requires TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET environment variables to be set.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if eventSubSecretArg != "" {
			if len(eventSubSecretArg) < 10 || len(eventSubSecretArg) > 100 {
				return fmt.Errorf("eventsub-secret must be between 10 and 100 characters")
			}
			// Only ASCII characters are allowed in the eventsub-secret
			for _, c := range eventSubSecretArg {
				if c > 127 {
					return fmt.Errorf("eventsub-secret must only contain ASCII characters")
				}
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var bufferDirectory string
		if bufferDirectoryArg != "" {
			bufferDirectory = bufferDirectoryArg
			err := os.MkdirAll(bufferDirectory, 0755)
			if err != nil {
				return fmt.Errorf("failed to create buffer directory: %w", err)
			}
		} else {
			tmpDir, cleanup, err := fsutil.TemporaryDirectory()
			if err != nil {
				return err
			}
			defer cleanup()
			bufferDirectory = tmpDir
		}
		logging.Info("using buffer directory", "dir", bufferDirectory)

		var eventSubSecret string
		if eventSubSecretArg != "" {
			eventSubSecret = eventSubSecretArg
		} else {
			logging.Info("no eventsub-secret provided, generating random secret")
			eventSubSecret = stringutil.RandomString(32)
		}

		if before, ok := strings.CutSuffix(liveBufferPublicURL, "/"); ok {
			liveBufferPublicURL = before
		}
		eventSubURL := fmt.Sprintf("%s/api/v1/twitch-event-sub", liveBufferPublicURL)
		logging.Info("using eventsub callback URL", "url", eventSubURL)

		twitchAPI, err := getTwitchAPIClient(eventSubSecret, eventSubURL)
		if err != nil {
			return fmt.Errorf("failed to create twitch API client: %w", err)
		}
		defer funcutils.LogError(twitchAPI.Close, "failed to close twitch API client")

		userID, err := twitchAPI.GetUserID(username)
		if err != nil {
			return fmt.Errorf("failed to get user ID for username %s: %w", username, err)
		}

		twitchClient, err := twitch.NewClient(twitchAPI, userID)
		if err != nil {
			return err
		}

		director, err := buffer.NewDirector(maxStreams, bufferDirectory, username, twitchClient.OnlineChannel())
		if err != nil {
			return fmt.Errorf("failed to create director: %w", err)
		}
		defer funcutils.LogError(director.Close, "failed to close director")

		mux := api.GetMux(twitchAPI, director)

		err = twitchClient.StartEventSub()
		if err != nil {
			return fmt.Errorf("failed to start eventsub: %w", err)
		}

		addr := fmt.Sprintf("%s:%d", httpHost, httpPort)
		logging.Info("starting livebuffer server", "host", httpHost, "port", httpPort)
		err = http.ListenAndServe(
			addr,
			mux,
		)

		return fmt.Errorf("failed to start server: %w", err)
	},
}
