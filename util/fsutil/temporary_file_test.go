package fsutil_test

import (
	"os"
	"strings"
	"testing"

	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemporaryFile(t *testing.T) {
	t.Run("creates a temporary file", func(t *testing.T) {
		tmpFile, cleanup, err := fsutil.TemporaryFile()
		require.NoError(t, err)
		defer cleanup()

		assert.FileExists(t, tmpFile)
	})

	t.Run("writes and reads data to/from the temporary file", func(t *testing.T) {
		tmpFile, cleanup, err := fsutil.TemporaryFile()
		require.NoError(t, err)
		defer cleanup()

		data := []byte("Hello, World!")
		err = os.WriteFile(tmpFile, data, 0644)
		require.NoError(t, err)

		readData, err := os.ReadFile(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, data, readData)
	})

	t.Run("cleans up the temporary file", func(t *testing.T) {
		tmpFile, cleanup, err := fsutil.TemporaryFile()
		require.NoError(t, err)

		cleanup()

		assert.NoFileExists(t, tmpFile)
	})

	t.Run("creates a temporary file with a specific ending", func(t *testing.T) {
		tmpFile, cleanup, err := fsutil.TemporaryFile(fsutil.TemporaryFileWithEnding(".txt"))
		require.NoError(t, err)
		defer cleanup()

		assert.True(t, strings.HasSuffix(tmpFile, ".txt"))
	})
}
