// API definitions. The core implementation is delegated to printer.go.

package pp

import (
	"fmt"
	"io"
	"os"

	"github.com/rprtr258/assert/internal/fun"
)

// BufferFoldThreshold - If the length of array or slice is larger than this,
// the buffer will be shorten as {...}.
const BufferFoldThreshold = 1024

type PrettyPrinter struct {
	// To support WithLineInfo, we need to know which frame we should look at.
	// Thus callerLevel sets the number of frames it needs to skip.
	callerLevel int
	out         io.Writer
}

// New creates a new PrettyPrinter that can be used to pretty print values
func New() *PrettyPrinter {
	return newPrettyPrinter(2) // PrettyPrinter.* => formatAll
}

func newPrettyPrinter(callerLevel int) *PrettyPrinter {
	return &PrettyPrinter{
		callerLevel: callerLevel,
		out:         os.Stdout,
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

func (pp *PrettyPrinter) formatAll(objects []any) []any {
	return fun.SliceMap(objects, func(object any) any {
		return pp.format(object)
	})
}

// Default pretty printer. It's public so that you can modify config globally.
var Default = newPrettyPrinter(3) // pp.* => PrettyPrinter.* => formatAll

// Sprint formats given arguments and returns the result as string.
func Sprint(a ...any) string {
	return Default.Sprint(a...)
}
