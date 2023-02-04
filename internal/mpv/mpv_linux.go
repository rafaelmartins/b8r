package mpv

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func ipcServerName(id string) string {
	if id == "" {
		id = "UNK"
	}
	dir := os.TempDir()
	if dir == "" {
		dir = "/tmp"
	}
	return filepath.Join(dir, fmt.Sprintf("b8r-mpv-%s.socket", id))
}

func dial(id string) (net.Conn, error) {
	return net.Dial("unix", ipcServerName(id))
}
