package gitea

import (
	"fmt"
	"net/url"
	"strings"
)

// Summary returns a brief description of the Gitea SCM configuration
func (g *Gitea) Summary() string {
	var URL *url.URL
	var err error

	urlStr := "gitea.com"
	if g.Spec.URL != "" {
		urlStr = g.Spec.URL
	}

	URL, err = url.Parse(strings.Join([]string{urlStr, g.Spec.Owner, g.Spec.Repository}, "/"))
	if err != nil || URL == nil {
		return ""
	}

	URL.User = nil
	URL.Scheme = ""

	return fmt.Sprintf("%s@%s", strings.TrimPrefix(URL.String(), "//"), g.Spec.Branch)
}
