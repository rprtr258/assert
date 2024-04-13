package scuf

import "fmt"

type Mod = string

const (
	// Foreground colors
	FgBlack   = "30"
	FgRed     = "31"
	FgGreen   = "32"
	FgYellow  = "33"
	FgBlue    = "34"
	FgMagenta = "35"
	FgCyan    = "36"
	FgWhite   = "37"
	// Foreground bright colors
	FgHiBlack   = "90"
	FgHiRed     = "91"
	FgHiGreen   = "92"
	FgHiYellow  = "93"
	FgHiBlue    = "94"
	FgHiMagenta = "95"
	FgHiCyan    = "96"
	FgHiWhite   = "97"

	// Background colors
	BgBlack   = "40"
	BgRed     = "41"
	BgGreen   = "42"
	BgYellow  = "43"
	BgBlue    = "44"
	BgMagenta = "45"
	BgCyan    = "46"
	BgWhite   = "47"
	// Background bright colors
	BgHiBlack   = "100"
	BgHiRed     = "101"
	BgHiGreen   = "102"
	BgHiYellow  = "103"
	BgHiBlue    = "104"
	BgHiMagenta = "105"
	BgHiCyan    = "106"
	BgHiWhite   = "107"

	// Common consts
	_esc              byte = '\x1b'   // Escape character
	_csi                   = "\x1b["  // Control Sequence Introducer
	_osc                   = "\x1b]"  // Operating System Command
	_stringTerminator      = "\x1b\\" // String Terminator
	ModReset               = "0"
	ModBold                = "1"
	ModFaint               = "2"
	ModItalic              = "3"
	ModUnderline           = "4"
	ModBlink               = "5"
	ModReverse             = "7"
	ModCrossout            = "9"
	ModOverline            = "53"
)

// r,g,b are 0-255
func FgRGB(r, g, b uint8) Mod {
	return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}

func String(s string, mod Mod) string {
	return _csi + mod + "m" +
		s +
		_csi + ModReset + "m"
}
