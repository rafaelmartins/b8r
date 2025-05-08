package androidtv

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/rafaelmartins/b8r/internal/androidtv/pb"
	"google.golang.org/protobuf/proto"
)

type Pairing struct {
	c          *connection
	errCh      chan struct{}
	err        error
	dumpEvents bool

	SecretCallback   func() (string, error)
	CompleteCallback func()
}

func NewPairing(host string, cert *tls.Certificate, dumpEvents bool) (*Pairing, error) {
	c, err := newConnection(host+":6467", cert)
	if err != nil {
		return nil, err
	}

	rv := &Pairing{
		c:          c,
		errCh:      make(chan struct{}),
		dumpEvents: dumpEvents,
	}
	go func() {
		rv.err = rv.c.Listen(rv)
		close(rv.errCh)
	}()
	return rv, nil
}

func (r *Pairing) Close() error {
	return r.c.Close()
}

func (r *Pairing) Request() error {
	return r.Write(&pb.PairingMessage{
		PairingRequest: &pb.PairingRequest{
			ServiceName: "io.rgm.b8r",
			ClientName:  "b8r",
		},
		Status:          pb.PairingMessage_STATUS_OK,
		ProtocolVersion: 2,
	})
}

func (r *Pairing) Listen() error {
	<-r.errCh
	return r.err
}

func (r *Pairing) Write(msg proto.Message) error {
	return r.c.Write(msg)
}

func (Pairing) alloc() proto.Message {
	return &pb.PairingMessage{}
}

func (r *Pairing) handle(msg proto.Message) error {
	pm := msg.(*pb.PairingMessage)

	if r.dumpEvents {
		fmt.Fprintf(os.Stderr, "event: androidtv: %+v\n", pm)
	}

	if pm.PairingRequestAck != nil {
		if pm.Status != pb.PairingMessage_STATUS_OK {
			return fmt.Errorf("androidtv: pairing request failed")
		}

		return r.Write(&pb.PairingMessage{
			PairingOption: &pb.PairingOption{
				PreferredRole: pb.RoleType_ROLE_TYPE_INPUT,
				InputEncodings: []*pb.PairingEncoding{{
					Type:         pb.PairingEncoding_ENCODING_TYPE_HEXADECIMAL,
					SymbolLength: 6,
				}},
			},
			Status:          pb.PairingMessage_STATUS_OK,
			ProtocolVersion: 2,
		})
	}

	if pm.PairingOption != nil {
		if pm.Status != pb.PairingMessage_STATUS_OK {
			return fmt.Errorf("androidtv: pairing option failed")
		}

		return r.Write(&pb.PairingMessage{
			PairingConfiguration: &pb.PairingConfiguration{
				ClientRole: pb.RoleType_ROLE_TYPE_INPUT,
				Encoding: &pb.PairingEncoding{
					Type:         pb.PairingEncoding_ENCODING_TYPE_HEXADECIMAL,
					SymbolLength: 6,
				},
			},
			Status:          pb.PairingMessage_STATUS_OK,
			ProtocolVersion: 2,
		})
	}

	if pm.PairingConfigurationAck != nil {
		if pm.Status != pb.PairingMessage_STATUS_OK {
			return fmt.Errorf("androidtv: pairing configuration failed")
		}

		if r.SecretCallback == nil {
			return fmt.Errorf("androidtv: SecretCallback is required")
		}

		code, err := r.SecretCallback()
		if err != nil {
			return err
		}

		cpk, err := r.c.getClientPublicKey()
		if err != nil {
			return err
		}

		spk, err := r.c.getServerPublicKey()
		if err != nil {
			return err
		}

		hash := sha256.New()
		hash.Write(cpk.N.Bytes())
		hash.Write([]byte{1, 0, 1})
		hash.Write(spk.N.Bytes())
		hash.Write([]byte{1, 0, 1})
		e, err := hex.DecodeString(code[2:6])
		if err != nil {
			return err
		}
		hash.Write(e)
		secret := hash.Sum(nil)

		return r.Write(&pb.PairingMessage{
			PairingSecret: &pb.PairingSecret{
				Secret: secret,
			},
			Status:          pb.PairingMessage_STATUS_OK,
			ProtocolVersion: 2,
		})
	}

	if pm.PairingSecretAck != nil {
		if pm.Status != pb.PairingMessage_STATUS_OK {
			return fmt.Errorf("androidtv: pairing secret failed")
		}
		if r.CompleteCallback != nil {
			r.CompleteCallback()
		}
		return nil
	}

	if pm.Status != pb.PairingMessage_STATUS_OK {
		return fmt.Errorf("androidtv: something related to pairing failed")
	}
	return nil
}
