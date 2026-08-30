package twitch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type Recorder struct {
	username string
	cmd      *exec.Cmd
	mu       *sync.Mutex
}

func isStreamlinkInstalled() bool {
	command := exec.Command("streamlink", "--version")
	err := command.Run()
	return err == nil
}

func NewRecorder(username string) (*Recorder, error) {
	if !isStreamlinkInstalled() {
		return nil, errors.New("streamlink is not installed. Please install streamlink to use the recorder")
	}
	return &Recorder{
		username: username,
		mu:       &sync.Mutex{},
	}, nil
}

func (r *Recorder) Record(ctx context.Context) (io.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	streamlinkCmd := exec.CommandContext(ctx, "streamlink", "--stdout", "twitch.tv/"+r.username, "best")

	reader, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create streamlink stdout pipe: %w", err)
	}

	err = streamlinkCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start streamlink command: %w", err)
	}
	r.cmd = streamlinkCmd

	return reader, nil
}

func (r *Recorder) WaitFinished() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return errors.New("streamlink command is not running")
	}
	err := r.cmd.Wait()
	if err != nil {
		return fmt.Errorf("streamlink command failed: %w", err)
	}
	return nil
}
