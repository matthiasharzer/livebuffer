package stream

import (
	"fmt"
	"time"

	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
)

type StreamState string

const (
	StreamStateArchived StreamState = "archived"
	StreamStateLive     StreamState = "live"
)

type Info struct {
	ID                  string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
	Duration            time.Duration
	StreamState         StreamState
	FilePath            string
	Size                int64
}

type ClipInfo struct {
	Stream    Info
	StartTime time.Duration
	EndTime   time.Duration
	Duration  time.Duration
}

func BuildInfo(streamDirectory string, size int64, state StreamState) (Info, error) {
	metadata, err := ReadMetadata(streamDirectory)
	if err != nil {
		return Info{}, err
	}

	duration, err := ffmpegutil.GetDuration(File(streamDirectory))
	if err != nil {
		return Info{}, fmt.Errorf("failed to get duration of stream file: %w", err)
	}

	return Info{
		ID:                  metadata.ID,
		Title:               metadata.Title,
		BroadcasterUserName: metadata.BroadcasterUserName,
		StartedAt:           metadata.StartedAt,
		Duration:            duration,
		StreamState:         state,
		FilePath:            File(streamDirectory),
		Size:                size,
	}, nil
}
