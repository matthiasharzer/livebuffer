package twitch

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

const streamlinkTimeout = 10 * time.Second

func getHLSURL(videoURL string) (string, error) {
	if !isStreamlinkInstalled() {
		return "", errors.New("streamlink is not installed. Please install streamlink to use the HLS URL retrieval")
	}

	args := []string{
		videoURL,
		"best",
		"--stream-url",
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(streamlinkTimeout))
	defer cancel()
	streamlinkCmd := exec.CommandContext(ctx, "streamlink", args...)

	output, err := streamlinkCmd.Output()
	if err != nil {
		return "", errors.New("failed to retrieve HLS URL using streamlink: " + err.Error())
	}

	hlsURL := string(output)
	if hlsURL == "" {
		return "", errors.New("streamlink did not return a valid HLS URL")
	}

	return strings.TrimSpace(hlsURL), nil
}

func isSegmentMuted(segment hls.Segment) bool {
	// This is not a very robust way to determine if a segment is muted, but twitch seems
	// to use "muted" in the filename for muted segments, so this should work for now.
	return strings.Contains(segment.FileName, "muted")
}

func getMutedSegmentsFromHLS(videoURL string) ([]MutedSegment, error) {
	hlsURL, err := getHLSURL(videoURL)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Get(hlsURL)
	if err != nil {
		return nil, errors.New("failed to fetch HLS playlist: " + err.Error())
	}
	defer funcutils.LogError(resp.Body.Close, "failed to close HLS playlist response body")

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch HLS playlist: received non-200 status code")
	}

	hlsSegments, err := hls.ParseSegments(resp.Body)
	if err != nil {
		return nil, errors.New("failed to parse HLS playlist: " + err.Error())
	}

	var mutedSegments []MutedSegment
	var currentTime time.Duration
	var nextSegment *MutedSegment
	for _, segment := range hlsSegments {
		if isSegmentMuted(segment) {
			if nextSegment == nil {
				nextSegment = &MutedSegment{
					Start: currentTime,
					End:   currentTime + segment.Duration,
				}
			} else {
				nextSegment.End += segment.Duration
			}
		} else {
			if nextSegment != nil {
				mutedSegments = append(mutedSegments, *nextSegment)
			}
			nextSegment = nil
		}
		currentTime += segment.Duration
	}
	if nextSegment != nil {
		mutedSegments = append(mutedSegments, *nextSegment)
	}
	return mutedSegments, nil
}
