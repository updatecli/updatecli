package stash

import (
	"fmt"
	"net/url"
	"strings"
)

// Summary returns a brief description of the Stash SCM configuration
func (s *Stash) Summary() string {
	var URL *url.URL
	var err error

	URL, err = url.Parse(strings.Join([]string{s.Spec.URL, s.Spec.Owner, s.Spec.Repository}, "/"))
	if err != nil || URL == nil || s.Spec.URL == "" {
		return ""
	}

	URL.User = nil
	URL.Scheme = ""

	return fmt.Sprintf("%s@%s", strings.TrimPrefix(URL.String(), "//"), s.Spec.Branch)
}
