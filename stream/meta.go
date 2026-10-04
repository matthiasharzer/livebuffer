package stream

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var metaMutex = make(map[string]*sync.RWMutex)
var masterMu = &sync.RWMutex{}

func getMutex(streamDir string) *sync.RWMutex {
	masterMu.Lock()
	defer masterMu.Unlock()

	manager, exists := metaMutex[streamDir]
	if !exists {
		manager = &sync.RWMutex{}
		metaMutex[streamDir] = manager
	}
	return manager
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

func WriteMetadata(streamDir string, metadata Metadata) error {
	mu := getMutex(streamDir)
	mu.Lock()
	defer mu.Unlock()

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

func ReadMetadata(streamDir string) (Metadata, error) {
	mu := getMutex(streamDir)
	mu.RLock()
	defer mu.RUnlock()

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

func UpdateMetadata(streamDir string, updateFunc func(*Metadata) error) error {
	mu := getMutex(streamDir)
	mu.Lock()
	defer mu.Unlock()

	metadata, err := ReadMetadata(streamDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	err = updateFunc(&metadata)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	err = WriteMetadata(streamDir, metadata)
	if err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}

	return nil
}

func IsStreamDirectory(streamDir string) bool {
	mu := getMutex(streamDir)
	mu.Lock()
	defer mu.Unlock()

	metadataFile := MetadataFile(streamDir)
	info, err := os.Stat(metadataFile)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
