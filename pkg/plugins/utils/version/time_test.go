package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeSearch(t *testing.T) {
	data := []struct {
		name                 string
		versions             []string
		expected             string
		expectedError        bool
		expectedErrorMessage error
		layout               string
	}{
		{

			name: "All valid versions",
			versions: []string{
				"2021-01-01",
				"2021-01-02",
				"2022-03-01",
				"2021-02-01",
			},
			layout:   "2006-01-02",
			expected: "2022-03-01",
		},
		{

			name: "All valid versions but wrong layout",
			versions: []string{
				"2021-01-01",
				"2021-01-02",
				"2022-03-01",
				"2021-02-01",
			},
			layout:               "20060102",
			expectedError:        true,
			expectedErrorMessage: ErrNoValidDateFound,
		},
		{

			name:                 "no versions specified",
			versions:             []string{},
			expectedError:        true,
			expectedErrorMessage: ErrNoVersionsFound,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			dateVersion := Time{
				layout: d.layout,
			}

			mapVersions := make(map[string]string)
			for _, v := range d.versions {
				mapVersions[v] = v
			}

			err := dateVersion.Search(mapVersions)

			if d.expectedError {
				assert.ErrorIs(t, err, d.expectedErrorMessage)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dateVersion.FoundVersion.ParsedVersion, d.expected)
		})
	}
}
