package list

import "time"

type Response struct {
	Streams []ResponseStream `json:"streams"`
}

type ResponseStream struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Size                 string    `json:"size"`
	SizeBytes            int64     `json:"size_bytes"`
	Duration             string    `json:"duration"`
	DurationMilliseconds int64     `json:"duration_milliseconds"`
	StartedAt            time.Time `json:"started_at"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	StreamState          string    `json:"stream_state"`
}
