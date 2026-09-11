package hls

import (
	"net/http"
	"path/filepath"
	"strings"
)

func ServeHTTP(streamFilesDirectory string, filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(filename, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		filename = filepath.Clean(filename)

		isIndexFile := filename == indexFilename
		isChunkFile := filepath.Ext(filename) == chunkExtension

		if isIndexFile || isChunkFile {
			fullPath := filepath.Join(streamFilesDirectory, filename)
			http.ServeFile(w, r, fullPath)
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
