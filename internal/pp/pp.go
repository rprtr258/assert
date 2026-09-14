// API definitions. The core implementation is delegated to printer.go.

package pp

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/rprtr258/assert/internal/scuf"
)

type ColorScheme struct {
	Bool            string
	Integer         string
	Float           string
	String          string
	StringQuotation string
	EscapedChar     string
	FieldName       string
	PointerAdress   string
	Nil             string
	Time            string
	StructName      string
	ObjectLength    string
}

var defaultScheme = ColorScheme{
	Bool:            scuf.FgCyan + ";" + scuf.ModBold,
	Integer:         scuf.FgBlue + ";" + scuf.ModBold,
	Float:           scuf.FgMagenta + ";" + scuf.ModBold,
	String:          scuf.FgRed,
	StringQuotation: scuf.FgRed + ";" + scuf.ModBold,
	EscapedChar:     scuf.FgMagenta + ";" + scuf.ModBold,
	FieldName:       scuf.FgYellow,
	PointerAdress:   scuf.FgBlue + ";" + scuf.ModBold,
	Nil:             scuf.FgCyan + ";" + scuf.ModBold,
	Time:            scuf.FgBlue + ";" + scuf.ModBold,
	StructName:      scuf.FgGreen,
	ObjectLength:    scuf.FgBlue,
}

// Global variable API
var (
	// Default pretty printer. It's public so that you can modify config globally.
	Default = newPrettyPrinter(3) //nolint:mnd // pp.* => PrettyPrinter.* => formatAll
	// WithLineInfo add file name and line information to output
	// call this function with care, because getting stack has performance penalty
	WithLineInfo bool
)

const (
	// BufferFoldThreshold - If the length of array or slice is larger than this,
	// the buffer will be shorten as {...}.
	BufferFoldThreshold = 1024
	// PrintMapTypes when set to true will have map types will always appended to maps.
	PrintMapTypes = true
)

// Internals
var (
	defaultOut          = os.Stdout
	defaultWithLineInfo = false
)

type PrettyPrinter struct {
	// WithLineInfo adds file name and line information to output.
	// Call this function with care, because getting stack has performance penalty.
	WithLineInfo bool
	// To support WithLineInfo, we need to know which frame we should look at.
	// Thus callerLevel sets the number of frames it needs to skip.
	callerLevel        int
	out                io.Writer
	currentScheme      ColorScheme
	outLock            sync.Mutex
	maxDepth           int
	ColoringEnabled    bool
	DecimalUint        bool
	ThousandsSeparator bool
	// This skips unexported fields of structs.
	ExportedOnly bool
}

// New creates a new PrettyPrinter that can be used to pretty print values
func New() *PrettyPrinter {
	return newPrettyPrinter(2) //nolint:mnd // PrettyPrinter.* => formatAll
}

func newPrettyPrinter(callerLevel int) *PrettyPrinter {
	return &PrettyPrinter{
		WithLineInfo:    defaultWithLineInfo,
		callerLevel:     callerLevel,
		out:             defaultOut,
		currentScheme:   defaultScheme,
		maxDepth:        -1,
		ColoringEnabled: true,
		DecimalUint:     true,
		ExportedOnly:    false,
	}
}

// Print prints given arguments.
func (pp *PrettyPrinter) Print(a ...any) {
	fmt.Fprint(pp.out, pp.formatAll(a)...)
}

// Sprint formats given arguments and returns the result as string.
func (pp *PrettyPrinter) Sprint(a ...any) string {
	return fmt.Sprint(pp.formatAll(a)...)
}

// SetOutput sets pp's output
func (pp *PrettyPrinter) SetOutput(o io.Writer) {
	pp.outLock.Lock()
	defer pp.outLock.Unlock()

	pp.out = o
}

// GetOutput returns pp's output.
func (pp *PrettyPrinter) GetOutput() io.Writer {
	return pp.out
}

// ResetOutput sets pp's output back to the default output
func (pp *PrettyPrinter) ResetOutput() {
	pp.outLock.Lock()
	defer pp.outLock.Unlock()

	pp.out = defaultOut
}

func (pp *PrettyPrinter) formatAll(objects []any) []any {
	results := make([]any, 0, len(objects)+1)

	if pp.WithLineInfo || pp == Default && WithLineInfo { // fix for backwards capability
		_, fn, line, _ := runtime.Caller(pp.callerLevel)
		results = append(results, fmt.Sprintf("%s:%d\n", fn, line))
	}

	for _, object := range objects {
		results = append(results, pp.format(object))
	}

	return results
}

// Sprint formats given arguments and returns the result as string.
func Sprint(a ...any) string {
	return Default.Sprint(a...)
}
