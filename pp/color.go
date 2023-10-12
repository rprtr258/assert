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
