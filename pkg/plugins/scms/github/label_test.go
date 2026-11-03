package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchLabels(t *testing.T) {
	repositoryLabels := []repositoryLabelApi{
		{ID: "1", Name: "bug", Description: "Something isn't working"},
		{ID: "2", Name: "enhancement", Description: "New feature or request"},
	}

	tests := []struct {
		name        string
		wanted      []string
		wantMatched []repositoryLabelApi
		wantMissing []string
	}{
		{
			name:        "all labels already exist",
			wanted:      []string{"bug", "enhancement"},
			wantMatched: []repositoryLabelApi{repositoryLabels[0], repositoryLabels[1]},
			wantMissing: nil,
		},
		{
			name:        "some labels are missing",
			wanted:      []string{"bug", "updatecli"},
			wantMatched: []repositoryLabelApi{repositoryLabels[0]},
			wantMissing: []string{"updatecli"},
		},
		{
			name:        "no labels wanted",
			wanted:      nil,
			wantMatched: nil,
			wantMissing: nil,
		},
		{
			name:        "none of the wanted labels exist",
			wanted:      []string{"updatecli", "ci"},
			wantMatched: nil,
			wantMissing: []string{"updatecli", "ci"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, missing := matchLabels(tt.wanted, repositoryLabels)
			assert.Equal(t, tt.wantMatched, matched)
			assert.Equal(t, tt.wantMissing, missing)
		})
	}
}
