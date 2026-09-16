package broadcaster

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/stream"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
)

type RecordingState string

const (
	RecordingStateStopped   RecordingState = "stopped"
	RecordingStateRecording RecordingState = "recording"
)

type StreamDirectoryFunc = func(streamID string) string

type Monitor struct {
	maxStreams               int
	broadcasterUserName      string
	onlineChannel            observer.ReadonlyChannel[twitch.StreamOnlineState]
	unsubscribeOnlineChannel observer.UnsubscribeFunc
	liveRecordingSession     *stream.RecordingSession
	recordingStateChannel    observer.ReadWriteChannel[RecordingState]
	streamDirectory          StreamDirectoryFunc

	mu sync.RWMutex
}

func NewMonitor(maxStreams int, username string, onlineChannel observer.ReadonlyChannel[twitch.StreamOnlineState], streamDirectory StreamDirectoryFunc) (*Monitor, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
	}

	if !ffmpegutil.IsInstalled() {
		return nil, errors.New("ffmpeg is required. Please install ffmpeg to use the monitor")
	}

	monitor := &Monitor{
		maxStreams:            maxStreams,
		broadcasterUserName:   username,
		onlineChannel:         onlineChannel,
		streamDirectory:       streamDirectory,
		recordingStateChannel: observer.NewChannel[RecordingState](),
		mu:                    sync.RWMutex{},
	}
	monitor.subscribeToOnlineChannel()
	return monitor, nil
}

func (m *Monitor) subscribeToOnlineChannel() {
	m.unsubscribeOnlineChannel = m.onlineChannel.Subscribe(m.onlineStateChanged)
}

func (m *Monitor) onlineStateChanged(state twitch.StreamOnlineState) {
	if state.IsOnline {
		startedAt := time.Now()
		if state.StartedAt != nil {
			startedAt = *state.StartedAt
		}
		m.wentLive(stream.WentLiveEvent{
			StreamID:            state.StreamID,
			Title:               state.Title,
			BroadcasterUserName: state.BroadcasterUserName,
			StartedAt:           startedAt,
		})
	} else {
		m.wentOffline()
	}
}

func (m *Monitor) startRecording(event stream.WentLiveEvent) bool {
	if m.liveRecordingSession != nil {
		logging.Warn("received went live event while already recording (ignoring)", "stream_id", event.StreamID)
		return false
	}

	streamBufferDir := m.streamDirectory(event.StreamID)
	err := os.MkdirAll(streamBufferDir, 0755)
	if err != nil {
		logging.Error("failed to create stream buffer directory", "error", err)
		return false
	}

	recordingSession, err := stream.StartRecording(event, streamBufferDir)
	if err != nil {
		logging.Error("failed to create recording stream manager", "error", err)
		return false
	}
	m.liveRecordingSession = recordingSession

	return true
}

func (m *Monitor) wentLive(event stream.WentLiveEvent) {
	logging.Info("stream went live, starting recording session", "username", event.BroadcasterUserName)
	m.mu.Lock()
	success := m.startRecording(event)
	m.mu.Unlock()

	if !success {
		return
	}

	m.recordingStateChannel.Publish(RecordingStateRecording)
	logging.Info("started recording session", "username", event.BroadcasterUserName, "stream_id", event.StreamID)
}

func (m *Monitor) wentOffline() {
	logging.Info("stream went offline, stopping recording", "username", m.broadcasterUserName)
	m.mu.Lock()
	m.stopRecording()
	m.mu.Unlock()

	m.recordingStateChannel.Publish(RecordingStateStopped)
	logging.Info("stopped recording session", "username", m.broadcasterUserName)
}

func (m *Monitor) stopRecording() {
	if m.liveRecordingSession != nil {
		err := m.liveRecordingSession.Close()
		if err != nil {
			logging.Error("failed to close recording session", "error", err)
		}
		m.liveRecordingSession = nil
	}
}

func (m *Monitor) RecordingStateChannel() observer.ReadonlyChannel[RecordingState] {
	return m.recordingStateChannel
}

func (m *Monitor) IsLive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.liveRecordingSession != nil
}

func (m *Monitor) GetLiveStreamID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.liveRecordingSession == nil {
		return ""
	}
	return m.liveRecordingSession.StreamID()
}

func (m *Monitor) BroadcasterUserName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.broadcasterUserName
}

func (m *Monitor) MaxStreams() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxStreams
}

func (m *Monitor) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.unsubscribeOnlineChannel != nil {
		m.unsubscribeOnlineChannel()
		m.unsubscribeOnlineChannel = nil
	}
	m.recordingStateChannel.Clear()
	m.stopRecording()
	return nil
}
