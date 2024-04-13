package assert

import (
	"fmt"
	"reflect"

	"github.com/rprtr258/assert/internal/fun"
)

type diffLine struct {
	selector         string
	comment          string
	expected, actual any
}

func diffImpl(selectorPrefix string, expected, actual any) fun.Seq[diffLine] {
	etype, atype := reflect.TypeOf(expected), reflect.TypeOf(actual)
	switch {
	case etype == nil && atype == nil:
		return func(func(diffLine) bool) {}
	case etype == nil:
		return fun.FromMany(diffLine{
			selector: selectorPrefix,
			comment:  "expected nil, actual is not nil",
			expected: etype,
			actual:   atype,
		})
	case atype == nil:
		return fun.FromMany(diffLine{
			selector: selectorPrefix,
			comment:  "expected not nil, actual is nil",
			expected: etype,
			actual:   atype,
		})
	case etype != atype:
		return fun.FromMany(diffLine{
			selector: selectorPrefix,
			comment:  "different types",
			expected: etype.String(),
			actual:   atype.String(),
		})
	default:
		eval, aval := reflect.ValueOf(expected), reflect.ValueOf(actual)

		switch etype.Kind() {
		case reflect.Invalid:
			return func(f func(diffLine) bool) {}
		case reflect.Bool:
			if e, a := expected.(bool), actual.(bool); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if e, a := eval.Int(), aval.Int(); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			if e, a := eval.Uint(), aval.Uint(); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
		case reflect.Float32, reflect.Float64:
			if e, a := eval.Float(), aval.Float(); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
		case reflect.Complex64, reflect.Complex128:
			if e, a := eval.Complex(), aval.Complex(); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
		case reflect.String:
			if e, a := eval.String(), aval.String(); e != a {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "",
					expected: e,
					actual:   a,
				})
			}

			return func(f func(diffLine) bool) {}
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
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  fmt.Sprintf("len: %d != %d", lenExpected, lenActual),
					expected: expected,
					actual:   actual,
				})
			}

			// check if only one is nil
			if lenExpected == 0 {
				if (expected == nil) != (actual == nil) {
					return fun.FromMany(diffLine{
						selector: selectorPrefix,
						comment:  "one slice is nil, other is not",
						expected: expected,
						actual:   actual,
					})
				}

				return func(f func(diffLine) bool) {}
			}

			return fun.FlatMap(
				fun.FromRange(0, lenExpected),
				func(i int) fun.Seq[diffLine] {
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
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  fmt.Sprintf("len: %d != %d", lenExpected, lenActual),
					expected: expected,
					actual:   actual,
				})
			}

			return fun.FlatMap(
				fun.FromRange(0, lenExpected),
				func(i int) fun.Seq[diffLine] {
					return diffImpl(
						fmt.Sprintf("%s[%d]", selectorPrefix, i),
						eval.Index(i).Interface(),
						aval.Index(i).Interface(),
					)
				})
		case reflect.Struct:
			fields := etype.NumField()
			return fun.FlatMap(
				fun.FromRange(0, fields),
				// Filter(func(i int) bool {
				// 	return compareExported && typ.Field(i).IsExported()
				// }),
				func(i int) fun.Seq[diffLine] {
					ee := valueToInterface(eval.Field(i))
					aa := valueToInterface(aval.Field(i))
					if eval.Field(i).Comparable() && ee == aa {
						return func(f func(diffLine) bool) {}
					}

					return diffImpl(
						fmt.Sprintf("%s.%s", selectorPrefix, etype.Field(i).Name),
						ee,
						aa,
					)
				})
		case reflect.Map:
			expectedKeys := map[any]struct{}{}
			for _, k := range eval.MapKeys() {
				expectedKeys[k.Interface()] = struct{}{}
			}

			actualKeys := map[any]struct{}{}
			for _, k := range aval.MapKeys() {
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

			return func(yield func(diffLine) bool) {
				fun.FlatMap(
					fun.FromDictKeys(commonKeys),
					func(k any) fun.Seq[diffLine] {
						return diffImpl(
							fmt.Sprintf("%s[%v]", selectorPrefix, k),
							eval.MapIndex(reflect.ValueOf(k)),
							aval.MapIndex(reflect.ValueOf(k)),
						)
					})(yield)
				fun.Map(
					fun.FromDictKeys(expectedOnlyKeys),
					func(k any) diffLine {
						return diffLine{
							selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
							comment:  "not found key in actual",
							expected: eval.MapIndex(reflect.ValueOf(k)).Interface(),
							actual:   nil,
						}
					})(yield)
				fun.Map(
					fun.FromDictKeys(actualOnlyKeys),
					func(k any) diffLine {
						return diffLine{
							selector: fmt.Sprintf("%s[%v]", selectorPrefix, k),
							comment:  "unexpected key in actual",
							expected: nil,
							actual:   aval.MapIndex(reflect.ValueOf(k)).Interface(),
						}
					})(yield)
			}
		case reflect.Interface:
			if expected == nil && actual == nil {
				return func(f func(diffLine) bool) {}
			}

			if expected == nil || actual == nil {
				return fun.FromMany(diffLine{
					selector: selectorPrefix,
					comment:  "one is nil, one is not",
					expected: expected,
					actual:   actual,
				})
			}

			return diffImpl(selectorPrefix, eval.Elem(), aval.Elem())
		}

		// TODO: remove and return "" when other types are supported
		panic(fmt.Sprintf("unsupported kind: %s, %#v", eval.Kind().String(), eval.Interface()))
	}
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it panics.
func diff[T any](expected, actual T) fun.Seq[diffLine] {
	return diffImpl("", expected, actual)
}
