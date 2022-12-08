package mime

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/danwakefield/fnmatch"
)

var errNotFound = errors.New("mime: type not found")

func detectFromData(filename string) (string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	// read start of file and rewind
	var buf [512]byte
	n, _ := io.ReadFull(fp, buf[:])

	// then try signature detection from net/http
	if m := http.DetectContentType(buf[:n]); m != "" {
		return m, nil
	}

	return "", errNotFound
}

func detectFromFilename(filename string) (string, error) {
	index := -1
	var match *mimeType
	for _, t := range registry {
		for i, p := range t.patterns {
			if fnmatch.Match(p, filename, 0) {
				if i == 0 {
					return t.name, nil
				}
				if index == -1 || i < index {
					index = i
					match = t
				}
				break
			}
		}
	}
	if match != nil && index != -1 {
		return match.name, nil
	}
	return "", errNotFound
}

func Detect(filename string) (string, error) {
	if m, err := detectFromFilename(filename); err == nil && m != "" {
		return m, nil
	}

	return detectFromData(filename)
}
