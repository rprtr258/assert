package pp

import (
	"bytes"
	"io"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func TestDefaultOutput(t *testing.T) {
	testOutput := &bytes.Buffer{}
	init := GetDefaultOutput()
	SetDefaultOutput(testOutput)
	ass.Equal[io.Writer](t, testOutput, GetDefaultOutput())
	ass.Equal(t, "", testOutput.String())
	Print("abcde")
	ass.NotEqual(t, "", testOutput.String())
	ass.NotEqual(t, init, GetDefaultOutput())
	ResetDefaultOutput()
	ass.Equal(t, init, GetDefaultOutput())
}

func TestColorScheme(t *testing.T) {
	SetColorScheme(ColorScheme{})
	ass.NotEqual(t, 0, len(Default.currentScheme.FieldName))
}

func TestWithLineInfo(t *testing.T) {
	outputWithoutLineInfo := &bytes.Buffer{}
	SetDefaultOutput(outputWithoutLineInfo)
	Print("abcde")

	outputWithLineInfo := &bytes.Buffer{}
	SetDefaultOutput(outputWithLineInfo)
	WithLineInfo = true
	Print("abcde")

	ResetDefaultOutput()

	ass.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())
}

func TestWithLineInfoBackwardsCompatible(t *testing.T) {
	// Test that the global accessible field `WithLineInfo` does not mutate other instances

	outputWithLineInfo := new(bytes.Buffer)
	SetDefaultOutput(outputWithLineInfo)
	WithLineInfo = true
	Print("abcde")

	outputWithoutLineInfo := new(bytes.Buffer)
	pp := New()
	pp.SetOutput(outputWithoutLineInfo)
	pp.Print("abcde")

	ass.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())

	ResetDefaultOutput()
}

func TestStructPrintingWithTags(t *testing.T) {
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
		test := test
		t.Run(name, func(t *testing.T) {
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
