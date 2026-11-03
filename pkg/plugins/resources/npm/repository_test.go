package npm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepositoryUnmarshalJSON covers the "repository" shorthand that the npm
// registry serves for packages such as date-fns, which used to fail the whole
// packument decoding.
func TestRepositoryUnmarshalJSON(t *testing.T) {
	testdata := []struct {
		name     string
		payload  string
		expected Repository
	}{
		{
			name:     "object form",
			payload:  `{"type":"git","url":"git+https://github.com/date-fns/date-fns.git"}`,
			expected: Repository{Type: "git", URL: "git+https://github.com/date-fns/date-fns.git"},
		},
		{
			name:     "github shorthand",
			payload:  `"github:date-fns/date-fns"`,
			expected: Repository{Type: "git", URL: "git+https://github.com/date-fns/date-fns.git"},
		},
		{
			name:     "bare owner/repo shorthand",
			payload:  `"date-fns/date-fns"`,
			expected: Repository{Type: "git", URL: "git+https://github.com/date-fns/date-fns.git"},
		},
		{
			name:     "gitlab shorthand",
			payload:  `"gitlab:owner/repo"`,
			expected: Repository{Type: "git", URL: "git+https://gitlab.com/owner/repo.git"},
		},
		{
			name:     "url kept untouched",
			payload:  `"https://github.com/date-fns/date-fns"`,
			expected: Repository{Type: "git", URL: "https://github.com/date-fns/date-fns"},
		},
		{
			name:     "gist shorthand kept untouched",
			payload:  `"gist:11081aaa281"`,
			expected: Repository{Type: "git", URL: "gist:11081aaa281"},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			var got Repository
			require.NoError(t, json.Unmarshal([]byte(data.payload), &got))
			assert.Equal(t, data.expected, got)
		})
	}
}

// TestDataUnmarshalJSONWithRepositoryShorthand covers the packument decoding that
// the shorthand used to break.
func TestDataUnmarshalJSONWithRepositoryShorthand(t *testing.T) {
	var got Data
	require.NoError(t, json.Unmarshal([]byte(`{"repository":"github:date-fns/date-fns","dist-tags":{"latest":"4.1.0"}}`), &got))

	assert.Equal(t, "git+https://github.com/date-fns/date-fns.git", got.Repository.URL)
	assert.Equal(t, "4.1.0", got.DistTags.Latest)
}

// TestRepositoryUnmarshalJSONInvalid covers a payload that is neither form still
// reporting the decoding error.
func TestRepositoryUnmarshalJSONInvalid(t *testing.T) {
	var got Repository
	assert.Error(t, json.Unmarshal([]byte(`["github:date-fns/date-fns"]`), &got))
}
