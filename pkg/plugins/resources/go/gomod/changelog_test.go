package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		version        GoMod
		expectedResult string
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.version.Changelog())
		})
	}
}
