package run

import (
	"fmt"
	"net/http"
	"os"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/spf13/cobra"
)

var httpPort int
var httpHost string

func init() {
	Command.Flags().IntVarP(&httpPort, "port", "p", 4000, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
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
	RunE: func(cmd *cobra.Command, args []string) error {
		username := "pokemonxp"

		bufferDirectory := ".buffer"
		err := os.MkdirAll(bufferDirectory, 0777)
		if err != nil {
			return err
		}

		twitchAPI, err := getTwitchAPIClient("secret-dev", "http://localhost:4000/api/v1/twitch-event-sub")

		userID, err := twitchAPI.GetUserID(username)
		if err != nil {
			return err
		}

		twitchClient, err := twitch.NewClient(twitchAPI, userID)
		if err != nil {
			return err
		}

		director, err := buffer.NewDirector(2, bufferDirectory, username, twitchClient.OnlineChannel())
		if err != nil {
			return err
		}

		_ = director

		mux := api.GetMux(twitchAPI, director)

		addr := fmt.Sprintf("%s:%d", httpHost, httpPort)
		logging.Info("starting livebuffer server", "host", httpHost, "port", httpPort)
		err = http.ListenAndServe(
			addr,
			mux,
		)

		return fmt.Errorf("failed to start server: %w", err)

		//director.WentLive()
		//defer director.Close()
		//
		//fmt.Printf("Waiting for streams to be available...\n")
		//
		//time.Sleep(10 * time.Second)
		//

		return nil
	},
}
