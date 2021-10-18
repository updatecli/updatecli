package github

import (
	"os"
	"strings"
	"testing"
)

type DataSet []Data
type Data struct {
	github           Github
	expectedTags     []string
	expectedReleases []string
	expectedSource   string
}

var (
	dataSet = DataSet{
		{
			github: Github{
				spec: Spec{
					Owner:      "olblak",
					Repository: "nocode",
					Token:      os.Getenv("GITHUB_TOKEN"),
					Username:   os.Getenv("GITHUB_ACTOR"),
				},
			},
			expectedTags:     []string{"1.0.0"},
			expectedReleases: []string{"1.0.0"},
			expectedSource:   "1.0.0",
		},
	}
)

func TestGetTags(t *testing.T) {

	for _, data := range dataSet {
		tags, err := data.github.SearchTags()

		if len(tags) < len(data.expectedTags) {
			t.Errorf("Error missign tags, expected %v, got %v", data.expectedTags, tags)
		}

		if err != nil {
			t.Errorf("Something went wrong when retrieving tags: %q", err)
		}

		t.Logf("Tags:\n\tExpected:\t%v\n\tGot:\t\t%v\n", data.expectedTags, tags)

		for id, tag := range data.expectedTags {
			if strings.Compare(tag, tags[id]) != 0 {
				t.Errorf("At position %d, expected tag %q, got %q", id, tag, tags[id])
			}
		}

	}
}

func TestSource(t *testing.T) {

	for _, data := range dataSet {
		got, err := data.github.Source("")

		if err != nil {
			t.Errorf("Something went wrong when retrieving source: %q", err)
		}

		if strings.Compare(got, data.expectedSource) != 0 {
			t.Errorf("Expected source value %q, got %q", data.expectedSource, got)
		}
	}

}

func TestSearchReleases(t *testing.T) {

	for _, data := range dataSet {
		releases, err := data.github.SearchReleases()

		if err != nil {
			t.Errorf("Something went wrong when retrieving release: %q", err)
		}

		for id, release := range releases {
			if strings.Compare(release, data.expectedReleases[id]) != 0 {
				t.Errorf("At position %d, expected tag %q, got %q", id, release, data.expectedReleases[id])
			}
		}
	}
}
