package version

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DataSet struct {
	Semver                  Semver
	Versions                []string
	SortedVersions          []string
	ExpectedInitErr         error
	ExpectedParsedVersion   string
	ExpectedOriginalVersion string
	ExpectedSearchErr       error
}

var (
	dataset []DataSet = []DataSet{
		{
			Versions:                []string{"1.0", "2.0", "4.0", "3.0", "6.0", "5.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "6.0.0",
			ExpectedOriginalVersion: "6.0",
		},
		{
			Versions:                []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "6.0.0",
			ExpectedOriginalVersion: "6.0.0",
		},
		{
			Versions:                []string{"v1.0.0", "v2.0.0", "v4.0.0", "v3.0.0", "v6.0.0", "v5.0.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "6.0.0",
			ExpectedOriginalVersion: "v6.0.0",
		},
		{
			Versions:                []string{},
			SortedVersions:          []string{},
			ExpectedInitErr:         errors.New("no valid semantic version found"),
			ExpectedSearchErr:       ErrNoVersionsFound,
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Semver: Semver{
				Constraint: "~5",
			},
			Versions:                []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "5.0.0",
			ExpectedOriginalVersion: "5.0.0",
		},
		{
			Semver: Semver{
				Constraint: "~9",
			},
			Versions:                []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       ErrNoVersionFound,
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Semver: Semver{
				Constraint: "xyz",
			},
			Versions:                []string{"1.0.0", "2.0.0", "4.0.0", "3.0.0", "6.0.0", "5.0.0"},
			SortedVersions:          []string{"6.0.0", "5.0.0", "4.0.0", "3.0.0", "2.0.0", "1.0.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       fmt.Errorf("improper constraint: xyz"),
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Versions:                []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			SortedVersions:          []string{},
			ExpectedInitErr:         errors.New("no valid semantic version found"),
			ExpectedSearchErr:       errors.New("no valid semantic version found"),
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
	}
)

func TestSemVerInit(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)
	}
}

func TestSemVerSort(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)

		d.Semver.Sort()

		for id, version := range d.Semver.versions {
			assert.Equal(t, d.SortedVersions[id], version.String())
		}
	}
}

func TestSemVerSearch(t *testing.T) {
	for id, d := range dataset {
		t.Logf("Dataset position %d", id)
		err := d.Semver.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)

		d.Semver.Sort()
		err = d.Semver.Search(d.Versions)
		gotParsedVersion := d.Semver.FoundVersion.ParsedVersion
		gotOriginalVersion := d.Semver.FoundVersion.OriginalVersion

		assert.Equal(t, d.ExpectedSearchErr, err)
		assert.Equal(t, d.ExpectedParsedVersion, gotParsedVersion)
		assert.Equal(t, d.ExpectedOriginalVersion, gotOriginalVersion)
	}
}
