//go:build unix
// +build unix

package client

import (
	"net"
)

func dial(pipe string) (net.Conn, error) {
	return net.Dial("unix", pipe)
}
