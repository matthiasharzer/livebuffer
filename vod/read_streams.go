package vod

import (
	"fmt"
	"iter"
	"os"
	"slices"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/vod/filter"
)

type streamCommon struct {
	manager stream.Manager
	details stream.Details
}

func (r *Repository) readStream(streamID string) (*stream.Manager, error) {
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

//func (d *Director) readMeta(streamID string) (*stream.Metadata, error) {
//	streamDirectory := d.streamDirectory(streamID)
//	meta, err := stream.ReadMetadata(streamDirectory)
//	if os.IsNotExist(err) {
//		return nil, nil
//	} else if err != nil {
//		return nil, err
//	}
//	return &meta, err
//}

func (r *Repository) readAllStreams() iter.Seq2[stream.Manager, error] {
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

			manager, err := r.readStream(streamID)
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

func (r *Repository) readStreams(filterFunc filter.Func) iter.Seq2[stream.Manager, error] {
	return filter.Apply(r.readAllStreams(), filterFunc)
}

func (r *Repository) getStreamsSortedByStartTime(filterFunc filter.Func) ([]stream.Manager, error) {
	var streams []stream.Manager
	for streamInfo, err := range r.readStreams(filterFunc) {
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

func (r *Repository) getStreamCommon(streamID string) (*streamCommon, error) {
	manager, err := r.readStream(streamID)
	if err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, nil
	}
	details, err := manager.GetDetails()
	if err != nil {
		return nil, err
	}
	return &streamCommon{
		manager: *manager,
		details: details,
	}, nil
}

func (r *Repository) getStreamDetails(filterFunc filter.Func) ([]stream.Details, error) {
	managers, err := r.getStreamsSortedByStartTime(filterFunc)
	if err != nil {
		return nil, fmt.Errorf("failed reading streams: %w", err)
	}

	var allDetails []stream.Details
	for _, manager := range managers {
		details, err := manager.GetDetails()
		if err != nil {
			return nil, fmt.Errorf("failed to read stream details of %s: %w", manager.StreamID(), err)
		}
		allDetails = append(allDetails, details)
	}
	return allDetails, nil
}

func (r *Repository) getAllStreamDetails() ([]stream.Details, error) {
	return r.getStreamDetails(nil)
}
