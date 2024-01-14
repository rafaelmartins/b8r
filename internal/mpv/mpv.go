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

type state uint8

const (
	stateStopped state = iota
	stateStarting
	stateRunning
	stateExited
)

type EventHandler func(m *MPV, event string, data map[string]interface{}) error

type MPV struct {
	binary       string
	id           string
	cmd          *exec.Cmd
	cmdIdle      bool
	cmdExtraArgs []string
	cmdWait      chan bool
	socket       net.Conn
	state        state
	err          error

	mtx           sync.Mutex
	requestID     uint
	pending       map[uint]chan *result
	handlers      map[string][]EventHandler
	setupCommands [][]interface{}
}

func New(binary string, id string, idle bool, extraArgs ...string) (*MPV, error) {
	rv := &MPV{
		binary:       binary,
		id:           id,
		cmdIdle:      idle,
		cmdExtraArgs: extraArgs,
		pending:      make(map[uint]chan *result),
		handlers:     make(map[string][]EventHandler),
	}

	if err := rv.SetupCommand("disable_event", "all"); err != nil {
		return nil, err
	}
	return rv, nil
}

func (m *MPV) Start() error {
	if m.state != stateStopped && m.state != stateExited {
		return errors.New("mpv: already started")
	}
	m.state = stateStarting
	m.cmdWait = make(chan bool)

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

	go func() {
		m.state = stateRunning
		m.err = cmd.Run()
		m.state = stateExited
		close(m.cmdWait)
	}()

	var (
		socket net.Conn
		err    error
	)
	for i := 0; i < 20; i++ {
		if m.state == stateExited {
			return m.err
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
	setupCommands := m.setupCommands
	m.mtx.Unlock()

	go m.listen()

	for _, scmd := range setupCommands {
		if _, err := m.Command(scmd...); err != nil {
			return err
		}
	}
	return nil
}

func (m *MPV) Wait() error {
	<-m.cmdWait
	return m.err
}

func (m *MPV) listen() {
	for {
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
	eh, found := m.handlers[event]
	m.handlers[event] = append(eh, fn)
	m.mtx.Unlock()

	if !found {
		if err := m.SetupCommand("enable_event", event); err != nil {
			return err
		}
	}

	return nil
}

func (m *MPV) CommandWithContext(ctx context.Context, args ...interface{}) (interface{}, error) {
	switch m.state {
	case stateStopped:
		return nil, errors.New("mpv: not running")
	case stateExited:
		return nil, errors.New("mpv: exited")
	case stateStarting:
		return nil, errors.New("mpv: starting")
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

func (m *MPV) SetupCommand(args ...interface{}) error {
	if m.state != stateStopped && m.state != stateExited {
		if _, err := m.Command(args...); err != nil {
			return err
		}
	}

	m.mtx.Lock()
	m.setupCommands = append(m.setupCommands, args)
	m.mtx.Unlock()

	return nil
}

func (m *MPV) GetProperty(name string) (interface{}, error) {
	return m.Command("get_property", name)
}

func (m *MPV) GetPropertyBool(name string) (bool, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return false, err
	}

	if v, ok := rv.(bool); ok {
		return v, nil
	}
	return false, errors.New("mpv: returned property value is not a boolean")
}

func (m *MPV) GetPropertyFloat64(name string) (float64, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return 0, err
	}

	if v, ok := rv.(float64); ok {
		return v, nil
	}
	return 0, errors.New("mpv: returned property value is not a float64")
}

func (m *MPV) GetPropertyInt(name string) (int, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return 0, err
	}

	if v, ok := rv.(float64); ok {
		return int(v), nil
	}
	return 0, errors.New("mpv: returned property value is not an integer")
}

func (m *MPV) GetPropertyString(name string) (string, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return "", err
	}

	if v, ok := rv.(string); ok {
		return v, nil
	}
	return "", errors.New("mpv: returned property value is not a string")
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
