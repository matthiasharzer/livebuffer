package stream

import (
	"time"
)

type State string

const (
	StateArchived State = "archived"
	StateLive     State = "live"
)

type Info struct {
	ID                  string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
	Duration            time.Duration
	StreamState         State
	Directory           string
	Size                int64
}

type ClipInfo struct {
	Stream    Info
	StartTime time.Duration
	EndTime   time.Duration
	Duration  time.Duration
}
