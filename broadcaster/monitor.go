package broadcaster

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
)

type StreamDirectoryFunc = func(streamID string) string

type Monitor struct {
	maxStreams               int
	broadcasterUserName      string
	onlineChannel            observer.ReadonlyChannel[twitch.StreamOnlineState]
	unsubscribeOnlineChannel observer.UnsubscribeFunc
	liveRecordingSession     *stream.RecordingSession
	streamDirectory          StreamDirectoryFunc

	mu sync.Mutex
}

func NewMonitor(maxStreams int, username string, onlineChannel observer.ReadonlyChannel[twitch.StreamOnlineState], streamDirectory StreamDirectoryFunc) (*Monitor, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
	}

	if !ffmpegutil.IsInstalled() {
		return nil, errors.New("ffmpeg is required. Please install ffmpeg to use the monitor")
	}

	//bufferDir := filepath.Join(bufferBaseDirectory, username)
	//err := os.MkdirAll(bufferDir, 0777)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create buffer directory: %w", err)
	//}

	monitor := &Monitor{
		maxStreams:          maxStreams,
		broadcasterUserName: username,
		onlineChannel:       onlineChannel,
		streamDirectory:     streamDirectory,
		mu:                  sync.Mutex{},
	}
	//err := monitor.cleanupFiles()
	//if err != nil {
	//	return nil, err
	//}

	monitor.subscribeToOnlineChannel()
	return monitor, nil
}

//func (m *Monitor) cleanupFiles() error {
//	broadcasterStreams, err := m.state.GetStreamsByBroadcaster(m.broadcasterUserName)
//	if err != nil {
//		return fmt.Errorf("failed to read broadcaster streams: %w", err)
//	}
//	if len(broadcasterStreams) <= m.maxStreams {
//		return nil
//	}
//
//	slices.SortStableFunc(broadcasterStreams, func(a, b stream.Details) int {
//		return a.StartedAt.Compare(b.StartedAt)
//	})
//
//	var liveStreamID string
//	if m.liveRecordingSession != nil {
//		liveStreamID = m.liveRecordingSession.StreamID()
//	}
//
//	streamsToDelete := broadcasterStreams[:len(broadcasterStreams)-m.maxStreams]
//	for _, streamInfo := range streamsToDelete {
//		if streamInfo.ID == liveStreamID {
//			// This should never happen, but just in case, we skip deleting the live stream
//			logging.Warn("skipping deletion of live stream", "stream", streamInfo.ID, "directory", streamInfo.Directory)
//			continue
//		}
//		err := os.RemoveAll(streamInfo.Directory)
//		if err != nil {
//			return fmt.Errorf("failed to delete buffered stream %s: %w", streamInfo.Directory, err)
//		}
//		logging.Info("deleted buffered stream", "stream", streamInfo.ID, "directory", streamInfo.Directory)
//	}
//	return nil
//}

func (m *Monitor) subscribeToOnlineChannel() {
	m.unsubscribeOnlineChannel = m.onlineChannel.Subscribe(m.onlineStateChanged)
}

func (m *Monitor) onlineStateChanged(state twitch.StreamOnlineState) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
		err := m.stopRecordingStop()
		if err != nil {
			logging.Error("failed to stop recording session", "error", err)
		}
	}
}

func (m *Monitor) wentLive(event stream.WentLiveEvent) {
	logging.Info("stream went live, starting recording session", "username", event.BroadcasterUserName)

	if m.liveRecordingSession != nil {
		logging.Warn("received went live event while already recording (ignoring)", "stream_id", event.StreamID)
		return
	}

	streamBufferDir := m.streamDirectory(event.StreamID)
	err := os.MkdirAll(streamBufferDir, 0777)
	if err != nil {
		logging.Error("failed to create stream buffer directory", "error", err)
		return
	}

	recordingSession, err := stream.StartRecording(event, streamBufferDir)
	if err != nil {
		logging.Error("failed to create recording stream manager", "error", err)
		return
	}
	m.liveRecordingSession = recordingSession
	logging.Info("started recording session", "username", event.BroadcasterUserName, "stream_id", event.StreamID)

	//err = m.cleanupFiles()
	//if err != nil {
	//	logging.Warn("failed to cleanup files", "error", err)
	//}
}

func (m *Monitor) stopRecordingStop() error {
	logging.Info("stopping recording session", "username", m.broadcasterUserName)
	if m.liveRecordingSession != nil {
		err := m.liveRecordingSession.Close()
		if err != nil {
			logging.Error("failed to close recording session", "error", err)
		}
		m.liveRecordingSession = nil
	}
	return nil
	//return m.cleanupFiles()
}

func (m *Monitor) IsLive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.liveRecordingSession != nil
}

func (m *Monitor) GetLiveStreamID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.liveRecordingSession == nil {
		return ""
	}
	return m.liveRecordingSession.StreamID()
}

func (m *Monitor) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.unsubscribeOnlineChannel != nil {
		m.unsubscribeOnlineChannel()
		m.unsubscribeOnlineChannel = nil
	}
	return m.stopRecordingStop()
}
