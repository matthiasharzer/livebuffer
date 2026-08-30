package fsutil_test

import (
	"os"
	"testing"
	"time"

	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFilesOrdered(t *testing.T) {
	t.Run("returns files in order by time", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)
		defer cleanup()

		testFiles := []struct {
			fileName string
			time     time.Time
		}{
			{"file1.txt", time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)},
			{"file2.txt", time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)},
			{"file3.txt", time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
		}
		for _, tf := range testFiles {
			filePath := tmpDir + "/" + tf.fileName
			err := os.WriteFile(filePath, []byte("test content"), 0644)
			require.NoError(t, err)

			err = os.Chtimes(filePath, tf.time, tf.time)
			require.NoError(t, err)
		}

		orderedFiles, err := fsutil.ListFilesOrdered(tmpDir)
		require.NoError(t, err)

		assert.Equal(t, []string{"file3.txt", "file1.txt", "file2.txt"}, orderedFiles)
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, err := fsutil.ListFilesOrdered("/non/existent/directory")
		require.Error(t, err)
	})

	t.Run("does not include directories in the result", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)
		defer cleanup()

		err = os.Mkdir(tmpDir+"/subdir", 0755)
		require.NoError(t, err)

		err = os.WriteFile(tmpDir+"/file.txt", []byte("test content"), 0644)
		require.NoError(t, err)

		orderedFiles, err := fsutil.ListFilesOrdered(tmpDir)
		require.NoError(t, err)

		assert.Equal(t, []string{"file.txt"}, orderedFiles)
	})

	t.Run("returns empty slice for empty directory", func(t *testing.T) {
		tmpDir, cleanup, err := fsutil.TemporaryDirectory()
		require.NoError(t, err)
		defer cleanup()

		orderedFiles, err := fsutil.ListFilesOrdered(tmpDir)
		require.NoError(t, err)

		assert.Empty(t, orderedFiles)
	})
}
