package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/mpv"
	"github.com/rafaelmartins/b8r/internal/presets"
	"github.com/rafaelmartins/b8r/internal/source"
)

const (
	keySeek5Fwd  = "RIGHT"
	keySeek5Bwd  = "LEFT"
	keySeek60Fwd = "UP"
	keySeek60Bwd = "DOWN"
)

var (
	fdump   = flag.Bool("dump", false, "dump source entries (after fitlering) and exit")
	ffp     = flag.Bool("fp", false, "use fp source")
	fmute   = flag.Bool("mute", false, "mute by default")
	frand   = flag.Bool("random", false, "randomize items")
	fstart  = flag.Bool("start", false, "load first item during startup")
	ffilter = flag.String("filter", ".*", "regex to filter items")
	fpreset = flag.String("preset", "", "preset to load. overrides most of other options")
	fsn     = flag.String("sn", "", "serial number of device to use")

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

func loadNextFile(m *mpv.MPV, src *source.Source) error {
	next, err := src.Pop()
	if err != nil {
		return err
	}

	file, err := src.GetFile(next)
	if err != nil {
		return err
	}

	if err := m.SetProperty("osd-playing-msg", next); err != nil {
		return err
	}
	if err := m.SetProperty("pause", true); err != nil {
		return err
	}
	if err := m.SetProperty("fullscreen", true); err != nil {
		return err
	}

	waitingPlayback = true
	playing = next

	_, err = m.Command("loadfile", file)
	return err
}

func main() {
	pr, err := presets.New()
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	if *fpreset == "" && len(flag.Args()) > 0 {
		if p := pr.Get(flag.Arg(0)); p != nil {
			*fpreset = flag.Arg(0)
		}
	}

	if *fpreset != "" {
		p := pr.Get(*fpreset)
		if p == nil {
			log.Fatalf("error: preset not found: %s", *fpreset)
		}

		// FIXME
		*ffp = p.Source == "fp"

		*fmute = p.Mute
		*frand = p.Random
		*fstart = p.Start
		*ffilter = p.Filter
	}

	srcName := "local"
	if *ffp {
		srcName = "fp"
	}

	src, err := source.New(srcName, *frand, *ffilter)
	if err != nil {
		log.Fatal(err)
	}

	if srcName == "local" {
		if len(flag.Args()) > 0 {
			src.SetParameter("path", flag.Arg(0))
		} else {
			src.SetParameter("path", ".")
		}
	}

	if *fdump {
		entries, err := src.List()
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			fmt.Println(entry)
		}
		return
	}

	dev, err := b8.GetDevice(*fsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := dev.Open(); err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	for i := 0; i < 3; i++ {
		dev.Led(b8.LedFlash)
		time.Sleep(100 * time.Millisecond)
	}

	m, err := mpv.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := m.AddHandler("playback-restart", func(mp *mpv.MPV, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

		fmt.Printf("Playing: %s\n", playing)

		if err := mp.SetProperty("mute", *fmute); err != nil {
			return err
		}
		if err := mp.SetProperty("video-align-x", 0); err != nil {
			return err
		}
		if err := mp.SetProperty("video-align-y", 0); err != nil {
			return err
		}
		if err := mp.SetProperty("video-rotate", 0); err != nil {
			return err
		}
		if err := mp.SetProperty("video-zoom", 0); err != nil {
			return err
		}
		return mp.SetProperty("pause", false)
	}); err != nil {
		log.Fatal(err)
	}

	for k, v := range keybinds {
		if _, err := m.Command("keybind", k, v); err != nil {
			log.Fatal(err)
		}
	}

	if *fstart {
		if err := loadNextFile(m, src); err != nil {
			log.Fatal(err)
		}
	}

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
			return loadNextFile(m, src)
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
		nil,
	))

	dev.AddHandler(b8.BUTTON_2, b8Handler(
		func(b *b8.Button) error {
			return m.CycleProperty("mute")
		},
		func(b *b8.Button) error {
			return m.CyclePropertyValues("video-rotate", "90", "180", "270", "0")
		},
		func(b *b8.Button) error {
			return m.CycleProperty("fullscreen")
		},
		func(b *b8.Button) error {
			return m.CycleProperty("pause")
		},
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
			if err := m.SetProperty("mute", *fmute); err != nil {
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

	if err := dev.Listen(); err != nil {
		log.Fatal(err)
	}
}
