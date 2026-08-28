package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// concatenate builds the content handed to the templating engine out of ordered
// (file, content) pairs, the same way config.New does.
func concatenate(t *testing.T, files [][2]string) ([]byte, []manifestFragment) {
	t.Helper()

	var content []byte
	var fragments []manifestFragment

	for _, file := range files {
		var fragment []byte
		fragments, fragment = appendFragment(fragments, file[0], []byte(file[1]))
		content = append(content, fragment...)
	}

	return content, fragments
}

func TestAppendFragment(t *testing.T) {
	content, fragments := concatenate(t, [][2]string{
		{"_first.yaml", "a\nb\n"},
		// No trailing newline: the fragment must be normalized, otherwise `c` and
		// `d` would end up glued together on the same line.
		{"_second.yaml", "c"},
		{"main.yaml", "d\ne\n"},
	})

	require.Equal(t, "a\nb\nc\nd\ne\n", string(content))
	require.Equal(t, []manifestFragment{
		{file: "_first.yaml", firstLine: 1, lineCount: 2},
		{file: "_second.yaml", firstLine: 3, lineCount: 1},
		{file: "main.yaml", firstLine: 4, lineCount: 2},
	}, fragments)
}

func TestAppendFragmentEmptyContent(t *testing.T) {
	// An empty file must not be given a phantom line, or every fragment after it
	// would be reported one line too far.
	_, fragments := concatenate(t, [][2]string{
		{"_empty.yaml", ""},
		{"main.yaml", "a\n"},
	})

	require.Equal(t, []manifestFragment{
		{file: "_empty.yaml", firstLine: 1, lineCount: 0},
		{file: "main.yaml", firstLine: 1, lineCount: 1},
	}, fragments)
}

func TestLocateTemplateError(t *testing.T) {
	_, fragments := concatenate(t, [][2]string{
		{"_first.yaml", "a\nb\n"},
		{"_second.yaml", "c\nd\ne\n"},
		{"main.yaml", "f\ng\n"},
	})

	testdata := []struct {
		name            string
		err             error
		templateName    string
		fragments       []manifestFragment
		expectedMessage string
		expectedFile    string
		expectedLine    int
	}{
		{
			name:            "error in the first partial",
			err:             errors.New(`template: main.yaml:1: bad character U+007B '{'`),
			expectedMessage: `_first.yaml:1: bad character U+007B '{'`,
			expectedFile:    "_first.yaml",
			expectedLine:    1,
		},
		{
			name:            "error in a later partial",
			err:             errors.New(`template: main.yaml:4: unexpected "{" in command`),
			expectedMessage: `_second.yaml:2: unexpected "{" in command`,
			expectedFile:    "_second.yaml",
			expectedLine:    2,
		},
		{
			name:            "error in the manifest itself",
			err:             errors.New(`template: main.yaml:7: function "foo" not defined`),
			expectedMessage: `main.yaml:2: function "foo" not defined`,
			expectedFile:    "main.yaml",
			expectedLine:    2,
		},
		{
			name:            "column is preserved",
			err:             errors.New(`template: main.yaml:4:10: executing "main.yaml" at <.value>: nil data`),
			expectedMessage: `_second.yaml:2:10: executing "main.yaml" at <.value>: nil data`,
			expectedFile:    "_second.yaml",
			expectedLine:    2,
		},
		{
			name:         "a name containing a colon does not swallow the line",
			err:          errors.New(`template: C:\repo\main.yaml:4: unexpected "{" in command`),
			templateName: `C:\repo\main.yaml`,
			fragments: []manifestFragment{
				{file: `C:\repo\main.yaml`, firstLine: 1, lineCount: 10},
			},
			expectedMessage: `C:\repo\main.yaml:4: unexpected "{" in command`,
			expectedFile:    `C:\repo\main.yaml`,
			expectedLine:    4,
		},
		{
			// The templating helpers are also used without any fragment recorded,
			// in which case the error must survive untouched.
			name:            "no fragment recorded",
			err:             errors.New(`template: main.yaml:4: unexpected "{" in command`),
			fragments:       []manifestFragment{},
			expectedMessage: `template: main.yaml:4: unexpected "{" in command`,
		},
		{
			name:            "line beyond every fragment",
			err:             errors.New(`template: main.yaml:99: unexpected "{" in command`),
			expectedMessage: `template: main.yaml:99: unexpected "{" in command`,
		},
		{
			name:            "error carrying no location",
			err:             errors.New("no value found for environment variable FOO"),
			expectedMessage: "no value found for environment variable FOO",
		},
		{
			name:            "location of another template",
			err:             errors.New(`template: dummy:4: unexpected "{" in command`),
			expectedMessage: `template: dummy:4: unexpected "{" in command`,
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			templateName := data.templateName
			if templateName == "" {
				templateName = "main.yaml"
			}

			f := fragments
			if data.fragments != nil {
				f = data.fragments
			}

			location, got := locateTemplateError(data.err, templateName, f)

			require.EqualError(t, got, data.expectedMessage)

			if data.expectedFile == "" {
				require.Nil(t, location)
				// An untouched error keeps its identity so that callers can still
				// inspect it.
				require.ErrorIs(t, got, data.err)
				return
			}

			require.NotNil(t, location)
			assert.Equal(t, data.expectedFile, location.fragment.file)
			assert.Equal(t, data.expectedLine, location.line)
			require.ErrorIs(t, got, data.err)
		})
	}
}

func TestLocateTemplateErrorNil(t *testing.T) {
	location, got := locateTemplateError(nil, "main.yaml", nil)

	require.NoError(t, got)
	require.Nil(t, location)
}

func TestRenderExcerpt(t *testing.T) {
	content, fragments := concatenate(t, [][2]string{
		{"_partial.yaml", "p1\np2\np3\n"},
		{"main.yaml", "m1\nm2\nm3\nm4\nm5\nm6\nm7\nm8\nm9\nm10\n"},
	})

	testdata := []struct {
		name            string
		concatLine      int
		contextLines    int
		expectedExcerpt string
	}{
		{
			name:         "in the middle of a file",
			concatLine:   9, // main.yaml:6
			contextLines: 3,
			expectedExcerpt: "  3 | m3\n" +
				"  4 | m4\n" +
				"  5 | m5\n" +
				"> 6 | m6\n" +
				"  7 | m7\n" +
				"  8 | m8\n" +
				"  9 | m9\n",
		},
		{
			name:         "at the top of a file, not spilling into the partial before it",
			concatLine:   4, // main.yaml:1
			contextLines: 3,
			expectedExcerpt: "> 1 | m1\n" +
				"  2 | m2\n" +
				"  3 | m3\n" +
				"  4 | m4\n",
		},
		{
			name:         "at the bottom of a file",
			concatLine:   13, // main.yaml:10
			contextLines: 3,
			expectedExcerpt: "   7 | m7\n" +
				"   8 | m8\n" +
				"   9 | m9\n" +
				"> 10 | m10\n",
		},
		{
			name:         "in a partial, not spilling into the manifest after it",
			concatLine:   2, // _partial.yaml:2
			contextLines: 3,
			expectedExcerpt: "  1 | p1\n" +
				"> 2 | p2\n" +
				"  3 | p3\n",
		},
		{
			name:            "without any surrounding context",
			concatLine:      9,
			contextLines:    0,
			expectedExcerpt: "> 6 | m6\n",
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			location := resolveFragment(fragments, data.concatLine)
			require.NotNil(t, location)

			require.Equal(t, data.expectedExcerpt, renderExcerpt(content, location, data.contextLines))
		})
	}
}

func TestRenderExcerptSingleLineFile(t *testing.T) {
	content, fragments := concatenate(t, [][2]string{{"main.yaml", "only"}})

	location := resolveFragment(fragments, 1)
	require.NotNil(t, location)

	require.Equal(t, "> 1 | only\n", renderExcerpt(content, location, 3))
}

func TestRenderExcerptNoLocation(t *testing.T) {
	require.Empty(t, renderExcerpt([]byte("a\n"), nil, 3))
}

// TestNewStringTemplateLocatesError checks the whole path, from a template the
// engine cannot parse to an error naming the file and line the user wrote.
func TestNewStringTemplateLocatesError(t *testing.T) {
	content, fragments := concatenate(t, [][2]string{
		{"_partial.yaml", "sources:\n  first:\n    kind: shell\n"},
		{"main.yaml", "targets:\n  second:\n    spec:\n      command: {{{ .value }}\n"},
	})

	template := Template{CfgFile: "main.yaml", Fragments: fragments}

	_, err := template.NewStringTemplate(content)

	require.EqualError(t, err, fmt.Sprintf("%s\n\n%s",
		`main.yaml:4: unexpected "{" in command`,
		"  1 | targets:\n"+
			"  2 |   second:\n"+
			"  3 |     spec:\n"+
			"> 4 |       command: {{{ .value }}",
	))
}

// TestNewStringTemplateWithoutFragments checks that the helpers keep working for
// callers that template a standalone string, with no manifest behind it.
func TestNewStringTemplateWithoutFragments(t *testing.T) {
	template := Template{}

	_, err := template.NewStringTemplate([]byte("hello {{ .value }"))

	require.EqualError(t, err, `template: cfg:1: unexpected "}" in operand`)
}
