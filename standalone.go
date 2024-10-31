package main

import (
	"fmt"
	"strings"

	"github.com/rafaelmartins/b8r/internal/cleanup"
	"github.com/rafaelmartins/b8r/internal/cli"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/b8r/internal/mpv/server"
	"github.com/rafaelmartins/b8r/internal/presets"
	"github.com/rafaelmartins/b8r/internal/source"
	"github.com/rafaelmartins/b8r/internal/utils"
	"github.com/rafaelmartins/octokeyz/go/octokeyz"
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
	oEvents = &cli.BoolOption{
		Name:    'e',
		Default: false,
		Help:    "dump mpv events (useful for development)",
	}
	oInclude = &cli.StringOption{
		Name:    'i',
		Default: ".*",
		Help:    "regex to include all matched entries",
		Metavar: "REGEX",
	}
	oExclude = &cli.StringOption{
		Name:    'x',
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
			devs, err := octokeyz.Enumerate()
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
			oEvents,
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
	defer cleanup.Cleanup()

	cCli.Parse()

	pr, err := presets.New()
	cleanup.Check(err)

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
	cleanup.Check(err)

	singleEntry, err := src.SetEntries(entries, frecursive, frand, finclude, fexclude)
	cleanup.Check(err)

	hsrc := src
	if singleEntry {
		hsrc = nil
		fstart = true
	}

	if oDump.GetValue() {
		cleanup.Check(src.ForEachEntry(func(e string) {
			fmt.Println(e)
		}))
		return
	}

	dev, err := octokeyz.GetDevice(oSerialNumber.GetValue())
	cleanup.Check(err)

	cleanup.Check(dev.Open())
	cleanup.Register(dev)

	cleanup.Check(utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine1, "b8r", octokeyz.DisplayLineAlignCenter)))
	cleanup.Check(utils.LedFlash3Times(dev))

	s := server.New(
		"mpv",
		dev.SerialNumber(),
		true,
		"--fullscreen",
		"--image-display-duration=inf",
		"--loop",
		"--really-quiet",
		"--osd-duration=3000",
	)
	cleanup.Check(s.Start())

	c, err := client.NewFromSocket(s.GetSocket(), oEvents.GetValue())
	cleanup.Check(err)

	go func() {
		cleanup.Check(c.Listen(nil))
	}()

	cleanup.Check(handlers.RegisterMPVHandlers(dev, c, fmute, hsrc != nil))
	cleanup.Check(handlers.RegisterOctokeyzHandlers(dev, c, hsrc))

	if fstart {
		cleanup.Check(handlers.LoadNextFile(c, src))
	}

	go func() {
		cleanup.Check(dev.Listen(nil))
	}()

	cleanup.Check(s.Wait())
}
