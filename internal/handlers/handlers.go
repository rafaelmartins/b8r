package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/b8r/internal/source"
	"github.com/rafaelmartins/b8r/internal/utils"
	"github.com/rafaelmartins/octokeyz/go/octokeyz"
)

var (
	mod octokeyz.Modifier

	keySeek5Fwd  = []interface{}{"osd-bar", "seek", 5}
	keySeek5Bwd  = []interface{}{"osd-bar", "seek", -5}
	keySeek60Fwd = []interface{}{"osd-bar", "seek", 60}
	keySeek60Bwd = []interface{}{"osd-bar", "seek", -60}

	waitingPlayback = false
	current         = ""
	next            = ""
	idxTotal        = 0
	idxCurrent      = 0
)

func octokeyzHandler(dev *octokeyz.Device, short octokeyz.ButtonHandler, long octokeyz.ButtonHandler, modShort octokeyz.ButtonHandler, modLong octokeyz.ButtonHandler) octokeyz.ButtonHandler {
	return func(b *octokeyz.Button) error {
		lpDuration := 400 * time.Millisecond
		done := make(chan struct{})

		go func() {
			ticker := time.NewTicker(lpDuration)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if dev != nil {
						dev.Led(octokeyz.LedFlash)
					}
					return
				}
			}
		}()

		pressed := mod.Pressed()
		duration := b.WaitForRelease()

		close(done)

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

func octokeyzHoldKeyHandler(m *client.MpvIpcClient, cmd []interface{}, modCmd []interface{}) octokeyz.ButtonHandler {
	return func(b *octokeyz.Button) error {
		arDelay := 200 * time.Millisecond
		arRate := (1 * time.Second) / 40

		c := cmd
		if mod.Pressed() {
			c = modCmd
		}

		if _, err := m.Command(c...); err != nil && !errors.Is(err, client.ErrMpvCommand) {
			return err
		}
		time.Sleep(arDelay)

		done := make(chan struct{})

		go func() {
			ticker := time.NewTicker(arRate)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if _, err := m.Command(c...); err != nil && !errors.Is(err, client.ErrMpvCommand) {
						log.Printf("error: %s", err) // FIXME
						return
					}
				}
			}
		}()

		b.WaitForRelease()
		close(done)

		return nil
	}
}

func LoadNextFile(m *client.MpvIpcClient, src *source.Source) error {
	if m == nil {
		return errors.New("handlers: missing mpv")
	}
	if src == nil {
		return errors.New("handlers: missing source")
	}

	current = next

	var err error
	if current == "" {
		current, err = src.NextEntry()
		if err != nil {
			return err
		}
	}

	idxTotal = src.GetEntriesCount()
	idxCurrent = idxTotal - src.GetCurrentEntriesCount()

	next, err = src.NextEntry()
	if err != nil {
		return err
	}

	file, err := src.GetFile(current)
	if err != nil {
		return err
	}

	if err := m.SetProperty("osd-playing-msg", filepath.ToSlash(current)); err != nil {
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

	_, err = m.Command("loadfile", file)
	return err
}

func RegisterOctokeyzHandlers(dev *octokeyz.Device, m *client.MpvIpcClient, src *source.Source) error {
	if dev == nil {
		return errors.New("handlers: missing device")
	}
	if m == nil {
		return errors.New("handlers: missing mpv")
	}

	if src != nil {
		if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine3, fmt.Sprintf("Source: %s", src.GetBackendName()), octokeyz.DisplayLineAlignLeft)); err != nil {
			return err
		}
		if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine4, fmt.Sprintf("0 / %d", src.GetEntriesCount()), octokeyz.DisplayLineAlignLeft)); err != nil {
			return err
		}

		var err error
		next, err = src.NextEntry()
		if err != nil {
			return err
		}

		if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine7, fmt.Sprintf("N: %s", next), octokeyz.DisplayLineAlignLeft)); err != nil {
			return err
		}

		dev.AddHandler(octokeyz.BUTTON_1, octokeyzHandler(dev,
			func(b *octokeyz.Button) error {
				if paused, err := m.GetPropertyBool("pause"); err == nil && paused {
					if err := m.SetProperty("pause", false); err != nil {
						return err
					}
					return m.SetProperty("fullscreen", true)
				}
				return LoadNextFile(m, src)
			},
			func(b *octokeyz.Button) error {
				if err := m.SetProperty("pause", true); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", false)
			},
			func(b *octokeyz.Button) error {
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
		dev.AddHandler(octokeyz.BUTTON_1, octokeyzHandler(dev,
			func(b *octokeyz.Button) error {
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
			func(b *octokeyz.Button) error {
				_, err := m.Command("quit")
				return err
			},
			nil,
		))
	}

	dev.AddHandler(octokeyz.BUTTON_2, octokeyzHandler(dev,
		func(b *octokeyz.Button) error {
			return m.CycleProperty("mute")
		},
		func(b *octokeyz.Button) error {
			return m.CyclePropertyValues("video-rotate", "90", "180", "270", "0")
		},
		func(b *octokeyz.Button) error {
			rotate, err := m.GetPropertyInt("video-dec-params/rotate")
			if err != nil {
				if errors.Is(err, client.ErrMpvPropertyUnavailable) {
					return nil
				}
				return err
			}

			flip := "hflip"
			if rotate == 90 || rotate == 270 {
				flip = "vflip"
			}

			_, err = m.Command("vf", "toggle", flip)
			return err
		},
		func(b *octokeyz.Button) error {
			if _, err := m.Command("script-message", "osc-visibility", "always", "true"); err != nil {
				return err
			}

			go func() {
				time.Sleep(5 * time.Second)
				m.Command("script-message", "osc-visibility", "auto", "true")
			}()

			return nil
		},
	))

	dev.AddHandler(octokeyz.BUTTON_3, octokeyzHoldKeyHandler(m, keySeek5Bwd, keySeek60Bwd))

	dev.AddHandler(octokeyz.BUTTON_4, octokeyzHoldKeyHandler(m, keySeek5Fwd, keySeek60Fwd))

	dev.AddHandler(octokeyz.BUTTON_5, mod.Handler)
	dev.AddHandler(octokeyz.BUTTON_5, func(b *octokeyz.Button) error {
		return dev.Led(octokeyz.LedFlash)
	})

	dev.AddHandler(octokeyz.BUTTON_6, octokeyzHandler(dev,
		func(b *octokeyz.Button) error {
			data, err := m.GetPropertyFloat64("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data)*1.25))
		},
		func(b *octokeyz.Button) error {
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
		func(b *octokeyz.Button) error {
			data, err := m.GetPropertyFloat64("video-zoom")
			if err != nil {
				return err
			}
			return m.SetProperty("video-zoom", math.Log2(math.Pow(2, data)/1.25))
		},
		nil,
	))

	dev.AddHandler(octokeyz.BUTTON_7, octokeyzHandler(dev,
		func(b *octokeyz.Button) error {
			return m.AddProperty("video-align-y", -0.1)
		},
		func(b *octokeyz.Button) error {
			return m.SetProperty("video-align-y", -1)
		},
		func(b *octokeyz.Button) error {
			return m.AddProperty("video-align-y", 0.1)
		},
		func(b *octokeyz.Button) error {
			return m.SetProperty("video-align-y", 1)
		},
	))

	dev.AddHandler(octokeyz.BUTTON_8, octokeyzHandler(dev,
		func(b *octokeyz.Button) error {
			return m.AddProperty("video-align-x", 0.1)
		},
		func(b *octokeyz.Button) error {
			return m.SetProperty("video-align-x", 1)
		},
		func(b *octokeyz.Button) error {
			return m.AddProperty("video-align-x", -0.1)
		},
		func(b *octokeyz.Button) error {
			return m.SetProperty("video-align-x", -1)
		},
	))

	return nil
}

func RegisterMPVHandlers(dev *octokeyz.Device, m *client.MpvIpcClient, mute bool, withNext bool) error {
	if dev == nil {
		return errors.New("handlers: missing device")
	}
	if m == nil {
		return errors.New("handlers: missing mpv ipc client")
	}

	m.AddHandler("playback-restart", func(mp *client.MpvIpcClient, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

		if err := mp.SetProperty("mute", mute); err != nil {
			return err
		}
		if err := mp.SetProperty("pause", false); err != nil {
			return err
		}

		fmt.Printf("Playing: %s\n", current)
		if withNext {
			if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine4, fmt.Sprintf("%d / %d", idxCurrent, idxTotal), octokeyz.DisplayLineAlignLeft)); err != nil {
				return err
			}
		}
		if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine6, fmt.Sprintf("C: %s", current), octokeyz.DisplayLineAlignLeft)); err != nil {
			return err
		}
		if withNext {
			if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine7, fmt.Sprintf("N: %s", next), octokeyz.DisplayLineAlignLeft)); err != nil {
				return err
			}
		}
		return mp.SetProperty("force-media-title", current)
	})

	m.AddHandler("end-file", func(mp *client.MpvIpcClient, event string, data map[string]interface{}) error {
		if data["reason"].(string) == "stop" {
			return utils.IgnoreDisplayMissing(dev.DisplayClearLine(octokeyz.DisplayLine6))
		}
		return nil
	})

	return nil
}
