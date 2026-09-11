package twitch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
)

type Recorder struct {
	username      string
	streamlinkCmd *exec.Cmd
	mu            *sync.Mutex
}

func isStreamlinkInstalled() bool {
	command := exec.Command("streamlink", "--version")
	err := command.Run()
	return err == nil
}

func cleanupOnRecordingError(ffmpegCmd *exec.Cmd, ffmpegStdin io.ReadCloser, ffmpegStdout io.ReadCloser) {
	if ffmpegStdin != nil {
		closeErr := ffmpegStdin.Close()
		if closeErr != nil {
			logging.Warn("fialed to close ffmpeg stdin reader", "error", closeErr)
		}
	}
	if ffmpegStdout != nil {
		closeErr := ffmpegStdout.Close()
		if closeErr != nil {
			logging.Warn("failed to close ffmpeg stdout reader", "error", closeErr)
		}
	}
	if ffmpegCmd != nil {
		killErr := ffmpegCmd.Process.Kill()
		if killErr != nil {
			logging.Warn("failed to kill ffmpeg process after streamlink startup failure", "error", killErr)
		} else {
			waitErr := ffmpegCmd.Wait()
			if waitErr != nil {
				logging.Warn("failed to wait for ffmpeg process to finish after streamlink startup failure", "error", waitErr)
			}
		}
	}
}

func NewRecorder(username string) (*Recorder, error) {
	if !isStreamlinkInstalled() {
		return nil, errors.New("streamlink is not installed. Please install streamlink to use the recorder")
	}
	if !ffmpegutil.IsInstalled() {
		return nil, errors.New("ffmpeg is not installed. Please install ffmpeg to use the recorder")
	}
	return &Recorder{
		username: username,
		mu:       &sync.Mutex{},
	}, nil
}

func (r *Recorder) Record(ctx context.Context) (io.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	streamlinkArgs := []string{
		"--ffmpeg-copyts",
		"--ffmpeg-start-at-zero",
		"--stdout",
		"twitch.tv/" + r.username,
		"best",
	}
	streamlinkCmd := exec.CommandContext(ctx, "streamlink", streamlinkArgs...)

	streamlinkStdout, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create streamlink stdout pipe: %w", err)
	}

	err = streamlinkCmd.Start()
	if err != nil {
		closeErr := streamlinkStdout.Close()
		if closeErr != nil {
			logging.Warn("failed to close stdout reader", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to start streamlink command: %w", err)
	}
	r.streamlinkCmd = streamlinkCmd

	return streamlinkStdout, nil
}

func (r *Recorder) WaitFinished() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	if r.streamlinkCmd != nil {
		err := r.streamlinkCmd.Wait()
		if err != nil {
			errs = append(errs, fmt.Errorf("streamlink command failed: %w", err))
		}
	}
	r.streamlinkCmd = nil
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
