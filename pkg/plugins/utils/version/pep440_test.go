package version

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Pep440DataSet struct {
	Pep440                  Pep440
	Versions                []string
	SortedVersions          []string
	ExpectedInitErr         error
	ExpectedParsedVersion   string
	ExpectedOriginalVersion string
	ExpectedSearchErr       error
}

var (
	pep440Dataset []Pep440DataSet = []Pep440DataSet{
		{
			Versions:                []string{"1.0", "3.0", "2.0", "6.0", "4.0", "5.0"},
			SortedVersions:          []string{"6.0", "5.0", "4.0", "3.0", "2.0", "1.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "6.0",
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
			Versions:                []string{},
			SortedVersions:          []string{},
			ExpectedInitErr:         errors.New("no valid PEP 440 version found"),
			ExpectedSearchErr:       ErrNoVersionsFound,
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Pep440: Pep440{
				Constraint: ">=2.0,<4.0",
			},
			Versions:                []string{"1.0", "2.0", "3.0", "4.0", "5.0"},
			SortedVersions:          []string{"5.0", "4.0", "3.0", "2.0", "1.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "3.0",
			ExpectedOriginalVersion: "3.0",
		},
		{
			Pep440: Pep440{
				Constraint: ">=9.0",
			},
			Versions:                []string{"1.0", "2.0", "3.0", "4.0", "5.0"},
			SortedVersions:          []string{"5.0", "4.0", "3.0", "2.0", "1.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       ErrNoVersionFound,
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Pep440: Pep440{
				Constraint: "xyz",
			},
			Versions:                []string{"1.0", "2.0", "3.0", "4.0", "5.0"},
			SortedVersions:          []string{"5.0", "4.0", "3.0", "2.0", "1.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       fmt.Errorf("improper constraint: xyz"),
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			Versions:                []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			SortedVersions:          []string{},
			ExpectedInitErr:         errors.New("no valid PEP 440 version found"),
			ExpectedSearchErr:       errors.New("no valid PEP 440 version found"),
			ExpectedParsedVersion:   "",
			ExpectedOriginalVersion: "",
		},
		{
			// Pre-releases sort below their release counterpart per PEP 440.
			Versions:                []string{"1.0a1", "1.0b2", "1.0rc1", "1.0"},
			SortedVersions:          []string{"1.0", "1.0rc1", "1.0b2", "1.0a1"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "1.0",
			ExpectedOriginalVersion: "1.0",
		},
		{
			// Post-releases sort above their release counterpart per PEP 440.
			Versions:                []string{"2.0", "2.0.post1", "1.9"},
			SortedVersions:          []string{"2.0.post1", "2.0", "1.9"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "2.0.post1",
			ExpectedOriginalVersion: "2.0.post1",
		},
		{
			// Wildcard skips dev pre-release in favor of the highest stable version.
			Versions:                []string{"0.28.1", "1.0.dev3"},
			SortedVersions:          []string{"1.0.dev3", "0.28.1"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "0.28.1",
			ExpectedOriginalVersion: "0.28.1",
		},
		{
			// Wildcard skips alpha pre-release in favor of the highest stable version.
			Versions:                []string{"2.57.0", "3.0.0a7"},
			SortedVersions:          []string{"3.0.0a7", "2.57.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "2.57.0",
			ExpectedOriginalVersion: "2.57.0",
		},
		{
			// Wildcard falls back to the highest pre-release when no stable version exists.
			Versions:                []string{"1.0a1", "1.0b2"},
			SortedVersions:          []string{"1.0b2", "1.0a1"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "1.0b2",
			ExpectedOriginalVersion: "1.0b2",
		},
		{
			// Empty constraint behaves identically to wildcard: skips pre-releases.
			Pep440: Pep440{
				Constraint: "",
			},
			Versions:                []string{"2.57.0", "3.0.0a7"},
			SortedVersions:          []string{"3.0.0a7", "2.57.0"},
			ExpectedInitErr:         nil,
			ExpectedSearchErr:       nil,
			ExpectedParsedVersion:   "2.57.0",
			ExpectedOriginalVersion: "2.57.0",
		},
	}
)

func TestPep440Init(t *testing.T) {
	for id, d := range pep440Dataset {
		t.Logf("Dataset position %d", id)
		err := d.Pep440.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)
	}
}

func TestPep440Sort(t *testing.T) {
	for id, d := range pep440Dataset {
		t.Logf("Dataset position %d", id)
		err := d.Pep440.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)

		d.Pep440.Sort()

		for id, version := range d.Pep440.versions {
			assert.Equal(t, d.SortedVersions[id], version.String())
		}
	}
}

func TestPep440Search(t *testing.T) {
	for id, d := range pep440Dataset {
		t.Logf("Dataset position %d", id)
		err := d.Pep440.Init(d.Versions)
		assert.Equal(t, d.ExpectedInitErr, err)

		d.Pep440.Sort()
		err = d.Pep440.Search(d.Versions)
		gotParsedVersion := d.Pep440.FoundVersion.ParsedVersion
		gotOriginalVersion := d.Pep440.FoundVersion.OriginalVersion

		assert.Equal(t, d.ExpectedSearchErr, err)
		assert.Equal(t, d.ExpectedParsedVersion, gotParsedVersion)
		assert.Equal(t, d.ExpectedOriginalVersion, gotOriginalVersion)
	}
}
