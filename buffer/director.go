package buffer

import (
	"sync"

	"github.com/matthiasharzer/livebuffer/buffer/broadcaster"
	"github.com/matthiasharzer/livebuffer/buffer/vod"
)

type StreamState string

const (
	StreamStateLive     StreamState = "live"
	StreamStateArchived StreamState = "archived"
)

type Director struct {
	Repository          *vod.Repository
	broadcasterMonitors map[string]*broadcaster.Monitor

	mu sync.RWMutex
}

func NewDirector(repository *vod.Repository, broadcasterMonitors map[string]*broadcaster.Monitor) *Director {
	return &Director{
		Repository:          repository,
		broadcasterMonitors: broadcasterMonitors,
		mu:                  sync.RWMutex{},
	}
}

func (d *Director) GetStreamStateFunc() func(streamID string) StreamState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// pre calculate for performance reasons
	liveStreams := make(map[string]bool)
	for _, monitor := range d.broadcasterMonitors {
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

	liveMonitor, ok := d.broadcasterMonitors[username]
	if !ok {
		return "", nil
	}
	if !liveMonitor.IsLive() {
		return "", nil
	}
	return d.Repository.GetStreamFilesDirectory(liveMonitor.GetLiveStreamID())
}

func (d *Director) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	repositoryErr := d.Repository.Close()

	var firstMonitorErr error
	for _, monitor := range d.broadcasterMonitors {
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
