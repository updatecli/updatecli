package vprefix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name      string
		from, to  string
		found     string
		wantCalls [][2]string
		wantFound bool
	}{
		{name: "plain tags match", from: "1.0.0", to: "2.0.0", found: "2.0.0", wantCalls: [][2]string{{"1.0.0", "2.0.0"}}, wantFound: true},
		{name: "prefixed tags match", from: "1.0.0", to: "2.0.0", found: "v2.0.0", wantCalls: [][2]string{{"1.0.0", "2.0.0"}, {"v1.0.0", "v2.0.0"}}, wantFound: true},
		{name: "already prefixed is not retried", from: "v1.0.0", to: "v2.0.0", wantCalls: [][2]string{{"v1.0.0", "v2.0.0"}}},
		{name: "empty from keeps empty", to: "2.0.0", wantCalls: [][2]string{{"", "2.0.0"}, {"", "v2.0.0"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls [][2]string
			got := Search(tt.from, tt.to, func(from, to string) *result.Changelogs {
				calls = append(calls, [2]string{from, to})
				if to == tt.found {
					return &result.Changelogs{{Title: to}}
				}
				return nil
			})
			assert.Equal(t, tt.wantCalls, calls)
			assert.Equal(t, tt.wantFound, got != nil)
		})
	}
}

func TestAdd(t *testing.T) {
	assert.Equal(t, "", add(""))
	assert.Equal(t, "v2.31.0", add("v2.31.0"))
	assert.Equal(t, "v2.31.0", add("2.31.0"))
}
