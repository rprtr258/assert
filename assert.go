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
	"unsafe"

	"github.com/rprtr258/fun/iter"
	"github.com/rprtr258/scuf"

	"github.com/rprtr258/assert/pp"
	"github.com/rprtr258/assert/q"
)

//go:linkname valueInterface reflect.valueInterface
func valueInterface(v reflect.Value, safe bool) any

// use instead of v.Interface
func valueToInterface(v reflect.Value) any {
	return valueInterface(v, false)
}

var (
	_fgExpected = scuf.FgRGB(0x96, 0xf7, 0x59)
	_fgActual   = scuf.FgRGB(0xff, 0x40, 0x53)
)

func mapJoin[T any](seq iter.Seq[T], toString func(T) string, sep string) string {
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
func callerInfo() iter.Seq[caller] {
	return func(yield func(caller) bool) bool {
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
						return false
					}
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
		return true
	}
}

type labeledContent struct {
	label   string
	content string
}

type diffLine struct {
	selector         string
	comment          string
	expected, actual any
}

func diffImpl(selectorPrefix string, expected, actual any) iter.Seq[diffLine] {
	etype, atype := reflect.TypeOf(expected), reflect.TypeOf(actual)
	if etype != atype {
		if etype == nil || atype == nil {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "different types SHIT POOP SHIT POOP SHIT FUCK SHIT",
				expected: etype,
				actual:   atype,
			})
		}

		return iter.FromMany(diffLine{
			selector: selectorPrefix,
			comment:  "different types",
			expected: etype.String(),
			actual:   atype.String(),
		})
	}

	eval, aval := reflect.ValueOf(expected), reflect.ValueOf(actual)

	switch etype.Kind() {
	case reflect.Invalid:
		return iter.FromNothing[diffLine]()
	case reflect.Bool:
		if e, a := expected.(bool), actual.(bool); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if e, a := eval.Int(), aval.Int(); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if e, a := eval.Uint(), aval.Uint(); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.Float32, reflect.Float64:
		if e, a := eval.Float(), aval.Float(); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.Complex64, reflect.Complex128:
		if e, a := eval.Complex(), aval.Complex(); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.String:
		if e, a := eval.String(), aval.String(); e != a {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  "",
				expected: e,
				actual:   a,
			})
		}

		return iter.FromNothing[diffLine]()
	case reflect.Pointer:
		return diffImpl(
			"(*"+selectorPrefix+")",
			eval.Elem().Interface(),
			aval.Elem().Interface(),
		)
	case reflect.Slice:
		lenExpected := eval.Len()
		lenActual := aval.Len()
		if lenExpected != lenActual {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  fmt.Sprintf("len: %d != %d", lenExpected, lenActual),
				expected: expected,
				actual:   actual,
			})
		}

		// check if only one is nil
		if lenExpected == 0 {
			if (expected == nil) != (actual == nil) {
				return iter.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "one slice is nil, other is not",
					expected: expected,
					actual:   actual,
				})
			}

			return iter.FromNothing[diffLine]()
		}

		return iter.FlatMap(
			iter.FromRange(0, lenExpected, 1),
			func(i int) iter.Seq[diffLine] {
				return diffImpl(
					fmt.Sprintf("%s[%d]", selectorPrefix, i),
					eval.Index(i).Interface(),
					aval.Index(i).Interface(),
				)
			})
	case reflect.Array:
		lenExpected := etype.Len()
		lenActual := atype.Len()
		if lenExpected != lenActual {
			return iter.FromMany(diffLine{
				selector: selectorPrefix,
				comment:  fmt.Sprintf("len: %d != %d", lenExpected, lenActual),
				expected: expected,
				actual:   actual,
			})
		}

		return iter.FlatMap(
			iter.FromRange(0, lenExpected, 1),
			func(i int) iter.Seq[diffLine] {
				return diffImpl(
					fmt.Sprintf("%s[%d]", selectorPrefix, i),
					eval.Index(i).Interface(),
					aval.Index(i).Interface(),
				)
			})
	case reflect.Struct:
		fields := etype.NumField()
		return iter.FlatMap(
			iter.FromRange(0, fields, 1),
			// Filter(func(i int) bool {
			// 	return typ.Field(i).IsExported() // private fields not shown then ??????????????????????????????????
			// }),
			func(i int) iter.Seq[diffLine] {
				if !etype.Field(i).IsExported() { // SHIT POOP SHIT FUCK SHIT POOP FUCK SHIT POOP
					ef, af := eval.Field(i), aval.Field(i)

					switch ef.Kind() {
					case reflect.String:
						if e, a := ef.String(), af.String(); e != a {
							return iter.FromMany(diffLine{
								selector: fmt.Sprintf("%s.%s", selectorPrefix, etype.Field(i).Name),
								comment:  "",
								expected: e,
								actual:   a,
							})
						} else {
							return iter.FromNothing[diffLine]()
						}
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						if e, a := ef.Int(), af.Int(); e != a {
							return iter.FromMany(diffLine{
								selector: fmt.Sprintf("%s.%s", selectorPrefix, etype.Field(i).Name),
								comment:  "",
								expected: e,
								actual:   a,
							})
						} else {
							return iter.FromNothing[diffLine]()
						}
					default:
						if ee, aa := reflect.NewAt(
							eval.Field(i).Type(),
							unsafe.Pointer(eval.Field(i).UnsafePointer()),
						).Elem().Interface(), reflect.NewAt(
							aval.Field(i).Type(),
							unsafe.Pointer(aval.Field(i).UnsafePointer()),
						).Elem().Interface(); ee != aa {
							return iter.FromMany(diffLine{
								selector: fmt.Sprintf("%s.%s", selectorPrefix, etype.Field(i).Name),
								comment:  fmt.Sprintf("field values are different, or not, i can't see and can't show them to you because reflect is crap and golang is crap and i cant access private field values even for reading but for primitives it is allowed but Interface method is disallowed: %s", etype.Field(i).Name),
								expected: aa,
								actual:   ee,
							})
						}

						return iter.FromNothing[diffLine]()
					}
				}

				if eval.Field(i).Comparable() && eval.Field(i).Interface() == aval.Field(i).Interface() {
					return iter.FromNothing[diffLine]()
				}

				return diffImpl(
					fmt.Sprintf("%s.%s", selectorPrefix, etype.Field(i).Name),
					eval.Field(i).Interface(),
					aval.Field(i).Interface(),
				)
			})
		// case reflect.Map:
		// 	expectedKeys := map[any]struct{}{}
		// 	for _, k := range eval.MapKeys() {
		// 		expectedKeys[k.Interface()] = struct{}{}
		// 	}

		// 	actualKeys := map[any]struct{}{}
		// 	for _, k := range aval.MapKeys() {
		// 		actualKeys[k.Interface()] = struct{}{}
		// 	}

		// 	commonKeys := map[any]struct{}{}
		// 	expectedOnlyKeys := map[any]struct{}{}
		// 	for k := range expectedKeys {
		// 		if _, ok := actualKeys[k]; ok {
		// 			commonKeys[k] = struct{}{}
		// 		} else {
		// 			expectedOnlyKeys[k] = struct{}{}
		// 		}
		// 	}

		// 	actualOnlyKeys := map[any]struct{}{}
		// 	for k := range actualKeys {
		// 		if _, ok := expectedKeys[k]; !ok {
		// 			actualOnlyKeys[k] = struct{}{}
		// 		}
		// 	}

		// 	return iter.Flatten(iter.FromMany(
		// 		iter.FlatMap(
		// 			iter.Keys(iter.FromDict(commonKeys)),
		// 			func(k any) iter.Seq[diffLine] {
		// 				return diffImpl(
		// 					fmt.Sprintf("%s[%v]", selectorPrefix, k),
		// 					eval.MapIndex(reflect.ValueOf(k)),
		// 					aval.MapIndex(reflect.ValueOf(k)),
		// 				)
		// 			}),
		// 		iter.Map(
		// 			iter.Keys(iter.FromDict(expectedOnlyKeys)),
		// 			func(k any) diffLine {
		// 				return diffLine{
		// 					selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
		// 					comment:  "not found key in actual",
		// 					expected: eval.MapIndex(reflect.ValueOf(k)).Interface(),
		// 					actual:   nil,
		// 				}
		// 			}),
		// 		iter.Map(
		// 			iter.Keys(iter.FromDict(actualOnlyKeys)),
		// 			func(k any) diffLine {
		// 				return diffLine{
		// 					selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
		// 					comment:  "unexpected key in actual",
		// 					expected: nil,
		// 					actual:   aval.MapIndex(reflect.ValueOf(k)).Interface(),
		// 				}
		// 			}),
		// 	))
		// case reflect.Interface:
		// 	if (expected == nil) && (actual == nil) {
		// 		return iter.FromNothing[diffLine]()
		// 	}

		// 	if (expected == nil) || (actual == nil) {
		// 		return iter.FromMany(diffLine{
		// 			selector: selectorPrefix,
		// 			comment:  "one is nil, one is not",
		// 			expected: expected,
		// 			actual:   actual,
		// 		})
		// 	}

		// 	return diffImpl(selectorPrefix, eval.Elem(), aval.Elem())
	}

	// TODO: remove and return "" when other types are supported
	panic(fmt.Sprintf("unsupported kind: %s, %#v", eval.Kind().String(), eval.Interface()))
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it panics.
func diff[T any](expected, actual T) iter.Seq[diffLine] {
	return diffImpl("", expected, actual)
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

func fail(t testing.TB, lines []labeledContent) {
	t.Error("\n" + mapJoin(iter.FromMany(lines...), formatLabeledContent, "\n"))
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
			scuf.String("Value is zero", scuf.FgHiRed),
			fmt.Sprintf("%s is zero, asserted not to", scuf.String(actualName, _fgActual)),
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
			conditionName + scuf.String(" is false", scuf.FgHiRed),
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
			conditionName + scuf.String(" is false", scuf.FgHiRed),
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
			conditionName + scuf.String(" is true", scuf.FgHiRed),
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
			conditionName + scuf.String(" is true", scuf.FgHiRed),
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

func Regexp(t *testing.T, re, text string) {
	True(t, regexp.MustCompile(re).MatchString(text))
}

func EqualError(t *testing.T, errText string, err error) {
	Equal(t, errText, err.Error())
}

func Len[T any](t *testing.T, lenn int, slice []T) {
	Equal(t, lenn, len(slice))
}
