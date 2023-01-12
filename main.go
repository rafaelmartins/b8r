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
	fdump   = flag.Bool("dump", false, "dump source entries (after fitlering) and exit")
	ffp     = flag.Bool("fp", false, "use fp source")
	fmute   = flag.Bool("mute", false, "mute by default")
	frand   = flag.Bool("random", false, "randomize items")
	fstart  = flag.Bool("start", false, "load first item during startup")
	ffilter = flag.String("filter", ".*", "regex to filter items")
	fsn     = flag.String("sn", "", "serial number of device to use")

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

	waitingPlayback := false
	playing := ""

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

	loadFile := func() error {
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

	if *fstart {
		if err := loadFile(); err != nil {
			log.Fatal(err)
		}
	}

	dev.AddHandler(b8.BUTTON_1, func(b *b8.Button) error {
		pressed := mod.Pressed()
		if b.WaitForRelease() < 400*time.Millisecond {
			paused := false
			if pausedInt, err := m.GetProperty("pause"); err == nil {
				if p, ok := pausedInt.(bool); ok {
					paused = p
				}
			}
			if pressed || paused {
				if err := m.SetProperty("pause", false); err != nil {
					return err
				}
				return m.SetProperty("fullscreen", true)
			}

			return loadFile()
		}

		if pressed {
			if err := m.SetProperty("pause", true); err != nil {
				return err
			}
			err := m.SetProperty("fullscreen", false)
			return err
		}
		_, err := m.Command("stop")
		return err
	})

	dev.AddHandler(b8.BUTTON_2, func(b *b8.Button) error {
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

	dev.AddHandler(b8.BUTTON_3, func(b *b8.Button) error {
		if mod.Pressed() {
			return holdKey(m, b, keySeek60Bwd)
		}
		return holdKey(m, b, keySeek5Bwd)
	})

	dev.AddHandler(b8.BUTTON_4, func(b *b8.Button) error {
		if mod.Pressed() {
			return holdKey(m, b, keySeek60Fwd)
		}
		return holdKey(m, b, keySeek5Fwd)
	})

	dev.AddHandler(b8.BUTTON_5, mod.Handler)
	dev.AddHandler(b8.BUTTON_5, func(b *b8.Button) error {
		dev.Led(b8.LedOn)
		b.WaitForRelease()
		dev.Led(b8.LedOff)
		return nil
	})

	dev.AddHandler(b8.BUTTON_6, func(b *b8.Button) error {
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

	dev.AddHandler(b8.BUTTON_7, func(b *b8.Button) error {
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

	dev.AddHandler(b8.BUTTON_8, func(b *b8.Button) error {
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

	if err := dev.Listen(); err != nil {
		log.Fatal(err)
	}
}
