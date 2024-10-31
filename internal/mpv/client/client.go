package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	ErrClosed = errors.New("client is closed")
)

type result struct {
	RequestID uint        `json:"request_id"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

type EventHandler func(m *MpvIpcClient, event string, data map[string]interface{}) error

type MpvIpcClient struct {
	conn       io.ReadWriteCloser
	closed     bool
	dumpEvents bool

	mtx       sync.Mutex
	requestID uint
	pending   map[uint]chan *result
	handlers  map[string][]EventHandler
}

func NewFromSocket(socket string, dumpEvents bool) (*MpvIpcClient, error) {
	var (
		conn net.Conn
		err  error
	)

	for i := 0; i < 50; i++ {
		conn, err = dial(socket)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	return &MpvIpcClient{
		conn:       conn,
		dumpEvents: dumpEvents,
		pending:    make(map[uint]chan *result),
		handlers:   make(map[string][]EventHandler),
	}, nil
}

func NewFromFd(fd uintptr, dumpEvents bool) (*MpvIpcClient, error) {
	return &MpvIpcClient{
		conn:       os.NewFile(fd, "pipe"),
		dumpEvents: dumpEvents,
		pending:    make(map[uint]chan *result),
		handlers:   make(map[string][]EventHandler),
	}, nil
}

func (m *MpvIpcClient) Close() error {
	if m.conn != nil {
		m.closed = true
		return m.conn.Close()
	}
	return nil
}

func sendError(errCh chan error, err error) {
	if errCh == nil {
		log.Printf("error: %s", err)
		return
	}

	select {
	case errCh <- err:
	default:
	}
}

func (m *MpvIpcClient) Listen(errCh chan error) error {
	if m.closed {
		return fmt.Errorf("mpv: ipc: client: %w", ErrClosed)
	}

	scanner := bufio.NewScanner(m.conn)
	for scanner.Scan() {
		buf := result{}
		if err := json.Unmarshal(scanner.Bytes(), &buf); err != nil {
			sendError(errCh, err)
			continue
		}
		if buf.Error == "" {
			ebuf := map[string]interface{}{}
			if err := json.Unmarshal(scanner.Bytes(), &ebuf); err != nil {
				sendError(errCh, err)
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

			if m.dumpEvents {
				fmt.Fprintf(os.Stderr, "event: %s: %q\n", eventName, ebuf)
			}

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
						sendError(errCh, err)
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
	return scanner.Err()
}

func (m *MpvIpcClient) AddHandler(event string, fn EventHandler) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	eh, found := m.handlers[event]
	m.handlers[event] = append(eh, fn)
	return found
}

func (m *MpvIpcClient) CommandWithContext(ctx context.Context, args ...interface{}) (interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("mpv: ipc: client: %w", ErrClosed)
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

	n, err := m.conn.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("mpv: ipc: client: failed to write command")
	}

	dctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		var ret *result
		select {
		case <-dctx.Done():
			return nil, errors.New("mpv: ipc: client: command: timeout")

		case ret = <-c:
			err := matchError(ret.Error)
			if errors.Is(err, ErrMpvSuccess) {
				return ret.Data, nil
			}
			return nil, fmt.Errorf("mpv: ipc: client: command: %w", err)

		default:
		}
	}
}

func (m *MpvIpcClient) Command(args ...interface{}) (interface{}, error) {
	return m.CommandWithContext(context.Background(), args...)
}

func (m *MpvIpcClient) GetProperty(name string) (interface{}, error) {
	return m.Command("get_property", name)
}

func (m *MpvIpcClient) GetPropertyBool(name string) (bool, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return false, err
	}

	if v, ok := rv.(bool); ok {
		return v, nil
	}
	return false, errors.New("mpv: ipc: client: received property value is not a boolean")
}

func (m *MpvIpcClient) GetPropertyFloat64(name string) (float64, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return 0, err
	}

	if v, ok := rv.(float64); ok {
		return v, nil
	}
	return 0, errors.New("mpv: ipc: client: received property value is not a float64")
}

func (m *MpvIpcClient) GetPropertyInt(name string) (int, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return 0, err
	}

	if v, ok := rv.(float64); ok {
		return int(v), nil
	}
	return 0, errors.New("mpv: ipc: client: received property value is not an integer")
}

func (m *MpvIpcClient) GetPropertyString(name string) (string, error) {
	rv, err := m.GetProperty(name)
	if err != nil {
		return "", err
	}

	if v, ok := rv.(string); ok {
		return v, nil
	}
	return "", errors.New("mpv: ipc: client: received property value is not a string")
}

func (m *MpvIpcClient) SetProperty(name string, value interface{}) error {
	_, err := m.Command("set_property", name, value)
	return err
}

func (m *MpvIpcClient) AddProperty(name string, value interface{}) error {
	_, err := m.Command("add", name, value)
	return err
}

func (m *MpvIpcClient) CycleProperty(name string) error {
	_, err := m.Command("cycle", name)
	return err
}

func (m *MpvIpcClient) CyclePropertyValues(name string, value ...interface{}) error {
	_, err := m.Command(append([]interface{}{"cycle_values", name}, value...)...)
	return err
}
