package hls

import "path/filepath"

const chunkExtension = ".ts"
const chunkSizeSeconds = 5
const chunkFilename = "chunk_%05d" + chunkExtension
const indexFilename = "index.m3u8"

func indexFilePath(directory string) string {
	return filepath.Join(directory, indexFilename)
}
