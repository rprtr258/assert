// API definitions. The core implementation is delegated to printer.go.
package pp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	colorable "github.com/mattn/go-colorable"
	"github.com/rprtr258/scuf"
)

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
	Bool:            scuf.Combine(scuf.FgCyan, scuf.ModBold),
	Integer:         scuf.Combine(scuf.FgBlue, scuf.ModBold),
	Float:           scuf.Combine(scuf.FgMagenta, scuf.ModBold),
	String:          scuf.FgRed,
	StringQuotation: scuf.Combine(scuf.FgRed, scuf.ModBold),
	EscapedChar:     scuf.Combine(scuf.FgMagenta, scuf.ModBold),
	FieldName:       scuf.FgYellow,
	PointerAdress:   scuf.Combine(scuf.FgBlue, scuf.ModBold),
	Nil:             scuf.Combine(scuf.FgCyan, scuf.ModBold),
	Time:            scuf.Combine(scuf.FgBlue, scuf.ModBold),
	StructName:      scuf.FgGreen,
	ObjectLength:    scuf.FgBlue,
}

// Global variable API
var (
	// Default pretty printer. It's public so that you can modify config globally.
	Default = newPrettyPrinter(3) // pp.* => PrettyPrinter.* => formatAll
	// If the length of array or slice is larger than this,
	// the buffer will be shorten as {...}.
	BufferFoldThreshold = 1024
	// PrintMapTypes when set to true will have map types will always appended to maps.
	PrintMapTypes = true
	// WithLineInfo add file name and line information to output
	// call this function with care, because getting stack has performance penalty
	WithLineInfo bool
)

// Internals
var (
	defaultOut          = colorable.NewColorableStdout()
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
	return newPrettyPrinter(2) // PrettyPrinter.* => formatAll
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
func (pp *PrettyPrinter) Print(a ...any) (n int, err error) {
	return fmt.Fprint(pp.out, pp.formatAll(a)...)
}

// Printf prints a given format.
func (pp *PrettyPrinter) Printf(format string, a ...any) (n int, err error) {
	return fmt.Fprintf(pp.out, format, pp.formatAll(a)...)
}

// Println prints given arguments with newline.
func (pp *PrettyPrinter) Println(a ...any) (n int, err error) {
	return fmt.Fprintln(pp.out, pp.formatAll(a)...)
}

// Sprint formats given arguments and returns the result as string.
func (pp *PrettyPrinter) Sprint(a ...any) string {
	return fmt.Sprint(pp.formatAll(a)...)
}

// Sprintf formats with pretty print and returns the result as string.
func (pp *PrettyPrinter) Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, pp.formatAll(a)...)
}

// Sprintln formats given arguments with newline and returns the result as string.
func (pp *PrettyPrinter) Sprintln(a ...any) string {
	return fmt.Sprintln(pp.formatAll(a)...)
}

// Fprint prints given arguments to a given writer.
func (pp *PrettyPrinter) Fprint(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprint(w, pp.formatAll(a)...)
}

// Fprintf prints format to a given writer.
func (pp *PrettyPrinter) Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return fmt.Fprintf(w, format, pp.formatAll(a)...)
}

// Fprintln prints given arguments to a given writer with newline.
func (pp *PrettyPrinter) Fprintln(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprintln(w, pp.formatAll(a)...)
}

// Errorf formats given arguments and returns it as error type.
func (pp *PrettyPrinter) Errorf(format string, a ...any) error {
	return errors.New(pp.Sprintf(format, a...))
}

// Fatal prints given arguments and finishes execution with exit status 1.
func (pp *PrettyPrinter) Fatal(a ...any) {
	fmt.Fprint(pp.out, pp.formatAll(a)...)
	os.Exit(1)
}

// Fatalf prints a given format and finishes execution with exit status 1.
func (pp *PrettyPrinter) Fatalf(format string, a ...any) {
	fmt.Fprintf(pp.out, format, pp.formatAll(a)...)
	os.Exit(1)
}

// Fatalln prints given arguments with newline and finishes execution with exit status 1.
func (pp *PrettyPrinter) Fatalln(a ...any) {
	fmt.Fprintln(pp.out, pp.formatAll(a)...)
	os.Exit(1)
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

func or[T interface {
	scuf.Modifier | []scuf.Modifier
}](x, y T) T {
	if x == nil {
		return y
	}
	return x
}

// SetColorScheme takes a colorscheme used by all future Print calls.
func (pp *PrettyPrinter) SetColorScheme(scheme ColorScheme) {
	pp.currentScheme = ColorScheme{
		Bool:            or(scheme.Bool, defaultScheme.Bool),
		Integer:         or(scheme.Integer, defaultScheme.Integer),
		Float:           or(scheme.Float, defaultScheme.Float),
		String:          or(scheme.String, defaultScheme.String),
		StringQuotation: or(scheme.StringQuotation, defaultScheme.StringQuotation),
		EscapedChar:     or(scheme.EscapedChar, defaultScheme.EscapedChar),
		FieldName:       or(scheme.FieldName, defaultScheme.FieldName),
		PointerAdress:   or(scheme.PointerAdress, defaultScheme.PointerAdress),
		Nil:             or(scheme.Nil, defaultScheme.Nil),
		Time:            or(scheme.Time, defaultScheme.Time),
		StructName:      or(scheme.StructName, defaultScheme.StructName),
		ObjectLength:    or(scheme.ObjectLength, defaultScheme.ObjectLength),
	}

}

// ResetColorScheme resets colorscheme to default.
func (pp *PrettyPrinter) ResetColorScheme() {
	pp.currentScheme = defaultScheme
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

// Print prints given arguments.
func Print(a ...any) (n int, err error) {
	return Default.Print(a...)
}

// Printf prints a given format.
func Printf(format string, a ...any) (n int, err error) {
	return Default.Printf(format, a...)
}

// Println prints given arguments with newline.
func Println(a ...any) (n int, err error) {
	return Default.Println(a...)
}

// Sprint formats given arguments and returns the result as string.
func Sprint(a ...any) string {
	return Default.Sprint(a...)
}

// Sprintf formats with pretty print and returns the result as string.
func Sprintf(format string, a ...any) string {
	return Default.Sprintf(format, a...)
}

// Sprintln formats given arguments with newline and returns the result as string.
func Sprintln(a ...any) string {
	return Default.Sprintln(a...)
}

// Fprint prints given arguments to a given writer.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	return Default.Fprint(w, a...)
}

// Fprintf prints format to a given writer.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return Default.Fprintf(w, format, a...)
}

// Fprintln prints given arguments to a given writer with newline.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	return Default.Fprintln(w, a...)
}

// Errorf formats given arguments and returns it as error type.
func Errorf(format string, a ...any) error {
	return Default.Errorf(format, a...)
}

// Fatal prints given arguments and finishes execution with exit status 1.
func Fatal(a ...any) {
	Default.Fatal(a...)
}

// Fatalf prints a given format and finishes execution with exit status 1.
func Fatalf(format string, a ...any) {
	Default.Fatalf(format, a...)
}

// Fatalln prints given arguments with newline and finishes execution with exit status 1.
func Fatalln(a ...any) {
	Default.Fatalln(a...)
}

// Change Print* functions' output to a given writer.
// For example, you can limit output by ENV.
//
//	func init() {
//		if os.Getenv("DEBUG") == "" {
//			pp.SetDefaultOutput(ioutil.Discard)
//		}
//	}
func SetDefaultOutput(o io.Writer) {
	Default.SetOutput(o)
}

// GetOutput returns pp's default output.
func GetDefaultOutput() io.Writer {
	return Default.GetOutput()
}

// Change Print* functions' output to default one.
func ResetDefaultOutput() {
	Default.ResetOutput()
}

// SetColorScheme takes a colorscheme used by all future Print calls.
func SetColorScheme(scheme ColorScheme) {
	Default.SetColorScheme(scheme)
}

// ResetColorScheme resets colorscheme to default.
func ResetColorScheme() {
	Default.ResetColorScheme()
}

// SetMaxDepth sets the printer's Depth, -1 prints all
func SetDefaultMaxDepth(v int) {
	Default.maxDepth = v
}
