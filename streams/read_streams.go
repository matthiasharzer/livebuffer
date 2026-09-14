package streams

import (
	"fmt"
	"iter"
	"os"
	"slices"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/streams/filter"
)

type streamCommon struct {
	manager stream.Manager
	details stream.Details
}

func (d *Director) readStream(streamID string) (*stream.Manager, error) {
	streamDirectory := d.streamDirectory(streamID)
	_, err := os.Stat(streamDirectory)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	streamState := d.getLiveStateFunc(streamID)

	manager, err := stream.NewManager(streamDirectory, streamState)
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

func (d *Director) readAllStreams() iter.Seq2[stream.Manager, error] {
	return func(yield func(stream.Manager, error) bool) {
		dirEntries, err := os.ReadDir(d.bufferDirectory)
		if err != nil {
			yield(stream.Manager{}, fmt.Errorf("failed to list streams: %w", err))
			return
		}

		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}
			streamID := entry.Name()

			manager, err := d.readStream(streamID)
			if err != nil {
				yield(stream.Manager{}, fmt.Errorf("failed to read stream metadata: %w", err))
				return
			}

			if !yield(*manager, nil) {
				return
			}
		}
	}
}

func (d *Director) readStreams(filterFunc filter.Func) iter.Seq2[stream.Manager, error] {
	return filter.Apply(d.readAllStreams(), filterFunc)
}

func (d *Director) expandMeta(meta stream.Metadata) (stream.Details, error) {
	directory := d.streamDirectory(meta.ID)
	state := d.getLiveStateFunc(meta.ID)
	return stream.NewInfo(meta, directory, state)
}

func (d *Director) getStreamsSortedByStartTime(filterFunc filter.Func) ([]stream.Manager, error) {
	var streams []stream.Manager
	for streamInfo, err := range d.readStreams(filterFunc) {
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

func (d *Director) getStreamCommon(streamID string) (*streamCommon, error) {
	manager, err := d.readStream(streamID)
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

func (d *Director) getStreamDetails(filterFunc filter.Func) ([]stream.Details, error) {
	managers, err := d.getStreamsSortedByStartTime(filterFunc)
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

func (d *Director) getAllStreamDetails() ([]stream.Details, error) {
	return d.getStreamDetails(nil)
}
