package streams

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/streams/filter"
)

type GetLiveStateFunc = func(streamID string) stream.State

type Director struct {
	bufferDirectory  string
	getLiveStateFunc GetLiveStateFunc
}

func NewDirector(bufferDirectory string, isLiveFunc GetLiveStateFunc) *Director {
	return &Director{
		bufferDirectory:  bufferDirectory,
		getLiveStateFunc: isLiveFunc,
	}
}

func (d *Director) streamDirectory(streamID string) string {
	return filepath.Join(d.bufferDirectory, streamID)
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
	return d.getAllStreamDetails()
}

func (d *Director) GetStreamsByBroadcaster(username string) ([]stream.Details, error) {
	return d.getStreamDetails(filter.ByBroadcasterName(username))
}

func (d *Director) GetStreamFilesDirectory(streamID string) (string, error) {
	// TODO: needed?
	return d.streamDirectory(streamID), nil
}

func (d *Director) Close() error {
	// TODO: needed?
	return nil
}
