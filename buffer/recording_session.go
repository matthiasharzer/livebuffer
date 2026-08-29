package buffer

import (
	"context"
	"fmt"
	"io"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

type recordingSession struct {
	recorder *twitch.Recorder
	buffer   *VideoFileBuffer
}

func newRecordingSession(username string, bufferFilePath string) (*recordingSession, error) {
	recorder, err := twitch.NewRecorder(username)
	if err != nil {
		return nil, fmt.Errorf("failed to create twitch recorder: %w", err)
	}

	buffer, err := NewVideoFileBuffer(bufferFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create video file buffer: %w", err)
	}

	return &recordingSession{
		recorder: recorder,
		buffer:   buffer,
	}, nil
}

func (rs *recordingSession) Start(ctx context.Context) error {
	reader, err := rs.recorder.Record(ctx)
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	go func() {
		_, err := io.Copy(rs.buffer, reader)
		if err != nil {
			logging.Error("failed to write to video store", "error", err)
		}
	}()

	return nil
}

func (rs *recordingSession) Close() error {
	if rs.buffer != nil {
		funcutils.LogError(rs.buffer.Close, "failed to close video buffer")
	}
	return nil
}

func (rs *recordingSession) FilePath() string {
	return rs.buffer.filePath
}
