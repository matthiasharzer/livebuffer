package live

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/hls"
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
		logging.Warn("failed to clean up metadata file after session creation error", "metadata_file", stream.MetadataFile(streamDirectory), "error", metadataRemovalErr)
	}
	filesRemovalErr := os.RemoveAll(stream.FilesDirectory(streamDirectory))
	if filesRemovalErr != nil {
		logging.Warn("failed to clean up stream files directory after session creation error", "stream_files_directory", stream.FilesDirectory(streamDirectory), "error", filesRemovalErr)
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

	streamFilesDirectory := stream.FilesDirectory(streamDirectory)
	err = os.MkdirAll(streamFilesDirectory, 0755)
	if err != nil {
		cleanupOnFailure(streamDirectory, nil, nil)
		return nil, fmt.Errorf("failed to create stream files directory: %w", err)
	}

	recordingContext, cancel := context.WithCancel(ctx)

	broadcaster := newBroadcastWriter()
	session, err := newRecordingSession(recordingContext, event.BroadcasterUserName, streamFilesDirectory)
	if err != nil {
		cancel()
		cleanupOnFailure(streamDirectory, session, broadcaster)
		return nil, err
	}
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

func (sm *StreamManager) filesDirectory() string {
	return stream.FilesDirectory(sm.streamDirectory)
}

func (sm *StreamManager) StreamInfo() (stream.Info, error) {
	hlsInfo, err := hls.Stat(stream.FilesDirectory(sm.streamDirectory))
	if err != nil {
		return stream.Info{}, fmt.Errorf("failed to stat hls directory %s: %w", stream.FilesDirectory(sm.streamDirectory), err)
	}
	return stream.BuildInfo(sm.streamDirectory, hlsInfo.Size, stream.StreamStateLive)
}

func (sm *StreamManager) StreamID() string {
	return sm.id
}

func (sm *StreamManager) Reader() (io.ReadCloser, int64, error) {
	// TODO: move outside
	ctx := context.Background()
	reader, err := hls.NewReader(ctx, sm.filesDirectory())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create hls reader: %w", err)
	}
	// TODO: fix size (remove?)
	return reader, 0, nil
	//return sm.session.buffer.NewSnapshotReader()
}

func (sm *StreamManager) ClipReader(from time.Duration, to time.Duration) (io.ReadCloser, error) {
	return nil, nil
}

func (sm *StreamManager) StreamFilePath() (string, error) {
	return stream.File(sm.streamDirectory), nil
}

func (sm *StreamManager) LiveSubscribe(clientChan chan []byte) {
	if sm.broadcaster == nil {
		close(clientChan)
		return
	}
	sm.broadcaster.AddClient(clientChan)
}

func (sm *StreamManager) LiveUnsubscribe(clientChan chan []byte) {
	if sm.broadcaster == nil {
		close(clientChan)
		return
	}
	sm.broadcaster.RemoveClient(clientChan)
}

func (sm *StreamManager) Close() error {
	if sm.cancelRecording != nil {
		sm.cancelRecording()
		sm.cancelRecording = nil
	}
	var errs []error
	if sm.session != nil {
		err := sm.session.Close()
		if err != nil {
			errs = append(errs, err)
		}
		sm.session = nil
	}
	if sm.broadcaster != nil {
		err := sm.broadcaster.Close()
		if err != nil {
			errs = append(errs, err)
		}
		sm.broadcaster = nil
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
