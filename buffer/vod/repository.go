package vod

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"

	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/logging"
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

func (r *Repository) StreamFilesDirectory(streamID string) string {
	return stream.FilesDirectory(r.StreamDirectory(streamID))
}

// ReadStream reads the stream manager with the given streamID, if the given streamID resolves to a valid stream
// directory. Will return nil if the directory is not a stream directory. Errors are critical only
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

// ReadAllStreams reads all known streams from the buffer directory. Non-stream directories or files will be skipped
func (r *Repository) ReadAllStreams() iter.Seq2[stream.Manager, error] {
	return func(yield func(stream.Manager, error) bool) {
		dirEntries, err := os.ReadDir(r.bufferDirectory)
		if err != nil {
			yield(stream.Manager{}, fmt.Errorf("failed to list streams: %w", err))
			return
		}

		for _, entry := range dirEntries {
			if !entry.IsDir() {
				logging.Warn("found non-directory entry in buffer directory (skipping)", "entry_name", entry.Name())
				continue
			}
			streamID := entry.Name()

			manager, err := r.ReadStream(streamID)
			if err != nil {
				if !yield(stream.Manager{}, fmt.Errorf("failed to read stream metadata: %w", err)) {
					return
				}
				continue
			}
			if manager == nil {
				logging.Warn("found directory that is not a stream directory (skipping)", "directory_name", entry.Name())
				continue
			}
			if manager.StreamID() != entry.Name() {
				logging.Warn("found stream in unexpected directory (skipping)", "stream_id", manager.StreamID(), "directory_name", entry.Name())
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

// ReadAllBufferEntryPaths reads all directory entries and returns the path to the entry, even when it does not
// is a valid stream directory
func (r *Repository) ReadAllBufferEntryPaths() ([]string, error) {
	dirEntries, err := os.ReadDir(r.bufferDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to stats directory: %w", err)
	}

	var paths []string
	for _, entry := range dirEntries {
		paths = append(paths, r.StreamDirectory(entry.Name()))
	}
	return paths, nil
}

func (r *Repository) Close() error {
	//return d.state.Close()
	return nil
}
