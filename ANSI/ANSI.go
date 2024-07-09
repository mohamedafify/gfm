package ansi

import (
	"fmt"
)

type RGBColor struct {
	r, g, b byte
}

// onedark theme colors
var (
	Black  = RGBColor{40, 44, 52}
	Red    = RGBColor{224, 108, 117}
	Green  = RGBColor{152, 195, 121}
	Yellow = RGBColor{229, 192, 123}
	Blue   = RGBColor{97, 175, 239}
	Pink   = RGBColor{198, 120, 221}
	Cyan   = RGBColor{86, 182, 194}
	Grey   = RGBColor{171, 178, 191}
)

const Space byte = '\x20'

var (
	ResetBuffer         []byte = []byte("\x1b[?1049l")
	CleanBuffer         []byte = []byte("\x1b[?1049h")
	ClearScreen         []byte = []byte("\x1b[2J")
	MoveTop             []byte = []byte("\x1b[H")
	InverseColors       []byte = []byte("\x1b[7m")
	ResetTextAttributes []byte = []byte("\x1b[0m")
	MoveOnNewLine       []byte = []byte("\x1b[E")
	HideCursor          []byte = []byte("\x1b[?25l")
	ShowCursor          []byte = []byte("\x1b[?25h")
)

func SetFGColor(color RGBColor) []byte {
	return []byte(fmt.Sprintf("\x1b[38;2;%v;%v;%vm", color.r, color.g, color.b))
}

func SetBGColor(color RGBColor) []byte {
	return []byte(fmt.Sprintf("\x1b[48;2;%v;%v;%vm", color.r, color.g, color.b))
}

func MoveCursor(x, y int) []byte {
	return []byte(fmt.Sprintf("\x1b[%d;%dH", y, x))
}

func ScrollUp(n int) []byte {
	return []byte(fmt.Sprintf("\x1b[%dS", n))
}

func ScrollDown(n int) []byte {
	return []byte(fmt.Sprintf("\x1b[%dT", n))
}
