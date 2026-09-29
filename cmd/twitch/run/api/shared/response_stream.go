package shared

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/stream"
)

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

func ResponseStreamFromDetails(details stream.Details) ResponseStream {
	return ResponseStream{
		ID:                   details.ID,
		Title:                details.Title,
		Size:                 humanize.Bytes(uint64(details.Size)),
		SizeBytes:            details.Size,
		Duration:             details.Duration.String(),
		DurationMilliseconds: details.Duration.Milliseconds(),
		StartedAt:            details.StartedAt,
		Username:             details.BroadcasterUserName,
		State:                string(details.StreamState),
	}
}
