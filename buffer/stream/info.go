package stream

import (
	"time"
)

type StreamState string

const (
	StateArchived StreamState = "archived"
	StateLive     StreamState = "live"
)

type Info struct {
	ID                  string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
	Duration            time.Duration
	StreamState         StreamState
	Directory           string
	Size                int64
}

type ClipInfo struct {
	Stream    Info
	StartTime time.Duration
	EndTime   time.Duration
	Duration  time.Duration
}
