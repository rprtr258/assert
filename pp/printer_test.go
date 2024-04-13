package pp

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"
	"time"
	"unsafe"

	"github.com/rprtr258/scuf"
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

var c Circular = Circular{}

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

func color_bool(s string) string {
	return scuf.String(s, scuf.FgCyan, scuf.ModBold)
}

func color_number(s string) string {
	return scuf.String(s, scuf.FgBlue, scuf.ModBold)
}

func color_float(s string) string {
	return scuf.String(s, scuf.FgMagenta, scuf.ModBold)
}

func TestFormat(t *testing.T) {
	processTestCases(t, Default, []testCase{
		{nil, scuf.String("nil", scuf.FgCyan, scuf.ModBold)},
		{[]int(nil), scuf.NewString(func(b scuf.Buffer) {
			b.
				String("[]").
				String("int", scuf.FgGreen).
				InBytePair('(', ')', func(b scuf.Buffer) {
					b.String("nil", scuf.FgCyan, scuf.ModBold)
				})
		})},
		{true, color_bool("true")},
		{false, color_bool("false")},
		{int(4), color_number("4")},
		{int8(8), color_number("8")},
		{int16(16), color_number("16")},
		{int32(32), color_number("32")},
		{int64(64), color_number("64")},
		{uint(4), color_number("4")},
		{uint8(8), color_number("8")},
		{uint16(16), color_number("16")},
		{uint32(32), color_number("32")},
		{uint64(64), color_number("64")},
		{uintptr(128), color_number("128")},
		{float32(2.23), color_float("2.230000")},
		{float64(3.14), color_float("3.140000")},
		{complex64(complex(3, -4)), color_number("(3-4i)")},
		{complex128(complex(5, 6)), color_number("(5+6i)")},
		{"string", scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`"`, scuf.FgRed, scuf.ModBold).
				String("string", scuf.FgRed).
				String(`"`, scuf.FgRed, scuf.ModBold)
		})},
		{[]string{}, scuf.NewString(func(b scuf.Buffer) {
			b.
				String("[]").
				String("string", scuf.FgGreen).
				String("{}")
		})},
		{EmptyStruct{}, scuf.NewString(func(b scuf.Buffer) {
			b.
				String("pp.").
				String("EmptyStruct", scuf.FgGreen).
				String("{}")
		})},
		{[]*Piyo{nil, nil}, scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`[]*pp.`).
				String("Piyo", scuf.FgGreen).
				InBytePair('{', '}', func(b scuf.Buffer) {
					b.
						NL().
						String(`    (*pp.`).String("Piyo", scuf.FgGreen).String(`)(`).String("nil", scuf.FgCyan, scuf.ModBold).String(`),`).NL().
						String(`    (*pp.`).String("Piyo", scuf.FgGreen).String(`)(`).String("nil", scuf.FgCyan, scuf.ModBold).String(`),`).NL()
				})
		})},
		{&c, scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`&pp.`).
				String("Circular", scuf.FgGreen).
				InBytePair('{', '}', func(b scuf.Buffer) {
					b.NL().
						String(`    `).
						String("C", scuf.FgYellow).
						String(`: &pp.`).
						String("Circular", scuf.FgGreen).
						String(`{...},`).NL()
				})
		})},
		{"日本\t語\x00", scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`"`, scuf.FgRed, scuf.ModBold).
				String(`日本`, scuf.FgRed).
				String(`\t`, scuf.FgMagenta, scuf.ModBold).
				String(`語`, scuf.FgRed).
				String(`\x00`, scuf.FgMagenta, scuf.ModBold).
				String(`"`, scuf.FgRed, scuf.ModBold)
		})},
		{time.Date(2015, time.February, 14, 22, 15, 0, 0, time.UTC), scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`2015`, scuf.FgBlue, scuf.ModBold).
				String(`-`).
				String(`02`, scuf.FgBlue, scuf.ModBold).
				String(`-`).
				String(`14`, scuf.FgBlue, scuf.ModBold).
				SPC().
				String(`22`, scuf.FgBlue, scuf.ModBold).
				String(`:`).
				String(`15`, scuf.FgBlue, scuf.ModBold).
				String(`:`).
				String(`00`, scuf.FgBlue, scuf.ModBold).
				String(` `).
				String(`UTC`, scuf.FgBlue, scuf.ModBold)
		})},
		{LargeBuffer{}, scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`pp.`).String(`LargeBuffer`, scuf.FgGreen).InBytePair('{', '}', func(b scuf.Buffer) {
				b.
					NL().
					String(`    `).
					String(`Buf`, scuf.FgYellow).
					String(`: [`).
					String(`1025`, scuf.FgBlue).
					String(`]`).
					String(`uint8`, scuf.FgGreen).
					String(`{...},`).
					NL()
			})
		})},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, scuf.NewString(func(b scuf.Buffer) {
			b.
				String(`[]`).
				String(`uint8`, scuf.FgGreen).
				InBytePair('{', '}', func(b scuf.Buffer) {
					b.NL().
						String(`    `).
						String(color_number(`0`) + `, `).
						String(color_number(`1`) + `, `).
						String(color_number(`2`) + `, `).
						String(color_number(`3`) + `, `).
						String(color_number(`4`) + `, `).
						String(color_number(`5`) + `, `).
						String(color_number(`6`) + `, `).
						String(color_number(`7`) + `, `).
						String(color_number(`8`) + `, `).
						String(color_number(`9`) + `, `).
						String(color_number(`0`) + `, `).
						String(color_number(`1`) + `, `).
						String(color_number(`2`) + `, `).
						String(color_number(`3`) + `, `).
						String(color_number(`4`) + `, `).
						String(color_number(`5`) + `,`).
						NL().
						String(`    `).
						String(color_number(`6`) + `, `).
						String(color_number(`7`) + `, `).
						String(color_number(`8`) + `, `).
						String(color_number(`9`) + `,`).
						NL()
				})
		})},
		{[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[]` +
			scuf.String(`uint16`, scuf.FgGreen) +
			`{
    ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `, ` +
			color_number(`3`) + `, ` +
			color_number(`4`) + `, ` +
			color_number(`5`) + `, ` +
			color_number(`6`) + `, ` +
			color_number(`7`) + `,` +
			`
    ` +
			color_number(`8`) + `, ` +
			color_number(`9`) + `, ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `, ` +
			color_number(`3`) + `, ` +
			color_number(`4`) + `, ` +
			color_number(`5`) + `,` +
			`
    ` +
			color_number(`6`) + `, ` +
			color_number(`7`) + `, ` +
			color_number(`8`) + `, ` +
			color_number(`9`) + `,` +
			`
}`},
		{[]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[]` + scuf.String(`uint32`, scuf.FgGreen) + `{
    ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `, ` +
			color_number(`3`) + `, ` +
			color_number(`4`) + `, ` +
			color_number(`5`) + `, ` +
			color_number(`6`) + `, ` +
			color_number(`7`) + `,
    ` +
			color_number(`8`) + `, ` +
			color_number(`9`) + `, ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `, ` +
			color_number(`3`) + `, ` +
			color_number(`4`) + `, ` +
			color_number(`5`) + `,
    ` +
			color_number(`6`) + `, ` +
			color_number(`7`) + `, ` +
			color_number(`8`) + `, ` +
			color_number(`9`) + `,
}`},
		{[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, `[]` + scuf.String(`uint64`, scuf.FgGreen) + `{
    ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `, ` +
			color_number(`3`) + `,
    ` +
			color_number(`4`) + `, ` +
			color_number(`5`) + `, ` +
			color_number(`6`) + `, ` +
			color_number(`7`) + `,
    ` +
			color_number(`8`) + `, ` +
			color_number(`9`) + `, ` +
			color_number(`0`) + `,
}`},
		{[][]byte{{0, 1, 2}, {3, 4}, {255}}, `[]` + scuf.String(`[]uint8`, scuf.FgGreen) + `{
    []` + scuf.String(`uint8`, scuf.FgGreen) + `{
        ` +
			color_number(`0`) + `, ` +
			color_number(`1`) + `, ` +
			color_number(`2`) + `,
    },
    []` + scuf.String(`uint8`, scuf.FgGreen) + `{
        ` +
			color_number(`3`) + `, ` +
			color_number(`4`) + `,
    },
    []` + scuf.String(`uint8`, scuf.FgGreen) + `{
        ` +
			color_number(`255`) +
			`,
    },
}`},
		{map[string]any{"foo": 10, "bar": map[int]int{20: 30}}, scuf.String(`map[string]interface {}`, scuf.FgGreen) + `{
    ` +
			scuf.String(`"`, scuf.FgRed, scuf.ModBold) +
			scuf.String(`bar`, scuf.FgRed) +
			scuf.String(`"`, scuf.FgRed, scuf.ModBold) +
			`: ` +
			scuf.String(`map[int]int`, scuf.FgGreen) +
			`{
        ` +
			scuf.String(`20`, scuf.FgBlue, scuf.ModBold) +
			`: ` +
			scuf.String(`30`, scuf.FgBlue, scuf.ModBold) +
			`,
    },
    ` +
			scuf.String(`"`, scuf.FgRed, scuf.ModBold) +
			scuf.String(`foo`, scuf.FgRed) +
			scuf.String(`"`, scuf.FgRed, scuf.ModBold) +
			`: ` +
			scuf.String(`10`, scuf.FgBlue, scuf.ModBold) +
			`,
}`},
		{Private{b: false, i: 1, u: 2, f: 2.22, c: complex(5, 6)}, ""},
	})
}

func TestThousands(t *testing.T) {
	thousandsPrinter := newPrettyPrinter(3)
	thousandsPrinter.ThousandsSeparator = true
	thousandsPrinter.DecimalUint = true

	processTestCases(t, thousandsPrinter, []testCase{
		{int(4), color_number("4")},
		{int(4000), color_number("4,000")},
		{uint(1000), color_number("1,000")},
		{uint16(16000), color_number("16,000")},
		{uint32(32000), color_number("32,000")},
		{uint64(64000), color_number("64,000")},
		{float64(3000.14), color_float("3,000.140000")},
	})
}

func processTestCases(t *testing.T, printer *PrettyPrinter, tests []testCase) {
	t.Helper()

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%#v", test.object), func(t *testing.T) {
			t.Parallel()

			actual := printer.format(test.object)
			if test.expect != actual {
				t.Errorf(`
TestCase: %#[1]v
Type: %[2]s
Expect: %[3]q
      : %[3]s
Actual: %[4]q
      : %[4]s
`,
					test.object,
					reflect.ValueOf(test.object).Kind(),
					test.expect,
					actual,
				)
				return
			}
		})
	}

	for _, object := range []any{
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
		object := object
		t.Run(fmt.Sprintf("%#v", object), func(t *testing.T) {
			t.Parallel()

			logResult(t, object, printer.format(object))
		})
	}
}

func logResult(t *testing.T, object any, actual string) {
	t.Logf("%#v =>\n%s\n\n", object, actual)
}
