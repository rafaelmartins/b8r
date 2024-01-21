package client

import (
	"net"
)

func dial(pipe string) (net.Conn, error) {
	return net.Dial("unix", pipe)
}
