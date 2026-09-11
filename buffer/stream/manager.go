package stream

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/matthiasharzer/livebuffer/hls"
)

type Manager struct {
	state           State
	meta            Metadata
	streamDirectory string
}

func NewManager(streamDirectory string, state State) (*Manager, error) {
	metadata, err := ReadMetadata(streamDirectory)
	if err != nil {
		return nil, err
	}

	return &Manager{
		state:           state,
		meta:            metadata,
		streamDirectory: streamDirectory,
	}, nil
}

func (m *Manager) StreamFilesDirectory() string {
	return FilesDirectory(m.streamDirectory)
}

func (m *Manager) StreamInfo() (Info, error) {
	hlsInfo, err := hls.Stat(FilesDirectory(m.streamDirectory))
	if err != nil {
		return Info{}, fmt.Errorf("failed to stat hls directory %s: %w", m.StreamFilesDirectory(), err)
	}

	return Info{
		ID:                  m.meta.ID,
		Title:               m.meta.Title,
		BroadcasterUserName: m.meta.BroadcasterUserName,
		StartedAt:           m.meta.StartedAt,
		Duration:            hlsInfo.Duration,
		StreamState:         m.state,
		Directory:           m.streamDirectory,
		Size:                hlsInfo.Size,
	}, nil
}
func (m *Manager) StreamID() string {
	return m.meta.ID
}
func (m *Manager) StreamFilePath() (string, error) {
	return "", nil
}

func (m *Manager) Reader(ctx context.Context) (io.ReadCloser, error) {
	reader, err := hls.NewReader(ctx, m.StreamFilesDirectory())
	if err != nil {
		return nil, fmt.Errorf("failed to create hls reader: %w", err)
	}
	return reader, nil
}

func (m *Manager) ClipReader(ctx context.Context, from time.Duration, to time.Duration) (io.ReadCloser, error) {
	return hls.NewClipReader(ctx, m.StreamFilesDirectory(), from, to)
}
