package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/cli"
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
	oDump = &cli.BoolOption{
		Name:    'd',
		Default: false,
		Help:    "dump source entries (after filtering) and exit",
	}
	oMute = &cli.BoolOption{
		Name:    'm',
		Default: false,
		Help:    "mute by default",
	}
	oRand = &cli.BoolOption{
		Name:    'z',
		Default: false,
		Help:    "randomize items",
	}
	oRecursive = &cli.BoolOption{
		Name:    'r',
		Default: false,
		Help:    "list items recursively",
	}
	oStart = &cli.BoolOption{
		Name:    's',
		Default: false,
		Help:    "load first item during startup",
	}
	oInclude = &cli.StringOption{
		Name:    'i',
		Default: ".*",
		Help:    "regex to include all matched items",
		Metavar: "REGEX",
	}
	oExclude = &cli.StringOption{
		Name:    'e',
		Default: "$^",
		Help:    "regex to exclude all matched items (applied after -i)",
		Metavar: "REGEX",
	}
	oSerialNumber = &cli.StringOption{
		Name:    'n',
		Default: "",
		Help:    "serial number of device to use",
		Metavar: "SERIAL_NUMBER",
		CompletionHandler: func(cur string) []string {
			devs, err := b8.Enumerate()
			if err != nil {
				return nil
			}
			rv := []string{}
			for _, d := range devs {
				if sn := d.SerialNumber(); strings.HasPrefix(sn, cur) {
					rv = append(rv, sn)
				}
			}
			return rv
		},
	}
	aPresetOrSource = &cli.Argument{
		Name:     "preset-or-source",
		Required: true,
		Help:     "a preset or a source to use",
		CompletionHandler: func(prev string, cur string) []string {
			pr, _ := presets.New()
			rv := []string{}
			for _, c := range append(source.List(), pr.List()...) {
				if strings.HasPrefix(c, cur) {
					rv = append(rv, c)
				}
			}
			return rv
		},
	}
	aEntry = &cli.Argument{
		Name:              "entry",
		Required:          false,
		Help:              "a single entry to load (requires a source)",
		CompletionHandler: source.CompletionHandler,
	}

	cCli = &cli.Cli{
		Help: `¯\_(ツ)_/¯`,
		Options: []cli.Option{
			oDump,
			oMute,
			oRand,
			oRecursive,
			oStart,
			oInclude,
			oExclude,
			oSerialNumber,
		},
		Arguments: []*cli.Argument{
			aPresetOrSource,
			aEntry,
		},
	}

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

func check(err any) {
	if err != nil {
		log.Fatal("error: ", err)
	}
}

func main() {
	cCli.Parse()

	pr, err := presets.New()
	check(err)

	srcName := ""
	entry := ""
	fmute := oMute.Default
	frand := oRand.Default
	frecursive := oRecursive.Default
	fstart := oStart.Default
	finclude := oInclude.Default
	fexclude := oExclude.Default

	if p := pr.Get(aPresetOrSource.GetValue()); p != nil {
		srcName = p.Source
		if p.Entry != nil {
			entry = *p.Entry
		}
		if p.Mute != nil {
			fmute = *p.Mute
		}
		if p.Random != nil {
			frand = *p.Random
		}
		if p.Recursive != nil {
			frecursive = *p.Recursive
		}
		if p.Start != nil {
			fstart = *p.Start
		}
		if p.Include != nil {
			finclude = *p.Include
		}
		if p.Exclude != nil {
			fexclude = *p.Exclude
		}
	} else {
		srcName = aPresetOrSource.GetValue()
		if aEntry.IsSet() {
			entry = aEntry.GetValue()
		}
	}

	if oMute.IsSet() {
		fmute = oMute.GetValue()
	}
	if oRand.IsSet() {
		frand = oRand.GetValue()
	}
	if oRecursive.IsSet() {
		frecursive = oRecursive.GetValue()
	}
	if oStart.IsSet() {
		fstart = oStart.GetValue()
	}
	if oInclude.IsSet() {
		finclude = oInclude.GetValue()
	}
	if oExclude.IsSet() {
		fexclude = oExclude.GetValue()
	}

	src, err := source.New(srcName, frand, finclude, fexclude)
	check(err)

	check(src.SetParameter("entry", entry))
	check(src.SetParameter("recursive", frecursive))

	if oDump.GetValue() {
		entries, err := src.List()
		check(err)

		for _, e := range entries {
			fmt.Println(e)
		}
		return
	}

	dev, err := b8.GetDevice(oSerialNumber.GetValue())
	check(err)

	check(dev.Open())
	defer dev.Close()

	for i := 0; i < 3; i++ {
		dev.Led(b8.LedFlash)
		time.Sleep(100 * time.Millisecond)
	}

	m, err := mpv.New(
		"mpv",
		dev.SerialNumber(),
		true,
		"--fullscreen",
		"--image-display-duration=inf",
		"--loop",
		"--ontop",
		"--really-quiet",
	)
	check(err)

	go func() {
		check(m.Wait())

		for {
			check(m.Start())
			check(m.Wait())
			time.Sleep(100 * time.Millisecond)
		}
	}()

	check(m.AddHandler("playback-restart", func(mp *mpv.MPV, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

		fmt.Printf("Playing: %s\n", playing)

		if err := mp.SetProperty("mute", fmute); err != nil {
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
	}))

	for k, v := range keybinds {
		_, err := m.Command("keybind", k, v)
		check(err)
	}

	if fstart {
		check(loadNextFile(m, src))
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
			if err := m.SetProperty("mute", fmute); err != nil {
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

	check(dev.Listen())
}
