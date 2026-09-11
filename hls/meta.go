package hls

import "path/filepath"

const ChunkExtension = ".ts"
const ChunkSizeSeconds = 5
const ChunkFilename = "chunk_%05d" + ChunkExtension
const IndexFilename = "index.m3u8"

func IndexFilePath(directory string) string {
	return filepath.Join(directory, IndexFilename)
}

func ResolveFile(directory string, filename string) string {
	filename = filepath.Clean(filename)

	if filename == IndexFilename || filename == "." {
		return IndexFilePath(directory)
	}
	if filepath.Ext(filename) == ChunkExtension {
		return filepath.Join(directory, filename)
	}
	return ""
}
