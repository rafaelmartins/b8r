package mpv

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

type result struct {
	RequestID uint        `json:"request_id"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

type EventHandler func(m *MPV, event string, data map[string]interface{}) error

type MPV struct {
	binary       string
	id           string
	cmd          *exec.Cmd
	cmdIdle      bool
	cmdExtraArgs []string
	cmdErr       chan error
	socket       net.Conn

	mtx       sync.Mutex
	requestID uint
	pending   map[uint]chan *result
	handlers  map[string][]EventHandler
}

func New(binary string, id string, idle bool, extraArgs ...string) (*MPV, error) {
	rv := &MPV{
		binary:       binary,
		id:           id,
		cmdIdle:      idle,
		cmdExtraArgs: extraArgs,
		cmdErr:       make(chan error, 1),
		pending:      make(map[uint]chan *result),
		handlers:     make(map[string][]EventHandler),
	}

	if err := rv.Start(); err != nil {
		return nil, err
	}

	// this will leak...
	go rv.listen()

	return rv, nil
}

func (m *MPV) Start() error {
	binary := m.binary
	if binary == "" {
		binary = "mpv"
	}
	idle := "once"
	if m.cmdIdle {
		idle = "yes"
	}
	cmd := exec.Command(binary, append(
		[]string{
			"--idle=" + idle,
			"--input-ipc-server=" + ipcServerName(m.id),
		}, m.cmdExtraArgs...)...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	select {
	case <-m.cmdErr:
	default:
	}

	go func() {
		m.cmdErr <- cmd.Run()
	}()

	var (
		socket net.Conn
		err    error
	)
	for i := 0; i < 20; i++ {
		done := false
		select {
		case <-m.cmdErr:
			done = true
		default:
		}
		if done {
			break
		}

		socket, err = dial(m.id)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return err
	}

	m.mtx.Lock()
	m.cmd = cmd
	m.socket = socket
	m.mtx.Unlock()

	return nil
}

func (m *MPV) Wait() error {
	return <-m.cmdErr
}

func (m *MPV) listen() {
	for {
		if m.socket == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		scanner := bufio.NewScanner(m.socket)
		for scanner.Scan() {
			buf := result{}
			if err := json.Unmarshal(scanner.Bytes(), &buf); err != nil {
				log.Printf("error: %s", err)
				continue
			}
			if buf.Error == "" {
				ebuf := map[string]interface{}{}
				if err := json.Unmarshal(scanner.Bytes(), &ebuf); err != nil {
					log.Printf("error: %s", err)
					continue
				}
				if ebuf["event"] == nil {
					continue
				}
				eventName, ok := ebuf["event"].(string)
				if !ok {
					continue
				}
				delete(ebuf, "event")

				handlers := []EventHandler{}

				m.mtx.Lock()
				if _, ok := m.handlers[eventName]; ok {
					for _, fn := range m.handlers[eventName] {
						if fn == nil {
							continue
						}
						handlers = append(handlers, fn)
					}
				}
				m.mtx.Unlock()

				go func(hnd []EventHandler) {
					for _, fn := range hnd {
						if err := fn(m, eventName, ebuf); err != nil {
							log.Printf("error: event %q: %s", eventName, err)
						}
					}
				}(handlers)

				continue
			}

			m.mtx.Lock()
			if res, ok := m.pending[buf.RequestID]; ok {
				res <- &buf
				delete(m.pending, buf.RequestID)
			}
			m.mtx.Unlock()
		}
	}
}

func (m *MPV) AddHandler(event string, fn EventHandler) error {
	if fn == nil {
		return nil
	}

	m.mtx.Lock()
	m.handlers[event] = append(m.handlers[event], fn)
	m.mtx.Unlock()
	return nil
}

func (m *MPV) waitReady() bool {
	for i := 0; i < 20; i++ {
		if m.cmd != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return m.cmd != nil
}

func (m *MPV) CommandWithContext(ctx context.Context, args ...interface{}) (interface{}, error) {
	if m.cmd == nil {
		if !m.waitReady() {
			return nil, errors.New("mpv: not running")
		}
	}

	m.mtx.Lock()
	m.requestID++
	cmd := map[string]interface{}{
		"request_id": m.requestID,
		"command":    args,
		"async":      true,
	}
	c := make(chan *result)
	m.pending[m.requestID] = c
	m.mtx.Unlock()

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	n, err := m.socket.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("mpv: failed to write command to ipc socket")
	}

	dctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		var ret *result
		select {
		case <-dctx.Done():
			return nil, errors.New("mpv: command: timeout")

		case ret = <-c:
			if ret.Error == "success" {
				return ret.Data, nil
			}
			return nil, fmt.Errorf("mpv: command: %s", ret.Error)

		default:
		}
	}
}

func (m *MPV) Command(args ...interface{}) (interface{}, error) {
	return m.CommandWithContext(context.Background(), args...)
}

func (m *MPV) GetProperty(name string) (interface{}, error) {
	return m.Command("get_property", name)
}

func (m *MPV) SetProperty(name string, value interface{}) error {
	_, err := m.Command("set_property", name, value)
	return err
}

func (m *MPV) AddProperty(name string, value interface{}) error {
	_, err := m.Command("add", name, value)
	return err
}

func (m *MPV) CycleProperty(name string) error {
	_, err := m.Command("cycle", name)
	return err
}

func (m *MPV) CyclePropertyValues(name string, value ...interface{}) error {
	_, err := m.Command(append([]interface{}{"cycle_values", name}, value...)...)
	return err
}
