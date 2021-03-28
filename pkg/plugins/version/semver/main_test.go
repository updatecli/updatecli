package semver

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type DataSet struct {
	Semver            Semver
	Versions          []string
	SortedVersions    []string
	ExpectedInitErr   error
	ExpectedVersion   string
	ExpectedSearchErr error
}

var (
	dataset []DataSet = []DataSet{
		{
			Versions:          []string{"1.0", "2.0", "4.0", "3.0", "6.0", "5.0"},
			SortedVersions:    []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:   nil,
			ExpectedSearchErr: nil,
			ExpectedVersion:   "6.0.0",
		},
		{
			Versions:          []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:    []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:   nil,
			ExpectedSearchErr: nil,
			ExpectedVersion:   "6.0.0",
		},
		{
			Versions:          []string{},
			SortedVersions:    []string{},
			ExpectedInitErr:   errors.New("No valid semantic version found"),
			ExpectedSearchErr: ErrNoVersionsFound,
			ExpectedVersion:   "",
		},
		{
			Semver: Semver{
				Constraint: "~5",
			},
			Versions:          []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:    []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:   nil,
			ExpectedSearchErr: nil,
			ExpectedVersion:   "5.0.0",
		},
		{
			Semver: Semver{
				Constraint: "~9",
			},
			Versions:          []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:    []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:   nil,
			ExpectedSearchErr: ErrNoVersionFound,
			ExpectedVersion:   "",
		},
		{
			Semver: Semver{
				Constraint: "xyz",
			},
			Versions:          []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:    []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:   nil,
			ExpectedSearchErr: fmt.Errorf("improper constraint: xyz"),
			ExpectedVersion:   "",
		},
		{
			Versions:          []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			SortedVersions:    []string{},
			ExpectedInitErr:   errors.New("No valid semantic version found"),
			ExpectedSearchErr: errors.New("No valid semantic version found"),
			ExpectedVersion:   "",
		},
	}
)

func TestInit(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		if err != nil && d.ExpectedInitErr != nil {
			if strings.Compare(err.Error(), d.ExpectedInitErr.Error()) != 0 {
				t.Errorf("Unexpected error %q, got %q", err.Error(), d.ExpectedInitErr)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q", err)
		}
	}
}

func TestSort(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		if err != nil && d.ExpectedInitErr != nil {
			if strings.Compare(err.Error(), d.ExpectedInitErr.Error()) != 0 {
				t.Errorf("Unexpected error %q, got %q", err.Error(), d.ExpectedInitErr)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q", err)
		}

		d.Semver.Sort()

		for id, version := range d.Semver.versions {
			if version.String() != d.SortedVersions[id] {
				t.Errorf("At position %d, version should be %q instead of %q", id, version, d.SortedVersions[id])
			}
		}
	}
}

func TestSearch(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		if err != nil && d.ExpectedInitErr != nil {
			if strings.Compare(err.Error(), d.ExpectedInitErr.Error()) != 0 {
				t.Errorf("Unexpected error %q, got %q", err.Error(), d.ExpectedInitErr)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q", err)
		}

		d.Semver.Sort()
		got, err := d.Semver.Search(d.Versions)
		if err != nil && d.ExpectedSearchErr != nil {
			if strings.Compare(err.Error(), d.ExpectedSearchErr.Error()) != 0 {
				t.Errorf("Unexpected error %q, got %q", err.Error(), d.ExpectedSearchErr)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q", err)
		}

		if got != d.ExpectedVersion {
			t.Errorf("Expected version %q, got %q", d.ExpectedVersion, got)
		}
	}
}
