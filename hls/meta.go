package hls

import "path/filepath"

const chunkExtension = ".ts"
const chunkSizeSeconds = 5
const chunkFilename = "chunk_%05d" + chunkExtension
const indexFilename = "index.m3u8"

func IndexFilePath(directory string) string {
	return filepath.Join(directory, indexFilename)
}

func ResolveFile(directory string, filename string) string {
	filename = filepath.Clean(filename)

	if filename == indexFilename || filename == "." {
		return IndexFilePath(directory)
	}
	if filepath.Ext(filename) == chunkExtension {
		return filepath.Join(directory, filename)
	}
	return ""
}
