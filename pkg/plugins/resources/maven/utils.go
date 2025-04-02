package maven

import (
	"net/url"
	"path"
	"strings"
)

// isRepositoriesContainsMavenCentral return if the repositories list contains the maven central url
func isRepositoriesContainsMavenCentral(repositories []string) (bool, error) {

	mcu, err := url.Parse(MavenCentralRepository)
	if err != nil {
		return false, err
	}

	for _, repository := range repositories {
		r, err := url.Parse(repository)
		if err != nil {
			return false, err
		}

		if r.Hostname() == mcu.Hostname() {
			return true, nil
		}
	}

	return false, nil
}

// trimUsernamePasswordFromURL parse a maven URL then remove the username/password component from it
func trimUsernamePasswordFromURL(URL string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	result := u.Scheme + "://" + u.Hostname()

	if len(u.Path) > 0 {
		result = result + u.Path
	}

	return result, nil
}

// joinURL returns an url
func joinURL(elems []string) (string, error) {

	var URL string

	if len(elems) == 0 {
		return "", nil
	}

	URL = elems[0]
	if !strings.HasPrefix(elems[0], "https://") && !strings.HasPrefix(elems[0], "http://") {
		URL = "https://" + elems[0]
	}
	URL = strings.TrimSuffix(URL, "/")

	if len(elems) == 1 {
		return URL, nil
	}

	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	URL = u.String()

	remainingElements := elems[1:]

	args := path.Join(remainingElements...)

	return URL + "/" + args, nil
}
