package buffer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

// Director manages the livebuffer for twitch streams
type Director struct {
	maxStreams      int
	bufferDirectory string
	username        string
	onlineChannel   observer.ReadonlyChannel[twitch.StreamOnlineState]
	session         *recordingSession
	cancelRecording func()

	mu sync.Mutex
}

func NewDirector(maxStreams int, bufferBaseDirectory string, username string, onlineChannel observer.ReadonlyChannel[twitch.StreamOnlineState]) (*Director, error) {
	if maxStreams <= 0 {
		return nil, errors.New("maxStreams must be greater than 0")
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
		session:         nil,
		cancelRecording: nil,
		mu:              sync.Mutex{},
	}
	err = director.cleanupFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to initially cleanup buffer files: %w", err)
	}

	onlineChannel.Subscribe(director.onlineStateChanged)
	return director, nil
}

func (d *Director) cleanupFiles() error {
	files, err := fsutil.ListFilesOrdered(d.bufferDirectory)
	if err != nil {
		return fmt.Errorf("failed to list buffer files: %w", err)
	}

	if len(files) <= d.maxStreams {
		return nil
	}

	filesToDelete := files[:len(files)-d.maxStreams]
	for _, fileName := range filesToDelete {
		filePath := filepath.Join(d.bufferDirectory, fileName)
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to delete buffer file %s: %w", fileName, err)
		}
		logging.Info("deleted buffer file", "file", fileName)
	}
	return nil
}

func (d *Director) stopRecordingStop() error {
	logging.Info("stopping recording session", "username", d.username)
	if d.cancelRecording != nil {
		d.cancelRecording()
		d.cancelRecording = nil
	}
	if d.session != nil {
		err := d.session.Close()
		if err != nil {
			logging.Error("failed to close recording session", "error", err)
		}
		d.session = nil
	}

	return d.cleanupFiles()
}

func (d *Director) createRecordingContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, func() {
		cancel()
		d.cancelRecording = nil // we are the cancel function, so we clear it here

		err := d.stopRecordingStop()
		if err != nil {
			logging.Error("failed to stop recording session", "error", err)
		}
	}
}

func (d *Director) onlineStateChanged(state twitch.StreamOnlineState) {
	if state.IsOnline {
		d.wentLive()
	} else {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.cancelRecording != nil {
			d.cancelRecording()
			d.cancelRecording = nil
		}
	}
}

func (d *Director) wentLive() {
	logging.Info("stream went live, starting recording session", "username", d.username)
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cancelRecording != nil {
		d.cancelRecording()
		d.cancelRecording = nil
	}

	err := d.startRecording()
	if err != nil {
		logging.Error("failed to start recording session", "error", err)
	}
}

func (d *Director) startRecording() error {
	if d.session != nil {
		return fmt.Errorf("recording session already exists")
	}

	ctx, cancel := d.createRecordingContext()
	d.cancelRecording = cancel

	bufferFilePath := filepath.Join(d.bufferDirectory, fmt.Sprintf("%s_%d.m3u8", d.username, time.Now().Unix()))
	session, err := newRecordingSession(d.username, bufferFilePath)
	if err != nil {
		return fmt.Errorf("failed to create recording session: %w", err)
	}
	err = session.Start(ctx)
	if err != nil {
		funcutils.LogError(session.Close, "failed to close recording session after start error")
		return fmt.Errorf("failed to start recording session: %w", err)
	}

	d.session = session
	return nil
}

func (d *Director) GetStreams() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	files, err := fsutil.ListFilesOrdered(d.bufferDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to list buffer files: %w", err)
	}
	var streams []string
	for _, file := range files {
		streams = append(streams, filepath.Base(file))
	}
	return streams, nil
}

func (d *Director) GetStream(streamName string) (io.ReadCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	streamPath := filepath.Join(d.bufferDirectory, streamName)
	_, err := os.Stat(streamPath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat stream %s: %w", streamName, err)
	}

	if d.session != nil && d.session.FilePath() == streamPath {
		return d.session.buffer.NewSnapshotReader()
	}

	f, err := os.Open(streamPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream %s: %w", streamName, err)
	}
	return f, nil
}

func (d *Director) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onlineChannel.Unsubscribe(d.onlineStateChanged)
	return d.stopRecordingStop()
}
