package buffer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/buffer/stream/archived"
	"github.com/matthiasharzer/livebuffer/buffer/stream/live"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

const streamFileName = "stream.ts"
const metadataFileName = "metadata.json"

type StreamState string

const (
	StreamStateArchived StreamState = "archived"
	StreamStateLive     StreamState = "live"
)

type StreamInfo struct {
	ID                  string
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
	Duration            time.Duration
	StreamState         StreamState
	Size                int64
}

type ClipInfo struct {
	Stream    StreamInfo
	StartTime time.Duration
	EndTime   time.Duration
	Duration  time.Duration
}

type wentLiveEvent struct {
	Title               string
	BroadcasterUserName string
	StartedAt           time.Time
}

type streamMetadata struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	BroadcasterUserName string    `json:"broadcaster_user_name"`
	StartedAt           time.Time `json:"started_at"`
}

func isFfmpegInstalled() bool {
	command := exec.Command("ffmpeg", "-h")
	err := command.Run()
	return err == nil
}

type ffmpegReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (f *ffmpegReadCloser) Close() error {
	err := f.ReadCloser.Close()
	if err != nil {
		return fmt.Errorf("failed to close ffmpeg stdout pipe: %w", err)
	}

	err = f.cmd.Wait()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	return nil
}

// Director manages the livebuffer for twitch streams
type Director struct {
	maxStreams               int
	bufferDirectory          string
	username                 string
	onlineChannel            observer.ReadonlyChannel[twitch.StreamOnlineState]
	unsubscribeOnlineChannel observer.UnsubscribeFunc
	liveStreamManager        *live.StreamManager

	mu sync.Mutex
}

func NewDirector(maxStreams int, bufferBaseDirectory string, username string, onlineChannel observer.ReadonlyChannel[twitch.StreamOnlineState]) (*Director, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
	}

	if !isFfmpegInstalled() {
		return nil, errors.New("ffmpeg is not installed. Please install ffmpeg to use the director")
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
		return nil, fmt.Errorf("failed to initially cleanup buffer files: %w", err)
	}

	director.subscribeToOnlineChannel()
	return director, nil
}

func (d *Director) subscribeToOnlineChannel() {
	d.unsubscribeOnlineChannel = d.onlineChannel.Subscribe(d.onlineStateChanged)
}

func (d *Director) cleanupFiles() error {
	streams, err := d.GetStreams()
	if err != nil {
		return fmt.Errorf("failed to read streams: %w", err)
	}

	if len(streams) <= d.maxStreams {
		return nil
	}

	streamsToDelete := streams[:len(streams)-d.maxStreams]
	for _, streamInfo := range streamsToDelete {
		filePath := filepath.Join(d.bufferDirectory, streamInfo.FilePath)
		//err := os.Remove(filePath)
		//if err != nil {
		//	return fmt.Errorf("failed to delete buffer file %s: %w", streamInfo, err)
		//}
		_ = filePath
		logging.Info("deleted buffer file", "file", streamInfo)
	}
	return nil
}

func (d *Director) stopRecordingStop() error {
	logging.Info("stopping recording session", "username", d.username)
	if d.liveStreamManager != nil {
		err := d.liveStreamManager.Close()
		if err != nil {
			logging.Error("failed to close live stream manager", "error", err)
		}
		d.liveStreamManager = nil
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
			Title:               state.Title,
			BroadcasterUserName: state.BroadcasterUserName,
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

func (d *Director) writeMetadataFile(streamBufferDir string, event wentLiveEvent) error {
	metadataFilePath := filepath.Join(streamBufferDir, metadataFileName)
	id := fmt.Sprintf("%s_%s", event.BroadcasterUserName, event.StartedAt.Format("20060102_150405"))
	metadata := streamMetadata{
		ID:                  id,
		Title:               event.Title,
		BroadcasterUserName: event.BroadcasterUserName,
		StartedAt:           event.StartedAt,
	}

	file, err := os.Create(metadataFilePath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer funcutils.LogError(file.Close, "failed to close metadata file")

	encoder := json.NewEncoder(file)
	err = encoder.Encode(metadata)
	if err != nil {
		return fmt.Errorf("failed to write metadata to file: %w", err)
	}

	return nil
}

func (d *Director) wentLive(event stream.WentLiveEvent) {
	logging.Info("stream went live, starting recording session", "username", d.username)
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.liveStreamManager != nil {
		err := d.liveStreamManager.Close()
		if err != nil {
			logging.Error("failed to close existing live stream manager", "error", err)
			return
		}
		d.liveStreamManager = nil
	}

	streamBufferDir := filepath.Join(d.bufferDirectory, fmt.Sprintf("%s_%s", d.username, event.StartedAt.Format("20060102_150405")))
	err := os.MkdirAll(streamBufferDir, 0777)
	if err != nil {
		logging.Error("failed to create stream buffer directory", "error", err)
		return
	}

	manager, err := live.NewRecordingStreamManager(context.Background(), event, d.username, streamBufferDir)
	if err != nil {
		logging.Error("failed to create recording stream manager", "error", err)
		return
	}
	d.liveStreamManager = manager
	logging.Info("started recording session", "username", d.username, "stream_id", manager.StreamID())
}

func (d *Director) getManager(streamID string) (stream.Manager, error) {
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

func (d *Director) getManagerFromFolderName(folderName string) (stream.Manager, error) {
	streamDirectory := filepath.Join(d.bufferDirectory, folderName)
	metadata, err := stream.ReadMetadata(streamDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for stream: %w", err)
	}
	if d.liveStreamManager != nil {
		liveStreamID := d.liveStreamManager.StreamID()
		if metadata.ID == liveStreamID {
			return d.liveStreamManager, nil
		}
	}

	manager, err := archived.NewStreamManager(streamDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream manager for archived stream: %w", err)
	}
	return manager, nil
}

func (d *Director) readStreamManagers() iter.Seq2[stream.Manager, error] {
	return func(yield func(stream.Manager, error) bool) {
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

func (d *Director) GetStreams() ([]stream.Info, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

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

func (d *Director) GetStream(streamID string) (stream.Info, io.ReadCloser, error) {
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

	reader, size, err := streamManager.Reader()
	if err != nil {
		return stream.Info{}, nil, fmt.Errorf("failed to get stream reader: %w", err)
	}
	streamInfo.Size = size
	return streamInfo, reader, nil
}

func (d *Director) GetClip(streamID string, startTime, endTime time.Duration) (stream.ClipInfo, io.ReadCloser, error) {
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

	streamPath, err := streamManager.StreamFilePath()
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to resolve stream path: %w", err)
	}
	if streamPath == "" {
		return stream.ClipInfo{}, nil, nil
	}

	startStr := fmt.Sprintf("%.3f", startTime.Seconds())
	durationStr := fmt.Sprintf("%.3f", endTime.Seconds()-startTime.Seconds())

	args := []string{
		"-y",
		"-i", streamPath,
		"-ss", startStr,
		"-t", durationStr,
		"-c", "copy",
		"-f", "mpegts",
		"pipe:1",
	}

	cmd := exec.Command("ffmpeg", args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to create stdout pipe for ffmpeg: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return stream.ClipInfo{}, nil, fmt.Errorf("failed to start ffmpeg command: %w", err)
	}

	readCloser := ffmpegReadCloser{
		ReadCloser: stdoutPipe,
		cmd:        cmd,
	}
	clipInfo := stream.ClipInfo{
		Stream:    streamInfo,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime - startTime,
	}
	return clipInfo, &readCloser, nil
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
