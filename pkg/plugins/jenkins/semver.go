package jenkins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Semver struct to store a semver.
type Semver struct {
	Major      string
	Minor      string
	Patch      string
	Prerelease string
}

// NewVersion creates a new Semver.
func NewVersion(in string) Semver {
	s := Semver{}

	if strings.Contains(in, "+") {
		panic("this library does not support build numbers '" + in + "'")
	}

	if strings.Contains(in, ".JENKINS") {
		in = strings.ReplaceAll(in, ".JENKINS", "-JENKINS")
	}

	if strings.Contains(in, "-") {
		parts := strings.SplitN(in, "-", 2)
		s.Prerelease = parts[1]
		// parse the first part
		in = parts[0]
	}

	parts := strings.Split(in, ".")
	if len(parts) == 1 {
		s.Major = parts[0]
	} else if len(parts) == 2 {
		s.Major = parts[0]
		s.Minor = parts[1]
	} else if len(parts) == 3 {
		s.Major = parts[0]
		s.Minor = parts[1]
		s.Patch = parts[2]
	} else {
		panic("invalid number of parts '" + in + "'")
	}

	return s
}

func toInt(in string) (int, error) {
	if in == "" {
		return 0, nil
	}
	s, err := strconv.Atoi(in)
	if err != nil {
		return -1, errors.Wrap(err, "unable to parse as int: "+in)
	}
	return s, nil
}

// String string representation of this struct.
func (v *Semver) String() string {
	if v.Minor == "" {
		return v.Major
	} else if v.Patch == "" {
		return fmt.Sprintf("%s.%s", v.Major, v.Minor)
	} else if v.Prerelease == "" {
		return fmt.Sprintf("%s.%s.%s", v.Major, v.Minor, v.Patch)
	} else {
		return fmt.Sprintf("%s.%s.%s-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
}

// LessThan returns true if this Semver is less than the supplied one.
func (v Semver) LessThan(o Semver) bool {
	if v.Major != o.Major {
		val, err := v.lessThan(v.Major, o.Major)
		if err != nil {
			fmt.Printf("unable to compare '%s' & '%s' - %s", v.String(), o.String(), err)
			return false
		}
		return val
	}

	if v.Minor != o.Minor {
		val, err := v.lessThan(v.Minor, o.Minor)
		if err != nil {
			fmt.Printf("unable to compare '%s' & '%s' - %s", v.String(), o.String(), err)
			return false
		}
		return val
	}

	if v.Patch != o.Patch {
		val, err := v.lessThan(v.Patch, o.Patch)
		if err != nil {
			fmt.Printf("unable to compare '%s' & '%s' - %s", v.String(), o.String(), err)
			return false
		}
		return val
	}

	if v.Prerelease == o.Prerelease {
		return false
	}

	if v.Prerelease == "" {
		return false
	}

	if o.Prerelease == "" {
		return true
	}

	return strings.Compare(v.Prerelease, o.Prerelease) < 0
}

func (v Semver) lessThan(v1 string, v2 string) (bool, error) {
	i1, err := toInt(v1)
	if err != nil {
		return false, err
	}
	i2, err := toInt(v2)
	if err != nil {
		return false, err
	}
	return i1 < i2, nil
}

type bySemVer []string

func (s bySemVer) Len() int {
	return len(s)
}
func (s bySemVer) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bySemVer) Less(i, j int) bool {
	v1 := NewVersion(s[i])
	v2 := NewVersion(s[j])

	return v1.LessThan(v2)
}
