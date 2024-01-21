package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rafaelmartins/b8/go/b8"
	"github.com/rafaelmartins/b8r/internal/handlers"
	"github.com/rafaelmartins/b8r/internal/mpv/client"
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
	check := func(err any, fatal bool) {
		if err != nil {
			fmt.Println("[b8r]", err)
			if fatal {
				os.Exit(0)
			}
		}
	}

	m, err := client.NewFromFd(fd)
	check(err, true)

	dev, err := b8.GetDevice("")
	if err == nil {
		err = dev.Open()
		if err == nil {
			defer func() {
				dev.Led(b8.LedOff)
				dev.Close()
			}()

			for i := 0; i < 3; i++ {
				dev.Led(b8.LedFlash)
				time.Sleep(100 * time.Millisecond)
			}

			err = handlers.RegisterB8Handlers(dev, m, nil, func(b *b8.Button) error {
				_, err := m.Command("quit")
				return err
			})

			if err == nil {
				go func() {
					check(dev.Listen(), false)
				}()
			}
		}
	}
	check(err, false)

	check(m.Listen(nil), true)

	os.Exit(0)
}
