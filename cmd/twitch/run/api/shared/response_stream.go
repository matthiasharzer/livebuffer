package shared

import "time"

type ResponseStream struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Size                 string    `json:"size"`
	SizeBytes            int64     `json:"size_bytes"`
	Duration             string    `json:"duration"`
	DurationMilliseconds int64     `json:"duration_milliseconds"`
	StartedAt            time.Time `json:"started_at"`
	Username             string    `json:"username"`
	State                string    `json:"state"`
}
