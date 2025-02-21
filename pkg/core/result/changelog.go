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
	var changelogs string

	changelogs = fmt.Sprintf("\n\n%s:\n", strings.ToTitle("Changelog"))
	changelogs = changelogs + fmt.Sprintf("%s\n", strings.Repeat("-", len("Changelog")+1))

	for _, changelog := range *c {
		changelogs = changelogs + "\n" + changelog.String()
	}

	return changelogs
}

func (c Changelog) String() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n", c.Title, c.PublishedAt, c.URL, c.Body)
}
