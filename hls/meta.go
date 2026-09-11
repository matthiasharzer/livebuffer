package hls

import "path/filepath"

const hlsChunkExtension = ".ts"
const hlsChunkSizeSeconds = 5
const hlsChunkFilename = "chunk_%05d" + hlsChunkExtension
const hlsPlaylistFilename = "index.m3u8"

func IndexFilePath(directory string) string {
	return filepath.Join(directory, hlsPlaylistFilename)
}
