package archived

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

type StreamManager struct {
	streamDirectory string
}

func NewStreamManager(streamDirectory string) (stream.Manager, error) {
	return &StreamManager{
		streamDirectory: streamDirectory,
	}, nil
}

func (sm *StreamManager) StreamInfo() (stream.Info, error) {
	fileInfo, err := os.Stat(stream.File(sm.streamDirectory))
	if err != nil {
		return stream.Info{}, fmt.Errorf("failed to stat stream file: %w", err)
	}
	size := fileInfo.Size()

	return stream.BuildInfo(sm.streamDirectory, size, stream.StreamStateArchived)
}

func (sm *StreamManager) StreamID() (string, error) {
	metadata, err := stream.ReadMetadata(sm.streamDirectory)
	if err != nil {
		return "", err
	}
	return metadata.ID, nil
}

func (sm *StreamManager) Close() error {
	return nil
}
