package gitea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepositoryURL(t *testing.T) {
	tests := []struct {
		raw        string
		wantURL    string
		wantOwner  string
		wantRepo   string
		wantParsed bool
	}{
		{raw: "https://codeberg.org/forgejo/forgejo", wantURL: "https://codeberg.org", wantOwner: "forgejo", wantRepo: "forgejo", wantParsed: true},
		{raw: "https://codeberg.org/forgejo/forgejo.git", wantURL: "https://codeberg.org", wantOwner: "forgejo", wantRepo: "forgejo", wantParsed: true},
		{raw: "https://codeberg.org/forgejo/forgejo/src/branch/forgejo/README.md", wantURL: "https://codeberg.org", wantOwner: "forgejo", wantRepo: "forgejo", wantParsed: true},
		{raw: "https://Gitea.com/gitea/tea/", wantURL: "https://gitea.com", wantOwner: "gitea", wantRepo: "tea", wantParsed: true},
		{raw: "git+https://codeberg.org/owner/repo.git", wantURL: "https://codeberg.org", wantOwner: "owner", wantRepo: "repo", wantParsed: true},
		{raw: "git+ssh://git@codeberg.org/owner/repo.git", wantURL: "https://codeberg.org", wantOwner: "owner", wantRepo: "repo", wantParsed: true},
		{raw: "git@codeberg.org:owner/repo.git", wantURL: "https://codeberg.org", wantOwner: "owner", wantRepo: "repo", wantParsed: true},
		{raw: "codeberg.org/owner/repo/v2", wantURL: "https://codeberg.org", wantOwner: "owner", wantRepo: "repo", wantParsed: true},
		{raw: "https://github.com/updatecli/updatecli"},
		{raw: "https://gitea.example.com/owner/repo"},
		{raw: "https://codeberg.org/owner"},
		{raw: "https://codeberg.org/owner/.git"},
		{raw: "git@codeberg.org"},
		{raw: ""},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			gotURL, gotOwner, gotRepo, gotParsed := ParseRepositoryURL(tt.raw)
			assert.Equal(t, tt.wantParsed, gotParsed)
			assert.Equal(t, tt.wantURL, gotURL)
			assert.Equal(t, tt.wantOwner, gotOwner)
			assert.Equal(t, tt.wantRepo, gotRepo)
		})
	}
}
