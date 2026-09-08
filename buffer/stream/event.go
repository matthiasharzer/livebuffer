package stream

import "time"

type WentLiveEvent struct {
	StreamID            string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
}
