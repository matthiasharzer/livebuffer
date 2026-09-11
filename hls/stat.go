package hls

import (
	"fmt"
	"os"
	"time"
)

type Info struct {
	Size     int64
	Duration time.Duration
}

func Stat(directory string) (Info, error) {
	indexFile := indexFilePath(directory)
	_, err := os.Stat(indexFile)
	if os.IsNotExist(err) {
		return Info{}, fmt.Errorf("not a hls playlist directory")
	} else if err != nil {
		return Info{}, fmt.Errorf("failed to stat index file: %w", err)
	}

	var size int64
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Info{}, fmt.Errorf("failed to read hls directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// unexpected, but don't fail yet
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return Info{}, fmt.Errorf("failed to get file info of %s: %w", entry.Name(), err)
		}
		size += info.Size()
	}

	duration, err := GetDuration(indexFile)
	if err != nil {
		return Info{}, fmt.Errorf("failed to determine hls duration: %w", err)
	}

	return Info{
		Size:     size,
		Duration: duration,
	}, nil
}
