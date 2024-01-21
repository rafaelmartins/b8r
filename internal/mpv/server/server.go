package server

import (
	"errors"
	"os"
	"os/exec"
)

type MpvIpcServer struct {
	binary string
	args   []string
	socket string

	cmd  *exec.Cmd
	wait chan bool
	err  error
}

func New(binary string, id string, idle bool, extraArgs ...string) *MpvIpcServer {
	if binary == "" {
		binary = "mpv"
	}

	idleV := "once"
	if idle {
		idleV = "yes"
	}

	socket := getSocket(id)

	return &MpvIpcServer{
		binary: binary,
		args: append(
			[]string{
				"--load-scripts=no",
				"--idle=" + idleV,
				"--input-ipc-server=" + socket,
			}, extraArgs...),
		socket: socket,
	}
}

func (m *MpvIpcServer) Start() error {
	if m.cmd != nil {
		return errors.New("mpv: ipc: server: already started")
	}

	m.wait = make(chan bool)

	m.cmd = exec.Command(m.binary, m.args...)
	if errors.Is(m.cmd.Err, exec.ErrDot) {
		m.cmd.Err = nil
	}

	m.cmd.Stdin = os.Stdin
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	go func() {
		m.err = m.cmd.Run()
		m.cmd = nil
		close(m.wait)
	}()

	return nil
}

func (m *MpvIpcServer) Wait() error {
	if m.cmd == nil {
		return errors.New("mpv: ipc: server: not started")
	}

	<-m.wait
	return m.err
}

func (m *MpvIpcServer) GetSocket() string {
	return m.socket
}
