package shared

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/stream"
)

type ResponseMutedSegment struct {
	StartTimeMilliseconds int64 `json:"start_time_milliseconds"`
	EndTimeMilliseconds   int64 `json:"end_time_milliseconds"`
}

type ResponseStream struct {
	ID                   string                 `json:"id"`
	Title                string                 `json:"title"`
	Size                 string                 `json:"size"`
	SizeBytes            int64                  `json:"size_bytes"`
	Duration             string                 `json:"duration"`
	DurationMilliseconds int64                  `json:"duration_milliseconds"`
	StartedAt            time.Time              `json:"started_at"`
	Username             string                 `json:"username"`
	State                string                 `json:"state"`
	MutedSegments        []ResponseMutedSegment `json:"muted_segments"`
}

func formatMutedSegments(segments []stream.MutedSegment) []ResponseMutedSegment {
	formatted := make([]ResponseMutedSegment, len(segments))
	for i, segment := range segments {
		formatted[i] = ResponseMutedSegment{
			StartTimeMilliseconds: segment.Start.Milliseconds(),
			EndTimeMilliseconds:   segment.End.Milliseconds(),
		}
	}
	return formatted
}

func ResponseStreamFromDetails(details stream.Details) ResponseStream {
	title := details.Title
	if title == "" {
		title = "unknown title"
	}
	return ResponseStream{
		ID:                   details.ID,
		Title:                title,
		Size:                 humanize.Bytes(uint64(details.Size)),
		SizeBytes:            details.Size,
		Duration:             details.Duration.String(),
		DurationMilliseconds: details.Duration.Milliseconds(),
		StartedAt:            details.StartedAt,
		Username:             details.BroadcasterUserName,
		State:                string(details.StreamState),
		MutedSegments:        formatMutedSegments(details.MutedSegment),
	}
}
