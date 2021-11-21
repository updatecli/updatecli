package version

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SearchDataSet struct {
	Filter         Filter
	Versions       []string
	ExpectedResult string
	ExpectedError  error
}

type ValidateDataSet struct {
	Filter Filter
	Err    error
}

var (
	searchDataSet []SearchDataSet = []SearchDataSet{
		{
			Filter: Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			Versions:       []string{"1.0", "2.0", "3.0"},
			ExpectedResult: "3.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind: "latest",
			},
			Versions:       []string{"1.0", "2.0", "3.0"},
			ExpectedResult: "3.0",
			ExpectedError:  nil,
		},
		{
			Filter:         Filter{},
			Versions:       []string{"1.0", "2.0", "3.0"},
			ExpectedResult: "3.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind:    "semver",
				Pattern: "~2",
			},
			Versions:       []string{"1.0", "2.0", "3.0"},
			ExpectedResult: "2.0.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind: "semver",
			},
			Versions:       []string{"1.0", "2.0", "3.0"},
			ExpectedResult: "3.0.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind:    "regex",
				Pattern: "^updatecli-2.(\\d*)$",
			},
			Versions:       []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			ExpectedResult: "updatecli-2.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind: "regex",
			},
			Versions:       []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			ExpectedResult: "updatecli-3.0",
			ExpectedError:  nil,
		},
		{
			Filter: Filter{
				Kind:    "semver",
				Pattern: "~2",
			},
			Versions:       []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			ExpectedResult: "",
			ExpectedError:  errors.New("No valid semantic version found"),
		},
	}

	validateDataSet []ValidateDataSet = []ValidateDataSet{
		{
			Filter: Filter{
				Kind:    "semver",
				Pattern: "~2",
			},
			Err: nil,
		},
		{
			Filter: Filter{
				Kind:    "regex",
				Pattern: "~2",
			},
			Err: nil,
		},
		{
			Filter: Filter{
				Kind:    "noExist",
				Pattern: "~2",
			},
			Err: errors.New(`Unsupported version kind "noExist"`),
		},
	}
)

func TestSearch(t *testing.T) {
	for _, d := range searchDataSet {
		err := d.Filter.Validate()
		assert.NoError(t, err)

		err = d.Filter.Search(d.Versions)
		assert.Equal(t, d.ExpectedError, err)

		got := d.Filter.FoundVersion.ParsedVersion
		assert.Equal(t, d.ExpectedResult, got)
	}
}

func TestValidate(t *testing.T) {
	for _, v := range validateDataSet {
		err := v.Filter.Validate()
		assert.Equal(t, v.Err, err)
	}
}
