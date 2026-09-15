package buffer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/stream"
)

type streamCommon struct {
	manager stream.Manager
	details stream.Details
}

func (d *Director) getStreamCommon(streamID string) (*streamCommon, error) {
	manager, err := d.repository.ReadStream(streamID)
	if err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, nil
	}

	state := d.getStreamState(streamID)
	details, err := manager.GetDetails(state)
	if err != nil {
		return nil, err
	}
	return &streamCommon{
		manager: *manager,
		details: details,
	}, nil
}

func (d *Director) getStreamDetails(filterFunc filter.Func) ([]stream.Details, error) {
	managers, err := d.repository.GetStreamsSortedByStartTime(filterFunc)
	if err != nil {
		return nil, fmt.Errorf("failed reading streams: %w", err)
	}

	var allDetails []stream.Details
	for _, manager := range managers {
		state := d.getStreamState(manager.StreamID())
		details, err := manager.GetDetails(state)
		if err != nil {
			return nil, fmt.Errorf("failed to read stream details of %s: %w", manager.StreamID(), err)
		}
		allDetails = append(allDetails, details)
	}
	return allDetails, nil
}

func (d *Director) GetStreamReader(ctx context.Context, streamID string) (stream.Details, io.ReadCloser, error) {
	common, err := d.getStreamCommon(streamID)
	if err != nil {
		return stream.Details{}, nil, fmt.Errorf("failed reading stream %s: %w", streamID, err)
	}
	if common == nil {
		return stream.Details{}, nil, nil
	}
	reader, err := common.manager.Reader(ctx)
	if err != nil {
		return stream.Details{}, nil, fmt.Errorf("failed creating stream reader: %w", err)
	}
	return common.details, reader, nil
}

func (d *Director) GetClipReader(ctx context.Context, streamID string, startTime, endTime time.Duration) (stream.ClipInfo, io.ReadCloser, error) {
	common, err := d.getStreamCommon(streamID)
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed reading stream %s: %w", streamID, err)
	}
	if common == nil {
		return stream.ClipInfo{}, nil, nil
	}

	reader, err := common.manager.ClipReader(ctx, startTime, endTime)
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to create clip reader: %w", err)
	}

	clipInfo := stream.ClipInfo{
		StreamDetails: common.details,
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      endTime - startTime,
	}
	return clipInfo, reader, nil
}

func (d *Director) GetStream(streamID string) (*stream.Details, error) {
	common, err := d.getStreamCommon(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream details for %s: %w", streamID, err)
	}
	if common == nil {
		return nil, nil
	}
	return &common.details, nil
}

func (d *Director) GetStreams() ([]stream.Details, error) {
	return d.getStreamDetails(filter.None())
}

func (d *Director) GetStreamsByBroadcaster(username string) ([]stream.Details, error) {
	return d.getStreamDetails(filter.ByBroadcasterName(username))
}
