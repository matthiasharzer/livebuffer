package hls

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

func createSnapshot(directory string) (string, func(), error) {
	playlistFilepath := filepath.Join(directory, indexFilename)

	playlistBytes, err := os.ReadFile(playlistFilepath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read playlist file %s: %w", playlistFilepath, err)
	}
	tmpFile, err := os.CreateTemp(directory, "snapshot_*.m3u8")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create hls snapshot: %w", err)
	}
	defer funcutils.LogError(tmpFile.Close, "failed to close snapshot file")
	tmpFileName := tmpFile.Name()

	_, err = tmpFile.Write(playlistBytes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write to snapshot paylist: %w", err)
	}
	_, err = tmpFile.WriteString("\n#EXT-X-ENDLIST\n")
	if err != nil {
		return "", nil, fmt.Errorf("failed to write to snapshot paylist: %w", err)
	}

	return tmpFileName, func() {
		_ = os.Remove(tmpFileName)
	}, nil
}

type reader struct {
	cmd     *exec.Cmd
	pipe    io.ReadCloser
	cleanup func()
}

func createReader(ctx context.Context, args []string, cleanup func()) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create ffmpeg pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		_ = cmdStdout.Close()
		return nil, fmt.Errorf("failed to start ffmpeg command: %w", err)
	}

	return &reader{
		cmd:     cmd,
		pipe:    cmdStdout,
		cleanup: cleanup,
	}, nil
}

func NewReader(ctx context.Context, directory string) (io.ReadCloser, error) {
	snapshotFile, cleanup, err := createSnapshot(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create hls snapshot: %w", err)
	}

	args := []string{
		"-i", snapshotFile,
		"-c", "copy",
		"-f", "mpegts",
		"pipe:1",
	}
	return createReader(ctx, args, cleanup)
	//cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	//
	//cmdStdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create ffmpeg pipe: %w", err)
	//}
	//
	//err = cmd.Start()
	//if err != nil {
	//	_ = cmdStdout.Close()
	//	return nil, fmt.Errorf("failed to start ffmpeg command: %w", err)
	//}
	//
	//return &reader{
	//	cmd:     cmd,
	//	pipe:    cmdStdout,
	//	cleanup: cleanup,
	//}, nil
}

func NewClipReader(ctx context.Context, directory string, from time.Duration, to time.Duration) (io.ReadCloser, error) {
	snapshotFile, cleanup, err := createSnapshot(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create hls snapshot: %w", err)
	}
	startStr := fmt.Sprintf("%.3f", from.Seconds())
	durationStr := fmt.Sprintf("%.3f", to.Seconds()-from.Seconds())

	args := []string{
		"-i", snapshotFile,
		"-ss", startStr,
		"-t", durationStr,
		"-c", "copy",
		"-f", "mpegts",
		"pipe:1",
	}
	return createReader(ctx, args, cleanup)
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.pipe.Read(p)
}

func (r *reader) Close() error {
	r.cleanup()
	pipeErr := r.pipe.Close()

	if r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
	}

	waitErr := r.cmd.Wait()
	if pipeErr != nil {
		return pipeErr
	}
	return waitErr
}
