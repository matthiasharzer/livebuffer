package buffer

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/matthiasharzer/livebuffer/logging"
)

type VideoFileBuffer struct {
	filePath string

	mu   sync.RWMutex
	size int64

	writeHandle *os.File
}

func NewVideoFileBuffer(filePath string) (*VideoFileBuffer, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	info, err := f.Stat()
	if err != nil {
		closeErr := f.Close()
		if closeErr != nil {
			logging.Error("failed to close file after stat error", "file", filePath, "error", closeErr)
		}
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	return &VideoFileBuffer{
		filePath:    filePath,
		writeHandle: f,
		size:        info.Size(),
	}, nil
}

func (vs *VideoFileBuffer) Write(p []byte) (int, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	n, err := vs.writeHandle.Write(p)
	vs.size += int64(n)
	return n, err
}

func (vs *VideoFileBuffer) Close() error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	return vs.writeHandle.Close()
}

func (vs *VideoFileBuffer) NewSnapshotReader() (io.ReadCloser, error) {
	// How much did we write to the file?
	vs.mu.RLock()
	currentSize := vs.size
	vs.mu.RUnlock()

	f, err := os.Open(vs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s for snapshot: %w", vs.filePath, err)
	}

	// Limit the reader to the current size of the file at the time of snapshot creation
	limitedReader := io.LimitReader(f, currentSize)

	return &readCloserWrapper{
		Reader: limitedReader,
		Closer: f, // Ensure calling Close() closes the underlying *os.File
	}, nil
}

type readCloserWrapper struct {
	io.Reader
	io.Closer
}
