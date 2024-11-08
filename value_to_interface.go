package assert

import (
	"reflect"
	_ "unsafe" // to use go:linkname
)

//go:linkname valueInterface reflect.valueInterface
func valueInterface(v reflect.Value, safe bool) any

// use instead of v.Interface
func valueToInterface(v reflect.Value) any {
	return valueInterface(v, false)
}
