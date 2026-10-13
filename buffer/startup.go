package buffer

import (
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/scheduler"
	"github.com/matthiasharzer/livebuffer/stream"
)

func (d *Director) StartupTasks() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.checkDetailsInitially()
	d.checkMutedSegmentsInitially()
}

func (d *Director) isStreamFinished(meta stream.Metadata) bool {
	if meta.Details == nil {
		return false
	}
	state := d.getStreamState(meta.ID)
	if state == stream.StateLive {
		return false
	}
	return true
}

func (d *Director) checkDetailsInitially() {
	for s, err := range d.readKnownStreams() {
		if err != nil {
			logging.Warn("failed to read stream", "error", err)
			continue
		}
		meta := s.Meta()
		if !d.isStreamFinished(meta) {
			continue
		}

		monitor, ok := d.monitorByBroadcasterUserName[meta.BroadcasterUserName]
		if !ok {
			logging.Warn("no monitor found for broadcaster", "broadcaster", meta.BroadcasterUserName, "stream_id", s.StreamID())
			continue
		}

		monitor.SupplementStreamMetadata(meta.ID)
	}
}

func (d *Director) checkMutedSegmentsInitially() {
	for s, err := range d.readKnownStreams() {
		if err != nil {
			logging.Warn("failed to read stream", "error", err)
			continue
		}

		meta := s.Meta()
		if !d.isStreamFinished(meta) {
			continue
		}
		if meta.MutedSegments != nil {
			continue
		}
		monitor, ok := d.monitorByBroadcasterUserName[meta.BroadcasterUserName]
		if !ok {
			logging.Warn("no monitor found for broadcaster", "broadcaster", meta.BroadcasterUserName, "stream_id", s.StreamID())
			continue
		}

		streamEnd := meta.StartedAt.Add(meta.Details.Duration)
		checkAt := streamEnd.Add(checkMutedSegmentsDelay)
		checkIn := time.Until(checkAt)

		checkNow := checkIn < 0
		if checkNow {
			go monitor.CheckMutedSegments(meta.ID)
		} else {
			scheduler.In(checkIn).Run(func() error {
				monitor.CheckMutedSegments(meta.ID)
				return nil
			})
		}
	}
}
