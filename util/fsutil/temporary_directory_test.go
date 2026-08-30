package fsutil_test

import (
	"os"
	"testing"

	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemporaryDirectory(t *testing.T) {
	t.Run("returns a valid temporary directory path", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)
		defer cleanup()

		assert.DirExists(t, tmpDir)
	})

	t.Run("returns a writable temporary directory", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)
		defer cleanup()

		testFilePath := tmpDir + "/testfile.txt"
		err = os.WriteFile(testFilePath, []byte("test content"), 0644)
		require.NoError(t, err)

		content, err := os.ReadFile(testFilePath)
		require.NoError(t, err)

		assert.Equal(t, "test content", string(content))
	})

	t.Run("cleanup removes the temporary directory", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)

		// Create a file in the temporary directory to ensure it's not empty
		testFilePath := tmpDir + "/testfile.txt"
		err = os.WriteFile(testFilePath, []byte("test content"), 0644)
		require.NoError(t, err)

		cleanup()

		assert.NoDirExists(t, tmpDir)
	})
}
