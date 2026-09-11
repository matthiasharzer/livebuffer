package stream

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
)

type RecordingSession struct {
	streamID        string
	recorder        *twitch.Recorder
	hlsWriter       io.WriteCloser
	cancelRecording context.CancelFunc
	ctx             context.Context
}

func StartRecording(event WentLiveEvent, streamDirectory string) (*RecordingSession, error) {
	err := WriteMetadata(streamDirectory, Metadata{
		ID:                  event.StreamID,
		Title:               event.Title,
		BroadcasterUserName: event.BroadcasterUserName,
		StartedAt:           event.StartedAt,
	})
	if err != nil {
		return nil, err
	}

	streamFilesDirectory := FilesDirectory(streamDirectory)
	err = os.MkdirAll(streamFilesDirectory, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream files directory: %w", err)
	}

	recordingContext, cancel := context.WithCancel(context.Background())

	recorder, err := twitch.NewRecorder(event.BroadcasterUserName)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create twitch recorder: %w", err)
	}

	buffer, err := hls.NewWriter(recordingContext, streamFilesDirectory)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create hls writer: %w", err)
	}

	session := &RecordingSession{
		streamID:        event.StreamID,
		recorder:        recorder,
		ctx:             recordingContext,
		cancelRecording: cancel,
		hlsWriter:       buffer,
	}
	err = session.start()
	if err != nil {
		cancel()
		sessionCloseErr := session.Close()
		if sessionCloseErr != nil {
			logging.Warn("failed to close recording session after start error", "error", sessionCloseErr)
		}
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	return session, nil
}

func (rs *RecordingSession) StreamID() string {
	return rs.streamID
}

func (rs *RecordingSession) start() error {
	reader, err := rs.recorder.Record(rs.ctx)
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	go func() {
		_, err := io.Copy(rs.hlsWriter, reader)
		if err != nil {
			logging.Error("failed to write to video store", "error", err)
		}
		err = rs.recorder.WaitFinished()
		if err != nil {
			logging.Warn("streamlink command finished with error", "error", err)
		}
	}()

	return nil
}

func (rs *RecordingSession) Close() error {
	if rs.hlsWriter != nil {
		err := rs.hlsWriter.Close()
		if err != nil {
			return fmt.Errorf("failed to close hls writer: %w", err)
		}
	}
	if rs.cancelRecording != nil {
		rs.cancelRecording()
		rs.cancelRecording = nil
	}
	return nil
}
