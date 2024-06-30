package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Size struct {
	Width  int
	Height int
}

type Terminal struct {
	size Size
}

func (term *Terminal) Init() {
	size, err := getTerminalSize()
	if err != nil {
		log.Fatal(err)
	}
	term.size = size
	term.startSizeListener()
}

func getTerminalSize() (Size, error) {
	ws := &struct {
		rows uint16
		cols uint16
		x    uint16
		y    uint16
	}{}

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdout),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)

	if errno != 0 {
		return Size{}, errors.New("Syscall failed to get window size")
	}

	s := Size{
		Width:  int(ws.cols),
		Height: int(ws.rows),
	}
	return s, nil
}

func (size Size) String() string {
	return fmt.Sprintf("%d x %d", size.Height, size.Width)
}

// creates a new size change listener
func (term *Terminal) startSizeListener() {
	ch := make(chan os.Signal, 1)
	sig := unix.SIGWINCH
	signal.Notify(ch, sig)
	go func() {
		for {
			select {
			case <-ch:
				s, err := getTerminalSize()
				if err != nil {
					log.Println(err)
				}
				term.size = s
			}
		}
	}()
}
