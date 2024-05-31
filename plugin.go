package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
	"github.com/rafaelmartins/octokeyz/go/octokeyz"
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

func plugin(fd uintptr) {
	log.SetPrefix("[b8r] ")
	log.SetFlags(0)

	check := func(err any, fatal bool) {
		if err != nil {
			log.Print(err)
			if fatal {
				os.Exit(1)
			}
		}
	}

	m, err := client.NewFromFd(fd)
	check(err, true)

	// according to documentation, mpv is supposed to send a shutdown event when closing.
	// i never saw it happening (mpv just sends an EOF in the fd), but lets support it.
	m.AddHandler("shutdown", func(m *client.MpvIpcClient, event string, data map[string]interface{}) error {
		os.Exit(0)
		return nil
	})

	wait := make(chan bool)
	go func() {
		check(m.Listen(nil), true)
		wait <- true
	}()

	dev, err := octokeyz.GetDevice("")
	if err == nil {
		if err = dev.Open(); err == nil {
			defer func() {
				dev.Led(octokeyz.LedOff)
				dev.Close()
			}()

			for i := 0; i < 3; i++ {
				dev.Led(octokeyz.LedFlash)
				time.Sleep(100 * time.Millisecond)
			}

			if err = handlers.RegisterOctokeyzHandlers(dev, m, nil); err == nil {
				go func() {
					check(dev.Listen(nil), false)
				}()
			}
		}
	}
	check(err, false)

	<-wait
}
