package run

import (
	"fmt"
	"net/http"
	"os"

	"github.com/matthiasharzer/livebuffer/buffer"
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

		bufferDirectory := ".buffer"
		err := os.MkdirAll(bufferDirectory, 0777)
		if err != nil {
			return err
		}

		director, err := buffer.NewDirector(2, bufferDirectory, "lars_tm")
		if err != nil {
			return err
		}

		_ = director

		twitchAPI, err := getTwitchAPIClient("secret-dev", "http://localhost:4000/api/v1/twitch-event-sub")

		userID, err := twitchAPI.GetUserID("lars_tm")
		if err != nil {
			return err
		}

		twitchClient, err := twitch.NewClient(twitchAPI, userID)
		if err != nil {
			return err
		}

		_ = twitchClient

		twitchClient.OnlineChannel().Subscribe(func(data twitch.StreamOnlineState) {
			logging.Info("stream online state changed", "is_online", data.IsOnline, "started_at", data.StartedAt)

		})

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		mux.Handle("POST /api/v1/twitch-event-sub", twitchAPI.EventSubHTTPHandler())

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
		//streams, err := director.GetStreams()
		//if err != nil {
		//	return err
		//}
		//fmt.Printf("Streams: %v\n", streams)
		//
		//if len(streams) == 0 {
		//	return fmt.Errorf("no streams found")
		//}
		//
		//fmt.Printf("Downloading stream: %s\n", streams[0])
		//
		//reader, err := director.GetStream(streams[0])
		//if err != nil {
		//	return err
		//}
		//defer reader.Close()
		//
		//fmt.Printf("Writing stream to output.ts\n")
		//
		//fi, err := os.Create("output.ts")
		//if err != nil {
		//	return err
		//}
		//defer fi.Close()
		//
		//fmt.Printf("Copying stream to output.ts\n")
		//
		//_, err = io.Copy(fi, reader)
		//if err != nil {
		//	return err
		//}
		//
		//fmt.Printf("Finished writing stream to output.ts\n")

		return nil
	},
}
