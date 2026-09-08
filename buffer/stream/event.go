package stream

import "time"

type WentLiveEvent struct {
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
}
