package mpv

import (
	"net"

	"gopkg.in/natefinch/npipe.v2"
)

func ipcServerName(id string) string {
	if id == "" {
		id = "UNK"
	}
	return `\\.\pipe\b8r-mpv-` + id
}

func dial(id string) (net.Conn, error) {
	return npipe.Dial(ipcServerName(id))
}
