package androidtv

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/rafaelmartins/b8r/internal/androidtv/pb"
	"google.golang.org/protobuf/proto"
)

type ppStatus byte

const (
	ppUnknown ppStatus = iota
	ppPlay
	ppPause
)

type Remote struct {
	c          *connection
	errCh      chan struct{}
	err        error
	dumpEvents bool
	startedCh  chan struct{}
	started    bool
	ignore     bool

	appPackage string

	volMax   uint32
	volLevel uint32
	volMuted bool

	pp ppStatus
	m  bool
}

func NewRemote(host string, cert *tls.Certificate, dumpEvents bool) (*Remote, error) {
	c, err := newConnection(host+":6466", cert)
	if err != nil {
		return nil, err
	}

	rv := &Remote{
		c:          c,
		errCh:      make(chan struct{}),
		dumpEvents: dumpEvents,
		startedCh:  make(chan struct{}),
		pp:         ppUnknown,
	}
	go func() {
		rv.err = rv.c.Listen(rv)
		close(rv.errCh)
	}()
	return rv, nil
}

func (r *Remote) Close() error {
	if r.m && r.volMuted {
		r.Unmute()
	}
	if r.pp == ppPause {
		r.Play()
	}

	err := r.c.Close()
	r.startedCh = make(chan struct{})
	r.started = false
	r.ignore = false
	return err
}

func (r *Remote) Listen() error {
	<-r.errCh
	return r.err
}

func (r *Remote) Write(msg proto.Message) error {
	return r.c.Write(msg)
}

func (Remote) alloc() proto.Message {
	return &pb.RemoteMessage{}
}

func (r *Remote) handle(msg proto.Message) error {
	rm := msg.(*pb.RemoteMessage)

	if r.dumpEvents {
		fmt.Fprintf(os.Stderr, "event: androidtv: %+v\n", rm)
	}

	if rm.RemoteError != nil && rm.RemoteError.Value {
		return fmt.Errorf("androidtv: remote error: %+v", rm.RemoteError.Message)
	}

	if rm.RemoteConfigure != nil {
		return r.Write(&pb.RemoteMessage{
			RemoteConfigure: &pb.RemoteConfigure{
				Code1: 622,
				DeviceInfo: &pb.RemoteDeviceInfo{
					Model:       "b8r",
					Vendor:      "rgm.io",
					Unknown1:    1,
					Unknown2:    "1",
					PackageName: "b8r",
					AppVersion:  "1.0.0",
				},
			},
		})
	}

	if rm.RemoteSetActive != nil {
		return r.Write(&pb.RemoteMessage{
			RemoteSetActive: &pb.RemoteSetActive{
				Active: 622,
			},
		})
	}

	if rm.RemoteStart != nil {
		if rm.RemoteStart.Started {
			if r.started {
				r.ignore = false
			} else {
				close(r.startedCh)
				r.started = true
			}
		} else if r.started {
			r.ignore = true
		}
		return nil
	}

	if rm.RemotePingRequest != nil {
		return r.Write(&pb.RemoteMessage{
			RemotePingResponse: &pb.RemotePingResponse{
				Val1: rm.RemotePingRequest.Val1,
			},
		})
	}

	if rm.RemoteImeKeyInject != nil {
		if rm.RemoteImeKeyInject.AppInfo != nil {
			r.appPackage = rm.RemoteImeKeyInject.AppInfo.AppPackage
		} else {
			r.appPackage = ""
		}
		return nil
	}

	if rm.RemoteSetVolumeLevel != nil {
		r.volMax = rm.RemoteSetVolumeLevel.VolumeMax
		r.volLevel = rm.RemoteSetVolumeLevel.VolumeLevel
		r.volMuted = rm.RemoteSetVolumeLevel.VolumeMuted
		return nil
	}

	return nil
}

func (r *Remote) SendKeyCode(code string) error {
	if r.ignore {
		return nil // avoid spamming commands while the device is not available for it
	}

	<-r.startedCh

	keycode, found := pb.RemoteKeyCode_value[code]
	if !found {
		return fmt.Errorf("androidtv: key code not found: %s", code)
	}

	if err := r.Write(&pb.RemoteMessage{
		RemoteKeyInject: &pb.RemoteKeyInject{
			KeyCode:   pb.RemoteKeyCode(keycode),
			Direction: pb.RemoteDirection_SHORT,
		},
	}); err != nil {
		return err
	}

	switch code {
	case "KEYCODE_VOLUME_MUTE":
		r.m = true
	case "KEYCODE_MEDIA_PLAY":
		r.pp = ppPlay
	case "KEYCODE_MEDIA_PAUSE":
		r.pp = ppPause
	}

	return nil
}

func (r *Remote) Mute() error {
	if r.volMuted {
		return nil
	}
	return r.SendKeyCode("KEYCODE_VOLUME_MUTE")
}

func (r *Remote) Unmute() error {
	if !r.volMuted {
		return nil
	}
	return r.SendKeyCode("KEYCODE_VOLUME_MUTE")
}

func (r *Remote) isLauncher() bool {
	return strings.HasPrefix(r.appPackage, "com.google.android.apps.tv.") || strings.HasPrefix(r.appPackage, "com.android.tv.")
}

func (r *Remote) Play() error {
	if r.isLauncher() {
		return nil
	}
	return r.SendKeyCode("KEYCODE_MEDIA_PLAY")
}

func (r *Remote) Pause() error {
	if r.isLauncher() {
		return nil
	}
	return r.SendKeyCode("KEYCODE_MEDIA_PAUSE")
}
