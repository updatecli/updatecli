package github

import (
	"os"
	"testing"
)

func TestGetRepositoryLabels(t *testing.T) {

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

	err := g.getRepositoryLabelsInformation()
	if err != nil {
		t.Errorf("unexpected error: %q", err.Error())
	}

	for _, expectedLabel := range expectedLabels {
		found := false
		for _, gotLabel := range g.repositoryLabels {
			if gotLabel.Name == expectedLabel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("couldn't find label %q in %s", expectedLabel, g.repositoryLabels)
		}
	}

	g.getRepositoryLabels()
}
