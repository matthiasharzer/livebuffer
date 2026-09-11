package hls

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

// GetDuration returns the duration of the HLS stream
func GetDuration(indexFile string) (time.Duration, error) {
	file, err := os.Open(indexFile)
	if err != nil {
		return 0, err
	}
	defer funcutils.LogError(file.Close, "failed to close hls index file")

	var totalSeconds float64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		after, ok := strings.CutPrefix(line, "#EXTINF:")
		if !ok {
			continue
		}

		valStr := after
		valStr = strings.TrimSuffix(valStr, ",")

		duration, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue // Skip malformed lines, though FFmpeg writes them perfectly
		}

		totalSeconds += duration
	}

	err = scanner.Err()
	if err != nil {
		return 0, err
	}

	return time.Duration(totalSeconds * float64(time.Second)), nil
}
