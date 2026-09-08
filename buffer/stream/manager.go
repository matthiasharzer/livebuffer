package stream

type Manager interface {
	StreamInfo() (Info, error)
	StreamID() (string, error)
	Close() error
}
