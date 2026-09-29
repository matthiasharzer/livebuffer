package broadcasters

import "github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"

type ResponseBroadcaster struct {
	Username string                  `json:"username"`
	Streams  []shared.ResponseStream `json:"streams"`
}

type Response struct {
	Broadcasters []ResponseBroadcaster `json:"broadcasters"`
}
