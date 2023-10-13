package pp

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/rprtr258/fun"
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
		{time.Date(2015, time.February, 14, 22, 15, 0, 0, time.UTC), scuf.String(`2015`, scuf.FgBlue, scuf.ModBold) + `-` + scuf.String(`02`, scuf.FgBlue, scuf.ModBold) + `-` + scuf.String(`14`, scuf.FgBlue, scuf.ModBold) + ` ` + scuf.String(`22`, scuf.FgBlue, scuf.ModBold) + `:` + scuf.String(`15`, scuf.FgBlue, scuf.ModBold) + `:` + scuf.String(`00`, scuf.FgBlue, scuf.ModBold) + ` ` + scuf.String(`UTC`, scuf.FgBlue, scuf.ModBold)},
		{LargeBuffer{}, `pp.` + `[green]LargeBuffer[reset]` + `{
    ` + `[yellow]Buf[reset]` + `: [` + `[blue]1025[reset]` + `]` + `[green]uint8[reset]` + `{...},
}`},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[]` + `[green]uint8[reset]` + `{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`},
		{[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `
			[]` + `[green]uint16[reset]` + `{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`},
		{[]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, `[]` + `[green]uint32[reset]` + `{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `,
			}`},
		{[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, `[]` + `[green]uint64[reset]` + `{
			    ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`5`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`6`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`7`, scuf.FgBlue, scuf.ModBold) + `,
			    ` + scuf.String(`8`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`9`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `,
			}`},
		{[][]byte{{0, 1, 2}, {3, 4}, {255}}, `[]` + `[green][]uint8[reset]` + `{
			    []` + `[green]uint8[reset]` + `{
			        ` + scuf.String(`0`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`1`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`2`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    []` + `[green]uint8[reset]` + `{
			        ` + scuf.String(`3`, scuf.FgBlue, scuf.ModBold) + `, ` + scuf.String(`4`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    []` + `[green]uint8[reset]` + `{
			        ` + scuf.String(`255`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			}`},
		{map[string]any{"foo": 10, "bar": map[int]int{20: 30}}, `[green]map[string]interface {}[reset]` + `{
			    ` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `[red]bar[reset]` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `: ` + `[green]map[int]int[reset]` + `{
			        ` + scuf.String(`20`, scuf.FgBlue, scuf.ModBold) + `: ` + scuf.String(`30`, scuf.FgBlue, scuf.ModBold) + `,
			    },
			    ` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `[red]foo[reset]` + scuf.String(`"`, scuf.FgRed, scuf.ModBold) + `: ` + scuf.String(`10`, scuf.FgBlue, scuf.ModBold) + `,
			}`},
	})
}

func TestThousands(t *testing.T) {
	thousandsPrinter := newPrettyPrinter(3)
	thousandsPrinter.ThousandsSeparator = true
	thousandsPrinter.DecimalUint = true

	processTestCases(t, thousandsPrinter, []testCase{
		{int(4), scuf.String("4", scuf.FgBlue, scuf.ModBold)},
		{int(4000), scuf.String("4,000", scuf.FgBlue, scuf.ModBold)},
		{uint(1000), scuf.String("1,000", scuf.FgBlue, scuf.ModBold)},
		{uint16(16000), scuf.String("16,000", scuf.FgBlue, scuf.ModBold)},
		{uint32(32000), scuf.String("32,000", scuf.FgBlue, scuf.ModBold)},
		{uint64(64000), scuf.String("64,000", scuf.FgBlue, scuf.ModBold)},
		{float64(3000.14), scuf.String("3,000.140000", scuf.FgMagenta, scuf.ModBold)},
	})
}

func processTestCases(t *testing.T, printer *PrettyPrinter, cases []testCase) {
	t.Helper()

	for _, test := range cases {
		actual := printer.format(test.object)

		trimmed := strings.Trim(strings.Replace(test.expect, "\t", "", -1), "\n")
		expect := colorString(trimmed)
		if expect != actual {
			t.Errorf(`
TestCase: %#v
Type: %s
Expect: %# v
Actual: %# v
`,
				test.object,
				reflect.ValueOf(test.object).Kind(),
				expect,
				actual,
			)
			return
		}
		logResult(t, test.object, actual)
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
		logResult(t, object, printer.format(object))
	}
}

func logResult(t *testing.T, object any, actual string) {
	format := fun.IF(strings.Contains(actual, "\n"), "%#v =>\n%s\n", "%#v => %s\n")
	t.Logf(format, object, actual)
}

func colorString(text string) string {
	b := &bytes.Buffer{}
	colored := false

	lastMatch := []int{0, 0}
	for _, match := range colorRe.FindAllStringIndex(text, -1) {
		b.WriteString(text[lastMatch[1]:match[0]])
		lastMatch = match

		var colorText string
		color := text[lastMatch[0]+1 : lastMatch[1]-1]
		if code, ok := colors[color]; ok {
			colored = (color != "reset")
			colorText = fmt.Sprintf("\033[%sm", code)
		} else {
			colorText = text[lastMatch[0]:lastMatch[1]]
		}
		b.WriteString(colorText)
	}
	b.WriteString(text[lastMatch[1]:])

	if colored {
		b.WriteString("\033[0m")
	}
	return b.String()
}

var (
	colorRe = regexp.MustCompile(`(?i)\[[a-z0-9_-]+\]`)
	colors  = map[string]string{
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"bold":    "1",
		"reset":   "0",
	}
)
