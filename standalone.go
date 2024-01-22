package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/cli"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/b8r/internal/mpv/server"
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
		Help:    "randomize entries",
	}
	oRecursive = &cli.BoolOption{
		Name:    'r',
		Default: false,
		Help:    "list entries recursively",
	}
	oStart = &cli.BoolOption{
		Name:    's',
		Default: false,
		Help:    "load first entry during startup",
	}
	oInclude = &cli.StringOption{
		Name:    'i',
		Default: ".*",
		Help:    "regex to include all matched entries",
		Metavar: "REGEX",
	}
	oExclude = &cli.StringOption{
		Name:    'e',
		Default: "$^",
		Help:    "regex to exclude all matched entries (applied after -i)",
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
	aEntries = &cli.Argument{
		Name:              "entry",
		Required:          false,
		Remaining:         true,
		Help:              "one or more entries to load (requires a source. if only one, forces -s)",
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
			aEntries,
		},
	}
)

func standalone() {
	check := func(err any) {
		if err != nil {
			log.Fatal("error: ", err)
		}
	}

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
		if aEntries.IsSet() {
			entries = aEntries.GetValues()
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
	defer func() {
		dev.Led(b8.LedOff)
		dev.Close()
	}()

	for i := 0; i < 3; i++ {
		dev.Led(b8.LedFlash)
		time.Sleep(100 * time.Millisecond)
	}

	s := server.New(
		"mpv",
		dev.SerialNumber(),
		true,
		"--fullscreen",
		"--image-display-duration=inf",
		"--loop",
		"--ontop",
		"--really-quiet",
		"--osd-duration=3000",
	)
	check(s.Start())

	c, err := client.NewFromSocket(s.GetSocket())
	check(err)

	go func() {
		check(c.Listen(nil))
	}()

	check(handlers.RegisterMPVHandlers(c, fmute))
	check(handlers.RegisterB8Handlers(dev, c, hsrc))

	if fstart {
		check(handlers.LoadNextFile(c, src))
	}

	go func() {
		check(dev.Listen())
	}()

	check(s.Wait())
}
