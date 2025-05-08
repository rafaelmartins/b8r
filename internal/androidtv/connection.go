package androidtv

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"

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
	addr   string
	conf   *tls.Config
	conn   *tls.Conn
	closed bool
}

func newConnection(addr string, cert *tls.Certificate) (*connection, error) {
	conn := &connection{
		addr: addr,
		conf: &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			InsecureSkipVerify: true,
		},
	}

	var err error
	conn.conn, err = tls.Dial("tcp", addr, conn.conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *connection) Close() error {
	if c.closed {
		return fmt.Errorf("androidtv: %w", ErrConnectionClosed)
	}

	c.closed = true
	return c.conn.Close()
}

func (c *connection) Listen(ch connectionHandler) error {
	if c.closed {
		return fmt.Errorf("androidtv: %w", ErrConnectionClosed)
	}

	buf := make([]byte, 256)

	for {
		rv := []byte{}
		toRead := 0

		for {
			n, err := c.conn.Read(buf)
			if err == io.EOF || c.closed {
				return nil
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
		return fmt.Errorf("androidtv: failed to send protobuf message length: %w", err)
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
	if cs := c.conf.Certificates; len(cs) != 0 && len(cs[0].Certificate) != 0 {
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
