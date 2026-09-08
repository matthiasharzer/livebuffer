package live

import (
	"context"
	"errors"
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
	broadcaster     *broadcastWriter
}

func cleanupOnFailure(streamDirectory string, session *recordingSession, broadcaster *broadcastWriter) {
	metadataRemovalErr := os.Remove(stream.MetadataFile(streamDirectory))
	if metadataRemovalErr != nil {
		logging.Warn("failed to clean up metadata file after session creation error", "metadataFile", stream.MetadataFile(streamDirectory), "error", metadataRemovalErr)
	}
	if session != nil {
		sessionCloseErr := session.Close()
		if sessionCloseErr != nil {
			logging.Warn("failed to close recording session after start error", "error", sessionCloseErr)
		}
	}
	if broadcaster != nil {
		broadcastCleanupErr := broadcaster.Close()
		if broadcastCleanupErr != nil {
			logging.Warn("failed to clean up broadcaster after session creation error", "error", broadcastCleanupErr)
		}
	}
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

	broadcaster := newBroadcastWriter()
	session, err := newRecordingSession(event.BroadcasterUserName, stream.File(streamDirectory))
	if err != nil {
		cleanupOnFailure(streamDirectory, session, broadcaster)
		return nil, err
	}
	recordingContext, cancel := context.WithCancel(ctx)
	err = session.Start(recordingContext, broadcaster)
	if err != nil {
		cancel()
		cleanupOnFailure(streamDirectory, session, broadcaster)
		return nil, err
	}

	return &StreamManager{
		id:              event.StreamID,
		streamDirectory: streamDirectory,
		cancelRecording: cancel,
		session:         session,
		broadcaster:     broadcaster,
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

func (sm *StreamManager) LiveSubscribe(clientChan chan []byte) {
	sm.broadcaster.AddClient(clientChan)
}

func (sm *StreamManager) LiveUnsubscribe(clientChan chan []byte) {
	sm.broadcaster.RemoveClient(clientChan)
}

func (sm *StreamManager) Close() error {
	if sm.cancelRecording != nil {
		sm.cancelRecording()
	}
	var errs []error
	if sm.session != nil {
		err := sm.session.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if sm.broadcaster != nil {
		err := sm.broadcaster.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
