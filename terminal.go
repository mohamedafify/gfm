package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

const space = byte(' ')

type Size struct {
	width, height int
}

type Terminal struct {
	// size of the current terminal window
	size *Size

	// the state in which the terminal was before changes
	oldState *unix.Termios

	// current termina state
	state *unix.Termios

	// contains the data to be sent to Stdin
	outBuf []byte

	// lock protects the terminal and the state in this object from
	// concurrent processing of a key press and a Write() call.
	lock sync.Mutex

	// cursorX contains the current X value of the cursor where the left
	// edge is 0. cursorY contains the row number where the first row of
	// the current line is 0.
	cursorX, cursorY int

	// the buffer to write to
	c io.ReadWriter

	// current []File
	files []File
}

type File struct {
	name  string
	path  string
	isDir bool
	size  int64
}

func (term *Terminal) String() string {
	return fmt.Sprintf("Size: %d x %d\n", term.size.width, term.size.height)
}

func (term *Terminal) New() {

	term.size = new(Size)
	term.oldState = new(unix.Termios)
	term.state = new(unix.Termios)
	term.outBuf = make([]byte, 0)
	term.cursorX = 1
	term.cursorY = 1
	term.c = os.Stdin
	term.files = make([]File, 0)

	term.updateTerminalSize()
	term.startSizeListener()
	term.enableRawMode()

	term.hideCursor()
	// 	term.addBackground()
	term.updateScreen()
	term.Write()
}

// queue appends data to the end of t.outBuf
func (term *Terminal) queue(data []byte) {
	term.outBuf = append(term.outBuf, data...)
}

func (term *Terminal) Write() (n int, err error) {
	term.lock.Lock()
	defer term.lock.Unlock()

	if _, err = term.c.Write(term.outBuf); err != nil {
		return
	}
	term.outBuf = term.outBuf[:0]
	return
}

func (term *Terminal) move(x, y int) {
	if x < 1 {
		term.cursorX = 1
	} else if x > int(term.size.width) {
		term.cursorX = int(term.size.width)
	} else {
		term.cursorX = x
	}

	if y < 1 {
		term.cursorY = 1
	} else if y > int(term.size.height) {
		term.cursorY = int(term.size.height)
	} else {
		term.cursorY = y
	}

	log.Printf("current position: %dx%d\n", term.cursorX, term.cursorY)
	term.queue([]byte(fmt.Sprintf("\x1b[%d;%dH", y, x)))
}

func (term *Terminal) moveUp(n int) {
	if term.cursorY-n <= 1 {
		term.cursorY = 1
	} else {
		term.cursorY -= n
	}

	log.Printf("current position: %dx%d\n", term.cursorX, term.cursorY)
	term.queue([]byte(fmt.Sprintf("\x1b[%dA", n)))
}

func (term *Terminal) moveDown(n int) {
	if term.cursorY+n >= int(term.size.height) {
		term.cursorY = int(term.size.height)
	} else {
		term.cursorY += n
	}

	log.Printf("current position: %dx%d\n", term.cursorX, term.cursorY)
	term.queue([]byte(fmt.Sprintf("\x1b[%dB", n)))
}

func (term *Terminal) moveIn() {
	//TODO
	// if file open depending on extension
	// else if directory go into and load files
}

func (term *Terminal) moveOut() {
	//TODO
	// if directory go outof and load files
}

func (term *Terminal) moveTop() {
	term.queue([]byte("\x1b[H"))
}

func (term *Terminal) scrollUp(n int) {
	term.queue([]byte(fmt.Sprintf("\x1b[%dS", n)))
}

func (term *Terminal) scrollDown(n int) {
	term.queue([]byte(fmt.Sprintf("\x1b[%dT", n)))
}

func (term *Terminal) HandleInput() error {
	for {
		char := make([]byte, 1)
		os.Stdin.Read(char)
		stringChar := string(char[0])

		switch stringChar {
		case "q":
			term.quit()
		case "h":
			term.moveOut()
			term.Write()
		case "j":
			term.moveDown(1)
			term.Write()
		case "k":
			term.moveUp(1)
			term.Write()
		case "l":
			term.moveIn()
			term.Write()
		case "G":
			term.moveTop()
			term.Write()
		}
	}
	return nil
}

func (term *Terminal) quit() {
	log.Print("Quitting...")
	term.clearScreen()
	term.disableRawMode()
	stopLogger()
	os.Exit(0)
}

func (term *Terminal) updateTerminalSize() error {
	winSize := new(struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	})

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(unix.Stdout),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(winSize)),
	)

	term.size.width = int(winSize.Col)
	term.size.height = int(winSize.Row)

	if errno != 0 {
		err := errors.New("Syscall failed to get window size")
		log.Printf("failed to update terminal size %v", err)
		return err
	}

	log.Printf("my terminal size is %dx%d\n", term.size.width, term.size.height)
	return nil
}

func (term *Terminal) onTermSizeChanged() {
	term.updateTerminalSize()
	term.updateScreen()
	term.Write()
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

func (term *Terminal) clearScreen() {
	term.queue([]byte("\x1b[2J"))
	term.moveTop()
}

func (term *Terminal) addBackground() error {
	term.clearScreen()
	entireScreenSize := int(term.size.height) * int(term.size.width)

	// set color
	color := []byte(fmt.Sprintf("\x1b[48;2;%v", getColorANSI(black)))
	term.queue(color)

	// draw background
	for i := 0; i < entireScreenSize; i++ {
		term.queue([]byte("\x20"))
	}

	return nil
}

func (term *Terminal) updateScreen() error {
	term.clearScreen()
	if len(term.files) == 0 {
		log.Println("Getting directory info...")
		path, found := unix.Getenv("HOME")

		if !found {
			log.Println("Failed to get $HOME enviroment variable, using current working directory")
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalln("Failed to get cwd")
			}
			path = cwd
		}

		folder, err := os.Open(path)
		if err != nil {
			log.Println("failed to open folder", path)
			return err
		}

		defer folder.Close()

		folderContent, err := folder.ReadDir(0)
		if err != nil {
			log.Println("failed to get folder content", path)
			return err
		}

		for i := 0; i < len(folderContent); i++ {

			details, err := folderContent[i].Info()
			if err != nil {
				log.Println("failed to get file info", path)
				return err
			}

			name := folderContent[i].Name()
			isDir := folderContent[i].IsDir()
			size := details.Size()
			if name[0] == '.' || name[0:1] == ".." {
				continue
			}
			unix.Getenv("HOME")

			term.files = append(term.files, File{
				name:  name,
				isDir: isDir,
				size:  size,
			})
		}
	}

	for i := 0; i < len(term.files); i++ {
		term.queue([]byte(term.files[i].name))
		term.queue([]byte("\x1b[E"))
	}

	term.move(term.cursorX, term.cursorY)

	return nil
}

func (term *Terminal) hideCursor() {
	term.queue([]byte("\x1b[?25l"))
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
