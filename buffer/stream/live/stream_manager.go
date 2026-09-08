package live

import (
	"context"
	"fmt"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

type StreamManager struct {
	id              string
	streamDirectory string
	cancelRecording context.CancelFunc
	session         *recordingSession
}

func NewRecordingStreamManager(ctx context.Context, event stream.WentLiveEvent, username string, streamDirectory string) (stream.Manager, error) {
	id := fmt.Sprintf("%s_%s", event.BroadcasterUserName, event.StartedAt.Format("20060102_150405"))
	err := stream.WriteMetadata(streamDirectory, stream.Metadata{
		ID:                  id,
		Title:               event.Title,
		BroadcasterUserName: username,
		StartedAt:           event.StartedAt,
	})
	if err != nil {
		return nil, err
	}

	session, err := newRecordingSession(username, stream.File(streamDirectory))
	if err != nil {
		return nil, err
	}
	recordingContext, cancel := context.WithCancel(ctx)
	err = session.Start(recordingContext)
	if err != nil {
		cancel()
		return nil, err
	}

	return &StreamManager{
		id:              id,
		streamDirectory: streamDirectory,
		cancelRecording: cancel,
		session:         session,
	}, nil
}

func (sm *StreamManager) StreamInfo() (stream.Info, error) {
	size := sm.session.buffer.size
	return stream.BuildInfo(sm.streamDirectory, size, stream.StreamStateLive)
}

func (sm *StreamManager) StreamID() (string, error) {
	return sm.id, nil
}

func (sm *StreamManager) Close() error {
	if sm.session != nil {
		return sm.session.Close()
	}
	if sm.cancelRecording != nil {
		sm.cancelRecording()
	}
	return nil
}
