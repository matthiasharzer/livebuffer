package vod

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"slices"

	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/stream"
)

type BroadcasterName = string
type StreamID = string

type Repository struct {
	bufferDirectory string
}

func NewRepository(bufferDirectory string) *Repository {
	return &Repository{
		bufferDirectory: bufferDirectory,
	}
}

func (r *Repository) StreamDirectory(streamID string) string {
	return filepath.Join(r.bufferDirectory, streamID)
}

func (r *Repository) ReadStream(streamID string) (*stream.Manager, error) {
	streamDirectory := r.StreamDirectory(streamID)
	_, err := os.Stat(streamDirectory)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if !stream.IsStreamDirectory(streamDirectory) {
		return nil, nil
	}

	manager, err := stream.NewManager(streamDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream manager: %w", err)
	}
	return manager, nil
}

func (r *Repository) ReadAllStreams() iter.Seq2[stream.Manager, error] {
	return func(yield func(stream.Manager, error) bool) {
		dirEntries, err := os.ReadDir(r.bufferDirectory)
		if err != nil {
			yield(stream.Manager{}, fmt.Errorf("failed to list streams: %w", err))
			return
		}

		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}
			streamID := entry.Name()

			manager, err := r.ReadStream(streamID)
			if err != nil {
				yield(stream.Manager{}, fmt.Errorf("failed to read stream metadata: %w", err))
				return
			}
			if manager == nil {
				continue
			}

			if !yield(*manager, nil) {
				return
			}
		}
	}
}

func (r *Repository) ReadStreams(filterFunc filter.Func) iter.Seq2[stream.Manager, error] {
	return filter.Apply(r.ReadAllStreams(), filterFunc)
}

func (r *Repository) GetStreamsSortedByStartTime(filterFunc filter.Func) ([]stream.Manager, error) {
	var streams []stream.Manager
	for streamInfo, err := range r.ReadStreams(filterFunc) {
		if err != nil {
			return nil, fmt.Errorf("failed to read stream infos: %w", err)
		}
		streams = append(streams, streamInfo)
	}

	slices.SortStableFunc(streams, func(a, b stream.Manager) int {
		return a.Meta().StartedAt.Compare(b.Meta().StartedAt)
	})

	return streams, nil
}

func (r *Repository) Close() error {
	//return d.state.Close()
	return nil
}
