// color.go: Color API and implementation
package pp

import (
	"github.com/rprtr258/scuf"
)

// To use with SetColorScheme.
type ColorScheme struct {
	Bool            scuf.Modifier
	Integer         scuf.Modifier
	Float           scuf.Modifier
	String          scuf.Modifier
	StringQuotation scuf.Modifier
	EscapedChar     scuf.Modifier
	FieldName       scuf.Modifier
	PointerAdress   scuf.Modifier
	Nil             scuf.Modifier
	Time            scuf.Modifier
	StructName      scuf.Modifier
	ObjectLength    scuf.Modifier
}

var defaultScheme = ColorScheme{
	Bool:            scuf.CombineModifiers(scuf.FgCyan, scuf.ModBold),
	Integer:         scuf.CombineModifiers(scuf.FgBlue, scuf.ModBold),
	Float:           scuf.CombineModifiers(scuf.FgMagenta, scuf.ModBold),
	String:          scuf.FgRed,
	StringQuotation: scuf.CombineModifiers(scuf.FgRed, scuf.ModBold),
	EscapedChar:     scuf.CombineModifiers(scuf.FgMagenta, scuf.ModBold),
	FieldName:       scuf.FgYellow,
	PointerAdress:   scuf.CombineModifiers(scuf.FgBlue, scuf.ModBold),
	Nil:             scuf.CombineModifiers(scuf.FgCyan, scuf.ModBold),
	Time:            scuf.CombineModifiers(scuf.FgBlue, scuf.ModBold),
	StructName:      scuf.FgGreen,
	ObjectLength:    scuf.FgBlue,
}
