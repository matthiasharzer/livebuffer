package stream

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const fileName = "stream.ts"
const metadataFileName = "metadata.json"

func FilesDirectory(streamDir string) string {
	return filepath.Join(streamDir, "stream")
}

func MetadataFile(streamDir string) string {
	return filepath.Join(streamDir, metadataFileName)
}

type Metadata struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	BroadcasterUserName string    `json:"broadcaster_user_name"`
	StartedAt           time.Time `json:"started_at"`
}

func WriteMetadata(streamDir string, metadata Metadata) error {
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
