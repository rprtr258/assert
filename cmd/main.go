package main

import (
	"fmt"
	"math/big"
	"regexp"
	"time"
	"unsafe"

	"github.com/rprtr258/scuf"

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

type testCase struct {
	object any
	expect string
}

func TestFormat() {
	processTestCases(pp.Default, []testCase{
		{nil, scuf.String("nil", scuf.FgCyan, scuf.ModBold)},
		{[]int(nil), scuf.NewString(func(b scuf.Buffer) {
			b.
				String("[]").
				String("int", scuf.FgGreen).
				InBytePair('(', ')', func(b scuf.Buffer) {
					b.String("nil", scuf.FgCyan, scuf.ModBold)
				})
		})},
		{true, scuf.String("true", scuf.FgCyan, scuf.ModBold)},
		{false, scuf.String("false", scuf.FgCyan, scuf.ModBold)},
		{int(4), scuf.String("4", scuf.FgBlue, scuf.ModBold)},
		{int8(8), scuf.String("8", scuf.FgBlue, scuf.ModBold)},
		{int16(16), scuf.String("16", scuf.FgBlue, scuf.ModBold)},
		{int32(32), scuf.String("32", scuf.FgBlue, scuf.ModBold)},
		{int64(64), scuf.String("64", scuf.FgBlue, scuf.ModBold)},
		{uint(4), scuf.String("4", scuf.FgBlue, scuf.ModBold)},
		{uint8(8), scuf.String("8", scuf.FgBlue, scuf.ModBold)},
		{uint16(16), scuf.String("16", scuf.FgBlue, scuf.ModBold)},
		{uint32(32), scuf.String("32", scuf.FgBlue, scuf.ModBold)},
		{uint64(64), scuf.String("64", scuf.FgBlue, scuf.ModBold)},
		{uintptr(128), scuf.String("128", scuf.FgBlue, scuf.ModBold)},
		{float32(2.23), scuf.String("2.230000", scuf.FgMagenta, scuf.ModBold)},
		{float64(3.14), scuf.String("3.140000", scuf.FgMagenta, scuf.ModBold)},
		{complex64(complex(3, -4)), scuf.String("(3-4i)", scuf.FgBlue, scuf.ModBold)},
		{complex128(complex(5, 6)), scuf.String("(5+6i)", scuf.FgBlue, scuf.ModBold)},
		{"string", scuf.String(`"`, scuf.FgRed, scuf.ModBold) + scuf.String("string", scuf.FgRed) + scuf.String(`"`, scuf.FgRed, scuf.ModBold)},
		{[]string{}, "[]" + scuf.String("string", scuf.FgGreen) + "{}"},
		{EmptyStruct{}, "pp." + scuf.String("EmptyStruct", scuf.FgGreen) + "{}"},
		{
			[]*Piyo{nil, nil}, `[]*pp.` + scuf.String("Piyo", scuf.FgGreen) + `{
    (*pp.` + scuf.String("Piyo", scuf.FgGreen) + `)(` + scuf.String("nil", scuf.FgCyan, scuf.ModBold) + `),
    (*pp.` + scuf.String("Piyo", scuf.FgGreen) + `)(` + scuf.String("nil", scuf.FgCyan, scuf.ModBold) + `),
}`,
		},
		{
			&c, `&pp.` + scuf.String("Circular", scuf.FgGreen) + `{
    ` + scuf.String("C", scuf.FgYellow) + `: &pp.` + scuf.String("Circular", scuf.FgGreen) + `{...},
}`,
		},
		{
			"日本\t語\x00",
			scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `[red]日本[reset]` + scuf.String(`\t`, scuf.FgMagenta, scuf.ModBold) + `[red]語[reset]` + scuf.String(`\x00`, scuf.FgMagenta, scuf.ModBold) + scuf.String(`"`, scuf.FgRed, scuf.ModBold),
		},
		{
			time.Date(2015, time.February, 14, 22, 15, 0, 0, time.UTC),
			scuf.String(`2015`, scuf.FgBlue, scuf.ModBold) + `-` + scuf.String(`02`, scuf.FgBlue, scuf.ModBold) + `-` + scuf.String(`14`, scuf.FgBlue, scuf.ModBold) + ` ` + scuf.String(`22`, scuf.FgBlue, scuf.ModBold) + `:` + scuf.String(`15`, scuf.FgBlue, scuf.ModBold) + `:` + scuf.String(`00`, scuf.FgBlue, scuf.ModBold) + ` ` + scuf.String(`UTC`, scuf.FgBlue, scuf.ModBold),
		},
		{
			LargeBuffer{}, `pp.[green]LargeBuffer[reset]{
    [yellow]Buf[reset]: [[blue]1025[reset]][green]uint8[reset]{...},
}`,
		},
		{
			[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[][green]uint8[reset]{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`,
		},
		{
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[][green]uint16[reset]{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`,
		},
		{
			[]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[][green]uint32[reset]{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`,
		},
		{
			[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, `[][green]uint64[reset]{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `,
			}`,
		},
		{
			[][]byte{{0, 1, 2}, {3, 4}, {255}}, `[][green][]uint8[reset]{
			    [][green]uint8[reset]{
			        ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    [][green]uint8[reset]{
			        ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    [][green]uint8[reset]{
			        ` + scuf.String(`255`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			}`,
		},
		{
			map[string]any{"foo": 10, "bar": map[int]int{20: 30}}, `[green]map[string]interface {}[reset]{
			    ` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `[red]bar[reset]` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `: [green]map[int]int[reset]{
			        ` + scuf.String(`20`, scuf.FgBlue, scuf.ModBold) + `: ` + scuf.String(`30`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    ` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `[red]foo[reset]` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `: ` + scuf.String(`10`, scuf.FgBlue, scuf.ModBold) + `,
			}`,
		},
	})
}

func processTestCases(printer *pp.PrettyPrinter, cases []testCase) {
	for _, test := range cases {
		fmt.Println(printer.Sprint(test.object))
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
