package stream

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/matthiasharzer/livebuffer/hls"
)

func Reader(ctx context.Context, streamDir string) (io.ReadCloser, error) {
	reader, err := hls.NewReader(ctx, FilesDirectory(streamDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create hls reader: %w", err)
	}
	return reader, nil
}

func ClipReader(ctx context.Context, streamDir string, from time.Duration, to time.Duration) (io.ReadCloser, error) {
	return hls.NewClipReader(ctx, FilesDirectory(streamDir), from, to)
}
