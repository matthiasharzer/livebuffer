package httputil

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const indexFile = "index.html"

func serveFile(w http.ResponseWriter, r *http.Request, fileSystem http.FileSystem, fileName string) error {
	file, err := fileSystem.Open(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return errors.New("cannot serve directory")
	}

	http.ServeContent(w, r, fileName, stat.ModTime(), file)
	return nil
}

func HandleStaticSite(fileSystem fs.FS) http.HandlerFunc {
	httpFileSystem := http.FS(fileSystem)

	return func(w http.ResponseWriter, r *http.Request) {
		fileName := r.URL.Path
		if !strings.HasPrefix(fileName, "/") {
			fileName = "/" + fileName
		}

		if strings.HasSuffix(fileName, "/") {
			fileName += indexFile
		}

		err := serveFile(w, r, httpFileSystem, path.Clean(fileName))
		if err == nil {
			return
		}

		err = serveFile(w, r, httpFileSystem, path.Clean(indexFile))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
