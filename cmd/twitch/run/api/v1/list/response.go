package list

import (
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"
)

type Response struct {
	Streams []shared.ResponseStream `json:"streams"`
}
