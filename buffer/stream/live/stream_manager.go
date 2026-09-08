package live

import (
	"context"
	"io"
	"os"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/logging"
)

type StreamManager struct {
	id              string
	streamDirectory string
	cancelRecording context.CancelFunc
	session         *recordingSession
}

func NewRecordingStreamManager(ctx context.Context, event stream.WentLiveEvent, streamDirectory string) (*StreamManager, error) {
	err := stream.WriteMetadata(streamDirectory, stream.Metadata{
		ID:                  event.StreamID,
		Title:               event.Title,
		BroadcasterUserName: event.BroadcasterUserName,
		StartedAt:           event.StartedAt,
	})
	if err != nil {
		return nil, err
	}

	session, err := newRecordingSession(event.BroadcasterUserName, stream.File(streamDirectory))
	if err != nil {
		cleanupErr := os.Remove(stream.MetadataFile(streamDirectory))
		if cleanupErr != nil {
			logging.Warn("failed to clean up metadata file after session creation error", "metadataFile", stream.MetadataFile(streamDirectory), "error", cleanupErr)
		}
		return nil, err
	}
	recordingContext, cancel := context.WithCancel(ctx)
	err = session.Start(recordingContext)
	if err != nil {
		cancel()
		cleanupErr := os.Remove(stream.MetadataFile(streamDirectory))
		if cleanupErr != nil {
			logging.Warn("failed to clean up metadata file after session creation error", "metadataFile", stream.MetadataFile(streamDirectory), "error", cleanupErr)
		}
		return nil, err
	}

	return &StreamManager{
		id:              event.StreamID,
		streamDirectory: streamDirectory,
		cancelRecording: cancel,
		session:         session,
	}, nil
}

func (sm *StreamManager) StreamInfo() (stream.Info, error) {
	size := sm.session.buffer.Size()
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
