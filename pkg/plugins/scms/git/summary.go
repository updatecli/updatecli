package git

import (
	"fmt"
	"net/url"
	"strings"
)

// Summary returns a brief description of the Git SCM configuration
func (g *Git) Summary() string {
	var URL *url.URL
	var err error

	URL, err = url.Parse(strings.TrimSuffix(g.spec.URL, ".git"))
	if err != nil || URL == nil || g.spec.URL == "" {
		return g.spec.URL
	}

	URL.User = nil
	URL.Scheme = ""

	return fmt.Sprintf("%s@%s", strings.TrimPrefix(URL.String(), "//"), g.spec.Branch)
}
