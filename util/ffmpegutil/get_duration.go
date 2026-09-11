package ffmpegutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func GetDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe", "-v", "info", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	output, err := cmd.Output()
	fmt.Printf("Err: %s\n", stderrBuf.String())
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
