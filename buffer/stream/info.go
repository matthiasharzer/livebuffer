package stream

import (
	"time"
)

type State string

type Details struct {
	ID                  string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
	Duration            time.Duration
	Directory           string
	Size                int64
}

type ClipInfo struct {
	StreamDetails Details
	StartTime     time.Duration
	EndTime       time.Duration
	Duration      time.Duration
}
