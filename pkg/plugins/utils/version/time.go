package version

import (
	"maps"
	"slices"
	"time"

	"github.com/sirupsen/logrus"
)

type Time struct {
	// layout is the date layout used to parse the versions
	// it should be a valid Go time layout
	// https://golang.org/pkg/time/#pkg-constants
	// Here is a summary of the components of a layout string.
	// Each element shows by example the formatting of an element of the reference time.
	// Only these values are recognized.
	// Text in the layout string that is not recognized as part of the reference time is echoed verbatim during Format and expected to appear verbatim in the input to Parse.
	// Year: "2006" "06"
	// Month: "Jan" "January" "01" "1"
	// Day of the month: "2" "_2" "02"
	layout       string
	versions     []time.Time
	FoundVersion Version
}

// Init creates a new date object
func (d *Time) Init(versions []string) error {

	for _, version := range versions {
		t, err := time.Parse(d.layout, version)
		if err != nil {
			logrus.Debugf("Skipping %q because %s, skipping", version, err)
			continue
		}

		d.versions = append(d.versions, t)
	}

	if len(d.versions) == 0 {
		return ErrNoValidDateFound
	}

	return nil
}

// Search returns the version matching pattern from a sorted list.
func (d *Time) Search(rawversions map[string]string) error {

	versions := slices.Collect(maps.Keys(rawversions))

	// We need to be sure that at least one version exist
	if len(versions) == 0 {
		return ErrNoVersionsFound
	}

	err := d.Init(versions)
	if err != nil {
		return err
	}
	d.Sort()

	id := d.versions[len(d.versions)-1].Format(d.layout)

	d.FoundVersion.ParsedVersion = rawversions[id]
	d.FoundVersion.OriginalVersion = d.FoundVersion.ParsedVersion

	return nil
}

// Sort re-order a list of versions with the newest version last
func (d *Time) Sort() {
	slices.SortFunc(
		d.versions,
		func(a, b time.Time) int {
			return a.Compare(b)
		})
}
