package pp

import (
	"bytes"
	"io"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func TestDefaultOutput(t *testing.T) {
	t.Parallel()

	testOutput := &bytes.Buffer{}
	pp := newPrettyPrinter(3)
	init := pp.out
	pp.out = testOutput
	ass.Equal[io.Writer](t, testOutput, pp.out)
	ass.Equal(t, "", testOutput.String())
	pp.Print("abcde")
	ass.NotEqual(t, "", testOutput.String())
	ass.NotEqual(t, init, pp.out)
}

func TestColorScheme(t *testing.T) {
	t.Parallel()

	ass.NotEqual(t, 0, len(scheme.FieldName))
}

func TestStructPrintingWithTags(t *testing.T) {
	t.Parallel()

	type Foo struct {
		IgnoreMe     any    `pp:"-"`
		ChangeMyName string `pp:"NewName"`
		OmitIfEmpty  string `pp:",omitempty"`
		Full         string `pp:"full,omitempty"`
	}

	for name, test := range map[string]struct {
		foo                  Foo
		omitIfEmptyIsPresent bool
		fullIsPresent        bool
	}{
		"all set": {
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "i'm not empty",
				Full:         "hello",
			},
			omitIfEmptyIsPresent: true,
			fullIsPresent:        true,
		},
		"omit if empty not set": {
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "",
				Full:         "hello",
			},
			omitIfEmptyIsPresent: false,
			fullIsPresent:        true,
		},
		"both omitted": {
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "",
				Full:         "",
			},
			omitIfEmptyIsPresent: false,
			fullIsPresent:        false,
		},
		"zero": {
			foo:                  Foo{},
			omitIfEmptyIsPresent: false,
			fullIsPresent:        false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output := &bytes.Buffer{}
			pp := New()
			pp.out = output

			pp.Print(test.foo)

			result := output.String()

			ass.SContainsNot(t, "IgnoreMe", result)
			ass.SContainsIs(t, test.omitIfEmptyIsPresent, "OmitIfEmpty", result)

			// field Full is renamed to full by the tag
			ass.SContainsIs(t, test.fullIsPresent, "full", result)
		})
	}
}
