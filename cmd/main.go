package main

import (
	"fmt"
	"math/big"
	"regexp"
	"time"
	"unsafe"

	"github.com/rprtr258/assert/pp"
	"github.com/rprtr258/assert/q"
)

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
	deletedAt time.Time
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

var c = Circular{}

func init() {
	c.C = &c
}

var (
	tm = time.Date(2015, time.January, 2, 0, 0, 0, 0, time.UTC)

	bigInt, _      = new(big.Int).SetString("-908f8474ea971baf", 16)
	bigFloat, _, _ = big.ParseFloat("3.1415926535897932384626433832795028", 10, 10, big.ToZero)
)

func TestFormat() {
	processTestCases(pp.Default, []any{
		nil,
		[]int(nil),
		true,
		false,
		int(4),
		int8(8),
		int16(16),
		int32(32),
		int64(64),
		uint(4),
		uint8(8),
		uint16(16),
		uint32(32),
		uint64(64),
		uintptr(128),
		float32(2.23),
		float64(3.14),
		complex64(complex(3, -4)),
		complex128(complex(5, 6)),
		"string",
		[]string{},
		EmptyStruct{},
		[]*Piyo{nil, nil},
		&c,
		"日本\t語\x00",
		time.Date(2015, time.February, 14, 22, 15, 0, 0, time.UTC),
		LargeBuffer{},
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		[][]byte{{0, 1, 2}, {3, 4}, {255}},
		map[string]any{"foo": 10, "bar": map[int]int{20: 30}},
	})
}

func processTestCases(printer *pp.PrettyPrinter, objects []any) {
	for _, object := range objects {
		fmt.Println(printer.Sprint(object))
	}

	for _, object := range []any{
		Private{b: false, i: 1, u: 2, f: 2.22, c: complex(5, 6)},
		map[string]int{"hell": 23, "world": 34},
		map[string]map[string]string{"s1": {"v1": "m1", "va1": "me1"}, "si2": {"v2": "m2"}},
		Foo{Bar: 1, Hoge: "a", Hello: map[string]string{"hel": "world", "a": "b"}, HogeHoges: []HogeHoge{{Hell: "a", World: 1}, {Hell: "bbb", World: 100}}},
		[3]int{},
		[]string{"aaa", "bbb", "ccc"},
		make(chan bool, 10),
		func(a string, b float32) int { return 0 },
		&HogeHoge{},
		&Piyo{Field1: map[string]string{"a": "b", "cc": "dd"}, F2: &Foo{}, Fie3: 128},
		[]any{1, 3},
		any(1),
		HogeHoge{A: "test"},
		FooPri{Public: "hello", private: "world"},
		new(regexp.Regexp),
		unsafe.Pointer(new(regexp.Regexp)),
		"日本\t語\n\000\U00101234a",
		bigInt,
		bigFloat,
		&tm,
		&User{Name: "k0kubun", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), deletedAt: time.Now().UTC()},
	} {
		fmt.Println(printer.Sprint(object))
	}
}

func dump(
	int,
	string,
	float64,
	bool,
	[]int,
	[]byte,
	[]int,
) string {
	return fmt.Sprint(q.Q("main", "dump"))
}

func main() {
	e := []int{1, 2, 3}

	fmt.Println(dump(
		123,
		"hello world",
		3.1415926,
		func(n int) bool { return n > 0 }(1),
		[]int{1, 2, 3},
		[]byte("goodbye world"),
		e[1:],
	))
	TestFormat()
}
