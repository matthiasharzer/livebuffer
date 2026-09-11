package stream

import "io"

type Manager interface {
	StreamInfo() (Info, error)
	StreamID() string
	StreamFilePath() (string, error)
	Reader() (io.ReadCloser, int64, error)
}
