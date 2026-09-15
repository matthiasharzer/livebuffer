package vod

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

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

func (r *Repository) GetStreamReader(ctx context.Context, streamID string) (stream.Details, io.ReadCloser, error) {
	common, err := r.getStreamCommon(streamID)
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

func (r *Repository) GetClipReader(ctx context.Context, streamID string, startTime, endTime time.Duration) (stream.ClipInfo, io.ReadCloser, error) {
	common, err := r.getStreamCommon(streamID)
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

func (r *Repository) GetStream(streamID string) (*stream.Details, error) {
	common, err := r.getStreamCommon(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream details for %s: %w", streamID, err)
	}
	if common == nil {
		return nil, nil
	}
	return &common.details, nil
}

func (r *Repository) GetStreams() ([]stream.Details, error) {
	return r.getAllStreamDetails()
}

func (r *Repository) GetStreamsByBroadcaster(username string) ([]stream.Details, error) {
	return r.getStreamDetails(filter.ByBroadcasterName(username))
}

func (r *Repository) GetStreamFilesDirectory(streamID string) (string, error) {
	// TODO: needed?
	return r.StreamDirectory(streamID), nil
}

//func (d *Director) GetLiveStreamFilesDirectory(username string) (string, error) {
//	streamID := d.state.GetLiveStreamID(username)
//	if streamID == "" {
//		return "", nil
//	}
//	return d.streamDirectory(streamID), nil
//}

//func (d *Director) AddOnlineChannel(channels ...observer.ReadonlyChannel[twitch.StreamOnlineState]) {
//	for _, channel := range channels {
//		d.state.Add(channel)
//	}
//}

func (r *Repository) Close() error {
	//return d.state.Close()
	return nil
}
