package fsutil

import (
	"io/fs"
	"os"
	"sort"
	"time"
)

type fileInfo struct {
	fs.FileInfo
	ModTime time.Time
}

func ListFilesOrdered(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]fileInfo, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, fileInfo{
			FileInfo: info,
			ModTime:  info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}
