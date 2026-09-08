package live

import (
	"context"
	"fmt"
	"io"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

type StreamManager struct {
	id              string
	streamDirectory string
	cancelRecording context.CancelFunc
	session         *recordingSession
}

func NewRecordingStreamManager(ctx context.Context, event stream.WentLiveEvent, username string, streamDirectory string) (*StreamManager, error) {
	id := fmt.Sprintf("%s_%s", username, event.StartedAt.Format("20060102_150405"))
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

func (sm *StreamManager) StreamID() string {
	return sm.id
}

func (sm *StreamManager) Reader() (io.ReadCloser, int64, error) {
	return sm.session.buffer.NewSnapshotReader()
}

func (sm *StreamManager) StreamFilePath() (string, error) {
	return stream.File(sm.streamDirectory), nil
}

func (sm *StreamManager) Close() error {
	if sm.cancelRecording != nil {
		sm.cancelRecording()
	}
	if sm.session != nil {
		return sm.session.Close()
	}
	return nil
}
