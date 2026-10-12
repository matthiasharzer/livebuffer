package hls

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Segment struct {
	Duration time.Duration
	FileName string
}

func ParseSegments(playlistReader io.Reader) ([]Segment, error) {
	var segments []Segment
	scanner := bufio.NewScanner(playlistReader)
	var currentDuration float64
	var expectingAFileNameNext bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if after, ok := strings.CutPrefix(line, "#EXTINF:"); ok {
			valueStr := strings.TrimSuffix(after, ",")
			var err error
			currentDuration, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration from line '%s': %w", line, err)
			}

			expectingAFileNameNext = true
		} else if expectingAFileNameNext {
			segments = append(segments, Segment{
				Duration: time.Duration(int64(currentDuration * float64(time.Second))),
				FileName: line,
			})
			expectingAFileNameNext = false
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading playlist: %w", err)
	}

	return segments, nil
}
