package pp

import (
	"testing"

	"github.com/rprtr258/scuf"
)

type colorTest struct {
	input  string
	mods   []scuf.Modifier
	result string
}

func TestColorize(t *testing.T) {
	for _, test := range []colorTest{
		{
			"blue on red",
			[]scuf.Modifier{scuf.FgBlue, scuf.BgRed},
			"\x1b[34;41mblue on red\x1b[0m",
		},
		{
			"magenta on white",
			[]scuf.Modifier{scuf.FgMagenta, scuf.BgWhite},
			"\x1b[35;47mmagenta on white\x1b[0m",
		},
		{
			"cyan",
			[]scuf.Modifier{scuf.FgCyan},
			"\x1b[36mcyan\x1b[0m",
		},
		{
			"default on red",
			[]scuf.Modifier{scuf.BgRed},
			"\x1b[41mdefault on red\x1b[0m",
		},
		{
			"default bold on yellow",
			[]scuf.Modifier{scuf.ModBold, scuf.BgYellow},
			"\x1b[1;43mdefault bold on yellow\x1b[0m",
		},
		{
			"bold",
			[]scuf.Modifier{scuf.ModBold},
			"\x1b[1mbold\x1b[0m",
		},
		{
			"no color at all",
			[]scuf.Modifier{},
			"no color at all",
		},
	} {
		t.Run(test.input, func(t *testing.T) {
			if output := scuf.String(test.input, test.mods...); output != test.result {
				t.Errorf("Expected %q, got %q", test.result, output)
			}
		})
	}
}
