package pp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOutput(t *testing.T) {
	testOutput := &bytes.Buffer{}
	init := GetDefaultOutput()
	SetDefaultOutput(testOutput)
	assert.Equal(t, testOutput, GetDefaultOutput())
	assert.Equal(t, "", testOutput.String())
	Print("abcde")
	assert.NotEqual(t, "", testOutput.String())
	assert.NotEqual(t, init, GetDefaultOutput())
	ResetDefaultOutput()
	assert.Equal(t, init, GetDefaultOutput())
}

func TestColorScheme(t *testing.T) {
	SetColorScheme(ColorScheme{})
	assert.NotEqual(t, 0, len(Default.currentScheme.FieldName))
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

	assert.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())
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

	assert.NotEqual(t, outputWithLineInfo.Bytes(), outputWithoutLineInfo.Bytes())

	ResetDefaultOutput()
}

func TestStructPrintingWithTags(t *testing.T) {
	type Foo struct {
		IgnoreMe     any    `pp:"-"`
		ChangeMyName string `pp:"NewName"`
		OmitIfEmpty  string `pp:",omitempty"`
		Full         string `pp:"full,omitempty"`
	}

	for _, tc := range []struct {
		name               string
		foo                Foo
		omitIfEmptyOmitted bool
		fullOmitted        bool
	}{
		{
			name: "all set",
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "i'm not empty",
				Full:         "hello",
			},
			omitIfEmptyOmitted: false,
			fullOmitted:        false,
		},
		{
			name: "omit if empty not set",
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "",
				Full:         "hello",
			},
			omitIfEmptyOmitted: true,
			fullOmitted:        false,
		},
		{
			name: "both omitted",
			foo: Foo{
				IgnoreMe:     "i'm a secret",
				ChangeMyName: "i'm an alias",
				OmitIfEmpty:  "",
				Full:         "",
			},
			omitIfEmptyOmitted: true,
			fullOmitted:        true,
		},
		{
			name:               "zero",
			foo:                Foo{},
			omitIfEmptyOmitted: true,
			fullOmitted:        true,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			pp := New()
			pp.SetOutput(output)

			pp.Print(tc.foo)

			result := output.String()

			assert.NotContains(t, result, "IgnoreMe")
			assert.True(t, strings.Contains(result, "OmitIfEmpty") != tc.omitIfEmptyOmitted)

			// field Full is renamed to full by the tag
			assert.True(t, strings.Contains(result, "full") != tc.fullOmitted)
		})
	}

}
