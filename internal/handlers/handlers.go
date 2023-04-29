package handlers

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
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

func LoadNextFile(m *mpv.MPV, src *source.Source) error {
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

func RegisterB8Handlers(dev *b8.Device, m *mpv.MPV, src *source.Source, exit b8.ButtonHandler) error {
	if dev == nil {
		return errors.New("handlers: missing device")
	}
	if m == nil {
		return errors.New("handlers: missing mpv")
	}

	for k, v := range keybinds {
		if err := m.SetupCommand("keybind", k, v); err != nil {
			return err
		}
	}

	if src != nil {
		dev.AddHandler(b8.BUTTON_1, b8Handler(dev,
			func(b *b8.Button) error {
				if pausedInt, err := m.GetProperty("pause"); err == nil {
					if p, ok := pausedInt.(bool); ok && p {
						if err := m.SetProperty("pause", false); err != nil {
							return err
						}
						return m.SetProperty("fullscreen", true)
					}
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
				if cntInt, err := m.GetProperty("playlist-count"); err == nil {
					if cnt, ok := cntInt.(float64); ok && cnt == 0 && exit != nil {
						return exit(b)
					}
				}
				_, err := m.Command("stop")
				return err
			},
			nil,
		))
	} else {
		dev.AddHandler(b8.BUTTON_1, b8Handler(dev,
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
			nil,
			func(b *b8.Button) error {
				if exit != nil {
					return exit(b)
				}
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
			_, err := m.Command("vf", "toggle", "hflip")
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
			data, err := m.GetProperty("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data.(float64))*1.25))
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
			data, err := m.GetProperty("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data.(float64))/1.25))
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

func RegisterMPVHandlers(m *mpv.MPV, mute bool) error {
	if m == nil {
		return errors.New("handlers: missing mpv")
	}

	return m.AddHandler("playback-restart", func(mp *mpv.MPV, event string, data map[string]interface{}) error {
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
}
