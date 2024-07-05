package main

import (
	"fmt"
)

// onedark theme colors
var (
	black  = []byte{40, 44, 52}
	red    = []byte{224, 108, 117}
	green  = []byte{152, 195, 121}
	yellow = []byte{229, 192, 123}
	blue   = []byte{97, 175, 239}
	pink   = []byte{198, 120, 221}
	cyan   = []byte{86, 182, 194}
	grey   = []byte{171, 178, 191}
)

func getColorANSI(color []byte) string {
	return fmt.Sprintf("%v;%v;%vm", color[0], color[1], color[2])
}
