// color.go: Color API and implementation
package pp

import (
	"github.com/rprtr258/scuf"
)

const (
	// No color
	NoColor uint16 = 1 << 15
)

const (
	// Foreground colors for ColorScheme.
	_ uint16 = iota | NoColor
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	bitsForeground       = 0
	maskForeground       = 0xf
	ansiForegroundOffset = 30 - 1
)

const (
	// Background colors for ColorScheme.
	_ uint16 = iota<<bitsBackground | NoColor
	BackgroundBlack
	BackgroundRed
	BackgroundGreen
	BackgroundYellow
	BackgroundBlue
	BackgroundMagenta
	BackgroundCyan
	BackgroundWhite
	bitsBackground       = 4
	maskBackground       = 0xf << bitsBackground
	ansiBackgroundOffset = 40 - 1
)

const (
	// Bold flag for ColorScheme.
	Bold     uint16 = 1<<bitsBold | NoColor
	bitsBold        = 8
	maskBold        = 1 << bitsBold
	ansiBold        = 1
)

// To use with SetColorScheme.
type ColorScheme struct {
	Bool            []scuf.Modifier
	Integer         []scuf.Modifier
	Float           []scuf.Modifier
	String          scuf.Modifier
	StringQuotation []scuf.Modifier
	EscapedChar     []scuf.Modifier
	FieldName       scuf.Modifier
	PointerAdress   []scuf.Modifier
	Nil             []scuf.Modifier
	Time            []scuf.Modifier
	StructName      scuf.Modifier
	ObjectLength    scuf.Modifier
}

var defaultScheme = ColorScheme{
	Bool:            []scuf.Modifier{scuf.FgCyan, scuf.ModBold},
	Integer:         []scuf.Modifier{scuf.FgBlue, scuf.ModBold},
	Float:           []scuf.Modifier{scuf.FgMagenta, scuf.ModBold},
	String:          scuf.FgRed,
	StringQuotation: []scuf.Modifier{scuf.FgRed, scuf.ModBold},
	EscapedChar:     []scuf.Modifier{scuf.FgMagenta, scuf.ModBold},
	FieldName:       scuf.FgYellow,
	PointerAdress:   []scuf.Modifier{scuf.FgBlue, scuf.ModBold},
	Nil:             []scuf.Modifier{scuf.FgCyan, scuf.ModBold},
	Time:            []scuf.Modifier{scuf.FgBlue, scuf.ModBold},
	StructName:      scuf.FgGreen,
	ObjectLength:    scuf.FgBlue,
}

func (cs *ColorScheme) fixColors() {
	if cs.Bool == nil {
		cs.Bool = defaultScheme.Bool
	}
	if cs.Integer == nil {
		cs.Integer = defaultScheme.Integer
	}
	if cs.Float == nil {
		cs.Float = defaultScheme.Float
	}
	if cs.String == nil {
		cs.String = defaultScheme.String
	}
	if cs.StringQuotation == nil {
		cs.StringQuotation = defaultScheme.StringQuotation
	}
	if cs.EscapedChar == nil {
		cs.EscapedChar = defaultScheme.EscapedChar
	}
	if cs.FieldName == nil {
		cs.FieldName = defaultScheme.FieldName
	}
	if cs.PointerAdress == nil {
		cs.PointerAdress = defaultScheme.PointerAdress
	}
	if cs.Nil == nil {
		cs.Nil = defaultScheme.Nil
	}
	if cs.Time == nil {
		cs.Time = defaultScheme.Time
	}
	if cs.StructName == nil {
		cs.StructName = defaultScheme.StructName
	}
	if cs.ObjectLength == nil {
		cs.ObjectLength = defaultScheme.ObjectLength
	}
}

func colorizeText(text string, mods ...scuf.Modifier) string {
	return scuf.String(text, mods...)
}
