package buffer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/ffmpegutil"
)

// Director manages the livebuffer for twitch streams
type Director struct {
	maxStreams               int
	bufferDirectory          string
	username                 string
	onlineChannel            observer.ReadonlyChannel[twitch.StreamOnlineState]
	unsubscribeOnlineChannel observer.UnsubscribeFunc
	liveRecordingSession     *stream.RecordingSession

	mu sync.Mutex
}

func NewDirector(maxStreams int, bufferBaseDirectory string, username string, onlineChannel observer.ReadonlyChannel[twitch.StreamOnlineState]) (*Director, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
	}

	if !ffmpegutil.IsInstalled() {
		return nil, errors.New("ffmpeg and ffprobe are required. Please install both to use the director")
	}

	bufferDir := filepath.Join(bufferBaseDirectory, username)
	err := os.MkdirAll(bufferDir, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create buffer directory: %w", err)
	}

	director := &Director{
		maxStreams:      maxStreams,
		bufferDirectory: bufferDir,
		username:        username,
		onlineChannel:   onlineChannel,
		mu:              sync.Mutex{},
	}
	err = director.cleanupFiles()
	if err != nil {
		return nil, err
	}

	director.subscribeToOnlineChannel()
	return director, nil
}

func (d *Director) subscribeToOnlineChannel() {
	d.unsubscribeOnlineChannel = d.onlineChannel.Subscribe(d.onlineStateChanged)
}

func (d *Director) cleanupUnknownDirectories(knownStreamDirectories []string) error {
	dirEntries, err := os.ReadDir(d.bufferDirectory)
	if err != nil {
		return fmt.Errorf("failed to list buffer directory: %w", err)
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}
		streamDir := filepath.Join(d.bufferDirectory, entry.Name())
		if !slices.Contains(knownStreamDirectories, streamDir) {
			err := os.RemoveAll(streamDir)
			if err != nil {
				logging.Warn("failed to delete unknown stream directory", "directory", streamDir, "error", err)
				continue
			}
			logging.Info("deleted unknown stream directory", "directory", streamDir)
		}
	}
	return nil
}

func (d *Director) cleanupFiles() error {
	dirEntries, err := os.ReadDir(d.bufferDirectory)
	if err != nil {
		return fmt.Errorf("failed to list buffer directory: %w", err)
	}

	var knownStreamDirectories []string

	type basicStreamInfo struct {
		stream.Metadata
		Directory string
	}

	var streams []basicStreamInfo
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}
		streamDirectory := filepath.Join(d.bufferDirectory, entry.Name())
		meta, err := stream.ReadMetadata(streamDirectory)
		if err != nil {
			logging.Warn("failed to read metadata for stream directory", "directory", entry.Name(), "error", err)
			continue
		}
		streams = append(streams, basicStreamInfo{
			Metadata:  meta,
			Directory: streamDirectory,
		})
		knownStreamDirectories = append(knownStreamDirectories, streamDirectory)
	}
	err = d.cleanupUnknownDirectories(knownStreamDirectories)
	if err != nil {
		logging.Warn("failed to cleanup unknown directories", "error", err)
	}

	slices.SortStableFunc(streams, func(a, b basicStreamInfo) int {
		return a.StartedAt.Compare(b.StartedAt)
	})

	if len(streams) <= d.maxStreams {
		return nil
	}

	var liveStreamID string
	if d.liveRecordingSession != nil {
		liveStreamID = d.liveRecordingSession.StreamID()
	}

	streamsToDelete := streams[:len(streams)-d.maxStreams]
	for _, streamInfo := range streamsToDelete {
		if streamInfo.ID == liveStreamID {
			// This should never happen, but just in case, we skip deleting the live stream
			logging.Warn("skipping deletion of live stream", "stream", streamInfo.ID, "directory", streamInfo.Directory)
			continue
		}
		err := os.RemoveAll(streamInfo.Directory)
		if err != nil {
			return fmt.Errorf("failed to delete buffered stream %s: %w", streamInfo.Directory, err)
		}
		logging.Info("deleted buffered stream", "stream", streamInfo.ID, "directory", streamInfo.Directory)
	}
	return nil
}

func (d *Director) stopRecordingStop() error {
	logging.Info("stopping recording session", "username", d.username)
	if d.liveRecordingSession != nil {
		err := d.liveRecordingSession.Close()
		if err != nil {
			logging.Error("failed to close recording session", "error", err)
		}
		d.liveRecordingSession = nil
	}
	return d.cleanupFiles()
}

func (d *Director) onlineStateChanged(state twitch.StreamOnlineState) {
	if state.IsOnline {
		startedAt := time.Now()
		if state.StartedAt != nil {
			startedAt = *state.StartedAt
		}
		d.wentLive(stream.WentLiveEvent{
			StreamID:            state.StreamID,
			Title:               state.Title,
			BroadcasterUserName: d.username,
			StartedAt:           startedAt,
		})
	} else {
		d.mu.Lock()
		defer d.mu.Unlock()

		err := d.stopRecordingStop()
		if err != nil {
			logging.Error("failed to stop recording session", "error", err)
		}
	}
}

func (d *Director) wentLive(event stream.WentLiveEvent) {
	logging.Info("stream went live, starting recording session", "username", event.BroadcasterUserName)
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.liveRecordingSession != nil {
		logging.Warn("received went live event while already recording (ignoring)", "stream_id", event.StreamID)
		return
	}

	streamBufferDir := filepath.Join(d.bufferDirectory, fmt.Sprintf("%s_%s", event.BroadcasterUserName, event.StartedAt.Format("20060102_150405")))
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
	d.liveRecordingSession = recordingSession
	logging.Info("started recording session", "username", event.BroadcasterUserName, "stream_id", event.StreamID)

	err = d.cleanupFiles()
	if err != nil {
		logging.Warn("failed to cleanup files", "error", err)
	}
}

func (d *Director) getManager(streamID string) (*stream.Manager, error) {
	for manager, err := range d.readStreamManagers() {
		if err != nil {
			return nil, fmt.Errorf("failed to get stream managers: %w", err)
		}
		if streamID == manager.StreamID() {
			return manager, nil
		}
	}
	return nil, nil
}

func (d *Director) getManagerFromFolderName(folderName string) (*stream.Manager, error) {
	streamDirectory := filepath.Join(d.bufferDirectory, folderName)
	metadata, err := stream.ReadMetadata(streamDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for stream: %w", err)
	}

	streamState := stream.StateArchived
	if d.liveRecordingSession != nil {
		liveStreamID := d.liveRecordingSession.StreamID()
		if metadata.ID == liveStreamID {
			streamState = stream.StateLive
		}
	}

	manager, err := stream.NewManager(streamDirectory, streamState)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream manager: %w", err)
	}
	return manager, nil
}

func (d *Director) readStreamManagers() iter.Seq2[*stream.Manager, error] {
	return func(yield func(*stream.Manager, error) bool) {
		dirEntries, err := os.ReadDir(d.bufferDirectory)
		if err != nil {
			yield(nil, fmt.Errorf("failed to list streams: %w", err))
			return
		}
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}
			manager, err := d.getManagerFromFolderName(entry.Name())
			if err != nil {
				logging.Warn("failed to get stream manager for stream", "stream", entry.Name(), "error", err)
				continue
			}
			if !yield(manager, nil) {
				break
			}
		}
	}
}

func (d *Director) readStreamInfos() iter.Seq2[stream.Info, error] {
	return func(yield func(stream.Info, error) bool) {
		for manager, err := range d.readStreamManagers() {
			if err != nil {
				yield(stream.Info{}, err)
				return
			}
			streamInfo, err := manager.StreamInfo()
			if err != nil {
				logging.Warn("failed to get stream info from manager", "stream", streamInfo.Title, "error", err)
				continue
			}
			if !yield(streamInfo, nil) {
				break
			}
		}
	}
}

func (d *Director) getStreamsSortedByStartTime() ([]stream.Info, error) {
	var streams []stream.Info
	for streamInfo, err := range d.readStreamInfos() {
		if err != nil {
			return nil, fmt.Errorf("failed to read stream infos: %w", err)
		}
		streams = append(streams, streamInfo)
	}

	slices.SortStableFunc(streams, func(a, b stream.Info) int {
		return a.StartedAt.Compare(b.StartedAt)
	})

	return streams, nil
}

func (d *Director) GetStreams() ([]stream.Info, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.getStreamsSortedByStartTime()
}

func (d *Director) GetStream(ctx context.Context, streamID string) (stream.Info, io.ReadCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	streamManager, err := d.getManager(streamID)
	if err != nil {
		return stream.Info{}, nil, fmt.Errorf("failed to get stream manager: %w", err)
	}
	if streamManager == nil {
		return stream.Info{}, nil, nil
	}

	streamInfo, err := streamManager.StreamInfo()
	if err != nil {
		return stream.Info{}, nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	reader, err := streamManager.Reader(ctx)
	if err != nil {
		return stream.Info{}, nil, fmt.Errorf("failed to get stream reader: %w", err)
	}
	return streamInfo, reader, nil
}

func (d *Director) GetClip(ctx context.Context, streamID string, startTime, endTime time.Duration) (stream.ClipInfo, io.ReadCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	streamManager, err := d.getManager(streamID)
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to get stream manager: %w", err)
	}
	if streamManager == nil {
		return stream.ClipInfo{}, nil, nil
	}

	streamInfo, err := streamManager.StreamInfo()
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	reader, err := streamManager.ClipReader(ctx, startTime, endTime)
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to create clip reader: %w", err)
	}

	clipInfo := stream.ClipInfo{
		Stream:    streamInfo,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime - startTime,
	}
	return clipInfo, reader, nil
}

func (d *Director) GetBroadcaster() (*stream.BroadcastWriter, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.liveRecordingSession == nil {
		return nil, nil
	}
	return d.liveRecordingSession.Broadcaster, nil
}

func (d *Director) HasLiveStream() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.liveRecordingSession != nil
}

func (d *Director) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.unsubscribeOnlineChannel != nil {
		d.unsubscribeOnlineChannel()
		d.unsubscribeOnlineChannel = nil
	}
	return d.stopRecordingStop()
}
