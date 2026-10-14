package log

import (
	"strings"
)

// Title logs the title of a major step, such as "Pipeline" or "Auto Discovery".
func Title(title string) {
	Infof("\n\n%s\n", strings.Repeat("+", len(title)+4))
	Infof("+ %s +\n", strings.ToTitle(title))
	Infof("%s\n\n", strings.Repeat("+", len(title)+4))
}

// Section logs the header of a pipeline section, such as a source or a target.
func Section(category, name string) {
	Infof("\n%s: %s\n", category, name)
	Infof("%s\n\n", strings.Repeat("-", len(name)+len(category)+2))
}
