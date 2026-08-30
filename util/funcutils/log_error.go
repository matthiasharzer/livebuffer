package funcutils

import "github.com/matthiasharzer/livebuffer/logging"

func LogError(fn func() error, message string) {
	err := fn()
	if err != nil {
		logging.Error(message, "error", err)
	}
}
