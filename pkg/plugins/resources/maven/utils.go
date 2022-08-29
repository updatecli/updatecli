package maven

import (
	"net/url"
	"strings"
)

// isRepositoriesContainsMavenCentral return if the repositories list contains the maven central url
func isRepositoriesContainsMavenCentral(repositories []string) (bool, error) {

	mcu, err := url.Parse("https://" + MavenCentralRepository)
	if err != nil {
		return false, err
	}

	for i := range repositories {
		// Ensure we only interact with https repositories
		u := strings.TrimPrefix(repositories[i], "https://")
		u = strings.TrimPrefix(u, "http://")

		r, err := url.Parse("https://" + u)
		if err != nil {
			return false, err
		}

		if r.Hostname() == mcu.Hostname() {
			return true, nil
		}
	}

	return false, nil
}
