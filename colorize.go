//go:build dlg

package dlg

import (
	"bytes"
	"strconv"
	"strings"
)

// colorizeOrDont gets set when -ldflags "-X 'github.com/vvvvv/dlg.DLG_COLOR=1'"
var colorizeOrDont = func(buf *[]byte) {}

// resetColorOrDont gets set when -ldflags "-X 'github.com/vvvvv/dlg.DLG_COLOR=1'"
var resetColorOrDont = func(buf *[]byte) {}

func colorize(buf *[]byte) {
	*buf = append(*buf, termColor...)
}

// colorResets resets the terminal color sequence by appending \033[0m .
func colorReset(buf *[]byte) {
	// *buf = append(*buf, resetColor...)
	*buf = append(*buf, 27, 91, 48, 109)
}

// colorArgToTermColor converts a string to a terminal color escape sequence.
// It accepts:
//
//	"red"
//	"YELLOW"
//	"Black"
//	"\033[38;5;2m"
//	"\x1b[38;5;2m"
//	string([]byte{27, 91, 51, 56, 59, 53, 59, 49, 109})
//	"3"
//
// The function does only naive validation and doesn't guaranty the returned sequence is a valid terminal escape sequence.
func colorArgToTermColor(arg string) (color []byte, ok bool) {
	// \033, \x1b or \u001b become a raw ESC byte
	if s, err := strconv.Unquote(`"` + arg + `"`); err == nil {
		arg = s
	}
	c := bytes.TrimSpace([]byte(arg))

	if len(c) > 0 {
		if c[0] == 27 {
			return c, true
		} else {
			isNumber := (c[0] >= 48 && c[0] <= 57)
			isAlpha := (c[0] >= 65 && c[0] <= 122)

			color = make([]byte, 0, 32)
			// ESC
			color = append(color, 27)
			// Start of esc sequence
			color = append(color, "[38;5;"...)
			if isNumber {
				// Assume it's a terminal color code e.g.
				// 1 for red
				ok = true
				color = append(color, c...)
			} else if isAlpha {
				// Assume it's a color name
				name := strings.ToLower(arg)
				for i, n := range []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"} {
					if name == n {
						ok = true
						color = append(color, byte('0'+i))
					}
				}
			}
			// Add suffix
			if color[len(color)-1] == ';' {
				color = append(color, "1m"...)
			} else {
				color = append(color, ";1m"...)
			}
		}
	}
	return
}
