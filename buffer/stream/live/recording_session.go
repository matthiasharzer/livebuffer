package live

import (
	"context"
	"fmt"
	"io"

	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/twitch"
)

type recordingSession struct {
	recorder *twitch.Recorder
	//buffer   *hlsFileWriter
	hlsWriter io.WriteCloser
}

func newRecordingSession(ctx context.Context, username string, streamRecordingDirectory string) (*recordingSession, error) {
	recorder, err := twitch.NewRecorder(username)
	if err != nil {
		return nil, fmt.Errorf("failed to create twitch recorder: %w", err)
	}

	buffer, err := hls.NewWriter(ctx, streamRecordingDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create hls writer: %w", err)
	}

	return &recordingSession{
		recorder:  recorder,
		hlsWriter: buffer,
	}, nil
}

func (rs *recordingSession) Start(ctx context.Context, broadcaster *broadcastWriter) error {
	reader, err := rs.recorder.Record(ctx)
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	go func() {
		target := io.MultiWriter(rs.hlsWriter, broadcaster)
		_, err := io.Copy(target, reader)
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

func (rs *recordingSession) Close() error {
	if rs.hlsWriter != nil {
		err := rs.hlsWriter.Close()
		if err != nil {
			return fmt.Errorf("failed to close hls writer: %w", err)
		}
	}
	return nil
}

func (rs *recordingSession) FilePath() string {
	// TODO: verify if this is correct
	//return rs.hlsWriter.filePath
	return "hls.IndexFilePath()"
}
