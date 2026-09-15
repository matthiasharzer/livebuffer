package stream

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/matthiasharzer/livebuffer/hls"
)

type Manager struct {
	StreamDirectory string
	meta            Metadata
}

func NewManager(streamDirectory string) (*Manager, error) {
	metadata, err := ReadMetadata(streamDirectory)
	if err != nil {
		return nil, err
	}

	return &Manager{
		meta:            metadata,
		StreamDirectory: streamDirectory,
	}, nil
}

func (m *Manager) StreamFilesDirectory() string {
	return FilesDirectory(m.StreamDirectory)
}

func (m *Manager) Meta() Metadata {
	return m.meta
}

func (m *Manager) GetDetails(state State) (Details, error) {
	hlsInfo, err := hls.Stat(m.StreamFilesDirectory())
	if err != nil {
		return Details{}, fmt.Errorf("failed to stat hls directory %s: %w", m.StreamFilesDirectory(), err)
	}

	return Details{
		ID:                  m.meta.ID,
		Title:               m.meta.Title,
		BroadcasterUserName: m.meta.BroadcasterUserName,
		StartedAt:           m.meta.StartedAt,
		Duration:            hlsInfo.Duration,
		Directory:           m.StreamDirectory,
		Size:                hlsInfo.Size,
		StreamState:         state,
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
