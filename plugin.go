package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/rafaelmartins/b8r/internal/cleanup"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/b8r/internal/utils"
	"rafaelmartins.com/p/octokeyz"
)

func calledAsPlugin() (bool, uintptr) {
	if len(os.Args) < 2 {
		return false, 0
	}

	if !strings.HasSuffix(os.Args[0], ".run") {
		return false, 0
	}

	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--mpv-ipc-fd=") {
			if fd, err := strconv.Atoi(strings.Split(arg, "=")[1]); err == nil {
				return true, uintptr(fd)
			}
		}
	}
	return false, 0
}

func pluginInternal(m *client.MpvIpcClient) error {
	dev, err := octokeyz.GetDevice("")
	if err != nil {
		return err
	}

	if err := dev.Open(); err != nil {
		return err
	}
	cleanup.Register(dev)

	if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine1, "b8r plugin", octokeyz.DisplayLineAlignCenter)); err != nil {
		return err
	}

	m.AddHandler("property-change", func(m *client.MpvIpcClient, event string, data map[string]interface{}) error {
		switch int(data["id"].(float64)) {
		case 1:
			if fn := data["data"]; fn != nil {
				if err := utils.IgnoreDisplayMissing(dev.DisplayLine(octokeyz.DisplayLine4, fn.(string), octokeyz.DisplayLineAlignLeft)); err != nil {
					return err
				}
			}
		}
		return nil
	})

	if _, err := m.Command("observe_property", 1, "filename"); err != nil {
		return err
	}

	if err := utils.LedFlash3Times(dev); err != nil {
		return err
	}

	if err := handlers.RegisterOctokeyzHandlers(dev, m, nil); err != nil {
		return err
	}

	go func() {
		if err := dev.Listen(nil); err != nil {
			log.Print(err)
		}
	}()

	return nil
}

func plugin(fd uintptr) {
	defer cleanup.Cleanup()

	log.SetPrefix("[b8r] ")
	log.SetFlags(0)

	check := func(err any, fatal bool) {
		if err != nil {
			log.Print(err)
			if fatal {
				cleanup.Exit(1)
			}
		}
	}

	m, err := client.NewFromFd(fd, false)
	check(err, true)

	// according to documentation, mpv is supposed to send a shutdown event when closing.
	// i never saw it happening (mpv just sends an EOF in the fd), but lets support it.
	m.AddHandler("shutdown", func(m *client.MpvIpcClient, event string, data map[string]interface{}) error {
		cleanup.Exit(0)
		return nil
	})

	wait := make(chan bool)
	go func() {
		check(m.Listen(nil), true)
		wait <- true
	}()

	check(pluginInternal(m), false)
	<-wait
}
