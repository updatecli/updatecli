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
		{
			// goccy answers a wildcard query with a detached sequence wrapping the
			// matched nodes, so appending to it used to report a change while
			// leaving the file untouched.
			name:             "Append a value to every sequence matched by a wildcard",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `agents:
  - name: first
    tags:
      - a
  - name: second
    tags:
      - a
`,
			wantedContent: `agents:
  - name: first
    tags:
      - a
      - b
  - name: second
    tags:
      - a
      - b
`,
			wantedResult: true,
		},
		{
			name:             "Append to a wildcard skips the sequences already holding the value",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `agents:
  - name: first
    tags:
      - a
      - b
  - name: second
    tags:
      - a
`,
			wantedContent: `agents:
  - name: first
    tags:
      - a
      - b
  - name: second
    tags:
      - a
      - b
`,
			wantedResult: true,
		},
		{
			name:             "Append to a wildcard reports no change when every sequence holds the value",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `agents:
  - name: first
    tags:
      - b
  - name: second
    tags:
      - b
`,
			wantedContent: `agents:
  - name: first
    tags:
      - b
  - name: second
    tags:
      - b
`,
			wantedResult: false,
		},
		{
			name:             "Append a value to every sequence matched by a recursive descent",
			spec:             Spec{File: "test.yaml", Key: "$..tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `spec:
  tags:
    - a
status:
  tags:
    - a
`,
			wantedContent: `spec:
  tags:
    - a
    - b
status:
  tags:
    - a
    - b
`,
			wantedResult: true,
		},
		{
			name:             "Appending to a wildcard matching a value that is not a sequence fails",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `agents:
  - name: first
    tags: a
`,
			wantedError: true,
		},
		{
			name:             "Appending to a recursive descent matching nothing fails",
			spec:             Spec{File: "test.yaml", Key: "$..tags", AppendToArray: true},
			inputSourceValue: "b",
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

// Test_TargetWildcardKey covers keys selecting several nodes. goccy answers such a
// query with a detached sequence holding one entry per selected position, and a nil
// entry where the key is absent, so a non-nil node used to be mistaken for a key
// that exists everywhere: the target then reported a change that ReplaceWithNode
// had not made.
func Test_TargetWildcardKey(t *testing.T) {
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
			name:             "Update every node selected by a wildcard",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tag"},
			inputSourceValue: "v2",
			mockedContent: `agents:
  - name: first
    tag: v1
  - name: second
    tag: v1
`,
			wantedContent: `agents:
  - name: first
    tag: v2
  - name: second
    tag: v2
`,
			wantedResult: true,
		},
		{
			name:             "Updating a wildcard reports no change when every selected node already holds the value",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tag"},
			inputSourceValue: "v2",
			mockedContent: `agents:
  - name: first
    tag: v2
  - name: second
    tag: v2
`,
			wantedContent: `agents:
  - name: first
    tag: v2
  - name: second
    tag: v2
`,
			wantedResult: false,
		},
		{
			name:             "Updating a wildcard reports a change when only some selected nodes hold the value",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tag"},
			inputSourceValue: "v2",
			mockedContent: `agents:
  - name: first
    tag: v2
  - name: second
    tag: v1
`,
			wantedContent: `agents:
  - name: first
    tag: v2
  - name: second
    tag: v2
`,
			wantedResult: true,
		},
		{
			name:             "A key missing from every node selected by a wildcard is not found",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tag"},
			inputSourceValue: "v2",
			mockedContent: `agents:
  - name: first
  - name: second
`,
			wantedError: true,
		},
		{
			name:             "A key missing from some of the nodes selected by a wildcard fails",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tag"},
			inputSourceValue: "v2",
			mockedContent: `agents:
  - name: first
    tag: v1
  - name: second
`,
			wantedError: true,
		},
		{
			name:             "A recursive selector matching nothing is not found",
			spec:             Spec{File: "test.yaml", Key: "$..tag"},
			inputSourceValue: "v2",
			mockedContent:    "foo: bar\n",
			wantedError:      true,
		},
		{
			name:             "Update every node selected by a recursive selector",
			spec:             Spec{File: "test.yaml", Key: "$..tag"},
			inputSourceValue: "v2",
			mockedContent: `spec:
  tag: v1
status:
  tag: v1
`,
			wantedContent: `spec:
  tag: v2
status:
  tag: v2
`,
			wantedResult: true,
		},
		{
			name:             "Appending to a sequence missing from every node selected by a wildcard is not found",
			spec:             Spec{File: "test.yaml", Key: "$.agents[*].tags", AppendToArray: true},
			inputSourceValue: "b",
			mockedContent: `agents:
  - name: first
  - name: second
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

// Test_CreateMissingKeyRejectsMultiMatchKey covers createmissingkey refusing a key
// that selects several nodes: goccy's selector replacement only rewrites the
// positions already holding the key, so creating through such a key writes nothing.
func Test_CreateMissingKeyRejectsMultiMatchKey(t *testing.T) {
	for _, key := range []string{"$.agents[*].tag", "$..tag", "$.agents[*].spec.tag"} {
		t.Run(key, func(t *testing.T) {
			_, err := New(Spec{File: "test.yaml", Key: key, CreateMissingKey: true})
			require.ErrorContains(t, err, "`spec.createmissingkey` does not support the wildcard or recursive selector")
		})
	}

	// A trailing "[*]" addresses the sequence itself, which goccy resolves to the
	// live node, so it stays supported.
	_, err := New(Spec{File: "test.yaml", Key: "$.tags[*]", CreateMissingKey: true})
	require.NoError(t, err)
}

// Test_TargetSearchPatternIgnoresMissingWildcardKey covers spec.searchpattern being
// honored when a wildcard selects nodes that do not hold the key, which used to be
// unreachable because the detached wrapper goccy returns is not nil.
func Test_TargetSearchPatternIgnoresMissingWildcardKey(t *testing.T) {
	content := "agents:\n  - name: first\n  - name: second\n"

	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": content},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.agents[*].tag", SearchPattern: true})
	require.NoError(t, err)

	y.contentRetriever = &mockedText

	gotResult := result.Target{}
	require.NoError(t, y.Target(context.Background(), "v2", nil, false, &gotResult))

	assert.False(t, gotResult.Changed)
	assert.Equal(t, result.SKIPPED, gotResult.Result)
	assert.Equal(t, content, mockedText.Contents["test.yaml"])
}

// Test_TargetSearchPatternUpdatesPartialWildcardKey covers spec.searchpattern
// relaxing the requirement that every position selected by a wildcard holds the
// key: the ones holding it are updated instead of failing the target. A wildcard
// selecting nothing at all is skipped, so failing on a wildcard selecting some
// would be incoherent.
func Test_TargetSearchPatternUpdatesPartialWildcardKey(t *testing.T) {
	content := "agents:\n  - name: first\n    tag: v1\n  - name: second\n"

	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": content},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.agents[*].tag", SearchPattern: true})
	require.NoError(t, err)

	y.contentRetriever = &mockedText

	gotResult := result.Target{}
	require.NoError(t, y.Target(context.Background(), "v2", nil, false, &gotResult))

	assert.True(t, gotResult.Changed)
	assert.Equal(t, result.ATTENTION, gotResult.Result)
	assert.Equal(t, "agents:\n  - name: first\n    tag: v2\n  - name: second\n", mockedText.Contents["test.yaml"])
}

// Test_TargetDocumentIndexOutOfRange covers a documentindex addressing no document
// of the file: the keys loop then evaluates nothing, which used to leave the target
// reporting a success it had not made. The yamlpath engine and the source both
// report this as a missing key.
func Test_TargetDocumentIndexOutOfRange(t *testing.T) {
	documentIndex := 5

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": "image:\n  tag: v1\n"},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.image.tag", DocumentIndex: &documentIndex})
	require.NoError(t, err)

	y.contentRetriever = &mockedText
	y.files = map[string]file{
		"test.yaml": {filePath: "test.yaml", originalFilePath: "test.yaml"},
	}

	gotResult := result.Target{}
	gotErr := y.Target(context.Background(), "v2", nil, false, &gotResult)

	require.ErrorContains(t, gotErr, "documentindex 5 addresses no document of file")
	assert.False(t, gotResult.Changed)
	assert.Equal(t, "image:\n  tag: v1\n", mockedText.Contents["test.yaml"])
}

// Test_TargetSearchPatternIgnoresDocumentIndexOutOfRange covers spec.searchpattern
// relaxing that check, as the matched files need not all hold the same number of
// documents.
func Test_TargetSearchPatternIgnoresDocumentIndexOutOfRange(t *testing.T) {
	content := "image:\n  tag: v1\n"
	documentIndex := 5

	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

	mockedText := text.MockTextRetriever{
		Contents: map[string]string{"test.yaml": content},
	}

	y, err := New(Spec{File: "test.yaml", Key: "$.image.tag", DocumentIndex: &documentIndex, SearchPattern: true})
	require.NoError(t, err)

	y.contentRetriever = &mockedText

	gotResult := result.Target{}
	require.NoError(t, y.Target(context.Background(), "v2", nil, false, &gotResult))

	assert.False(t, gotResult.Changed)
	assert.Equal(t, result.SKIPPED, gotResult.Result)
	assert.Equal(t, content, mockedText.Contents["test.yaml"])
}
