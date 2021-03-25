package github

import (
	"os"
	"strings"
	"testing"
)

type DataSet struct {
	github           Github
	expectedTags     []string
	expectedReleases []string
	expectedSource   string
}

var (
	dataSet = DataSet{
		github: Github{
			Owner:      "olblak",
			Repository: "nocode",
			Token:      os.Getenv("GITHUB_TOKEN"),
		},
		expectedTags:     []string{"1.0.0"},
		expectedReleases: []string{"1.0.0"},
		expectedSource:   "1.0.0",
	}
)

func TestGetTags(t *testing.T) {

	tags, err := dataSet.github.SearchTags()

	if err != nil {
		t.Errorf("Something went wrong when retrieving tags: %q", err)
	}

	for id, tag := range tags {
		if strings.Compare(tag, dataSet.expectedTags[id]) != 0 {
			t.Errorf("At position %d, expected tag %q, got %q", id, tag, dataSet.expectedTags[id])
		}
	}
}

func TestSource(t *testing.T) {

	got, err := dataSet.github.Source("")

	if err != nil {
		t.Errorf("Something went wrong when retrieving tags: %q", err)
	}

	if strings.Compare(got, dataSet.expectedSource) != 0 {
		t.Errorf("Expected source value %q, got %q", dataSet.expectedSource, got)
	}
}

func TestSearchReleases(t *testing.T) {

	releases, err := dataSet.github.SearchReleases()

	if err != nil {
		t.Errorf("Something went wrong when retrieving tags: %q", err)
	}

	for id, release := range releases {
		if strings.Compare(release, dataSet.expectedReleases[id]) != 0 {
			t.Errorf("At position %d, expected tag %q, got %q", id, release, dataSet.expectedReleases[id])
		}
	}
}
