package github

import (
	"os"
	"testing"
)

func TestGetRepositoryLabelsInformation(t *testing.T) {

	// Short mode also skip integration test that require credentials
	if testing.Short() {
		t.Skip("Skipping test in short mode when it requires specific credentials")
		return
	}

	g := Github{
		spec: Spec{
			Owner:      "olblak",
			Repository: "nocode",
			Username:   os.Getenv("GITHUB_ACTOR"),
			Token:      os.Getenv("GITHUB_TOKEN"),
		},
	}
	expectedLabels := []string{
		"bug",
		"documentation",
		"duplicate",
	}

	gotLabels, err := g.getRepositoryLabelsInformation()
	if err != nil {
		t.Errorf("unexpected error: %q", err.Error())
	}

	for _, expectedLabel := range expectedLabels {
		found := false
		for _, gotLabel := range gotLabels {
			if gotLabel.Name == expectedLabel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("couldn't find label %q in %s", expectedLabel, gotLabels)
		}
	}
}
