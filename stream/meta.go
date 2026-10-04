package stream

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/matthiasharzer/livebuffer/util/mutexutil"
)

var mappedMutex = mutexutil.NewMappedMutex()

func lock(streamDir string) func() {
	mu, cleanupFunc := mappedMutex.GetMutex(streamDir)
	mu.Lock()
	return func() {
		mu.Unlock()
		cleanupFunc()
	}
}

const filesDirectory = "stream"
const metadataFileName = "metadata.json"

func FilesDirectory(streamDir string) string {
	return filepath.Join(streamDir, filesDirectory)
}

func MetadataFile(streamDir string) string {
	return filepath.Join(streamDir, metadataFileName)
}

type MetadataDetails struct {
	Duration time.Duration
	Size     int64
}

type Metadata struct {
	ID                  string           `json:"id"`
	Title               string           `json:"title"`
	BroadcasterUserName string           `json:"broadcaster_user_name"`
	StartedAt           time.Time        `json:"started_at"`
	Details             *MetadataDetails `json:"details,omitempty"`
}

func writeMetadata(streamDir string, metadata Metadata) error {
	metadataFile := MetadataFile(streamDir)
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = os.WriteFile(metadataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func readMetadata(streamDir string) (Metadata, error) {
	metadataFile := MetadataFile(streamDir)
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata Metadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

func updateMetadata(streamDir string, updateFunc func(*Metadata) error) error {
	metadata, err := readMetadata(streamDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	err = updateFunc(&metadata)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	err = writeMetadata(streamDir, metadata)
	if err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}

	return nil
}

func WriteMetadata(streamDir string, metadata Metadata) error {
	unlock := lock(streamDir)
	defer unlock()

	return writeMetadata(streamDir, metadata)
}

func ReadMetadata(streamDir string) (Metadata, error) {
	unlock := lock(streamDir)
	defer unlock()

	return readMetadata(streamDir)
}

func UpdateMetadata(streamDir string, updateFunc func(*Metadata) error) error {
	unlock := lock(streamDir)
	defer unlock()

	return updateMetadata(streamDir, updateFunc)
}

func IsStreamDirectory(streamDir string) bool {
	unlock := lock(streamDir)
	defer unlock()

	metadataFile := MetadataFile(streamDir)
	info, err := os.Stat(metadataFile)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
