package result

import (
	"fmt"
	"strings"
)

// Changelog represents a changelog entry
type Changelog struct {
	// Title represents the title of the changelog. Typically a version number.
	Title string
	// Body represents the body of the changelog.
	Body string
	// PublishedAt represents the date the changelog was published.
	PublishedAt string
	// URL represents the URL to the changelog.
	URL string
}

// Changelogs represents a list of Changelog
type Changelogs []Changelog

func (c *Changelogs) String() string {
	var changelogs []string

	for _, changelog := range *c {
		changelogs = append(changelogs, changelog.String())
	}

	return strings.Join(changelogs, "\n")
}

func (c Changelog) String() string {
	return fmt.Sprintf("Title: %s\nPublished At: %s\nLink: %s\nDescription: %s\n", c.Title, c.PublishedAt, c.URL, c.Body)
}
