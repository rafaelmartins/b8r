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
	"slices"
	"sync"
	"time"
)

var (
	ErrClosed = errors.New("client is closed")
)

type result struct {
	RequestID uint   `json:"request_id"`
	Error     string `json:"error"`
	Data      any    `json:"data"`
}

type PropertyHandler func(m *MpvIpcClient, property string, value any) error
type EventHandler func(m *MpvIpcClient, event string, data map[string]any) error

type MpvIpcClient struct {
	conn       io.ReadWriteCloser
	closed     bool
	dumpEvents bool

	pmtx       sync.Mutex
	propertyID uint
	phandlers  map[uint][]PropertyHandler

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

	for range 50 {
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
		phandlers:  make(map[uint][]PropertyHandler),
		pending:    make(map[uint]chan *result),
		handlers:   make(map[string][]EventHandler),
	}, nil
}

func NewFromFd(fd uintptr, dumpEvents bool) (*MpvIpcClient, error) {
	return &MpvIpcClient{
		conn:       os.NewFile(fd, "pipe"),
		dumpEvents: dumpEvents,
		phandlers:  make(map[uint][]PropertyHandler),
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

	m.AddHandler("property-change", func(m *MpvIpcClient, event string, data map[string]any) error {
		m.pmtx.Lock()
		defer m.pmtx.Unlock()

		handlers, ok := m.phandlers[uint(data["id"].(float64))]
		if !ok || data["data"] == nil {
			return nil
		}

		go func(handlers []PropertyHandler) {
			for _, fn := range handlers {
				if err := fn(m, data["name"].(string), data["data"]); err != nil {
					sendError(errCh, err)
				}
			}
		}(slices.Clone(handlers))
		return nil
	})

	scanner := bufio.NewScanner(m.conn)
	for scanner.Scan() {
		buf := result{}
		if err := json.Unmarshal(scanner.Bytes(), &buf); err != nil {
			sendError(errCh, err)
			continue
		}
		if buf.Error == "" {
			ebuf := map[string]any{}
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
				fmt.Fprintf(os.Stderr, "event: mpv: %s: %+v\n", eventName, ebuf)
			}

			m.mtx.Lock()
			go func(hnd []EventHandler) {
				for _, fn := range hnd {
					if err := fn(m, eventName, ebuf); err != nil {
						sendError(errCh, err)
					}
				}
			}(slices.Clone(m.handlers[eventName]))
			m.mtx.Unlock()

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
	if fn == nil {
		return false
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	eh, found := m.handlers[event]
	m.handlers[event] = append(eh, fn)
	return found
}

func (m *MpvIpcClient) CommandWithContext(ctx context.Context, args ...any) (any, error) {
	if m.closed {
		return nil, fmt.Errorf("mpv: ipc: client: %w", ErrClosed)
	}

	m.mtx.Lock()
	m.requestID++
	cmd := map[string]any{
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

func (m *MpvIpcClient) Command(args ...any) (any, error) {
	return m.CommandWithContext(context.Background(), args...)
}

func (m *MpvIpcClient) GetProperty(name string) (any, error) {
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

func (m *MpvIpcClient) SetProperty(name string, value any) error {
	_, err := m.Command("set_property", name, value)
	return err
}

func (m *MpvIpcClient) AddProperty(name string, value any) error {
	_, err := m.Command("add", name, value)
	return err
}

func (m *MpvIpcClient) CycleProperty(name string) error {
	_, err := m.Command("cycle", name)
	return err
}

func (m *MpvIpcClient) CyclePropertyValues(name string, value ...any) error {
	_, err := m.Command(append([]any{"cycle_values", name}, value...)...)
	return err
}

func (m *MpvIpcClient) ObserveProperty(name string, fn PropertyHandler) error {
	if fn == nil {
		return nil
	}

	m.pmtx.Lock()
	defer m.pmtx.Unlock()

	m.propertyID++
	m.phandlers[m.propertyID] = append(m.phandlers[m.propertyID], fn)

	_, err := m.Command("observe_property", m.propertyID, name)
	return err
}
