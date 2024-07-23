//go:build unix
// +build unix

package server

import (
	"fmt"
	"os"
	"path/filepath"
)

func getSocket(id string) string {
	if id == "" {
		id = "UNK"
	}
	dir := os.TempDir()
	if dir == "" {
		dir = "/tmp"
	}
	return filepath.Join(dir, fmt.Sprintf("b8r-mpv-%s.socket", id))
}
