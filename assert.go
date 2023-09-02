package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/k0kubun/pp"
	"github.com/muesli/termenv"

	"github.com/rprtr258/assert/q"
)

func mapJoin[T any](slice []T, toString func(T) string, sep string) string {
	parts := make([]string, len(slice))
	for i, v := range slice {
		parts[i] = toString(v)
	}
	return strings.Join(parts, sep)
}

// or returns the first non-zero value
func or[T comparable](xs ...T) T {
	var zero T
	for _, x := range xs {
		if x != zero {
			return x
		}
	}
	return zero
}

// Stolen from the `go test` tool.
// isTest tells whether name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name, prefix string) bool {
	switch {
	case !strings.HasPrefix(name, prefix):
		return false
	case len(name) == len(prefix): // "Test" is ok
		return true
	default:
		r, _ := utf8.DecodeRuneInString(name[len(prefix):])
		return !unicode.IsLower(r)
	}
}

/* CallerInfo is necessary because the assert functions use the testing object
internally, causing it to print the file:line of the assert method, rather than where
the problem actually occurred in calling code.*/

type caller struct {
	file     string
	line     int
	funcName string
}

// CallerInfo returns an array of strings containing the file and line number
// of each stack frame leading from the current test to the assert call that
// failed.
func CallerInfo() []caller {
	callers := []caller{}
	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		// This is a huge edge case, but it will panic if this is the case, see #180
		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name := f.Name()

		// testing.tRunner is the standard library function that calls
		// tests. Subtests are called directly by tRunner, without going through
		// the Test/Benchmark/Example function that contains the t.Run calls, so
		// with subtests we should break when we hit tRunner, without adding it
		// to the list of callers.
		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
		if len(parts) > 1 {
			dir := parts[len(parts)-2]
			if dir != "assert" && dir != "mock" && dir != "require" || file == "mock_test.go" {
				path, _ := filepath.Abs(file)
				callers = append(callers, caller{path, line, name})
			}
		}

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}
	return callers
}

type labeledContent struct {
	label   string
	content string
}

type diffLine struct {
	selector         string
	comment          string
	expected, actual reflect.Value
}

// TODO: change to iterators
func diffImpl(selectorPrefix string, expected, actual reflect.Value) []diffLine {
	switch expected.Kind() {
	case reflect.Int:
		if expected.Int() != actual.Int() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.Pointer:
		return diffImpl(
			"(*"+selectorPrefix+")",
			reflect.ValueOf(expected).Elem(),
			reflect.ValueOf(actual).Elem(),
		)
	case reflect.Slice:
		lenExpected := expected.Len()
		lenActual := actual.Len()
		if lenExpected != lenActual {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  fmt.Sprintf("len: %d != %d", lenExpected, lenActual),
				expected: expected,
				actual:   actual,
			}}
		}

		lines := make([]diffLine, lenExpected)
		for i := range lines {
			lines = append(lines, diffImpl(
				fmt.Sprintf("%s[%d]", selectorPrefix, i),
				expected.Index(i),
				actual.Index(i),
			)...)
		}
		return lines
	}

	panic(fmt.Sprintf("unsupported kind: %s", expected.Kind().String()))
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it returns an empty string.
func diff[T any](expected, actual T) []diffLine {
	return diffImpl("", reflect.ValueOf(expected), reflect.ValueOf(actual))
	// switch reflect.TypeOf(expected).Kind() {
	// case reflect.Pointer:
	// 	return
	// }

	// if ek != reflect.Struct &&
	// 	ek != reflect.Map &&
	// 	ek != reflect.Slice &&
	// 	ek != reflect.Array &&
	// 	ek != reflect.String {
	// 	return nil
	// }

	// return []diffLine{
	// 	{"[0]", "1", "2"},
	// 	{"[1]", "2", "3"},
	// 	{"[2]", "2", "4"},
	// }
}

func Equal[T any](t testing.TB, expected, actual T) {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "Equal")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	t.Error("\n" + mapJoin([]labeledContent{
		{
			termenv.String("Stacktrace").Faint().String(),
			mapJoin(CallerInfo(), func(v caller) string {
				j := strings.LastIndexByte(v.funcName, '/')
				shortFuncName := v.funcName[j+1:]
				return termenv.String(v.file).Foreground(termenv.ANSIBrightWhite).String() +
					":" +
					termenv.String(strconv.Itoa(v.line)).Foreground(termenv.ANSIGreen).String() +
					"\t" +
					termenv.String(shortFuncName).Foreground(termenv.ANSIBlue).String()

			}, "\n"),
		},
		{
			termenv.String(expectedName).Faint().String(),
			pp.Sprint(expected),
		},
		{
			termenv.String(actualName).Faint().String(),
			pp.Sprint(actual),
		},
		{
			termenv.String("Not equal").Faint().String(),
			mapJoin(diff(expected, actual), func(line diffLine) string {
				expectedStr := pp.Sprint(line.expected.Interface())
				actualStr := pp.Sprint(line.actual.Interface())

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					return strings.Join([]string{
						termenv.String(expectedName).Faint().String() + line.selector + " != " + termenv.String(actualName).Faint().String() + line.selector + ":",
						termenv.String(expectedStr).Foreground(termenv.ANSIBrightRed).String(),
						termenv.String(actualStr).Foreground(termenv.ANSIBrightGreen).String(),
					}, "\n")
				}

				return strings.Join([]string{
					termenv.String(expectedName).Faint().String() + line.selector + " != " + termenv.String(actualName).Faint().String() + line.selector + ":",
					"\t" + termenv.String(expectedStr).Foreground(termenv.ANSIBrightRed).String() + " != " + termenv.String(actualStr).Foreground(termenv.ANSIBrightGreen).String(),
				}, "\n")
			}, "\n"),
		},
	}, func(v labeledContent) string {
		return v.label +
			":\n    " +
			strings.ReplaceAll(v.content, "\n", "\n    ")
	}, "\n"))
}

// func Equalf[T any](t testing.TB, expected, actual T, format string, args ...any) {
// 	if a.ObjectsAreEqual(expected, actual) {
// 		return
// 	}

// 	diff := diff(expected, actual)
// 	Fail(t, fmt.Sprintf("Not equal:\n"+
// 		"expected: %q\n"+
// 		"actual  : %q%q", q.Q(expected), q.Q(actual), diff), append([]any{format}, args...))
// }

// func NotEqual[T any](t *testing.T, expected, actual T) {
// 	t.Helper()

// 	if !a.ObjectsAreEqual(expected, actual) {
// 		return
// 	}

// 	diff := diff(expected, actual)
// 	Fail(t, fmt.Sprintf("Equal: \n"+
// 		"expected: %s\n"+
// 		"actual  : %s%s", q.Q(expected), q.Q(actual), diff))
// }

func Zero[T any](t *testing.T, actual T) {
	var zero T
	Equal(t, zero, actual)
}

// func NotZero[T any](t *testing.T, actual T) {
// 	var zero T
// 	NotEqual(t, zero, actual)
// }

// func False(t *testing.T, actual bool) {
// 	Equal(t, false, actual)
// }

// func Falsef(t *testing.T, actual bool, format string, args ...any) {
// 	Equalf(t, false, actual, format, args...)
// }

func True(t *testing.T, actual bool) {
	Equal(t, true, actual)
}

// func Truef(t *testing.T, actual bool, format string, args ...any) {
// 	Equalf(t, true, actual, format, args...)
// }

// func NoError(t testing.TB, err error) {
// 	Equal(t, nil, err)
// }

// func Contains[T comparable](t *testing.T, slice []T, item T) {
// 	for _, v := range slice {
// 		if v == item {
// 			return
// 		}
// 	}

// 	Fail(t, fmt.Sprintf("Slice does not contain %s", spew.Sdump(item)))
// }

// func Substring(t *testing.T, text, needle string) {
// 	if strings.Contains(text, needle) {
// 		return
// 	}

// 	Fail(t, fmt.Sprintf("%s does not contain %s", spew.Sdump(text), spew.Sdump(needle)))
// }

// func Substringf(t *testing.T, text, needle string, format string, args ...any) {
// 	if strings.Contains(text, needle) {
// 		return
// 	}

// 	Fail(t, fmt.Sprintf("%s does not contain %s", spew.Sdump(text), spew.Sdump(needle)), append([]any{format}, args...))
// }
