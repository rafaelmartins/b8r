package client

import (
	"net"

	"gopkg.in/natefinch/npipe.v2"
)

func dial(socket string) (net.Conn, error) {
	return npipe.Dial(socket)
}
