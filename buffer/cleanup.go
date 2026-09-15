package buffer

import (
	"fmt"
	"os"
	"slices"

	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
	"github.com/matthiasharzer/livebuffer/util/iterutil"
)

func (d *Director) cleanup() {
	for username := range d.monitorByBroadcasterUserName {
		err := d.cleanupUser(username)
		if err != nil {
			logging.Warn("failed to cleanup user", "username", username, "error", err)
		}
	}
	d.cleanupUnknownDirectory()
	return
}

func (d *Director) cleanupUser(username string) error {
	monitor, ok := d.monitorByBroadcasterUserName[username]
	if !ok {
		return fmt.Errorf("user %s not managed by director", username)
	}
	maxStreams := monitor.MaxStreams()

	streamsIter := d.repository.ReadStreams(filter.ByBroadcasterName(username))
	streams, err := iterutil.Collect2(streamsIter)
	if err != nil {
		return fmt.Errorf("failed to read streams for user %s: %w", username, err)
	}

	slices.SortStableFunc(streams, func(a, b stream.Manager) int {
		return a.Meta().StartedAt.Compare(b.Meta().StartedAt)
	})

	if len(streams) <= maxStreams {
		return nil
	}

	liveStreamID := monitor.GetLiveStreamID()

	streamsToDelete := streams[:len(streams)-maxStreams]
	for _, streamManager := range streamsToDelete {
		if streamManager.StreamID() == liveStreamID {
			// This should never happen, but just in case, we skip deleting the live stream
			logging.Warn("skipping deletion of live stream", "stream", streamManager.StreamID(), "directory", streamManager.StreamDirectory)
			continue
		}
		err := os.RemoveAll(streamManager.StreamDirectory)
		if err != nil {
			return fmt.Errorf("failed to delete buffered stream %s: %w", streamManager.StreamDirectory, err)
		}
		logging.Info("deleted buffered stream", "stream", streamManager.StreamID(), "directory", streamManager.StreamDirectory)
	}
	return nil
}

func (d *Director) cleanupUnknownDirectory() {
	knownStreams := make(map[string]bool)

	streams := d.repository.ReadAllStreams()
	for streamManager, err := range streams {
		if err != nil {
			logging.Error("encountered critical stream reading error while cleaning up", "error", err)
			return
		}
		knownStreams[streamManager.StreamDirectory] = true
	}

	allPaths, err := d.repository.ReadAllBufferEntryPaths()
	if err != nil {
		logging.Error("failed to read buffer entries", "error", err)
		return
	}

	for _, path := range allPaths {
		_, known := knownStreams[path]
		if known {
			continue
		}
		err := os.RemoveAll(path)
		if err != nil {
			logging.Warn("failed to remove unknown file or directory", "path", path, "error", err)
			continue
		}
		logging.Info("removed unknown file or directory", "path", path)
	}
}
