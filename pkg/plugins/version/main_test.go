package version

import (
	"errors"
	"strings"
	"testing"
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
		if err != nil {
			t.Errorf("Unexpected err %q", err.Error())
		}
		got, err := d.Filter.Search(d.Versions)
		if err != nil && d.ExpectedError != nil {
			if strings.Compare(err.Error(), d.ExpectedError.Error()) != 0 {
				t.Errorf("Expected %q, got %q", d.ExpectedError.Error(), err.Error())
			}
		} else if err != nil {
			t.Errorf("Unexpected err %q", err.Error())
		}
		if got != d.ExpectedResult {
			t.Errorf("Expected version %q, got %q", d.ExpectedResult, got)
		}
	}

}

func TestValidate(t *testing.T) {
	for _, v := range validateDataSet {
		err := v.Filter.Validate()
		if err != nil && v.Err != nil {
			if strings.Compare(err.Error(), v.Err.Error()) != 0 {
				t.Errorf("Expected Error %q, got %q", v.Err, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected err %q", err.Error())
		}
	}
}
