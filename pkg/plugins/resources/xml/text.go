package xml

import (
	"strings"

	"github.com/beevik/etree"
)

// elemText returns the text of an element without the whitespace around it,
// which an element spanning several lines often has.
func elemText(elem *etree.Element) string {
	return strings.TrimSpace(elem.Text())
}

// setElemText replaces the text of an element while keeping the whitespace around it,
// so an element spanning several lines keeps its layout.
func setElemText(elem *etree.Element, value string) {
	current := elem.Text()
	trimmed := strings.TrimSpace(current)

	if trimmed == "" {
		elem.SetText(value)
		return
	}

	start := strings.Index(current, trimmed)
	elem.SetText(current[:start] + value + current[start+len(trimmed):])
}
