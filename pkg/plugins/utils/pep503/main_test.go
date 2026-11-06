package pep503

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	testdata := map[string]string{
		"Jinja2":         "jinja2",
		"zope.interface": "zope-interface",
		"My__Package":    "my-package",
		"a-._b":          "a-b",
		"friendly-bard":  "friendly-bard",
		"Friendly.Bard":  "friendly-bard",
		"FRIENDLY_BARD":  "friendly-bard",
		"friendly--bard": "friendly-bard",
	}

	for name, expected := range testdata {
		assert.Equal(t, expected, NormalizeName(name), name)
	}
}
