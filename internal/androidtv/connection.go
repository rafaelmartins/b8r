package androidtv

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
)

var (
	ErrConnectionClosed        = errors.New("connection is closed")
	ErrConnectionBadMessageLen = errors.New("message length must fit a byte")
)

type connectionHandler interface {
	alloc() proto.Message
	handle(msg proto.Message) error
}

type connection struct {
	m      sync.Mutex
	addr   string
	dialer *tls.Dialer
	conn   *tls.Conn
	closed bool
}

func newConnection(addr string, cert *tls.Certificate) (*connection, error) {
	rv := &connection{
		addr: addr,
		dialer: &tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout: time.Second,
			},
			Config: &tls.Config{
				Certificates:       []tls.Certificate{*cert},
				InsecureSkipVerify: true,
			},
		},
	}

	if err := rv.dial(); err != nil {
		return nil, err
	}
	return rv, nil
}

func (c *connection) dial() error {
	if !c.m.TryLock() {
		return nil
	}
	defer c.m.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	conn, err := c.dialer.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	if tc, ok := conn.(*tls.Conn); ok {
		c.conn = tc
		return nil
	}
	return fmt.Errorf("androidtv: invalid connection")
}

func (c *connection) redial() error {
	var err error
	for range 5 {
		err = c.dial()
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func (c *connection) Close() error {
	if c.closed {
		return fmt.Errorf("androidtv: %w", ErrConnectionClosed)
	}

	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *connection) Listen(ch connectionHandler) error {
	if c.closed {
		return fmt.Errorf("androidtv: %w", ErrConnectionClosed)
	}

	buf := make([]byte, 256)

	for {
		if c.closed {
			return nil
		}

		rv := []byte{}
		toRead := 0

		for {
			if c.closed {
				return nil
			}

			n, err := c.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
					if err := c.redial(); err != nil {
						return err
					}
					continue
				}
				return err
			}
			if n == 0 {
				continue
			}

			pos := 0
			if toRead == 0 {
				toRead = int(buf[0])
				pos++
			}

			tmp := buf[pos:n]
			if len(tmp) == 0 {
				continue
			}

			toRead -= len(tmp)
			rv = append(rv, tmp...)
			if toRead == 0 {
				break
			}
		}

		m := ch.alloc()
		if err := proto.Unmarshal(rv, m); err != nil {
			return err
		}
		if err := ch.handle(m); err != nil {
			return err
		}
	}
}

func (c *connection) Write(msg proto.Message) error {
	if c.closed {
		return fmt.Errorf("androidtv: %w", ErrConnectionClosed)
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("androidtv: failed to marshal protobuf message: %w", err)
	}

	l := len(data)
	if l > 0xff {
		return fmt.Errorf("androidtv: %w", ErrConnectionBadMessageLen)
	}

	if n, err := c.conn.Write([]byte{byte(l)}); err != nil {
		foundErr := true
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			if err := c.redial(); err == nil {
				foundErr = false
			}
		}
		if foundErr {
			return fmt.Errorf("androidtv: failed to send protobuf message length: %w", err)
		}
	} else if n != 1 {
		return fmt.Errorf("androidtv: failed to send protobuf message length: failed to write")
	}

	if n, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("androidtv: failed to send protobuf message: %w", err)
	} else if n != l {
		return fmt.Errorf("androidtv: failed to send protobuf message: failed to write")
	}
	return nil
}

func (c *connection) getClientPublicKey() (*rsa.PublicKey, error) {
	var crt *x509.Certificate
	if cs := c.dialer.Config.Certificates; len(cs) != 0 && len(cs[0].Certificate) != 0 {
		var err error
		crt, err = x509.ParseCertificate(cs[0].Certificate[0])
		if err != nil {
			return nil, err
		}
	}
	if crt == nil {
		return nil, fmt.Errorf("androidtv: failed to find client certificate")
	}

	pk, ok := crt.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("androidtv: client public key is not RSA")
	}
	return pk, nil
}

func (c *connection) getServerPublicKey() (*rsa.PublicKey, error) {
	pc := c.conn.ConnectionState().PeerCertificates
	if len(pc) == 0 {
		return nil, fmt.Errorf("androidtv: no server public key available")
	}
	pk, ok := pc[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("androidtv: server public key is not RSA")
	}
	return pk, nil
}
