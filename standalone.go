package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/rafaelmartins/b8r/internal/androidtv"
	"github.com/rafaelmartins/b8r/internal/cleanup"
	"github.com/rafaelmartins/b8r/internal/cli"
	"github.com/rafaelmartins/b8r/internal/config"
	"github.com/rafaelmartins/b8r/internal/dataset"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/b8r/internal/mpv/server"
	"github.com/rafaelmartins/b8r/internal/source"
	"github.com/rafaelmartins/b8r/internal/utils"
	"rafaelmartins.com/p/octokeyz"
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
		Help:    "dump mpv/android-tv events (useful for development)",
	}
	oPairAndroidTv = &cli.BoolOption{
		Name:    'p',
		Default: false,
		Help:    "pair with android-tv as remote control and exit",
	}
	oMuteAndroidTv = &cli.BoolOption{
		Name:    'u',
		Default: false,
		Help:    "mute/unmute android-tv device",
	}
	oPauseAndroidTv = &cli.BoolOption{
		Name:    'a',
		Default: false,
		Help:    "pause/unpause android-tv device",
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
	oTable = &cli.StringOption{
		Name:    't',
		Default: "",
		Help:    "create and load table from current source/preset",
		Metavar: "TABLE",
	}
	aPresetOrSourceOrTable = &cli.Argument{
		Name:     "preset-or-source-or-table",
		Required: false,
		Help:     "a preset or a source or a table to load from",
		CompletionHandler: func(prev string, cur string) []string {
			c, err := config.New()
			if err != nil {
				return nil
			}

			d, err := c.GetTablesDirectory()
			if err != nil {
				return nil
			}

			rv := []string{}
			for _, c := range append(append(source.List(), c.ListPresets()...), dataset.ListTables(d)...) {
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
			oPairAndroidTv,
			oMuteAndroidTv,
			oPauseAndroidTv,
			oInclude,
			oExclude,
			oSerialNumber,
			oTable,
		},
		Arguments: []*cli.Argument{
			aPresetOrSourceOrTable,
			aEntries,
		},
	}
)

func standalone() {
	defer cleanup.Cleanup()

	bi, ok := debug.ReadBuildInfo()
	if ok {
		cCli.Version = bi.Main.Version
	}
	cCli.Parse()

	conf, err := config.New()
	cleanup.Check(err)

	if oPairAndroidTv.GetValue() {
		if conf.AndroidTv.Host == "" {
			cleanup.Check("android-tv host not configured")
		}

		certFile, exists := conf.GetAndroidTvCertificate()
		if exists {
			cleanup.Check("android-tv certificate already exists, please remove it to pair again")
		}

		cert, err := androidtv.CreateCertificate(certFile)
		cleanup.Check(err)

		atv, err := androidtv.NewPairing(conf.AndroidTv.Host, cert, oEvents.GetValue())
		cleanup.Check(err)
		cleanup.Register(atv)

		atv.SecretCallback = func() (string, error) {
			fmt.Print("Please enter the code displayed on TV: ")
			rv, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			if s := strings.TrimSpace(rv); len(s) == 6 {
				return s, nil
			}
			return "", fmt.Errorf("invalid code, must be 6 hexadecimal digits")
		}

		atv.CompleteCallback = func() {
			fmt.Println("Paired successfully")
			cleanup.Exit(0)
		}

		cleanup.Check(atv.Request())
		cleanup.Check(atv.Listen())
		return
	}

	if !aPresetOrSourceOrTable.IsSet() {
		cCli.Usage(false, "PRESET_OR_SOURCE_OR_TABLE argument required unless setting `-p'")
		cleanup.Exit(1)
	}

	entries := []string{}
	fmute := oMute.Default
	frand := oRand.Default
	frecursive := oRecursive.Default
	fstart := oStart.Default
	finclude := oInclude.Default
	fexclude := oExclude.Default

	srcName := ""
	tableName := ""
	tableCreate := false
	if d, err := conf.GetTablesDirectory(); err == nil && dataset.TableExists(d, aPresetOrSourceOrTable.GetValue()) {
		tableName = aPresetOrSourceOrTable.GetValue()
		src, err := dataset.TableSource(d, aPresetOrSourceOrTable.GetValue())
		cleanup.Check(err)
		srcName = src
	} else if p := conf.GetPreset(aPresetOrSourceOrTable.GetValue()); p != nil {
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
		srcName = aPresetOrSourceOrTable.GetValue()
		if aEntries.IsSet() {
			entries = aEntries.GetValues()
		}
	}

	if tableName == "" && oTable.GetValue() != "" {
		tableName = oTable.GetValue()
		tableCreate = true
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

	tableDir, err := conf.GetTablesDirectory()
	cleanup.Check(err)

	singleEntry, err := src.SetEntries(tableDir, tableName, tableCreate, entries, frecursive, frand, finclude, fexclude)
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

	sn := conf.Standalone.SerialNumber
	if v := oSerialNumber.GetValue(); v != "" {
		sn = v
	}
	dev, err := octokeyz.GetDevice(sn)
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

	if conf.AndroidTv.Host != "" {
		certFile, exists := conf.GetAndroidTvCertificate()
		if !exists {
			cleanup.Check("android-tv certificate not found, please pair by calling this binary with `-p`")
		}

		cert, err := androidtv.OpenCertificate(certFile)
		cleanup.Check(err)

		atv, err := androidtv.NewRemote(conf.AndroidTv.Host, cert, oEvents.GetValue())
		cleanup.Check(err)
		cleanup.Register(atv)

		go func() {
			cleanup.Check(atv.Listen())
		}()

		handlers.AndroidTvInit(atv, oMuteAndroidTv.GetValue(), oPauseAndroidTv.GetValue())
	}

	cleanup.Check(handlers.RegisterMPVHandlers(dev, c, fmute, hsrc != nil))
	cleanup.Check(handlers.RegisterOctokeyzHandlers(dev, c, hsrc, false))

	if fstart {
		cleanup.Check(handlers.LoadNextFile(c, src))
	}

	go func() {
		cleanup.Check(dev.Listen(nil))
	}()

	cleanup.Check(s.Wait())
}
