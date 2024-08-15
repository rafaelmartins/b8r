package cleanup

import (
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	m       sync.Mutex
	closers = []io.Closer{}
	sig     chan os.Signal
)

func Cleanup() {
	m.Lock()
	defer m.Unlock()

	for _, c := range closers {
		c.Close()
	}
	closers = []io.Closer{}
}

func Register(c io.Closer) {
	if sig == nil {
		sig = make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			rv := <-sig
			Cleanup()
			os.Exit(128 + int(rv.(syscall.Signal)))
		}()
	}

	m.Lock()
	defer m.Unlock()
	closers = append(closers, c)
}

func Exit(code int) {
	Cleanup()
	os.Exit(code)
}

func Fatal(v ...any) {
	Cleanup()
	log.Fatal(v...)
}

func Check(err any) {
	if err != nil {
		Fatal("error: ", err)
	}
}
