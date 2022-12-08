package mpv

import (
	"net"

	"gopkg.in/natefinch/npipe.v2"
)

var (
	ipcServer = `\\.\pipe\b8-mpv`
)

func dial() (net.Conn, error) {
	return npipe.Dial(ipcServer)
}
