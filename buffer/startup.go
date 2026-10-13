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

func (d *Director) checkDetailsInitially() {
	for s := range d.readKnownStreams() {
		meta := s.Meta()
		if meta.Details != nil {
			continue
		}
		streamState := d.getStreamState(meta.ID)
		if streamState == stream.StateLive {
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
	for s := range d.readKnownStreams() {
		meta := s.Meta()
		details := meta.Details
		if meta.MutedSegments != nil {
			continue
		}
		if details == nil {
			logging.Warn("stream has no details, cannot check muted segments", "stream_id", s.StreamID())
			continue
		}
		monitor, ok := d.monitorByBroadcasterUserName[meta.BroadcasterUserName]
		if !ok {
			logging.Warn("no monitor found for broadcaster", "broadcaster", meta.BroadcasterUserName, "stream_id", s.StreamID())
			continue
		}

		streamEnd := meta.StartedAt.Add(details.Duration)
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
