package age

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrNoVersionMatchingAge is returned when a resource publishes versions but the age
// filter discarded all of them. It reports a cooldown still running, not a lookup
// failure, so callers are expected to skip rather than to fail.
var ErrNoVersionMatchingAge = errors.New("no version matching the age filter")

// Spec defines parameters to filter versions based on their release date
type Spec struct {
	// Minimum defines the minimum age of a release to be considered valid. It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// Accepted time units are "h" for hours, "d" for days, "w" for weeks, "mo" for months, and "y" for years. If no unit is provided, hours are assumed.
	Minimum string `yaml:",omitempty"`
	// Maximum defines the maximum age of a release to be considered valid. It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// Accepted time units are "h" for hours, "d" for days, "w" for weeks, "mo" for months, and "y" for years. If no unit is provided, hours are assumed.
	Maximum string `yaml:",omitempty"`
}

// Validate checks if the Spec fields are valid duration strings.
func (a Spec) Validate() error {
	if a.Minimum != "" {
		_, err := parseReleaseAge(a.Minimum)
		if err != nil {
			return fmt.Errorf("invalid MinimumReleaseAge %q: %w", a.Minimum, err)
		}
	}

	if a.Maximum != "" {
		_, err := parseReleaseAge(a.Maximum)
		if err != nil {
			return fmt.Errorf("invalid MaximumReleaseAge %q: %w", a.Maximum, err)
		}
	}

	return nil
}

// parseReleaseAge parses a duration string with support for "h", "d", "w", "mo", and "y" time units.
func parseReleaseAge(releaseAge string) (time.Duration, error) {
	releaseAge = strings.TrimSpace(releaseAge)

	if strings.HasSuffix(releaseAge, "d") {
		daysStr := strings.TrimSuffix(releaseAge, "d")
		dayInt, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		return time.ParseDuration(fmt.Sprintf("%dh", 24*dayInt))
	}

	if strings.HasSuffix(releaseAge, "w") {
		weeksStr := strings.TrimSuffix(releaseAge, "w")
		weekInt, err := strconv.Atoi(weeksStr)
		if err != nil {
			return 0, err
		}
		return time.ParseDuration(fmt.Sprintf("%dh", 24*7*weekInt))
	}
	if strings.HasSuffix(releaseAge, "y") {
		yearsStr := strings.TrimSuffix(releaseAge, "y")
		yearInt, err := strconv.Atoi(yearsStr)
		if err != nil {
			return 0, err
		}
		return time.ParseDuration(fmt.Sprintf("%dh", 24*365*yearInt))
	}
	if strings.HasSuffix(releaseAge, "mo") {
		monthsStr := strings.TrimSuffix(releaseAge, "mo")
		monthInt, err := strconv.Atoi(monthsStr)
		if err != nil {
			return 0, err
		}
		return time.ParseDuration(fmt.Sprintf("%dh", 24*365*monthInt/12))
	}

	return time.ParseDuration(releaseAge)
}

// IsOlderThan returns true if the version is older than the date period.
func (a Spec) IsOlderThan(t time.Time, since *time.Time) bool {
	if a.Minimum == "" {
		return false
	}

	if since == nil {
		now := time.Now()
		since = &now
	}

	duration, err := parseReleaseAge(a.Minimum)
	if err != nil {
		logrus.Errorf("invalid MinimumReleaseAge %q: %q\n", a.Minimum, err)
		return false
	}

	return since.Sub(t) < duration
}

// IsNewerThan returns true if the version is newer than the date period.
func (a Spec) IsNewerThan(t time.Time, since *time.Time) bool {
	if a.Maximum == "" {
		return false
	}

	if since == nil {
		now := time.Now()
		since = &now
	}

	duration, err := parseReleaseAge(a.Maximum)
	if err != nil {
		logrus.Errorf("invalid MaximumReleaseAge %q: %q\n", a.Maximum, err)
		return false
	}

	return since.Sub(t) > duration
}

// IsZero return true if filter is not initialized
func (a Spec) IsZero() bool {
	return a.Minimum == "" && a.Maximum == ""
}

// Matches returns true if a release published at t falls inside the age window.
// Both bounds are evaluated against the same instant.
func (a Spec) Matches(t time.Time) bool {
	now := time.Now()
	return !a.IsOlderThan(t, &now) && !a.IsNewerThan(t, &now)
}

// ManifestTemplate defines the "age" Go template used by autodiscovery crawlers to emit
// an age filter in the spec of a generated source. It takes a Spec as argument and emits
// nothing for an empty one:
//
//	tmpl, err := template.New("manifest").Parse(age.ManifestTemplate + myCrawlerTemplate)
//	// then, in myCrawlerTemplate:
//	//   {{- template "age" .Age }}
//
// The emitted keys are indented for a source spec, which is where crawlers use them.
const ManifestTemplate = `{{- define "age" }}
{{- if not .IsZero }}
      age:
        {{- if .Minimum }}
        minimum: '{{ .Minimum }}'
        {{- end }}
        {{- if .Maximum }}
        maximum: '{{ .Maximum }}'
        {{- end }}
{{- end }}
{{- end }}`
