package stream

import "time"

type WentLiveEvent struct {
	StreamID             string
	Title                string
	BroadcasterUserLogin string
	StartedAt            time.Time
}
