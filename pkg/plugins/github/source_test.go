package github

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTags(t *testing.T) {

	g := Github{
		Owner:      "updatecli",
		Repository: "updatecli",
		Token:      os.Getenv("GITHUB_TOKEN"),
		Directory:  filepath.Join(os.TempDir(), "tests", "updatecli"),
	}

	_, err := g.Clone()
	if err != nil {
		t.Errorf("Something went wrong when running git clone: %q", err)
	}

	tags, err := g.SearchTags()

	if err != nil {
		t.Errorf("Something went wrong when retrieving tags: %q", err)
	}

	if len(tags) == 0 {
		t.Errorf("No tags found: %q", tags)
	}
}
