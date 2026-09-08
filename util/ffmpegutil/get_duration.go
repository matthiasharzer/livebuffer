package ffmpegutil

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func GetDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get stream duration: %w", err)
	}

	durationStr := strings.Trim(string(output), "\n\r")
	duration, err := time.ParseDuration(fmt.Sprintf("%ss", durationStr))
	if err != nil {
		return 0, fmt.Errorf("failed to parse stream duration: %w", err)
	}

	return duration, nil
}
