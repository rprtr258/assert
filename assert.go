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

	col "github.com/rprtr258/col"

	"github.com/rprtr258/assert/pp"
	"github.com/rprtr258/assert/q"
)

var (
	_fgExpected = col.RGBColor("#96f759").Fg
	_fgActual   = col.RGBColor("#ff4053").Fg
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

// callerInfo returns an array of strings containing the file and line number
// of each stack frame leading from the current test to the assert call that
// failed.
func callerInfo() []caller {
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
	case reflect.Invalid:
		return nil
	case reflect.Bool:
		if expected.Bool() != actual.Bool() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if expected.Int() != actual.Int() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if expected.Uint() != actual.Uint() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.Float32, reflect.Float64:
		if expected.Float() != actual.Float() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.Complex64, reflect.Complex128:
		if expected.Complex() != actual.Complex() {
			return []diffLine{{
				selector: selectorPrefix,
				comment:  "",
				expected: expected,
				actual:   actual,
			}}
		}

		return nil
	case reflect.String:
		if expected.String() != actual.String() {
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

		// check if only one is nil
		if lenExpected == 0 {
			if expected.IsNil() != actual.IsNil() {
				return []diffLine{{
					selector: selectorPrefix,
					comment:  "",
					expected: expected,
					actual:   actual,
				}}
			}

			return nil
		}

		lines := []diffLine{}
		for i := 0; i < lenExpected; i++ {
			lines = append(lines, diffImpl(
				fmt.Sprintf("%s[%d]", selectorPrefix, i),
				expected.Index(i),
				actual.Index(i),
			)...)
		}
		return lines
	case reflect.Array:
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

		lines := []diffLine{}
		for i := 0; i < lenExpected; i++ {
			lines = append(lines, diffImpl(
				fmt.Sprintf("%s[%d]", selectorPrefix, i),
				expected.Index(i),
				actual.Index(i),
			)...)
		}
		return lines
	case reflect.Struct:
		typ := expected.Type()
		lines := []diffLine{}
		fields := typ.NumField()
		for i := 0; i < fields; i++ {
			lines = append(lines, diffImpl(
				fmt.Sprintf("%s.%s", selectorPrefix, typ.Field(i).Name),
				expected.Field(i),
				actual.Field(i),
			)...)
		}
		return lines
	case reflect.Map:
		expectedKeys := map[any]struct{}{}
		for _, k := range expected.MapKeys() {
			expectedKeys[k.Interface()] = struct{}{}
		}

		actualKeys := map[any]struct{}{}
		for _, k := range actual.MapKeys() {
			actualKeys[k.Interface()] = struct{}{}
		}

		commonKeys := map[any]struct{}{}
		expectedOnlyKeys := map[any]struct{}{}
		for k := range expectedKeys {
			if _, ok := actualKeys[k]; ok {
				commonKeys[k] = struct{}{}
			} else {
				expectedOnlyKeys[k] = struct{}{}
			}
		}

		actualOnlyKeys := map[any]struct{}{}
		for k := range actualKeys {
			if _, ok := expectedKeys[k]; !ok {
				actualOnlyKeys[k] = struct{}{}
			}
		}

		lines := []diffLine{}
		for k := range commonKeys {
			lines = append(lines, diffImpl(
				fmt.Sprintf("%s[%v]", selectorPrefix, k),
				expected.MapIndex(reflect.ValueOf(k)),
				actual.MapIndex(reflect.ValueOf(k)),
			)...)
		}
		for k := range expectedOnlyKeys {
			lines = append(lines, diffLine{
				selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
				comment:  "not found key in actual",
				expected: expected.MapIndex(reflect.ValueOf(k)),
				actual:   reflect.Value{},
			})
		}
		for k := range actualOnlyKeys {
			lines = append(lines, diffLine{
				selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
				comment:  "unexpected key in actual",
				expected: reflect.Value{},
				actual:   actual.MapIndex(reflect.ValueOf(k)),
			})
		}
		return lines
	}

	// TODO: remove and return "" when other types are supported
	panic(fmt.Sprintf("unsupported kind: %s, %#v", expected.Kind().String(), expected.Interface()))
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it panics.
func diff[T any](expected, actual T) []diffLine {
	return diffImpl("", reflect.ValueOf(expected), reflect.ValueOf(actual))
}

func formatLabeledContent(v labeledContent) string {
	return v.label +
		":\n    " +
		strings.ReplaceAll(v.content, "\n", "\n    ")
}

func messageLabeledContent(format string, arg ...any) labeledContent {
	return labeledContent{
		label:   "Message",
		content: fmt.Sprintf(format, arg...),
	}
}

func Equal[T any](t testing.TB, expected, actual T) {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "Equal")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			col.R("Not equal", col.ANSIBrightRed.Fg),
			mapJoin(diff(expected, actual), func(line diffLine) string {
				/*
					Bit complaining on go language: brackets on struct literal are
					required here because compiler authors can't fix parser
					and not interpret '{' as "if block" and that won't be fixed.
					See https://github.com/golang/go/issues/9181
				*/
				if line.expected == (reflect.Value{}) { // TODO: remove
					return line.selector
				}

				shorten := func(name, s string) string {
					// TODO: do string width if this code is kept
					short := strings.NewReplacer(
						"{\n    ", "{",
						",\n    ", ", ",
						",\n", "",
					).Replace(s)
					if len(name)+len(s) < 100 {
						return short
					}

					return s
				}

				expectedStr := shorten(expectedName, pp.Sprint(line.expected.Interface()))
				actualStr := shorten(actualName, pp.Sprint(line.actual.Interface()))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					comment := ""
					if line.comment != "" {
						comment = line.comment + ":"
					}

					return strings.Join([]string{
						comment,
						col.R(expectedName+line.selector, _fgExpected) + " = " + expectedStr,
						col.R(actualName+line.selector, _fgActual) + " = " + actualStr,
					}, "\n")
				}

				comment := ""
				if line.comment != "" {
					comment = ", " + line.comment
				}

				return strings.Join([]string{
					fmt.Sprintf(
						"%s != %s%s:",
						col.R(expectedName+line.selector, _fgExpected),
						col.R(actualName+line.selector, _fgActual),
						comment,
					),
					fmt.Sprintf(
						"\t%s !=\n\t%s",
						expectedStr,
						actualStr,
					),
				}, "\n")
			}, "\n\n"),
		},
	})
}

func Equalf[T any](t testing.TB, expected, actual T, format string, args ...any) {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "Equal")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			col.R("Not equal", col.ANSIBrightRed.Fg),
			mapJoin(diff(expected, actual), func(line diffLine) string {
				/*
					Bit complaining on go language: brackets on struct literal are
					required here because compiler authors can't fix parser
					and not interpret '{' as "if block" and that won't be fixed.
					See https://github.com/golang/go/issues/9181
				*/
				if line.expected == (reflect.Value{}) { // TODO: remove
					return line.selector
				}

				shorten := func(name, s string) string {
					// TODO: do string width if this code is kept
					short := strings.NewReplacer(
						"{\n    ", "{",
						",\n    ", ", ",
						",\n", "",
					).Replace(s)
					if len(name)+len(s) < 100 {
						return short
					}

					return s
				}

				expectedStr := shorten(expectedName, pp.Sprint(line.expected.Interface()))
				actualStr := shorten(actualName, pp.Sprint(line.actual.Interface()))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					comment := ""
					if line.comment != "" {
						comment = line.comment + ":"
					}

					return strings.Join([]string{
						comment,
						col.R(expectedName+line.selector, _fgExpected) + " = " + expectedStr,
						col.R(actualName+line.selector, _fgActual) + " = " + actualStr,
					}, "\n")
				}

				comment := ""
				if line.comment != "" {
					comment = ", " + line.comment
				}

				return strings.Join([]string{
					fmt.Sprintf(
						"%s != %s%s:",
						col.R(expectedName+line.selector, _fgExpected),
						col.R(actualName+line.selector, _fgActual),
						comment,
					),
					fmt.Sprintf(
						"\t%s !=\n\t%s",
						expectedStr,
						actualStr,
					),
				}, "\n")
			}, "\n\n"),
		},
	})
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

func fail(t testing.TB, lines []labeledContent) {
	t.Error("\n" + mapJoin(lines, formatLabeledContent, "\n"))
}

func stacktraceLabeledContent() labeledContent {
	return labeledContent{
		col.R("Stacktrace", col.Faint),
		mapJoin(callerInfo(), func(v caller) string {
			j := strings.LastIndexByte(v.funcName, '/')
			shortFuncName := v.funcName[j+1:]
			return col.R(v.file, col.ANSIBrightWhite.Fg) +
				":" +
				col.R(strconv.Itoa(v.line), col.ANSIGreen.Fg) +
				"\t" +
				col.R(shortFuncName, col.ANSIBlue.Fg)

		}, "\n"),
	}
}

func NotEqual[T any](t *testing.T, expected, actual T) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "NotEqual")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			col.R("Equal", col.ANSIBrightRed.Fg),
			strings.Join([]string{
				fmt.Sprintf(
					"%s and %s are equal, asserted not to, value is:",
					col.R(expectedName, _fgExpected),
					col.R(actualName, _fgActual),
				),
				"\t" + strings.ReplaceAll(pp.Sprint(expected), "\n", "\n\t"),
			}, "\n"),
		},
	})
}

func Zero[T any](t *testing.T, actual T) {
	t.Helper()

	var zero T
	if reflect.DeepEqual(zero, actual) {
		return
	}

	argNames := q.Q("assert", "Zero")
	expectedName := "Zero"
	actualName := or(argNames[1], "Actual")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			col.R("Not equal", col.ANSIBrightRed.Fg),
			mapJoin(diff(zero, actual), func(line diffLine) string {
				/*
					Bit complaining on go language: brackets on struct literal are
					required here because compiler authors can't fix parser
					and not interpret '{' as "if block" and that won't be fixed.
					See https://github.com/golang/go/issues/9181
				*/
				if line.expected == (reflect.Value{}) { // TODO: remove
					return line.selector
				}

				shorten := func(name, s string) string {
					// TODO: do string width if this code is kept
					short := strings.NewReplacer(
						"{\n    ", "{",
						",\n    ", ", ",
						",\n", "",
					).Replace(s)
					if len(name)+len(s) < 100 {
						return short
					}

					return s
				}

				expectedStr := shorten(expectedName, pp.Sprint(line.expected.Interface()))
				actualStr := shorten(actualName, pp.Sprint(line.actual.Interface()))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					comment := ""
					if line.comment != "" {
						comment = line.comment + ":"
					}

					return strings.Join([]string{
						comment,
						col.R(expectedName+line.selector, _fgExpected) + " = " + expectedStr,
						col.R(actualName+line.selector, _fgActual) + " = " + actualStr,
					}, "\n")
				}

				comment := ""
				if line.comment != "" {
					comment = ", " + line.comment
				}

				return strings.Join([]string{
					fmt.Sprintf(
						"%s != %s%s:",
						col.R(expectedName+line.selector, _fgExpected),
						col.R(actualName+line.selector, _fgActual),
						comment,
					),
					fmt.Sprintf("\t%s != %s", expectedStr, actualStr),
				}, "\n")
			}, "\n\n"),
		},
	})
}

func NotZero[T any](t *testing.T, actual T) {
	t.Helper()

	var zero T
	if !reflect.DeepEqual(zero, actual) {
		return
	}

	argNames := q.Q("assert", "NotZero")
	actualName := or(argNames[1], "Actual")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			col.R("Value is zero", col.ANSIBrightRed.Fg),
			fmt.Sprintf("%s is zero, asserted not to", col.R(actualName, _fgActual)),
		},
	})
}

func True(t *testing.T, condition bool) {
	t.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "True")
	conditionName := or(argNames[1], "Condition")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Condition is false",
			conditionName + col.R(" is false", col.ANSIBrightRed.Fg),
		},
	})
}

func Truef(t *testing.T, condition bool, format string, args ...any) {
	t.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "Truef")
	conditionName := or(argNames[1], "Condition")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Condition is false",
			conditionName + col.R(" is false", col.ANSIBrightRed.Fg),
		},
	})
}

func False(t *testing.T, condition bool) {
	t.Helper()

	if !condition {
		return
	}

	argNames := q.Q("assert", "False")
	conditionName := or(argNames[1], "Condition")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Condition is true",
			conditionName + col.R(" is true", col.ANSIBrightRed.Fg),
		},
	})
}

func Falsef(t *testing.T, condition bool, format string, args ...any) {
	t.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "Falsef")
	conditionName := or(argNames[1], "Condition")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Condition is true",
			conditionName + col.R(" is true", col.ANSIBrightRed.Fg),
		},
	})
}

func NoError(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		return
	}

	argNames := q.Q("assert", "NoError")
	errorName := or(argNames[1], "Error")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Unexpected error",
			errorName + " is " + pp.Sprint(err.Error()),
		},
	})
}

func NoErrorf(t testing.TB, err error, format string, args ...any) {
	t.Helper()

	if err == nil {
		return
	}

	argNames := q.Q("assert", "NoErrorf")
	errorName := or(argNames[1], "Error")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Unexpected error",
			errorName + " is " + pp.Sprint(err.Error()),
		},
	})
}

func Contains[T comparable](t *testing.T, slice []T, item T) {
	for _, v := range slice {
		if v == item {
			return
		}
	}

	argNames := q.Q("assert", "Contains")
	sliceName := or(argNames[1], "Slice")
	itemName := or(argNames[2], "Item")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			label: "Slice does not contain item",
			content: strings.Join([]string{
				sliceName + ": " + pp.Sprint(slice),
				itemName + ": " + pp.Sprint(item),
			}, "\n"),
		},
	})
}

func Substring(t *testing.T, text, needle string) {
	if strings.Contains(text, needle) {
		return
	}

	argNames := q.Q("assert", "Substring")
	textName := or(argNames[1], "Text")
	needleName := or(argNames[2], "Needle")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		{
			label: "String does not contain substring",
			content: strings.Join([]string{
				textName + ": " + pp.Sprint(text),
				needleName + ": " + pp.Sprint(needle),
			}, "\n"),
		},
	})
}

func Substringf(t *testing.T, text, needle string, format string, args ...any) {
	if strings.Contains(text, needle) {
		return
	}

	argNames := q.Q("assert", "Substringf")
	textName := or(argNames[1], "Text")
	needleName := or(argNames[2], "Needle")

	fail(t, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			label: "String does not contain substring",
			content: strings.Join([]string{
				textName + ": " + pp.Sprint(text),
				needleName + ": " + pp.Sprint(needle),
			}, "\n"),
		},
	})
}
