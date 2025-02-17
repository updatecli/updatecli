package changelog

import (
	"encoding/base64"
	"strings"

	"github.com/google/go-github/v69/github"
)

var (
	// Catalog is a map of github release where the
	// the id represents the repository owner, name and then
	// the range of releases
	// Catalog is an in memory cache to avoid multiple requests
	// to the github API
	Catalog map[string][]*github.RepositoryRelease
)

// generateCatalogID generates a unique id for the catalog
func generateCatalogID(url, owner, repository, from, to string) string {
	rawID := strings.Join([]string{url, owner, repository, from, to}, "|")

	return base64.StdEncoding.EncodeToString([]byte(rawID))
}

// getReleasesFromCatalog returns a list of releases from the catalog
// for a specific catalog id
func getReleasesFromCatalog(id string) []*github.RepositoryRelease {

	if _, ok := Catalog[id]; ok {
		return Catalog[id]
	}

	return nil
}
