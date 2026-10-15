package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "empty url defaults to gitlab.com",
			url:  "",
			want: "https://gitlab.com",
		},
		{
			name: "url without scheme gets https prepended",
			url:  "gitlab.example.com",
			want: "https://gitlab.example.com",
		},
		{
			name: "url with https scheme is left untouched",
			url:  "https://gitlab.example.com",
			want: "https://gitlab.example.com",
		},
		{
			name: "url with http scheme is left untouched",
			url:  "http://gitlab.example.com",
			want: "http://gitlab.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, EnsureValidURL(tt.url))
		})
	}
}
