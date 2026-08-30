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
	streamlinkCmd := exec.CommandContext(ctx, "streamlink", "--stdout", "twitch.tv/"+r.username, "best")

	reader, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create streamlink stdout pipe: %w", err)
	}

	err = streamlinkCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start streamlink command: %w", err)
	}

	go func() {
		err := streamlinkCmd.Wait()
		if err != nil {
			fmt.Printf("streamlink command exited with error: %v\n", err)
		}
	}()

	return reader, nil
}
