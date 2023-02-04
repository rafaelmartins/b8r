package handlers

import (
	"math"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/mpv"
	"github.com/rafaelmartins/b8r/internal/source"
)

const (
	keySeek5Fwd  = "RIGHT"
	keySeek5Bwd  = "LEFT"
	keySeek60Fwd = "UP"
	keySeek60Bwd = "DOWN"
)

var (
	mod b8.Modifier

	keybinds = map[string]string{
		keySeek5Fwd:  "seek 5",
		keySeek5Bwd:  "seek -5",
		keySeek60Fwd: "seek 60",
		keySeek60Bwd: "seek -60",
	}
)

type LoadNextFunc func(m *mpv.MPV, src *source.Source) error

func b8Handler(short b8.ButtonHandler, long b8.ButtonHandler, modShort b8.ButtonHandler, modLong b8.ButtonHandler) b8.ButtonHandler {
	return func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			if pressed {
				if modShort != nil {
					return modShort(b)
				}
				return nil
			}
			if short != nil {
				return short(b)
			}
			return nil
		}
		if pressed {
			if modLong != nil {
				return modLong(b)
			}
			return nil
		}
		if long != nil {
			return long(b)
		}
		return nil
	}
}

func b8HoldKeyHandler(m *mpv.MPV, key string, modKey string) b8.ButtonHandler {
	return func(b *b8.Button) error {
		k := key
		if mod.Pressed() {
			k = modKey
		}
		_, err := m.Command("keydown", k)
		if err != nil {
			return err
		}
		b.WaitForRelease()
		_, err = m.Command("keyup", k)
		return err
	}
}

func RegisterHandlers(dev *b8.Device, m *mpv.MPV, s *source.Source, loadNext LoadNextFunc, exit b8.ButtonHandler) error {
	for k, v := range keybinds {
		if _, err := m.Command("keybind", k, v); err != nil {
			return err
		}
	}

	if s != nil {
		dev.AddHandler(b8.BUTTON_1, b8Handler(
			func(b *b8.Button) error {
				if pausedInt, err := m.GetProperty("pause"); err == nil {
					if p, ok := pausedInt.(bool); ok && p {
						if err := m.SetProperty("pause", false); err != nil {
							return err
						}
						return m.SetProperty("fullscreen", true)
					}
				}
				if loadNext != nil {
					return loadNext(m, s)
				}
				return nil
			},
			func(b *b8.Button) error {
				if err := m.SetProperty("pause", true); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", false)
			},
			func(b *b8.Button) error {
				_, err := m.Command("quit")
				return err
			},
			exit,
		))
	} else {
		dev.AddHandler(b8.BUTTON_1, b8Handler(
			func(b *b8.Button) error {
				if pausedInt, err := m.GetProperty("pause"); err == nil {
					if p, ok := pausedInt.(bool); ok && p {
						if err := m.SetProperty("pause", false); err != nil {
							return err
						}
						return m.SetProperty("fullscreen", true)
					}
				}
				if err := m.SetProperty("pause", true); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", false)
			},
			func(b *b8.Button) error {
				_, err := m.Command("quit")
				return err
			},
			exit,
			nil,
		))
	}

	dev.AddHandler(b8.BUTTON_2, b8Handler(
		func(b *b8.Button) error {
			return m.CycleProperty("mute")
		},
		func(b *b8.Button) error {
			return m.CyclePropertyValues("video-rotate", "90", "180", "270", "0")
		},
		func(b *b8.Button) error {
			_, err := m.Command("vf", "toggle", "hflip")
			return err
		},
		nil,
	))

	dev.AddHandler(b8.BUTTON_3, b8HoldKeyHandler(m, keySeek5Bwd, keySeek60Bwd))

	dev.AddHandler(b8.BUTTON_4, b8HoldKeyHandler(m, keySeek5Fwd, keySeek60Fwd))

	dev.AddHandler(b8.BUTTON_5, mod.Handler)
	dev.AddHandler(b8.BUTTON_5, func(b *b8.Button) error {
		dev.Led(b8.LedOn)
		b.WaitForRelease()
		dev.Led(b8.LedOff)
		return nil
	})

	dev.AddHandler(b8.BUTTON_6, b8Handler(
		func(b *b8.Button) error {
			data, err := m.GetProperty("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data.(float64))*1.25))
		},
		func(b *b8.Button) error {
			if err := m.SetProperty("video-align-x", 0); err != nil {
				return err
			}
			if err := m.SetProperty("video-align-y", 0); err != nil {
				return err
			}
			if err := m.SetProperty("video-rotate", 0); err != nil {
				return err
			}
			return m.SetProperty("video-zoom", 0)
		},
		func(b *b8.Button) error {
			data, err := m.GetProperty("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data.(float64))/1.25))
		},
		nil,
	))

	dev.AddHandler(b8.BUTTON_7, b8Handler(
		func(b *b8.Button) error {
			return m.AddProperty("video-align-y", -0.1)
		},
		func(b *b8.Button) error {
			return m.SetProperty("video-align-y", -1)
		},
		func(b *b8.Button) error {
			return m.AddProperty("video-align-y", 0.1)
		},
		func(b *b8.Button) error {
			return m.SetProperty("video-align-y", 1)
		},
	))

	dev.AddHandler(b8.BUTTON_8, b8Handler(
		func(b *b8.Button) error {
			return m.AddProperty("video-align-x", 0.1)
		},
		func(b *b8.Button) error {
			return m.SetProperty("video-align-x", 1)
		},
		func(b *b8.Button) error {
			return m.AddProperty("video-align-x", -0.1)
		},
		func(b *b8.Button) error {
			return m.SetProperty("video-align-x", -1)
		},
	))

	return nil
}
