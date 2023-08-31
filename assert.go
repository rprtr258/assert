package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
	a "github.com/stretchr/testify/assert"

	"github.com/rprtr258/assert/q"
)

func typeAndKind(v any) (reflect.Type, reflect.Kind) {
	t := reflect.TypeOf(v)
	k := t.Kind()

	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}
	return t, k
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it returns an empty string.
func diff[T any](expected, actual T) string {
	// if expected == nil || actual == nil {
	// 	return ""
	// }

	et, ek := typeAndKind(expected)
	at, _ := typeAndKind(actual)

	if et != at {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array && ek != reflect.String {
		return ""
	}

	var e, a string

	switch et {
	case reflect.TypeOf(""):
		e = reflect.ValueOf(expected).String()
		a = reflect.ValueOf(actual).String()
	case reflect.TypeOf(time.Time{}):
		e = spew.Sdump(expected)
		a = spew.Sdump(actual)
	default:
		e = spew.Sdump(expected)
		a = spew.Sdump(actual)
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(e),
		B:        difflib.SplitLines(a),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})

	return "\n\nDiff:\n" + diff
}

func Equal[T any](t testing.TB, expected, actual T) {
	t.Helper()

	if a.ObjectsAreEqual(expected, actual) {
		return
	}

	diff := diff(expected, actual)
	a.Fail(t, fmt.Sprintf("Not equal: \n"+
		"expected: %s\n"+
		"actual  : %s%s", q.Q(expected), q.Q(actual), diff))
}

func Equalf[T any](t testing.TB, expected, actual T, format string, args ...any) {
	if a.ObjectsAreEqual(expected, actual) {
		return
	}

	diff := diff(expected, actual)
	a.Fail(t, fmt.Sprintf("Not equal:\n"+
		"expected: %q\n"+
		"actual  : %q%q", q.Q(expected), q.Q(actual), diff), append([]any{format}, args...))
}

func NotEqual[T any](t *testing.T, expected, actual T) {
	t.Helper()

	if !a.ObjectsAreEqual(expected, actual) {
		return
	}

	diff := diff(expected, actual)
	a.Fail(t, fmt.Sprintf("Equal: \n"+
		"expected: %s\n"+
		"actual  : %s%s", q.Q(expected), q.Q(actual), diff))
}

func Zero[T any](t *testing.T, actual T) {
	var zero T
	Equal(t, zero, actual)
}

func NotZero[T any](t *testing.T, actual T) {
	var zero T
	NotEqual(t, zero, actual)
}

func False(t *testing.T, actual bool) {
	Equal(t, false, actual)
}

func Falsef(t *testing.T, actual bool, format string, args ...any) {
	Equalf(t, false, actual, format, args...)
}

func True(t *testing.T, actual bool) {
	Equal(t, true, actual)
}

func Truef(t *testing.T, actual bool, format string, args ...any) {
	Equalf(t, true, actual, format, args...)
}

func NoError(t testing.TB, err error) {
	Equal(t, nil, err)
}

func Contains[T comparable](t *testing.T, slice []T, item T) {
	for _, v := range slice {
		if v == item {
			return
		}
	}

	a.Fail(t, fmt.Sprintf("Slice does not contain %s", spew.Sdump(item)))
}

func Substring(t *testing.T, text, needle string) {
	if strings.Contains(text, needle) {
		return
	}

	a.Fail(t, fmt.Sprintf("%s does not contain %s", spew.Sdump(text), spew.Sdump(needle)))
}

func Substringf(t *testing.T, text, needle string, format string, args ...any) {
	if strings.Contains(text, needle) {
		return
	}

	a.Fail(t, fmt.Sprintf("%s does not contain %s", spew.Sdump(text), spew.Sdump(needle)), append([]any{format}, args...))
}
