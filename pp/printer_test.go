package pp

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"
	"time"

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

	bigInt, _      = (&big.Int{}).SetString("-908f8474ea971baf", 16)
	bigFloat, _, _ = big.ParseFloat("3.1415926535897932384626433832795028", 10, 10, big.ToZero)
	_MSK, _        = time.LoadLocation("Europe/Moscow")
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
		{Private{b: false, i: 1, u: 2, f: 2.22, c: complex(5, 6)}, "pp.\x1b[32mPrivate\x1b[0m{\n    \x1b[33mb\x1b[0m: \x1b[36;1mfalse\x1b[0m,\n    \x1b[33mi\x1b[0m: \x1b[34;1m1\x1b[0m,\n    \x1b[33mu\x1b[0m: \x1b[34;1m2\x1b[0m,\n    \x1b[33mf\x1b[0m: \x1b[35;1m2.220000\x1b[0m,\n    \x1b[33mc\x1b[0m: \x1b[34;1m(5+6i)\x1b[0m,\n}"},
		{map[string]int{"hell": 23, "world": 34}, "\x1b[32mmap[string]int\x1b[0m{\n    \x1b[31;1m\"\x1b[0m\x1b[31mhell\x1b[0m\x1b[31;1m\"\x1b[0m:  \x1b[34;1m23\x1b[0m,\n    \x1b[31;1m\"\x1b[0m\x1b[31mworld\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[34;1m34\x1b[0m,\n}"},
		{map[string]map[string]string{"s1": {"v1": "m1", "va1": "me1"}, "si2": {"v2": "m2"}}, "\x1b[32mmap[string]map[string]string\x1b[0m{\n    \x1b[31;1m\"\x1b[0m\x1b[31ms1\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[32mmap[string]string\x1b[0m{\n        \x1b[31;1m\"\x1b[0m\x1b[31mv1\x1b[0m\x1b[31;1m\"\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31mm1\x1b[0m\x1b[31;1m\"\x1b[0m,\n        \x1b[31;1m\"\x1b[0m\x1b[31mva1\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[31;1m\"\x1b[0m\x1b[31mme1\x1b[0m\x1b[31;1m\"\x1b[0m,\n    },\n    \x1b[31;1m\"\x1b[0m\x1b[31msi2\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[32mmap[string]string\x1b[0m{\n        \x1b[31;1m\"\x1b[0m\x1b[31mv2\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[31;1m\"\x1b[0m\x1b[31mm2\x1b[0m\x1b[31;1m\"\x1b[0m,\n    },\n}"},
		{Foo{Bar: 1, Hoge: "a", Hello: map[string]string{"hel": "world", "a": "b"}, HogeHoges: []HogeHoge{{Hell: "a", World: 1}, {Hell: "bbb", World: 100}}}, "pp.\x1b[32mFoo\x1b[0m{\n    \x1b[33mBar\x1b[0m:   \x1b[34;1m1\x1b[0m,\n    \x1b[33mHoge\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31ma\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mHello\x1b[0m: \x1b[32mmap[string]string\x1b[0m{\n        \x1b[31;1m\"\x1b[0m\x1b[31ma\x1b[0m\x1b[31;1m\"\x1b[0m:   \x1b[31;1m\"\x1b[0m\x1b[31mb\x1b[0m\x1b[31;1m\"\x1b[0m,\n        \x1b[31;1m\"\x1b[0m\x1b[31mhel\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[31;1m\"\x1b[0m\x1b[31mworld\x1b[0m\x1b[31;1m\"\x1b[0m,\n    },\n    \x1b[33mHogeHoges\x1b[0m: []pp.\x1b[32mHogeHoge\x1b[0m{\n        pp.\x1b[32mHogeHoge\x1b[0m{\n            \x1b[33mHell\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31ma\x1b[0m\x1b[31;1m\"\x1b[0m,\n            \x1b[33mWorld\x1b[0m: \x1b[34;1m1\x1b[0m,\n            \x1b[33mA\x1b[0m:     \x1b[36;1mnil\x1b[0m,\n        },\n        pp.\x1b[32mHogeHoge\x1b[0m{\n            \x1b[33mHell\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31mbbb\x1b[0m\x1b[31;1m\"\x1b[0m,\n            \x1b[33mWorld\x1b[0m: \x1b[34;1m100\x1b[0m,\n            \x1b[33mA\x1b[0m:     \x1b[36;1mnil\x1b[0m,\n        },\n    },\n}"},
		{[3]int{}, "[\x1b[34m3\x1b[0m]\x1b[32mint\x1b[0m{\n    \x1b[34;1m0\x1b[0m,\n    \x1b[34;1m0\x1b[0m,\n    \x1b[34;1m0\x1b[0m,\n}"},
		{[]string{"aaa", "bbb", "ccc"}, "[]\x1b[32mstring\x1b[0m{\n    \x1b[31;1m\"\x1b[0m\x1b[31maaa\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[31;1m\"\x1b[0m\x1b[31mbbb\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[31;1m\"\x1b[0m\x1b[31mccc\x1b[0m\x1b[31;1m\"\x1b[0m,\n}"},
		{func(a string, b float32) int { return 0 }, "\x1b[32mfunc(string, float32) int\x1b[0m {...}"},
		{&HogeHoge{}, "&pp.\x1b[32mHogeHoge\x1b[0m{\n    \x1b[33mHell\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mWorld\x1b[0m: \x1b[34;1m0\x1b[0m,\n    \x1b[33mA\x1b[0m:     \x1b[36;1mnil\x1b[0m,\n}"},
		{&Piyo{Field1: map[string]string{"a": "b", "cc": "dd"}, F2: &Foo{}, Fie3: 128}, "&pp.\x1b[32mPiyo\x1b[0m{\n    \x1b[33mField1\x1b[0m: \x1b[32mmap[string]string\x1b[0m{\n        \x1b[31;1m\"\x1b[0m\x1b[31ma\x1b[0m\x1b[31;1m\"\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31mb\x1b[0m\x1b[31;1m\"\x1b[0m,\n        \x1b[31;1m\"\x1b[0m\x1b[31mcc\x1b[0m\x1b[31;1m\"\x1b[0m: \x1b[31;1m\"\x1b[0m\x1b[31mdd\x1b[0m\x1b[31;1m\"\x1b[0m,\n    },\n    \x1b[33mF2\x1b[0m: &pp.\x1b[32mFoo\x1b[0m{\n        \x1b[33mBar\x1b[0m:       \x1b[34;1m0\x1b[0m,\n        \x1b[33mHoge\x1b[0m:      \x1b[31;1m\"\x1b[0m\x1b[31;1m\"\x1b[0m,\n        \x1b[33mHello\x1b[0m:     \x1b[32mmap[string]string\x1b[0m{},\n        \x1b[33mHogeHoges\x1b[0m: []pp.\x1b[32mHogeHoge\x1b[0m(\x1b[36;1mnil\x1b[0m),\n    },\n    \x1b[33mFie3\x1b[0m: \x1b[34;1m128\x1b[0m,\n}"},
		{[]any{1, 3}, "[]\x1b[32minterface {}\x1b[0m{\n    \x1b[34;1m1\x1b[0m,\n    \x1b[34;1m3\x1b[0m,\n}"},
		{any(1), "\x1b[34;1m1\x1b[0m"},
		{HogeHoge{A: "test"}, "pp.\x1b[32mHogeHoge\x1b[0m{\n    \x1b[33mHell\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mWorld\x1b[0m: \x1b[34;1m0\x1b[0m,\n    \x1b[33mA\x1b[0m:     \x1b[31;1m\"\x1b[0m\x1b[31mtest\x1b[0m\x1b[31;1m\"\x1b[0m,\n}"},
		{FooPri{Public: "hello", private: "world"}, "pp.\x1b[32mFooPri\x1b[0m{\n    \x1b[33mPublic\x1b[0m:  \x1b[31;1m\"\x1b[0m\x1b[31mhello\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mprivate\x1b[0m: \x1b[31;1m\"\x1b[0m\x1b[31mworld\x1b[0m\x1b[31;1m\"\x1b[0m,\n}"},
		{&regexp.Regexp{}, "&regexp.\x1b[32mRegexp\x1b[0m{\n    \x1b[33mexpr\x1b[0m:           \x1b[31;1m\"\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mprog\x1b[0m:           (*syntax.\x1b[32mProg\x1b[0m)(\x1b[36;1mnil\x1b[0m),\n    \x1b[33monepass\x1b[0m:        (*regexp.\x1b[32monePassProg\x1b[0m)(\x1b[36;1mnil\x1b[0m),\n    \x1b[33mnumSubexp\x1b[0m:      \x1b[34;1m0\x1b[0m,\n    \x1b[33mmaxBitStateLen\x1b[0m: \x1b[34;1m0\x1b[0m,\n    \x1b[33msubexpNames\x1b[0m:    []\x1b[32mstring\x1b[0m(\x1b[36;1mnil\x1b[0m),\n    \x1b[33mprefix\x1b[0m:         \x1b[31;1m\"\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mprefixBytes\x1b[0m:    []\x1b[32muint8\x1b[0m(\x1b[36;1mnil\x1b[0m),\n    \x1b[33mprefixRune\x1b[0m:     \x1b[34;1m0\x1b[0m,\n    \x1b[33mprefixEnd\x1b[0m:      \x1b[34;1m0\x1b[0m,\n    \x1b[33mmpool\x1b[0m:          \x1b[34;1m0\x1b[0m,\n    \x1b[33mmatchcap\x1b[0m:       \x1b[34;1m0\x1b[0m,\n    \x1b[33mprefixComplete\x1b[0m: \x1b[36;1mfalse\x1b[0m,\n    \x1b[33mcond\x1b[0m:           \x1b[34;1m0\x1b[0m,\n    \x1b[33mminInputLen\x1b[0m:    \x1b[34;1m0\x1b[0m,\n    \x1b[33mlongest\x1b[0m:        \x1b[36;1mfalse\x1b[0m,\n}"},
		{"日本\t語\n\000\U00101234a", "\x1b[31;1m\"\x1b[0m\x1b[31m日本\x1b[0m\x1b[35;1m\\t\x1b[0m\x1b[31m語\x1b[0m\x1b[35;1m\\n\x1b[0m\x1b[35;1m\\x00\x1b[0m\x1b[35;1m\\U00101234\x1b[0m\x1b[31ma\x1b[0m\x1b[31;1m\"\x1b[0m"},
		{bigInt, "&\x1b[34;1m-10416690100818090927\x1b[0m"},
		{bigFloat, "&\x1b[35;1m3.140625\x1b[0m"},
		{&tm, "&\x1b[34;1m2015\x1b[0m-\x1b[34;1m01\x1b[0m-\x1b[34;1m02\x1b[0m \x1b[34;1m00\x1b[0m:\x1b[34;1m00\x1b[0m:\x1b[34;1m00\x1b[0m \x1b[34;1mUTC\x1b[0m"},
		{&User{Name: "k0kubun", CreatedAt: time.Date(2024, 04, 13, 9, 36, 49, 0, time.UTC), UpdatedAt: time.Date(2024, 04, 13, 9, 36, 49, 0, _MSK)}, "&pp.\x1b[32mUser\x1b[0m{\n    \x1b[33mName\x1b[0m:      \x1b[31;1m\"\x1b[0m\x1b[31mk0kubun\x1b[0m\x1b[31;1m\"\x1b[0m,\n    \x1b[33mCreatedAt\x1b[0m: \x1b[34;1m2024\x1b[0m-\x1b[34;1m04\x1b[0m-\x1b[34;1m13\x1b[0m \x1b[34;1m09\x1b[0m:\x1b[34;1m36\x1b[0m:\x1b[34;1m49\x1b[0m \x1b[34;1mUTC\x1b[0m,\n    \x1b[33mUpdatedAt\x1b[0m: \x1b[34;1m2024\x1b[0m-\x1b[34;1m04\x1b[0m-\x1b[34;1m13\x1b[0m \x1b[34;1m09\x1b[0m:\x1b[34;1m36\x1b[0m:\x1b[34;1m49\x1b[0m \x1b[34;1mEurope/Moscow\x1b[0m,\n}"},
		// {make(chan bool, 10), "(\x1b[32mchan bool\x1b[0m)(\x1b[34;1m0xc000068620\x1b[0m)"}, // TODO: flaky, depends on allocated address
		// {unsafe.Pointer(&regexp.Regexp{}), "unsafe.\x1b[32mPointer\x1b[0m(\x1b[34;1m0xc000108780\x1b[0m)"}, // TODO: flaky, depends on allocated address
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
}
