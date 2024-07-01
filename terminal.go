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

type Terminal struct {
	size     *unix.Winsize
	oldState *unix.Termios
	state    *unix.Termios
}

func (term *Terminal) String() string {
	return fmt.Sprintf("Size: %d x %d\n", term.size.Col, term.size.Row)
}

func (term *Terminal) New() {
	term.size = new(unix.Winsize)
	term.oldState = new(unix.Termios)
	term.state = new(unix.Termios)

	err := term.updateTerminalSize()

	if err != nil {
		log.Printf("failed to update terminal size %v", err)
	}

	term.startSizeListener()
	term.enableRawMode()
}

func (term *Terminal) HandleInput() error {
	for {
		char := make([]byte, 1)
		unix.Read(unix.Stdin, char)
		switch char[0] {
		case byte(113):
			term.quit()
		}
	}
	return nil
}

func (term *Terminal) quit() {
	term.disableRawMode()
	log.Print("Quitting...")
	stopLogger()
	os.Exit(0)
}

func (term *Terminal) updateTerminalSize() error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdout),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(term.size)),
	)

	if errno != 0 {
		return errors.New("Syscall failed to get window size")
	}
	return nil
}

func (term *Terminal) onTermSizeChanged() {
	err := term.updateTerminalSize()
	if err != nil {
		log.Printf("failed to update terminal size %v", err)
	}
}

// creates a new size change listener
func (term *Terminal) startSizeListener() {
	sizeChangeChannel := make(chan os.Signal, 1)
	sig := unix.SIGWINCH
	signal.Notify(sizeChangeChannel, sig)
	go func() {
		for {
			select {
			case <-sizeChangeChannel:
				term.onTermSizeChanged()
			}
		}
	}()
}

func (term *Terminal) disableRawMode() error {
	_, _, errno2 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdin),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(term.oldState)),
	)

	if errno2 != 0 {
		return errors.New("Syscall failed to set new term info")
	}

	return nil
}

func (term *Terminal) enableRawMode() error {
	_, _, errno1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdin),
		uintptr(unix.TCGETS),
		uintptr(unsafe.Pointer(term.oldState)),
	)

	if errno1 != 0 {
		return errors.New("Syscall failed to get term info")
	}

	*term.state = *term.oldState
	term.state.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	term.state.Oflag &^= unix.OPOST
	term.state.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	term.state.Cflag &^= unix.CSIZE | unix.PARENB
	term.state.Cflag |= unix.CS8
	term.state.Cc[unix.VMIN] = 1
	term.state.Cc[unix.VTIME] = 0

	_, _, errno2 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdin),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(term.state)),
	)

	if errno2 != 0 {
		return errors.New("Syscall failed to set new term info")
	}

	return nil
}
