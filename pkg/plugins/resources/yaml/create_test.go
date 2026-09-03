package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_splitYamlPathKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantedNames []string
		wantedRaws  []string
		wantedKeys  []bool
		wantedError bool
	}{
		{
			name:        "single key",
			key:         "$.a",
			wantedNames: []string{"a"},
			wantedRaws:  []string{".a"},
			wantedKeys:  []bool{true},
		},
		{
			name:        "nested keys",
			key:         "$.a.b.c",
			wantedNames: []string{"a", "b", "c"},
			wantedRaws:  []string{".a", ".b", ".c"},
			wantedKeys:  []bool{true, true, true},
		},
		{
			name:        "key with an index",
			key:         "$.a[0].b",
			wantedNames: []string{"a", "", "b"},
			wantedRaws:  []string{".a", "[0]", ".b"},
			wantedKeys:  []bool{true, false, true},
		},
		{
			name:        "key with a wildcard index",
			key:         "$.a[*].b",
			wantedNames: []string{"a", "", "b"},
			wantedRaws:  []string{".a", "[*]", ".b"},
			wantedKeys:  []bool{true, false, true},
		},
		{
			name:        "quoted key holding a dot",
			key:         "$.'a.b'.c",
			wantedNames: []string{"a.b", "c"},
			wantedRaws:  []string{".'a.b'", ".c"},
			wantedKeys:  []bool{true, true},
		},
		{
			name:        "quoted key holding an escaped quote",
			key:         `$.'a\'b'`,
			wantedNames: []string{"a'b"},
			wantedRaws:  []string{`.'a\'b'`},
			wantedKeys:  []bool{true},
		},
		{
			name:        "recursive descent is not a key",
			key:         "$..a",
			wantedNames: []string{""},
			wantedRaws:  []string{"..a"},
			wantedKeys:  []bool{false},
		},
		{
			name:        "root only",
			key:         "$",
			wantedNames: []string{},
			wantedRaws:  []string{},
			wantedKeys:  []bool{},
		},
		{
			name:        "missing root",
			key:         "a.b",
			wantedError: true,
		},
		{
			name:        "empty key",
			key:         "",
			wantedError: true,
		},
		{
			name:        "trailing dot",
			key:         "$.a.",
			wantedError: true,
		},
		{
			name:        "trailing recursive descent is parsed as a non creatable element",
			key:         "$.a..",
			wantedNames: []string{"a", ""},
			wantedRaws:  []string{".a", ".."},
			wantedKeys:  []bool{true, false},
		},
		{
			name:        "unterminated quote",
			key:         "$.'a",
			wantedError: true,
		},
		{
			name:        "unterminated index",
			key:         "$.a[0",
			wantedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitYamlPathKey(tt.key)

			if tt.wantedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, len(tt.wantedNames))

			for i := range got {
				assert.Equal(t, tt.wantedNames[i], got[i].name, "name of element %d", i)
				assert.Equal(t, tt.wantedRaws[i], got[i].raw, "raw of element %d", i)
				assert.Equal(t, tt.wantedKeys[i], got[i].isKey, "isKey of element %d", i)
			}

			// Concatenating the raw elements after "$" must rebuild the key, which
			// is what lets the walker query prefixes with goccy.
			assert.Equal(t, tt.key, pathPrefix(got, len(got)))
		})
	}
}

func Test_pathPrefix(t *testing.T) {
	elements, err := splitYamlPathKey("$.a[0].'b.c'")
	require.NoError(t, err)

	assert.Equal(t, "$", pathPrefix(elements, 0))
	assert.Equal(t, "$.a", pathPrefix(elements, 1))
	assert.Equal(t, "$.a[0]", pathPrefix(elements, 2))
	assert.Equal(t, "$.a[0].'b.c'", pathPrefix(elements, 3))
}
