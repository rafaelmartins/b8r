package main

import (
	"flag"
	"fmt"
	"log"
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
	ffp     = flag.Bool("fp", false, "use fp source")
	frand   = flag.Bool("random", false, "randomize items")
	fmute   = flag.Bool("mute", false, "mute by default")
	ffilter = flag.String("filter", ".*", "regex to filter items")

	mod b8.Modifier

	keybinds = map[string]string{
		keySeek5Fwd:  "seek 5",
		keySeek5Bwd:  "seek -5",
		keySeek60Fwd: "seek 60",
		keySeek60Bwd: "seek -60",
	}
)

func holdKey(m *mpv.MPV, b *b8.Button, key string) error {
	_, err := m.Command("keydown", key)
	if err != nil {
		return err
	}
	b.WaitForRelease()
	_, err = m.Command("keyup", key)
	return err
}

func main() {
	flag.Parse()

	dev, err := b8.ListDevices()
	if err != nil {
		log.Fatal(err)
	}

	if len(dev) != 1 {
		log.Fatal("b8: device not found")
	}

	d := dev[0]
	if err := d.Open(); err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	for i := 0; i < 3; i++ {
		d.Led(b8.LedFlash)
		time.Sleep(100 * time.Millisecond)
	}

	srcName := "folder"
	if *ffp {
		srcName = "fp"
	}

	src, err := source.New(srcName, *frand, *ffilter)
	if err != nil {
		log.Fatal(err)
	}

	if srcName == "folder" {
		if len(flag.Args()) > 0 {
			src.SetParameter("path", flag.Arg(0))
		} else {
			src.SetParameter("path", ".")
		}
	}

	m, err := mpv.New()
	if err != nil {
		log.Fatal(err)
	}

	waitingPlayback := false

	if err := m.AddHandler("playback-restart", func(mp *mpv.MPV, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

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

	paused := false

	d.AddHandler(b8.BUTTON_1, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			if pressed || paused {
				paused = false
				if err := m.SetProperty("pause", false); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", true)
			}

			next, err := src.Pop()
			if err != nil {
				return err
			}

			fmt.Printf("Playing: %s\n", next)

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

			_, err = m.Command("loadfile", file)
			return err
		}

		if pressed {
			if err := m.SetProperty("pause", true); err != nil {
				return err
			}
			err := m.SetProperty("fullscreen", false)
			paused = true
			return err
		}
		_, err := m.Command("stop")
		return err
	})

	d.AddHandler(b8.BUTTON_2, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			if pressed {
				return m.CyclePropertyValues("video-rotate", "90", "180", "270", "0")

			}
			return m.CycleProperty("mute")
		}
		if pressed {
			return m.CycleProperty("pause")
		}
		return m.CycleProperty("fullscreen")
	})

	d.AddHandler(b8.BUTTON_3, func(b *b8.Button) error {
		if mod.Pressed() {
			return holdKey(m, b, keySeek60Bwd)
		}
		return holdKey(m, b, keySeek5Bwd)
	})

	d.AddHandler(b8.BUTTON_4, func(b *b8.Button) error {
		if mod.Pressed() {
			return holdKey(m, b, keySeek60Fwd)
		}
		return holdKey(m, b, keySeek5Fwd)
	})

	d.AddHandler(b8.BUTTON_5, mod.Handler)
	d.AddHandler(b8.BUTTON_5, func(b *b8.Button) error {
		d.Led(b8.LedOn)
		b.WaitForRelease()
		d.Led(b8.LedOff)
		return nil
	})

	d.AddHandler(b8.BUTTON_6, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			data, err := m.GetProperty("video-zoom")
			if err != nil {
				return err
			}
			cur := math.Pow(2, data.(float64))
			if pressed {
				return m.SetProperty("video-zoom", math.Log2(cur/1.25))
			}
			return m.SetProperty("video-zoom", math.Log2(cur*1.25))
		}

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
	})

	d.AddHandler(b8.BUTTON_7, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			if pressed {
				return m.AddProperty("video-align-y", 0.1)
			}
			return m.AddProperty("video-align-y", -0.1)
		}
		if pressed {
			return m.SetProperty("video-align-y", 1)
		}
		return m.SetProperty("video-align-y", -1)
	})

	d.AddHandler(b8.BUTTON_8, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			if pressed {
				return m.AddProperty("video-align-x", -0.1)
			}
			return m.AddProperty("video-align-x", 0.1)
		}
		if pressed {
			return m.SetProperty("video-align-x", -1)
		}
		return m.SetProperty("video-align-x", 1)
	})

	if err := d.Listen(); err != nil {
		log.Fatal(err)
	}
}
