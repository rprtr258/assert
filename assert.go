package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/rprtr258/assert/internal/fun"
	"github.com/rprtr258/assert/internal/scuf"
	"github.com/rprtr258/assert/pp"
	"github.com/rprtr258/assert/q"
)

var (
	_fgExpected = scuf.FgRGB(0x96, 0xf7, 0x59)
	_fgActual   = scuf.FgRGB(0xff, 0x40, 0x53)
)

func mapJoin[T any](seq fun.Seq[T], toString func(T) string, sep string) string {
	var sb strings.Builder
	seq(func(v T) bool {
		if sb.Len() > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(toString(v))

		return true
	})
	return sb.String()
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

// isTest tells whether name looks like a test or benchmark, according to prefix.
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name string) bool {
	for _, prefix := range []string{"Test", "Benchmark", "Example"} {
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		if len(name) == len(prefix) { // "Test" is ok
			return true
		}

		r, _ := utf8.DecodeRuneInString(name[len(prefix):])
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// caller is necessary because the assert functions use the testing object
// internally, causing it to print the file:line of the assert method, rather
// than where the problem actually occurred in calling code.
type caller struct {
	file     string
	line     int
	funcName string
}

// callerInfo returns an array of strings containing the file and line number
// of each stack frame leading from the current test to the assert call that
// failed.
func callerInfo() fun.Seq[caller] {
	return func(yield func(caller) bool) {
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
					if !yield(caller{path, line, name}) {
						return
					}
				}
			}

			// Drop the package
			segments := strings.Split(name, ".")
			name = segments[len(segments)-1]
			if isTest(name) {
				break
			}
		}
	}
}

type labeledContent struct {
	label   string
	content string
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

func Equal[T any](tb testing.TB, expected, actual T) {
	tb.Helper()

	if reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "Equal")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			scuf.String("Not equal", scuf.FgHiRed),
			mapJoin(diff(expected, actual), func(line diffLine) string {
				if line.expected == nil { // TODO: remove
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

				expectedStr := shorten(expectedName, pp.Sprint(line.expected))
				actualStr := shorten(actualName, pp.Sprint(line.actual))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					return fun.Ternary(line.comment != "", line.comment+":\n", "") +
						scuf.String(expectedName+line.selector, _fgExpected) + " = " + expectedStr + "\n" +
						scuf.String(actualName+line.selector, _fgActual) + " = " + actualStr
				}

				comment := fun.Ternary(line.comment != "", ", "+line.comment, "")
				return scuf.String(expectedName+line.selector, _fgExpected) + " != " + scuf.String(actualName+line.selector, _fgActual) + comment + ":\n" +
					"\t" + expectedStr + " !=\n" +
					"\t" + actualStr
			}, "\n\n"),
		},
	})
}

func Equalf[T any](tb testing.TB, expected, actual T, format string, args ...any) {
	tb.Helper()

	if reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "Equal")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			scuf.String("Not equal", scuf.FgHiRed),
			mapJoin(diff(expected, actual), func(line diffLine) string {
				if line.expected == nil { // TODO: remove
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

				expectedStr := shorten(expectedName, pp.Sprint(line.expected))
				actualStr := shorten(actualName, pp.Sprint(line.actual))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					comment := ""
					if line.comment != "" {
						comment = line.comment + ":"
					}

					return strings.Join([]string{
						comment,
						scuf.String(expectedName+line.selector, _fgExpected) + " = " + expectedStr,
						scuf.String(actualName+line.selector, _fgActual) + " = " + actualStr,
					}, "\n")
				}

				comment := ""
				if line.comment != "" {
					comment = ", " + line.comment
				}

				return strings.Join([]string{
					fmt.Sprintf(
						"%s != %s%s:",
						scuf.String(expectedName+line.selector, _fgExpected),
						scuf.String(actualName+line.selector, _fgActual),
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

func fail(tb testing.TB, lines []labeledContent) {
	tb.Error("\n" + mapJoin(fun.FromMany(lines...), formatLabeledContent, "\n"))
}

func stacktraceLabeledContent() labeledContent {
	return labeledContent{
		scuf.String("Stacktrace", scuf.ModFaint),
		mapJoin(callerInfo(), func(v caller) string {
			j := strings.LastIndexByte(v.funcName, '/')
			shortFuncName := v.funcName[j+1:]
			return scuf.String(v.file, scuf.FgHiWhite) +
				":" +
				scuf.String(strconv.Itoa(v.line), scuf.FgGreen) +
				"\t" +
				scuf.String(shortFuncName, scuf.FgBlue)
		}, "\n"),
	}
}

func NotEqual[T any](tb testing.TB, expected, actual T) {
	tb.Helper()

	if !reflect.DeepEqual(expected, actual) {
		return
	}

	argNames := q.Q("assert", "NotEqual")
	expectedName := or(argNames[1], "Expected")
	actualName := or(argNames[2], "Actual")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			scuf.String("Equal", scuf.FgHiRed),
			strings.Join([]string{
				fmt.Sprintf(
					"%s and %s are equal, asserted not to, value is:",
					scuf.String(expectedName, _fgExpected),
					scuf.String(actualName, _fgActual),
				),
				"\t" + strings.ReplaceAll(pp.Sprint(expected), "\n", "\n\t"),
			}, "\n"),
		},
	})
}

func Zero[T any](tb testing.TB, actual T) {
	tb.Helper()

	var zero T
	if reflect.DeepEqual(zero, actual) {
		return
	}

	argNames := q.Q("assert", "Zero")
	expectedName := "Zero"
	actualName := or(argNames[1], "Actual")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			scuf.String("Not equal", scuf.FgHiRed),
			mapJoin(diff(zero, actual), func(line diffLine) string {
				if line.expected == nil { // TODO: remove
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

				expectedStr := shorten(expectedName, pp.Sprint(line.expected))
				actualStr := shorten(actualName, pp.Sprint(line.actual))

				if strings.ContainsRune(expectedStr, '\n') || strings.ContainsRune(actualStr, '\n') {
					return fun.Ternary(line.comment == "", "", line.comment+":") + "\n" +
						scuf.String(expectedName+line.selector, _fgExpected) + " = " + expectedStr + "\n" +
						scuf.String(actualName+line.selector, _fgActual) + " = " + actualStr
				}

				comment := fun.Ternary(line.comment == "", "", ", "+line.comment)
				return scuf.String(expectedName+line.selector, _fgExpected) + " != " + scuf.String(actualName+line.selector, _fgActual) + comment + ":\n" +
					"\t" + expectedStr + " != " + actualStr
			}, "\n\n"),
		},
	})
}

func NotZero[T any](tb testing.TB, actual T) {
	tb.Helper()

	var zero T
	if !reflect.DeepEqual(zero, actual) {
		return
	}

	argNames := q.Q("assert", "NotZero")
	actualName := or(argNames[1], "Actual")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			scuf.String("Value is zero", scuf.FgHiRed),
			fmt.Sprintf("%s is zero, asserted not to", scuf.String(actualName, _fgActual)),
		},
	})
}

func True(tb testing.TB, condition bool) {
	tb.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "True")
	conditionName := or(argNames[1], "Condition")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Condition is false",
			conditionName + scuf.String(" is false", scuf.FgHiRed),
		},
	})
}

func Truef(tb testing.TB, condition bool, format string, args ...any) {
	tb.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "Truef")
	conditionName := or(argNames[1], "Condition")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Condition is false",
			conditionName + scuf.String(" is false", scuf.FgHiRed),
		},
	})
}

func False(tb testing.TB, condition bool) {
	tb.Helper()

	if !condition {
		return
	}

	argNames := q.Q("assert", "False")
	conditionName := or(argNames[1], "Condition")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Condition is true",
			conditionName + scuf.String(" is true", scuf.FgHiRed),
		},
	})
}

func Falsef(tb testing.TB, condition bool, format string, args ...any) {
	tb.Helper()

	if condition {
		return
	}

	argNames := q.Q("assert", "Falsef")
	conditionName := or(argNames[1], "Condition")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Condition is true",
			conditionName + scuf.String(" is true", scuf.FgHiRed),
		},
	})
}

func NoError(tb testing.TB, err error) {
	tb.Helper()

	if err == nil {
		return
	}

	argNames := q.Q("assert", "NoError")
	errorName := or(argNames[1], "Error")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		{
			"Unexpected error",
			errorName + " is " + pp.Sprint(err.Error()),
		},
	})
}

func NoErrorf(tb testing.TB, err error, format string, args ...any) {
	tb.Helper()

	if err == nil {
		return
	}

	argNames := q.Q("assert", "NoErrorf")
	errorName := or(argNames[1], "Error")

	fail(tb, []labeledContent{
		stacktraceLabeledContent(),
		messageLabeledContent(format, args...),
		{
			"Unexpected error",
			errorName + " is " + pp.Sprint(err.Error()),
		},
	})
}

func Contains[T comparable](tb testing.TB, slice []T, item T) {
	for _, v := range slice {
		if v == item {
			return
		}
	}

	argNames := q.Q("assert", "Contains")
	sliceName := or(argNames[1], "Slice")
	itemName := or(argNames[2], "Item")

	fail(tb, []labeledContent{
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

func Substring(tb testing.TB, text, needle string) {
	if strings.Contains(text, needle) {
		return
	}

	argNames := q.Q("assert", "Substring")
	textName := or(argNames[1], "Text")
	needleName := or(argNames[2], "Needle")

	fail(tb, []labeledContent{
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

func Substringf(tb testing.TB, text, needle string, format string, args ...any) {
	if strings.Contains(text, needle) {
		return
	}

	argNames := q.Q("assert", "Substringf")
	textName := or(argNames[1], "Text")
	needleName := or(argNames[2], "Needle")

	fail(tb, []labeledContent{
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

func Regexp(tb testing.TB, re, text string) {
	True(tb, regexp.MustCompile(re).MatchString(text))
}

func EqualError(tb testing.TB, errText string, err error) {
	Equal(tb, errText, err.Error())
}

func Len[T any](tb testing.TB, lenn int, slice []T) {
	Equal(tb, lenn, len(slice))
}
