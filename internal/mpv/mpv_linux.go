package mpv

import (
	"net"
)

var (
	ipcServer = "/tmp/b8-mpv.socket"
)

func dial() (net.Conn, error) {
	return net.Dial("unix", ipcServer)
}
