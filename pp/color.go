// color.go: Color API and implementation
package pp

import (
	"github.com/rprtr258/scuf"
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

func or[T interface {
	scuf.Modifier | []scuf.Modifier
}](x, y T) T {
	if x == nil {
		return y
	}
	return x
}

func (cs *ColorScheme) fixColors() {
	*cs = ColorScheme{
		Bool:            or(cs.Bool, defaultScheme.Bool),
		Integer:         or(cs.Integer, defaultScheme.Integer),
		Float:           or(cs.Float, defaultScheme.Float),
		String:          or(cs.String, defaultScheme.String),
		StringQuotation: or(cs.StringQuotation, defaultScheme.StringQuotation),
		EscapedChar:     or(cs.EscapedChar, defaultScheme.EscapedChar),
		FieldName:       or(cs.FieldName, defaultScheme.FieldName),
		PointerAdress:   or(cs.PointerAdress, defaultScheme.PointerAdress),
		Nil:             or(cs.Nil, defaultScheme.Nil),
		Time:            or(cs.Time, defaultScheme.Time),
		StructName:      or(cs.StructName, defaultScheme.StructName),
		ObjectLength:    or(cs.ObjectLength, defaultScheme.ObjectLength),
	}
}

func colorizeText(text string, mods ...scuf.Modifier) string {
	return scuf.String(text, mods...)
}
