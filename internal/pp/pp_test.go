package pp

import (
	"bytes"
	"io"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func TestDefaultOutput(t *testing.T) {
	testOutput := &bytes.Buffer{}
	init := Default.GetOutput()
	Default.SetOutput(testOutput)
	ass.Equal[io.Writer](t, testOutput, Default.GetOutput())
	ass.Equal(t, "", testOutput.String())
	Default.Print("abcde")
	ass.NotEqual(t, "", testOutput.String())
	ass.NotEqual(t, init, Default.GetOutput())
	Default.ResetOutput()
	ass.Equal(t, init, Default.GetOutput())
}

func TestColorScheme(t *testing.T) {
	Default.currentScheme = defaultScheme
	ass.NotEqual(t, 0, len(Default.currentScheme.FieldName))
}

func TestWithLineInfo(t *testing.T) {
	outputWithoutLineInfo := &bytes.Buffer{}
	Default.SetOutput(outputWithoutLineInfo)
	Default.Print("abcde")

	outputWithLineInfo := &bytes.Buffer{}
	Default.SetOutput(outputWithLineInfo)

	WithLineInfo = true

	Default.Print("abcde")

	Default.ResetOutput()

	ass.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())
}

func TestWithLineInfoBackwardsCompatible(t *testing.T) {
	// Test that the global accessible field `WithLineInfo` does not mutate other instances
	outputWithLineInfo := &bytes.Buffer{}
	Default.SetOutput(outputWithLineInfo)

	WithLineInfo = true

	Default.Print("abcde")

	outputWithoutLineInfo := &bytes.Buffer{}
	pp := New()
	pp.SetOutput(outputWithoutLineInfo)
	pp.Print("abcde")

	ass.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())

	Default.ResetOutput()
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
			pp.SetOutput(output)

			pp.Print(test.foo)

			result := output.String()

			ass.SContainsNot(t, "IgnoreMe", result)
			ass.SContainsIs(t, test.omitIfEmptyIsPresent, "OmitIfEmpty", result)

			// field Full is renamed to full by the tag
			ass.SContainsIs(t, test.fullIsPresent, "full", result)
		})
	}
}
