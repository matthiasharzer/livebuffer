package hls

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
)

type writer struct {
	directory string
	cmd       *exec.Cmd
	pipe      io.WriteCloser
}

func NewWriter(ctx context.Context, directory string) (io.WriteCloser, error) {
	segmentFilepath := filepath.Join(directory, ChunkFilename)
	playlistFilepath := filepath.Join(directory, IndexFilename)

	args := []string{
		"-i", "pipe:0",
		"-c", "copy",

		// HLS options
		"-f", "hls",
		"-hls_time", strconv.Itoa(ChunkSizeSeconds),
		"-hls_list_size", "0",
		"-hls_playlist_type", "event",
		"-hls_segment_filename", segmentFilepath,
		"-hls_flags", "delete_segments+append_list",
		playlistFilepath,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create ffmpeg pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		_ = cmdStdin.Close()
		return nil, fmt.Errorf("failed to start ffmpeg command: %w", err)
	}

	return &writer{
		directory: directory,
		cmd:       cmd,
		pipe:      cmdStdin,
	}, nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.pipe.Write(p)
}

func (w *writer) Close() error {
	pipeErr := w.pipe.Close()

	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}

	waitErr := w.cmd.Wait()
	if pipeErr != nil {
		return pipeErr
	}
	return waitErr
}
