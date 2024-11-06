package pp

import (
	_ "embed"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"
	"time"

	"golang.org/x/tools/txtar"
)

//go:embed testdata.txtar
var _txtar []byte

var _expectedByObject = func() map[string]string {
	files := txtar.Parse(_txtar).Files

	tests := make(map[string]string, len(files))
	for _, file := range files {
		// check no object is not saved multiple times
		if _, ok := tests[file.Name]; ok {
			panic("duplicate file name: " + file.Name)
		}

		// remove trailing \n
		data := file.Data[:len(file.Data)-1]
		tests[file.Name] = string(data)
	}
	return tests
}()

type Foo struct {
	Bar       int
	Hoge      string
	Hello     map[string]string
	HogeHoges []HogeHoge
}

type FooPri struct {
	Public  string
	private string
}

type Piyo struct {
	Field1 map[string]string
	F2     *Foo
	Fie3   int
}

type HogeHoge struct {
	Hell  string
	World int
	A     any
}

type EmptyStruct struct{}

type User struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LargeBuffer struct {
	Buf [1025]byte
}

type Private struct {
	b bool
	i int
	u uint
	f float32
	c complex128
}

type Circular struct {
	C *Circular
}

var c = func() Circular {
	res := Circular{}
	res.C = &res
	return res
}()

var (
	tm = time.Date(2015, time.January, 2, 0, 0, 0, 0, time.UTC)

	bigInt = func() *big.Int {
		res, ok := (&big.Int{}).SetString("-908f8474ea971baf", 16)
		if !ok {
			panic("failed to set bigInt")
		}
		return res
	}()
	bigFloat = func() *big.Float {
		res, _, err := big.ParseFloat("3.1415926535897932384626433832795028", 10, 10, big.ToZero)
		if err != nil {
			panic(err.Error())
		}
		return res
	}()
	_MSK = func() *time.Location {
		res, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			panic(err.Error())
		}
		return res
	}()
)

func TestFormat(t *testing.T) {
	tests := []any{
		nil,
		[]int(nil),
		true,
		false,
		int(4),
		int(4000),
		int8(8),
		int16(16),
		int32(32),
		int64(64),
		uint(4),
		uint(1000),
		uint8(8),
		uint16(16),
		uint16(16000),
		uint32(32),
		uint32(32000),
		uint64(64),
		uint64(64000),
		uintptr(128),
		float32(2.23),
		float64(3.14),
		float64(3000.14),
		complex64(complex(3, -4)),
		complex128(complex(5, 6)),
		"string",
		[]string{},
		EmptyStruct{},
		[]*Piyo{nil, nil},
		"日本\t語\x00",
		time.Date(2015, time.February, 14, 22, 15, 0, 0, time.UTC),
		LargeBuffer{},
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		[][]byte{{0, 1, 2}, {3, 4}, {255}},
		map[string]any{"foo": 10, "bar": map[int]int{20: 30}},
		Private{b: false, i: 1, u: 2, f: 2.22, c: complex(5, 6)},
		map[string]int{"hell": 23, "world": 34},
		map[string]map[string]string{"s1": {"v1": "m1", "va1": "me1"}, "si2": {"v2": "m2"}},
		Foo{Bar: 1, Hoge: "a", Hello: map[string]string{"hel": "world", "a": "b"}, HogeHoges: []HogeHoge{{Hell: "a", World: 1}, {Hell: "bbb", World: 100}}},
		[3]int{},
		[]string{"aaa", "bbb", "ccc"},
		&HogeHoge{},
		[]any{1, 3},
		any(1),
		HogeHoge{A: "test"},
		FooPri{Public: "hello", private: "world"},
		&regexp.Regexp{},
		"日本\t語\n\000\U00101234a",
		bigInt,
		&tm,
		&User{Name: "k0kubun", CreatedAt: time.Date(2024, 4, 13, 9, 36, 49, 0, time.UTC), UpdatedAt: time.Date(2024, 4, 13, 9, 36, 49, 0, _MSK)},
		// TODO: flaky, depends on PRINTING HUGE BIG FLOAT AS JUST FUCKING 3.14
		// bigFloat,
		// TODO: flaky, depends on allocated address
		// func(a string, b float32) int { return 0 },
		// -- (func(string, float32) int)(0x549980) --
		// [32mfunc(string, float32) int[0m {...}
		// &Piyo{Field1: map[string]string{"a": "b", "cc": "dd"}, F2: &Foo{}, Fie3: 128},
		// -- &pp.Piyo{Field1:map[string]string{"a":"b", "cc":"dd"}, F2:(*pp.Foo)(0xc000028b80), Fie3:128} --
		// &pp.[32mPiyo[0m{
		//     [33mField1[0m: [32mmap[string]string[0m{
		//         [31;1m"[0m[31ma[0m[31;1m"[0m:  [31;1m"[0m[31mb[0m[31;1m"[0m,
		//         [31;1m"[0m[31mcc[0m[31;1m"[0m: [31;1m"[0m[31mdd[0m[31;1m"[0m,
		//     },
		//     [33mF2[0m: &pp.[32mFoo[0m{
		//         [33mBar[0m:       [34;1m0[0m,
		//         [33mHoge[0m:      [31;1m"[0m[31;1m"[0m,
		//         [33mHello[0m:     [32mmap[string]string[0m{},
		//         [33mHogeHoges[0m: []pp.[32mHogeHoge[0m([36;1mnil[0m),
		//     },
		//     [33mFie3[0m: [34;1m128[0m,
		// }
		// &c,
		// -- &pp.Circular{C:(*pp.Circular)(0x6ca188)} --
		// &pp.[32mCircular[0m{
		//     [33mC[0m: &pp.[32mCircular[0m{...},
		// }
		// make(chan bool, 10),
		// unsafe.Pointer(&regexp.Regexp{}),
	}
	keys := make(map[string]struct{}, len(tests))
	for _, object := range tests {
		name := fmt.Sprintf("%#v", object)
		keys[name] = struct{}{}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := Default.format(object)

			expected, ok := _expectedByObject[name]
			if !ok {
				t.Errorf(`no expected value found
Object: %s
Actual: %q
`, name, actual)
				t.FailNow()
			}

			if expected == actual {
				return
			}

			t.Errorf(`Type: %[1]s
Expect: %[2]q
Actual: %[3]q
Rendered Expect:
%[2]s
Rendered Actual:
%[3]s
`,
				reflect.ValueOf(object).Kind(),
				expected,
				actual,
			)
		})
	}

	for k := range _expectedByObject {
		if _, ok := keys[k]; !ok {
			t.Errorf("not tested record found: %s", k)
		}
	}
}
