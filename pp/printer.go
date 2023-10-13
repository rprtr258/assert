// printer.go: The actual pretty print implementation. Everything in this file should be private.
package pp

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/scuf"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const indentWidth = 2

func (pp *PrettyPrinter) format(object any) string {
	return newPrinter(object, &pp.currentScheme, pp.maxDepth, pp.ColoringEnabled, pp.DecimalUint, pp.ExportedOnly, pp.ThousandsSeparator).String()
}

func newPrinter(object any, currentScheme *ColorScheme, maxDepth int, coloringEnabled bool, decimalUint bool, exportedOnly bool, thousandsSeparator bool) *printer {
	buffer := &bytes.Buffer{}
	tw := &tabwriter.Writer{}
	tw.Init(buffer, indentWidth, 0, 1, ' ', 0)

	printer := &printer{
		Buffer:             buffer,
		tw:                 tw,
		depth:              0,
		maxDepth:           maxDepth,
		value:              reflect.ValueOf(object),
		visited:            map[uintptr]bool{},
		currentScheme:      currentScheme,
		coloringEnabled:    coloringEnabled,
		decimalUint:        decimalUint,
		exportedOnly:       exportedOnly,
		thousandsSeparator: thousandsSeparator,
	}

	if thousandsSeparator {
		printer.localizedPrinter = message.NewPrinter(language.English)
	}

	return printer
}

type printer struct {
	*bytes.Buffer
	tw                 *tabwriter.Writer
	depth              int
	maxDepth           int
	value              reflect.Value
	visited            map[uintptr]bool
	currentScheme      *ColorScheme
	coloringEnabled    bool
	decimalUint        bool
	exportedOnly       bool
	thousandsSeparator bool
	localizedPrinter   *message.Printer
}

func (p *printer) String() string {
	switch p.value.Kind() {
	case reflect.Bool:
		p.colorPrint(p.raw(), p.currentScheme.Bool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Complex64, reflect.Complex128:
		p.colorPrint(p.raw(), p.currentScheme.Integer)
	case reflect.Float32, reflect.Float64:
		p.colorPrint(p.raw(), p.currentScheme.Float)
	case reflect.String:
		p.printString()
	case reflect.Map:
		p.printMap()
	case reflect.Struct:
		p.printStruct()
	case reflect.Array:
		p.printArray()
	case reflect.Slice:
		p.printSlice()
	case reflect.Chan:
		p.printf("(%s)(%s)", p.colorizeType(p.value.Type()), p.pointerAddr())
	case reflect.Interface:
		p.printInterface()
	case reflect.Ptr:
		p.printPtr()
	case reflect.Func:
		p.printf("%s {...}", p.colorizeType(p.value.Type()))
	case reflect.UnsafePointer:
		p.printf("%s(%s)", p.colorizeType(p.value.Type()), p.pointerAddr())
	case reflect.Invalid:
		p.print(p.nil())
	default:
		p.print(p.raw())
	}

	p.tw.Flush()
	return p.Buffer.String()
}

func (p *printer) print(text string) {
	fmt.Fprint(p.tw, text)
}

func (p *printer) printf(format string, args ...any) {
	fmt.Fprintf(p.tw, format, args...)
}

func (p *printer) println(text string) {
	fmt.Fprintln(p.tw, text)
}

func (p *printer) indentPrint(text string) {
	p.print(p.indent() + text)
}

func (p *printer) indentPrintf(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	p.indentPrint(text)
}

func (p *printer) colorPrint(text string, mod scuf.Modifier) {
	p.print(p.colorize(text, mod))
}

func (p *printer) printString() {
	quoted := strconv.Quote(p.value.String())
	quoted = quoted[1 : len(quoted)-1]

	p.colorPrint(`"`, p.currentScheme.StringQuotation)
	for len(quoted) > 0 {
		pos := strings.IndexByte(quoted, '\\')
		if pos == -1 {
			p.colorPrint(quoted, p.currentScheme.String)
			break
		}
		if pos != 0 {
			p.colorPrint(quoted[0:pos], p.currentScheme.String)
		}

		n := 1
		switch quoted[pos+1] {
		case 'x': // "\x00"
			n = 3
		case 'u': // "\u0000"
			n = 5
		case 'U': // "\U00000000"
			n = 9
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // "\000"
			n = 3
		}
		p.colorPrint(quoted[pos:pos+n+1], p.currentScheme.EscapedChar)
		quoted = quoted[pos+n+1:]
	}
	p.colorPrint(`"`, p.currentScheme.StringQuotation)
}

func (p *printer) printMap() {
	if p.value.Len() == 0 {
		p.printf(p.colorizeType(p.value.Type()) + "{}")
		return
	}

	if p.visited[p.value.Pointer()] {
		p.printf(p.colorizeType(p.value.Type()) + "{...}")
		return
	}
	p.visited[p.value.Pointer()] = true

	if PrintMapTypes {
		p.printf(p.colorizeType(p.value.Type()) + "{\n")
	} else {
		p.println("{")
	}
	p.indented(func() {
		value := sortMap(p.value)
		for i := 0; i < value.Len(); i++ {
			p.indentPrintf(
				"%s:\t%s,\n",
				p.format(value.keys[i]),
				p.format(value.values[i]),
			)
		}
	})
	p.indentPrint("}")
}

func (p *printer) printStruct() {
	typ := p.value.Type()

	if p.value.CanInterface() {
		switch {
		case typ.String() == "time.Time" && typ.PkgPath() == "time":
			p.printTime()
			return
		case typ.String() == "big.Int":
			bigInt := p.value.Interface().(big.Int)
			p.print(p.colorize(bigInt.String(), p.currentScheme.Integer))
			return
		case typ.String() == "big.Float":
			bigFloat := p.value.Interface().(big.Float)
			p.print(p.colorize(bigFloat.String(), p.currentScheme.Float))
			return
		}
	}

	var fields []int
	for i := 0; i < p.value.NumField(); i++ {
		field := typ.Field(i)
		// ignore unexported if needed
		if p.exportedOnly && field.PkgPath != "" {
			continue
		}

		// ignore fields if zero value, or explicitly set
		if tag := field.Tag.Get("pp"); tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) == 2 && parts[1] == "omitempty" && valueIsZero(p.value.Field(i)) {
				continue
			}
			if parts[0] == "-" {
				continue
			}
		}
		fields = append(fields, i)
	}
	if len(fields) == 0 {
		p.print(p.colorizeType(p.value.Type()) + "{}")
		return
	}

	p.println(p.colorizeType(p.value.Type()) + "{")
	p.indented(func() {
		for _, i := range fields {
			field := p.value.Type().Field(i)

			fieldName := field.Name
			if tag := field.Tag.Get("pp"); tag != "" {
				tagName := strings.Split(tag, ",")
				if tagName[0] != "" {
					fieldName = tagName[0]
				}
			}

			p.indentPrintf(
				"%s:\t%s,\n",
				p.colorize(fieldName, p.currentScheme.FieldName),
				p.format(p.value.Field(i)),
			)
		}
	})
	p.indentPrint("}")
}

func (p *printer) printTime() {
	tm := p.value.Interface().(time.Time)
	p.printf(
		"%s-%s-%s %s:%s:%s %s",
		p.colorize(strconv.Itoa(tm.Year()), p.currentScheme.Time),
		p.colorize(fmt.Sprintf("%02d", tm.Month()), p.currentScheme.Time),
		p.colorize(fmt.Sprintf("%02d", tm.Day()), p.currentScheme.Time),
		p.colorize(fmt.Sprintf("%02d", tm.Hour()), p.currentScheme.Time),
		p.colorize(fmt.Sprintf("%02d", tm.Minute()), p.currentScheme.Time),
		p.colorize(fmt.Sprintf("%02d", tm.Second()), p.currentScheme.Time),
		p.colorize(tm.Location().String(), p.currentScheme.Time),
	)
}

func (p *printer) printSlice() {
	if p.value.IsNil() {
		p.printf(p.colorizeType(p.value.Type()) + "(" + p.nil() + ")")
		return
	}

	if p.value.Len() == 0 {
		p.printf(p.colorizeType(p.value.Type()) + "{}")
		return
	}

	if p.value.Kind() == reflect.Slice {
		if p.visited[p.value.Pointer()] {
			// Stop travarsing cyclic reference
			p.printf(p.colorizeType(p.value.Type()) + "{...}")
			return
		}
		p.visited[p.value.Pointer()] = true
	}

	// Fold a large buffer
	if p.value.Len() > BufferFoldThreshold {
		p.printf(p.colorizeType(p.value.Type()) + "{...}")
		return
	}

	p.println(p.colorizeType(p.value.Type()) + "{")
	p.indented(func() {
		groupsize := fun.
			SwitchZero[int](p.value.Type().Elem().Kind()).
			Case(reflect.Uint8, 16).
			Case(reflect.Uint16, 8).
			Case(reflect.Uint32, 8).
			Case(reflect.Uint64, 4).
			End()

		if groupsize > 0 {
			for i := 0; i < p.value.Len(); i++ {
				// indent for new group
				if i%groupsize == 0 {
					p.print(p.indent())
				}
				// slice element
				p.printf("%s,", p.format(p.value.Index(i)))
				// space or newline
				if (i+1)%groupsize == 0 || i+1 == p.value.Len() {
					p.print("\n")
				} else {
					p.print(" ")
				}
			}
		} else {
			for i := 0; i < p.value.Len(); i++ {
				p.indentPrintf(p.format(p.value.Index(i)) + ",\n")
			}
		}
	})
	p.indentPrint("}")
}

func (p *printer) printArray() {
	if p.value.Len() == 0 {
		p.printf("%s{}", p.colorizeType(p.value.Type()))
		return
	}

	if p.value.Kind() == reflect.Slice {
		if p.visited[p.value.Pointer()] {
			// Stop travarsing cyclic reference
			p.printf("%s{...}", p.colorizeType(p.value.Type()))
			return
		}
		p.visited[p.value.Pointer()] = true
	}

	// Fold a large buffer
	if p.value.Len() > BufferFoldThreshold {
		p.printf("%s{...}", p.colorizeType(p.value.Type()))
		return
	}

	p.println(p.colorizeType(p.value.Type()) + "{")
	p.indented(func() {
		groupsize := fun.
			SwitchZero[int](p.value.Type().Elem().Kind()).
			Case(reflect.Uint8, 16).
			Case(reflect.Uint16, 8).
			Case(reflect.Uint32, 8).
			Case(reflect.Uint64, 4).
			End()

		if groupsize > 0 {
			for i := 0; i < p.value.Len(); i++ {
				// indent for new group
				if i%groupsize == 0 {
					p.print(p.indent())
				}
				// slice element
				p.printf(p.format(p.value.Index(i)) + ",")
				// space or newline
				p.print(fun.IF((i+1)%groupsize == 0 || i+1 == p.value.Len(), "\n", " "))
			}
		} else {
			for i := 0; i < p.value.Len(); i++ {
				p.indentPrintf(p.format(p.value.Index(i)) + ",\n")
			}
		}
	})
	p.indentPrint("}")
}

func (p *printer) printInterface() {
	e := p.value.Elem()
	switch {
	case e.Kind() == reflect.Invalid:
		p.print(p.nil())
	case e.IsValid():
		p.print(p.format(e))
	default:
		p.printf("%s(%s)", p.colorizeType(p.value.Type()), p.nil())
	}
}

func (p *printer) printPtr() {
	if p.visited[p.value.Pointer()] {
		p.printf("&%s{...}", p.elemTypeString())
		return
	}
	if p.value.Pointer() != 0 {
		p.visited[p.value.Pointer()] = true
	}

	if p.value.Elem().IsValid() {
		p.printf("&%s", p.format(p.value.Elem()))
	} else {
		p.printf("(%s)(%s)", p.colorizeType(p.value.Type()), p.nil())
	}
}

func (p *printer) pointerAddr() string {
	return p.colorize(fmt.Sprintf("%#v", p.value.Pointer()), p.currentScheme.PointerAdress)
}

func (p *printer) elemTypeString() string {
	return p.colorizeType(p.value.Elem().Type())
}

func (p *printer) colorizeType(tt reflect.Type) string {
	t := tt.String()
	prefix := ""

	if p.matchRegexp(t, `^\[\].`) { // slice
		prefix = "[]"
		t = t[2:]
	}

	if p.matchRegexp(t, `^\[\d+\].`) { // array
		num := regexp.MustCompile(`\d+`).FindString(t)
		prefix = fmt.Sprintf("[%s]", p.colorize(num, p.currentScheme.ObjectLength))
		t = t[2+len(num):]
	}

	if p.matchRegexp(t, `^[^\.]+\.[^\.]+$`) { // struct
		ts := strings.Split(t, ".")
		t = fmt.Sprintf("%s.%s", ts[0], p.colorize(ts[1], p.currentScheme.StructName))
	} else {
		t = p.colorize(t, p.currentScheme.StructName)
	}
	return prefix + t
}

func (p *printer) matchRegexp(text, exp string) bool {
	return regexp.MustCompile(exp).MatchString(text)
}

func (p *printer) indented(proc func()) {
	p.depth++
	if p.maxDepth == -1 || p.depth <= p.maxDepth {
		proc()
	}
	p.depth--
}

func (p *printer) fmtOrLocalizedSprintf(format string, a ...any) string {
	if p.localizedPrinter == nil {
		return fmt.Sprintf(format, a...)
	}

	return p.localizedPrinter.Sprintf(format, a...)
}

func (p *printer) raw() string {
	// Some value causes panic when Interface() is called.
	switch p.value.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%#v", p.value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return p.fmtOrLocalizedSprintf("%v", p.value.Int())
	case reflect.Uint, reflect.Uintptr:
		if p.decimalUint {
			return p.fmtOrLocalizedSprintf("%d", p.value.Uint())
		} else {
			return fmt.Sprintf("%#v", p.value.Uint())
		}
	case reflect.Uint8:
		if p.decimalUint {
			return fmt.Sprintf("%d", p.value.Uint())
		} else {
			return fmt.Sprintf("0x%02x", p.value.Uint())
		}
	case reflect.Uint16:
		if p.decimalUint {
			return p.fmtOrLocalizedSprintf("%d", p.value.Uint())
		} else {
			return fmt.Sprintf("0x%04x", p.value.Uint())
		}
	case reflect.Uint32:
		if p.decimalUint {
			return p.fmtOrLocalizedSprintf("%d", p.value.Uint())
		} else {
			return fmt.Sprintf("0x%08x", p.value.Uint())
		}
	case reflect.Uint64:
		if p.decimalUint {
			return p.fmtOrLocalizedSprintf("%d", p.value.Uint())
		} else {
			return fmt.Sprintf("0x%016x", p.value.Uint())
		}
	case reflect.Float32, reflect.Float64:
		return p.fmtOrLocalizedSprintf("%f", p.value.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%#v", p.value.Complex())
	default:
		return fmt.Sprintf("%#v", p.value.Interface())
	}
}

func (p *printer) nil() string {
	return p.colorize("nil", p.currentScheme.Nil)
}

func (p *printer) colorize(text string, mod scuf.Modifier) string {
	if p.coloringEnabled {
		return scuf.String(text, mod)
	} else {
		return text
	}
}

func (p *printer) format(object any) string {
	pp := newPrinter(
		object,
		p.currentScheme,
		p.maxDepth,
		p.coloringEnabled,
		p.decimalUint,
		p.exportedOnly,
		p.thousandsSeparator,
	)
	pp.depth = p.depth
	pp.visited = p.visited
	if value, ok := object.(reflect.Value); ok {
		pp.value = value
	}
	return pp.String()
}

func (p *printer) indent() string {
	return strings.Repeat("    ", p.depth)
}

// valueIsZero reports whether v is the zero value for its type.
// It returns false if the argument is invalid.
// This is a copy paste of reflect#IsZero from go1.15. It is not present before go1.13 (source: https://golang.org/doc/go1.13#library)
// source: https://golang.org/src/reflect/value.go?s=34297:34325#L1090
// This will need to be updated for new types or the decision should be made to drop support for Go version pre go1.13
func valueIsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !valueIsZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !valueIsZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// this is the only difference between stdlib reflect#IsZero and this function. We're not going to
		// panic on the default cause, even
		return false
	}
}
