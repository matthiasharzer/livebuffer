package buffer

import (
	"sync"

	"github.com/matthiasharzer/livebuffer/buffer/broadcaster"
	"github.com/matthiasharzer/livebuffer/buffer/vod"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/stream"
)

type StreamState string

const (
	StreamStateLive     StreamState = "live"
	StreamStateArchived StreamState = "archived"
)

type Director struct {
	repository                   *vod.Repository
	monitorByBroadcasterUserName map[string]*broadcaster.Monitor
	unsubscriber                 []observer.UnsubscribeFunc

	mu sync.RWMutex
}

func NewDirector(repository *vod.Repository, broadcasterMonitors []*broadcaster.Monitor) (*Director, error) {
	director := &Director{
		repository:                   repository,
		monitorByBroadcasterUserName: make(map[string]*broadcaster.Monitor),
		unsubscriber:                 nil,
		mu:                           sync.RWMutex{},
	}
	director.cleanup()

	for _, monitor := range broadcasterMonitors {
		director.monitorByBroadcasterUserName[monitor.BroadcasterUserName()] = monitor

		unsubscribe := monitor.RecordingStateChannel().Subscribe(func(state broadcaster.RecordingState) {
			director.onRecordingStateChange(monitor.BroadcasterUserName(), state)
		})
		director.unsubscriber = append(director.unsubscriber, unsubscribe)
	}

	return director, nil
}

func (d *Director) onRecordingStateChange(username string, state broadcaster.RecordingState) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// we cleanup on all recording state updates, just to be sure, even though recording running should be enough
	err := d.cleanupUser(username)
	if err != nil {
		logging.Warn("failed to cleanup streams", "error", err, "username", username, "recording_state", state)
	}
}

func (d *Director) getStreamState(streamID string) stream.State {
	for _, monitor := range d.monitorByBroadcasterUserName {
		if monitor.GetLiveStreamID() == streamID {
			return stream.StateLive
		}
	}
	return stream.StateArchived
}

func (d *Director) GetStreamStateFunc() func(streamID string) StreamState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// pre calculate for performance reasons
	liveStreams := make(map[string]bool)
	for _, monitor := range d.monitorByBroadcasterUserName {
		streamID := monitor.GetLiveStreamID()
		if streamID == "" {
			continue
		}
		liveStreams[streamID] = true
	}

	return func(streamID string) StreamState {
		_, ok := liveStreams[streamID]
		if ok {
			return StreamStateLive
		}
		return StreamStateArchived
	}
}

func (d *Director) GetLiveStreamFilesDirectory(username string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	liveMonitor, ok := d.monitorByBroadcasterUserName[username]
	if !ok {
		return "", nil
	}
	liveStreamID := liveMonitor.GetLiveStreamID()
	if liveStreamID == "" {
		return "", nil
	}
	return d.repository.StreamFilesDirectory(liveStreamID), nil
}

// GetStreamFilesDirectory returns the directory for the hls playlist files or an empty string, if the streamID does
// not refer to an existing string
func (d *Director) GetStreamFilesDirectory(streamID string) (string, error) {
	directory := d.repository.StreamDirectory(streamID)
	if !stream.IsStreamDirectory(directory) {
		return "", nil
	}
	return d.repository.StreamFilesDirectory(streamID), nil
}

func (d *Director) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, unsubscribe := range d.unsubscriber {
		unsubscribe()
	}
	d.unsubscriber = nil

	repositoryErr := d.repository.Close()

	var firstMonitorErr error
	for _, monitor := range d.monitorByBroadcasterUserName {
		err := monitor.Close()
		if err != nil && firstMonitorErr == nil {
			firstMonitorErr = err
		}
	}

	if repositoryErr != nil {
		return repositoryErr
	}
	return firstMonitorErr
}
