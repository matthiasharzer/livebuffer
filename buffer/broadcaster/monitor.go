package broadcaster

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/stream"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
	"github.com/matthiasharzer/service-kit/logging"
)

type RecordingState string

const (
	RecordingStateStopped   RecordingState = "stopped"
	RecordingStateRecording RecordingState = "recording"
)

type RecordingStateChangeEvent struct {
	Username string
	State    RecordingState
	StreamID string
}

type StreamDirectoryFunc = func(streamID string) string

type Monitor struct {
	maxStreams               int
	twitchClient             *twitch.Client
	unsubscribeOnlineChannel observer.UnsubscribeFunc
	liveRecordingSession     *stream.RecordingSession
	recordingStateChannel    observer.ReadWriteChannel[RecordingStateChangeEvent]
	streamDirectory          StreamDirectoryFunc

	mu sync.RWMutex
}

func NewMonitor(maxStreams int, twitchClient *twitch.Client, streamDirectory StreamDirectoryFunc) (*Monitor, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
	}

	if !ffmpegutil.IsInstalled() {
		return nil, errors.New("ffmpeg is required. Please install ffmpeg to use the monitor")
	}

	monitor := &Monitor{
		maxStreams:            maxStreams,
		twitchClient:          twitchClient,
		streamDirectory:       streamDirectory,
		recordingStateChannel: observer.NewChannel[RecordingStateChangeEvent](),
		mu:                    sync.RWMutex{},
	}
	monitor.subscribeToOnlineChannel()
	return monitor, nil
}

func (m *Monitor) subscribeToOnlineChannel() {
	m.unsubscribeOnlineChannel = m.twitchClient.OnlineChannel().Subscribe(m.onlineStateChanged)
}

func (m *Monitor) onlineStateChanged(state twitch.StreamOnlineState) {
	if state.IsOnline {
		startedAt := time.Now()
		if state.StartedAt != nil {
			startedAt = *state.StartedAt
		}
		m.wentLive(stream.WentLiveEvent{
			StreamID:             state.StreamID,
			Title:                state.Title,
			BroadcasterUserLogin: state.BroadcasterUserLogin,
			StartedAt:            startedAt,
		})
	} else {
		m.wentOffline(stream.WentOfflineEvent{
			StreamID:             state.StreamID,
			Title:                state.Title,
			BroadcasterUserLogin: state.BroadcasterUserLogin,
		})
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
	logging.Info("stream went live, starting recording session", "username", event.BroadcasterUserLogin, "stream_id", event.StreamID)
	m.mu.Lock()
	success := m.startRecording(event)
	m.mu.Unlock()

	if !success {
		return
	}

	m.recordingStateChannel.Publish(RecordingStateChangeEvent{
		Username: event.BroadcasterUserLogin,
		StreamID: event.StreamID,
		State:    RecordingStateRecording,
	})
	logging.Info("started recording session", "username", event.BroadcasterUserLogin, "stream_id", event.StreamID)
}

func (m *Monitor) wentOffline(event stream.WentOfflineEvent) {
	logging.Info("stream went offline, stopping recording", "username", m.twitchClient.Username())
	m.mu.Lock()
	m.stopRecording()
	m.supplementStreamMetadata(event)
	m.mu.Unlock()

	m.recordingStateChannel.Publish(RecordingStateChangeEvent{
		Username: event.BroadcasterUserLogin,
		StreamID: event.StreamID,
		State:    RecordingStateStopped,
	})
	logging.Info("stopped recording session", "username", m.twitchClient.Username())
}

func (m *Monitor) supplementStreamMetadata(event stream.WentOfflineEvent) {
	streamDirectory := m.streamDirectory(event.StreamID)
	var err error

	if event.Title != "" {
		err = stream.UpdateMetadata(streamDirectory, func(meta *stream.Metadata) error {
			meta.Title = event.Title
			return nil
		})
		if err != nil {
			logging.Error("failed to update title", "error", err)
		}
	}

	err = m.supplementStreamMetadataDetails(event.StreamID)
	if err != nil {
		logging.Error("failed to supplement stream metadata details", "error", err)
	}
}

func (m *Monitor) supplementStreamMetadataDetails(streamID string) error {
	streamDirectory := m.streamDirectory(streamID)

	hlsInfo, err := hls.Stat(stream.FilesDirectory(streamDirectory))
	if err != nil {
		return err
	}

	err = stream.UpdateMetadata(streamDirectory, func(meta *stream.Metadata) error {
		meta.Details = &stream.MetadataDetails{
			Duration: hlsInfo.Duration,
			Size:     hlsInfo.Size,
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) stopRecording() {
	if m.liveRecordingSession == nil {
		logging.Warn("received went offline event while not recording (ignoring)")
		return
	}

	err := m.liveRecordingSession.Close()
	if err != nil {
		logging.Error("failed to close recording session", "error", err)
	}
	m.liveRecordingSession = nil
}

func (m *Monitor) SupplementStreamMetadataDetails(streamID string) {
	logging.Info("supplementing stream metadata", "stream_id", streamID)
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.supplementStreamMetadataDetails(streamID)
	if err != nil {
		logging.Error("failed to supplement stream metadata details", "error", err, "stream_id", streamID)
	}
}

func (m *Monitor) CheckMutedSegments(streamID string) {
	logging.Info("checking for muted segments", "stream_id", streamID)
	streamDirectory := m.streamDirectory(streamID)
	mutedSegments, err := m.twitchClient.GetMutedSegments(streamID)
	if err != nil {
		logging.Warn("failed to get muted segments", "error", err, "stream_id", streamID)
		return
	}

	if len(mutedSegments) == 0 {
		logging.Info("no muted segments found for stream", "stream_id", streamID)
	} else {
		logging.Info("found muted segments for stream", "stream_id", streamID, "count", len(mutedSegments))
	}

	//goland:noinspection ALL
	mutedSegmentsMeta := []stream.MutedSegment{}
	for _, mutedSegment := range mutedSegments {
		mutedSegmentsMeta = append(mutedSegmentsMeta, stream.MutedSegment{
			Start: mutedSegment.Start,
			End:   mutedSegment.End,
		})
	}

	err = stream.UpdateMetadata(streamDirectory, func(meta *stream.Metadata) error {
		meta.MutedSegments = mutedSegmentsMeta
		return nil
	})
	if err != nil {
		logging.Error("failed to update metadata with muted segments", "error", err, "stream_id", streamID)
		return
	}
}

func (m *Monitor) RecordingStateChannel() observer.ReadonlyChannel[RecordingStateChangeEvent] {
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
	return m.twitchClient.Username()
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
