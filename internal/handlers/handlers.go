package handlers

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
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

	waitingPlayback = false
	playing         = ""
)

func b8Handler(dev *b8.Device, short b8.ButtonHandler, long b8.ButtonHandler, modShort b8.ButtonHandler, modLong b8.ButtonHandler) b8.ButtonHandler {
	return func(b *b8.Button) error {
		lpDuration := 400 * time.Millisecond
		done := make(chan bool)

		go func() {
			ticker := time.NewTicker(lpDuration)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if dev != nil {
						dev.Led(b8.LedFlash)
					}
					return
				}
			}
		}()

		pressed := mod.Pressed()
		duration := b.WaitForRelease()

		select {
		case done <- true:
		default:
		}

		if duration < lpDuration {
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
			if modShort != nil {
				return modShort(b)
			}
			return nil
		}
		if long != nil {
			return long(b)
		}
		if short != nil {
			return short(b)
		}
		return nil
	}
}

func b8HoldKeyHandler(m *client.MpvIpcClient, key string, modKey string) b8.ButtonHandler {
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

func LoadNextFile(m *client.MpvIpcClient, src *source.Source) error {
	if m == nil {
		return errors.New("handlers: missing mpv")
	}
	if src == nil {
		return errors.New("handlers: missing source")
	}

	next, err := src.NextEntry()
	if err != nil {
		return err
	}

	file, err := src.GetFile(next)
	if err != nil {
		return err
	}

	if err := m.SetProperty("osd-playing-msg", filepath.ToSlash(next)); err != nil {
		return err
	}
	if err := m.SetProperty("pause", true); err != nil {
		return err
	}
	if err := m.SetProperty("fullscreen", true); err != nil {
		return err
	}
	if _, err := m.Command("vf", "remove", "hflip"); err != nil {
		return err
	}
	if _, err := m.Command("vf", "remove", "vflip"); err != nil {
		return err
	}
	if err := m.SetProperty("video-align-x", 0); err != nil {
		return err
	}
	if err := m.SetProperty("video-align-y", 0); err != nil {
		return err
	}
	if err := m.SetProperty("video-rotate", 0); err != nil {
		return err
	}
	if err := m.SetProperty("video-zoom", 0); err != nil {
		return err
	}

	waitingPlayback = true
	playing = next

	_, err = m.Command("loadfile", file)
	return err
}

func RegisterB8Handlers(dev *b8.Device, m *client.MpvIpcClient, src *source.Source) error {
	if dev == nil {
		return errors.New("handlers: missing device")
	}
	if m == nil {
		return errors.New("handlers: missing mpv")
	}

	for k, v := range keybinds {
		if _, err := m.Command("keybind", k, v); err != nil {
			return err
		}
	}

	if src != nil {
		dev.AddHandler(b8.BUTTON_1, b8Handler(dev,
			func(b *b8.Button) error {
				if paused, err := m.GetPropertyBool("pause"); err == nil && paused {
					if err := m.SetProperty("pause", false); err != nil {
						return err
					}
					return m.SetProperty("fullscreen", true)
				}
				return LoadNextFile(m, src)
			},
			func(b *b8.Button) error {
				if err := m.SetProperty("pause", true); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", false)
			},
			func(b *b8.Button) error {
				if cnt, err := m.GetPropertyInt("playlist-count"); err == nil && int(cnt) == 0 {
					_, err := m.Command("quit")
					return err
				}
				_, err := m.Command("stop")
				return err
			},
			nil,
		))
	} else {
		dev.AddHandler(b8.BUTTON_1, b8Handler(dev,
			func(b *b8.Button) error {
				if paused, err := m.GetPropertyBool("pause"); err == nil {
					if paused {
						if err := m.SetProperty("fullscreen", true); err != nil {
							return err
						}
						return m.SetProperty("pause", false)
					}
				} else {
					return err
				}

				if fs, err := m.GetPropertyBool("fullscreen"); err == nil {
					if fs {
						if err := m.SetProperty("pause", true); err != nil {
							return err
						}
					}
					return m.SetProperty("fullscreen", !fs)
				} else {
					return err
				}
			},
			nil,
			func(b *b8.Button) error {
				_, err := m.Command("quit")
				return err
			},
			nil,
		))
	}

	dev.AddHandler(b8.BUTTON_2, b8Handler(dev,
		func(b *b8.Button) error {
			return m.CycleProperty("mute")
		},
		func(b *b8.Button) error {
			return m.CyclePropertyValues("video-rotate", "90", "180", "270", "0")
		},
		func(b *b8.Button) error {
			rotate, err := m.GetPropertyInt("video-dec-params/rotate")
			if err != nil {
				return err
			}

			flip := "hflip"
			if rotate == 90 || rotate == 270 {
				flip = "vflip"
			}

			_, err = m.Command("vf", "toggle", flip)
			return err
		},
		nil,
	))

	dev.AddHandler(b8.BUTTON_3, b8HoldKeyHandler(m, keySeek5Bwd, keySeek60Bwd))

	dev.AddHandler(b8.BUTTON_4, b8HoldKeyHandler(m, keySeek5Fwd, keySeek60Fwd))

	dev.AddHandler(b8.BUTTON_5, mod.Handler)
	dev.AddHandler(b8.BUTTON_5, func(b *b8.Button) error {
		return dev.Led(b8.LedFlash)
	})

	dev.AddHandler(b8.BUTTON_6, b8Handler(dev,
		func(b *b8.Button) error {
			data, err := m.GetPropertyFloat64("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data)*1.25))
		},
		func(b *b8.Button) error {
			if _, err := m.Command("vf", "remove", "hflip"); err != nil {
				return err
			}
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
			data, err := m.GetPropertyFloat64("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data)/1.25))
		},
		nil,
	))

	dev.AddHandler(b8.BUTTON_7, b8Handler(dev,
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

	dev.AddHandler(b8.BUTTON_8, b8Handler(dev,
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

func RegisterMPVHandlers(m *client.MpvIpcClient, mute bool) error {
	if m == nil {
		return errors.New("handlers: missing mpv ipc client")
	}

	m.AddHandler("playback-restart", func(mp *client.MpvIpcClient, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

		fmt.Printf("Playing: %s\n", playing)

		if err := mp.SetProperty("mute", mute); err != nil {
			return err
		}
		return mp.SetProperty("pause", false)
	})

	return nil
}
