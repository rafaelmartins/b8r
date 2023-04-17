package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/cli"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv"
	"github.com/rafaelmartins/b8r/internal/presets"
	"github.com/rafaelmartins/b8r/internal/source"
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
		Remaining:         true,
		Help:              "item or collection to load (requires a source. if item, forces -s, ignores -i -e)",
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

	exit            = false
	waitingPlayback = false
	playing         = ""
)

func loadNextFile(m *mpv.MPV, src *source.Source) error {
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
	entries := []string{}
	fmute := oMute.Default
	frand := oRand.Default
	frecursive := oRecursive.Default
	fstart := oStart.Default
	finclude := oInclude.Default
	fexclude := oExclude.Default

	if p := pr.Get(aPresetOrSource.GetValue()); p != nil {
		srcName = p.Source
		if p.Entries != nil {
			entries = p.Entries
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
			entries = aEntry.GetValues()
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

	src, err := source.New(srcName)
	check(err)

	singleEntry, err := src.SetEntries(entries, frecursive, frand, finclude, fexclude)
	check(err)

	hsrc := src
	if singleEntry {
		hsrc = nil
		fstart = true
		exit = true
	}

	if oDump.GetValue() {
		check(src.ForEachEntry(func(e string) {
			fmt.Println(e)
		}))
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

	check(m.AddHandler("playback-restart", func(mp *mpv.MPV, event string, data map[string]interface{}) error {
		if !waitingPlayback {
			return nil
		}
		waitingPlayback = false

		fmt.Printf("Playing: %s\n", playing)

		if err := mp.SetProperty("mute", fmute); err != nil {
			return err
		}
		if _, err := m.Command("vf", "remove", "hflip"); err != nil {
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

	check(m.Start())

	check(handlers.RegisterHandlers(dev, m, hsrc, loadNextFile, func(b *b8.Button) error {
		exit = true
		_, err := m.Command("quit")
		return err
	}))

	if fstart {
		check(loadNextFile(m, src))
	}

	go func() {
		check(dev.Listen())
	}()

	check(m.Wait())
	for {
		if exit {
			// try to ensure that led is off
			dev.Led(b8.LedOff)
			return
		}
		check(m.Start())
		check(m.Wait())
		time.Sleep(100 * time.Millisecond)
	}
}
