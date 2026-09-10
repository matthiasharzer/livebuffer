package archived

import (
	"fmt"
	"io"
	"os"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

type StreamManager struct {
	ID              string
	streamDirectory string
}

func NewStreamManager(streamDirectory string) (stream.Manager, error) {
	metadata, err := stream.ReadMetadata(streamDirectory)
	if err != nil {
		return nil, err
	}

	return &StreamManager{
		ID:              metadata.ID,
		streamDirectory: streamDirectory,
	}, nil
}

func (sm *StreamManager) fileSize() (int64, error) {
	fileInfo, err := os.Stat(stream.File(sm.streamDirectory))
	if err != nil {
		return 0, fmt.Errorf("failed to stat stream file: %w", err)
	}
	return fileInfo.Size(), nil
}

func (sm *StreamManager) StreamInfo() (stream.Info, error) {
	size, err := sm.fileSize()
	if err != nil {
		return stream.Info{}, fmt.Errorf("failed to get stream file size: %w", err)
	}

	return stream.BuildInfo(sm.streamDirectory, size, stream.StreamStateArchived)
}

func (sm *StreamManager) StreamID() string {
	return sm.ID
}

func (sm *StreamManager) Reader() (io.ReadSeekCloser, int64, error) {
	size, err := sm.fileSize()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get stream file size: %w", err)
	}

	file, err := os.Open(stream.File(sm.streamDirectory))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open stream file: %w", err)
	}

	return file, size, nil
}

func (sm *StreamManager) StreamFilePath() (string, error) {
	return stream.File(sm.streamDirectory), nil
}
