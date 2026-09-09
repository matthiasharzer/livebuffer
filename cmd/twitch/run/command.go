package run

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/twitch/eventsub"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	"github.com/matthiasharzer/livebuffer/util/stringutil"
	"github.com/nicklaw5/helix/v2"
	"github.com/spf13/cobra"
)

var httpPort = 4000
var httpHost string
var usernames []string
var bufferDirectoryArg string
var liveBufferPublicURL string
var maxStreams = 2
var eventSubSecretArg string

func init() {
	Command.Flags().IntVarP(&httpPort, "port", "p", httpPort, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
	Command.Flags().StringSliceVarP(&usernames, "username", "u", []string{}, "Twitch username to buffer. Can be used multiple times (required)")
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

func getHelixClient() (*helix.Client, error) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID is not set")
	}

	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET is not set")
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	return helixClient, nil
}

func getUserContext(helixClient *helix.Client, eventSubClient *eventsub.Client, username string, bufferDirectory string) (userContext, error) {
	twitchClient, err := twitch.NewClient(helixClient, eventSubClient, username)
	if err != nil {
		return userContext{}, fmt.Errorf("failed to create twitch client: %w", err)
	}

	director, err := buffer.NewDirector(maxStreams, bufferDirectory, twitchClient.Username(), twitchClient.OnlineChannel())
	if err != nil {
		funcutils.LogError(twitchClient.Close, "failed to close twitch client on director initialization error")
		return userContext{}, fmt.Errorf("failed to create director: %w", err)
	}

	return userContext{
		username:     twitchClient.Username(),
		twitchClient: twitchClient,
		director:     director,
	}, nil
}

func getUserContexts(helixClient *helix.Client, eventSubClient *eventsub.Client, usernames []string, bufferDirectory string) ([]userContext, error) {
	var userContexts []userContext
	seen := make(map[string]bool)

	for _, usernameArg := range usernames {
		username := strings.ToLower(strings.TrimSpace(usernameArg))

		if username == "" {
			return nil, errors.New("twitch username cannot be empty")
		}
		_, isDuplicate := seen[username]
		if isDuplicate {
			return nil, fmt.Errorf("all provided twitch usernames must be unique. Found duplicated username %s", username)
		}
		context, err := getUserContext(helixClient, eventSubClient, username, bufferDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to create user context for %s: %w", username, err)
		}
		seen[username] = true
		userContexts = append(userContexts, context)
	}

	return userContexts, nil
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
		eventSubURL, err := url.Parse(fmt.Sprintf("%s/api/v1/twitch-event-sub", liveBufferPublicURL))
		if err != nil {
			return fmt.Errorf("failed to parse live buffer public URL: %w", err)
		}

		logging.Info("using eventsub callback URL", "url", eventSubURL.String())

		helixClient, err := getHelixClient()
		if err != nil {
			return err
		}
		eventSubClient := eventsub.NewClient(helixClient, *eventSubURL, eventSubSecret)

		userContexts, err := getUserContexts(helixClient, eventSubClient, usernames, bufferDirectory)
		if err != nil {
			return err
		}
		defer func() {
			for _, context := range userContexts {
				err := context.Close()
				if err != nil {
					logging.Warn("failed to close user context for", "username", context.username, "error", err)
				}
			}
		}()

		for _, context := range userContexts {
			err := context.twitchClient.StartEventSub()
			if err != nil {
				return fmt.Errorf("failed to start event sub for user %s: %w", context.username, err)
			}
		}

		mux := GetMux(userContexts, eventSubClient.HTTPHandler())

		addr := fmt.Sprintf("%s:%d", httpHost, httpPort)
		logging.Info("starting livebuffer server", "host", httpHost, "port", httpPort)
		err = http.ListenAndServe(
			addr,
			mux,
		)

		return fmt.Errorf("failed to start server: %w", err)
	},
}
