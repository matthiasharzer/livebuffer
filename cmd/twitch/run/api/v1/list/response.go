package list

import "time"

type Response struct {
	Streams []ResponseStream `json:"streams"`
}

type ResponseStream struct {
	ID                  string    `json:"id"`
	Title               string    `json:"name"`
	Size                int64     `json:"size"`
	Duration            string    `json:"duration"`
	StartedAt           time.Time `json:"started_at"`
	BroadcasterUserName string    `json:"broadcaster_user_name"`
	StreamState         string    `json:"stream_state"`
}
