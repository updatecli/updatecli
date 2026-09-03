package yaml

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_TargetCreateMissingKey(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		inputSourceValue string
		mockedContent    string
		wantedContent    string
		wantedResult     bool
		wantedError      bool
		dryRun           bool
	}{
		{
			name:             "Create a missing key at the root of the document",
			spec:             Spec{File: "test.yaml", Key: "$.newkey", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
foo: bar
`,
			wantedContent: `image:
  name: nginx
foo: bar
newkey: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a missing key under an existing mapping",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
foo: bar
`,
			wantedContent: `image:
  name: nginx
  tag: v1
foo: bar
`,
			wantedResult: true,
		},
		{
			name:             "Create a full missing branch",
			spec:             Spec{File: "test.yaml", Key: "$.a.b.c", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
`,
			wantedContent: `image:
  name: nginx
a:
  b:
    c: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a key under a null parent",
			spec:             Spec{File: "test.yaml", Key: "$.github.owner", CreateMissingKey: true},
			inputSourceValue: "olblak",
			mockedContent: `github:
foo: bar
`,
			wantedContent: `github:
  owner: olblak
foo: bar
`,
			wantedResult: true,
		},
		{
			name:             "Create a key inside an empty flow mapping",
			spec:             Spec{File: "test.yaml", Key: "$.github.owner", CreateMissingKey: true},
			inputSourceValue: "olblak",
			mockedContent:    "github: {}\n",
			wantedContent:    "github: {owner: olblak}\n",
			wantedResult:     true,
		},
		{
			name:             "Create a nested key matching a four spaces indentation",
			spec:             Spec{File: "test.yaml", Key: "$.root.a.b", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `root:
    child: 1
`,
			wantedContent: `root:
    child: 1
    a:
        b: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a key with a comment",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true, Comment: "updated by updatecli"},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
`,
			wantedContent: `image:
  name: nginx
  tag: v1 # updated by updatecli
`,
			wantedResult: true,
		},
		{
			name:             "Create a key while preserving the surrounding comments",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `# head comment
image:
  name: nginx # trailing comment
foo: bar
`,
			wantedContent: `# head comment
image:
  name: nginx # trailing comment
  tag: v1
foo: bar
`,
			wantedResult: true,
		},
		{
			name:             "Creating an already existing key set to the same value reports no change",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
  tag: v1
`,
			wantedContent: `image:
  name: nginx
  tag: v1
`,
			wantedResult: false,
		},
		{
			name:             "Creating an already existing key set to another value updates it",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true},
			inputSourceValue: "v2",
			mockedContent: `image:
  name: nginx
  tag: v1
`,
			wantedContent: `image:
  name: nginx
  tag: v2
`,
			wantedResult: true,
		},
		{
			name:             "Create a key in an empty document",
			spec:             Spec{File: "test.yaml", Key: "$.a.b", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent:    "",
			wantedContent: `a:
  b: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a key holding a dot in its name",
			spec:             Spec{File: "test.yaml", Key: `versions.1\.2`, CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `versions:
  a: 1
`,
			wantedContent: `versions:
  a: 1
  "1.2": v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a key in every document of a multi documents file",
			spec:             Spec{File: "test.yaml", Key: "$.newkey", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `a: 1
---
b: 2
`,
			wantedContent: `a: 1
newkey: v1
---
b: 2
newkey: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create a key in a single document of a multi documents file",
			spec:             Spec{File: "test.yaml", Key: "$.newkey", CreateMissingKey: true, DocumentIndex: ptrInt(1)},
			inputSourceValue: "v1",
			mockedContent: `a: 1
---
b: 2
`,
			wantedContent: `a: 1
---
b: 2
newkey: v1
`,
			wantedResult: true,
		},
		{
			name:             "Create multiple keys sharing the same value",
			spec:             Spec{File: "test.yaml", Keys: []string{"$.x.a", "$.z.b"}, CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent:    "foo: bar\n",
			wantedContent: `foo: bar
x:
  a: v1
z:
  b: v1
`,
			wantedResult: true,
		},
		{
			name:             "Creating a key in dry run mode does not write the file",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag", CreateMissingKey: true},
			inputSourceValue: "v1",
			dryRun:           true,
			mockedContent: `image:
  name: nginx
`,
			wantedContent: `image:
  name: nginx
`,
			wantedResult: true,
		},
		{
			name:             "A missing key without createmissingkey still fails",
			spec:             Spec{File: "test.yaml", Key: "$.image.tag"},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
`,
			wantedError: true,
		},
		{
			name:             "Creating a key through a scalar fails",
			spec:             Spec{File: "test.yaml", Key: "$.image.name.sub", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
`,
			wantedError: true,
		},
		{
			name:             "Creating a key through a missing array index fails",
			spec:             Spec{File: "test.yaml", Key: "$.missing[0].name", CreateMissingKey: true},
			inputSourceValue: "v1",
			mockedContent: `image:
  name: nginx
`,
			wantedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTargetTestCase(t, tt.spec, tt.inputSourceValue, tt.mockedContent, tt.wantedContent, tt.wantedResult, tt.wantedError, tt.dryRun)
		})
	}
}

func Test_TargetAppendToArray(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		inputSourceValue string
		mockedContent    string
		wantedContent    string
		wantedResult     bool
		wantedError      bool
		dryRun           bool
	}{
		{
			name:             "Append a value to a block sequence",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent: `tags:
  - v1.0.0
  - v1.1.0
`,
			wantedContent: `tags:
  - v1.0.0
  - v1.1.0
  - v1.2.3
`,
			wantedResult: true,
		},
		{
			name:             "Appending a value already present reports no change",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent: `tags:
  - v1.0.0
  - v1.2.3
`,
			wantedContent: `tags:
  - v1.0.0
  - v1.2.3
`,
			wantedResult: false,
		},
		{
			name:             "Appending a value already present as a folded scalar reports no change",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent: `tags:
  - >-
      v1.2.3
`,
			wantedContent: `tags:
  - >-
      v1.2.3
`,
			wantedResult: false,
		},
		{
			name:             "Appending a value already present as a quoted scalar reports no change",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent: `tags:
  - "v1.2.3"
`,
			wantedContent: `tags:
  - "v1.2.3"
`,
			wantedResult: false,
		},
		{
			name:             "Append a value to a flow sequence",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent:    "tags: [v1.0.0]\n",
			wantedContent:    "tags: [v1.0.0, v1.2.3]\n",
			wantedResult:     true,
		},
		{
			name:             "Append a value to an empty flow sequence",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1.2.3",
			mockedContent:    "tags: []\n",
			wantedContent:    "tags: [v1.2.3]\n",
			wantedResult:     true,
		},
		{
			name:             "Append a value preserving an unusual sequence indentation",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `tags:
   - a
`,
			wantedContent: `tags:
   - a
   - b
`,
			wantedResult: true,
		},
		{
			name:             "Append a value to a nested sequence",
			spec:             Spec{File: "test.yaml", Key: "$.spec.tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `spec:
  tags:
    - a
`,
			wantedContent: `spec:
  tags:
    - a
    - b
`,
			wantedResult: true,
		},
		{
			name:             "Append a value with a comment",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true, Comment: "added by updatecli"},
			inputSourceValue: "b",
			mockedContent: `tags:
  - a
`,
			wantedContent: `tags:
  - a
  - b # added by updatecli
`,
			wantedResult: true,
		},
		{
			name:             "Append creates the missing sequence when allowed",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true, CreateMissingKey: true},
			inputSourceValue: "v1.2.3",
			mockedContent:    "foo: bar\n",
			wantedContent: `foo: bar
tags:
  - v1.2.3
`,
			wantedResult: true,
		},
		{
			name:             "Append creates a missing branch ending on a sequence",
			spec:             Spec{File: "test.yaml", Key: "$.a.tags", AppendToArray: true, CreateMissingKey: true},
			inputSourceValue: "v1.2.3",
			mockedContent:    "foo: bar\n",
			wantedContent: `foo: bar
a:
  tags:
    - v1.2.3
`,
			wantedResult: true,
		},
		{
			name:             "Append replaces a null value with a sequence when allowed",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true, CreateMissingKey: true},
			inputSourceValue: "v1.2.3",
			mockedContent: `tags:
foo: bar
`,
			wantedContent: `tags:
  - v1.2.3
foo: bar
`,
			wantedResult: true,
		},
		{
			name:             "Append to a single document of a multi documents file",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true, DocumentIndex: ptrInt(0)},
			inputSourceValue: "b",
			mockedContent: `tags:
  - a
---
tags:
  - a
`,
			wantedContent: `tags:
  - a
  - b
---
tags:
  - a
`,
			wantedResult: true,
		},
		{
			name:             "Appending in dry run mode does not write the file",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "b",
			dryRun:           true,
			mockedContent: `tags:
  - a
`,
			wantedContent: `tags:
  - a
`,
			wantedResult: true,
		},
		{
			name:             "Appending to a key that is not a sequence fails",
			spec:             Spec{File: "test.yaml", Key: "$.image", AppendToArray: true},
			inputSourceValue: "v1",
			mockedContent:    "image: nginx\n",
			wantedError:      true,
		},
		{
			name:             "Appending to a missing key without createmissingkey fails",
			spec:             Spec{File: "test.yaml", Key: "$.tags", AppendToArray: true},
			inputSourceValue: "v1",
			mockedContent:    "foo: bar\n",
			wantedError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTargetTestCase(t, tt.spec, tt.inputSourceValue, tt.mockedContent, tt.wantedContent, tt.wantedResult, tt.wantedError, tt.dryRun)
		})
	}
}

// Test_TargetSearchPatternIgnoresMissingKey covers the go-yaml engine honoring
// spec.searchpattern when the key is absent, which used to be unreachable because
// goccy reports a missing key as a nil node rather than as ErrNotFoundNode.
func Test_TargetSearchPatternIgnoresMissingKey(t *testing.T) {
	// searchpattern resolves the file against the filesystem, and matches the
	// pattern on paths relative to the working directory, so the test runs from a
	// temporary directory holding the file.
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("test.yaml", []byte("foo: bar\n"), 0600))

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": "foo: bar\n"},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.image.tag", SearchPattern: true})
	require.NoError(t, err)

	y.contentRetriever = &mockedText

	gotResult := result.Target{}
	require.NoError(t, y.Target(context.Background(), "v1", nil, false, &gotResult))

	assert.False(t, gotResult.Changed)
	assert.Equal(t, result.SKIPPED, gotResult.Result)
	assert.Equal(t, "foo: bar\n", mockedText.Contents["test.yaml"])
}

// Test_TargetYamlPathMissingKeyErrors covers the yamlpath engine reporting a key
// missing from every document, which used to be silently skipped because the check
// sat inside the document loop.
func Test_TargetYamlPathMissingKeyErrors(t *testing.T) {
	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": "foo: bar\n---\nbar: foo\n"},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.image.tag", Engine: EngineYamlPath})
	require.NoError(t, err)

	y.contentRetriever = &mockedText
	y.files = map[string]file{
		"test.yaml": {filePath: "test.yaml", originalFilePath: "test.yaml"},
	}

	gotResult := result.Target{}
	assert.Error(t, y.Target(context.Background(), "v1", nil, false, &gotResult))
}

func runTargetTestCase(t *testing.T, spec Spec, inputSourceValue, mockedContent, wantedContent string, wantedResult, wantedError, dryRun bool) {
	t.Helper()

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": mockedContent},
	}

	y, err := New(spec)
	require.NoError(t, err)

	y.contentRetriever = &mockedText
	y.files = map[string]file{
		"test.yaml": {filePath: "test.yaml", originalFilePath: "test.yaml"},
	}

	gotResult := result.Target{}
	gotErr := y.Target(context.Background(), inputSourceValue, nil, dryRun, &gotResult)

	if wantedError {
		assert.Error(t, gotErr)
		return
	}

	require.NoError(t, gotErr)
	assert.Equal(t, wantedResult, gotResult.Changed)

	defer os.Remove("test.yaml")
	assert.Equal(t, wantedContent, mockedText.Contents["test.yaml"])
}

// Test_TargetYamlPathPartialDocumentMatch covers a key present in only some of the
// documents of a file: the documents carrying it are already up to date, so the
// file must be reported as unchanged.
func Test_TargetYamlPathPartialDocumentMatch(t *testing.T) {
	content := "tags: v1\n---\nother: foo\n"

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": content},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.tags", Engine: EngineYamlPath})
	require.NoError(t, err)

	y.contentRetriever = &mockedText
	y.files = map[string]file{
		"test.yaml": {filePath: "test.yaml", originalFilePath: "test.yaml"},
	}

	gotResult := result.Target{}
	require.NoError(t, y.Target(context.Background(), "v1", nil, false, &gotResult))

	assert.False(t, gotResult.Changed)
	assert.Equal(t, result.SUCCESS, gotResult.Result)
}
