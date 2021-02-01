package file

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Line hold rule that define if a specific line should be included or excluded
type Line struct {
	HasIncludes []string
	Includes    []string
	Excludes    []string
}

const (
	//ErrLineNotFound is the error message when no matching line found
	ErrLineNotFound string = "line not found"
)

// ContainsIncluded return a content with only matching lines
func (l *Line) ContainsIncluded(content string) (output string, err error) {
	for _, i := range l.Includes {
		found := false
		for _, line := range strings.Split(content, "\n") {

			if strings.Contains(line, i) {
				output = output + line + "\n"
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("Line '%v', not found in content:\n%v\n", i, content)
			return "", fmt.Errorf(ErrLineNotFound)
		}
	}
	return output, nil
}

// HasIncluded return an error content with only matching lines
func (l *Line) HasIncluded(content string) (found bool, err error) {
	for _, i := range l.HasIncludes {
		found = false
		for _, line := range strings.Split(content, "\n") {

			if strings.Contains(line, i) {
				found = true
			}
		}
		if !found {
			logrus.Errorf("Line '%v', not found in content:\n%v\n", i, content)
			return found, fmt.Errorf(ErrLineNotFound)
		}
	}
	return found, nil
}

// ContainsExcluded return a content without excluded lines
func (l *Line) ContainsExcluded(content string) (output string, err error) {
	output = content
	for _, exclude := range l.Excludes {
		tmp := ""
		for _, line := range strings.Split(output, "\n") {

			if strings.Contains(line, exclude) {
				continue
			}
			if len(line) != 0 {
				tmp = tmp + line + "\n"
			}
		}
		output = tmp
	}
	return output, nil
}
