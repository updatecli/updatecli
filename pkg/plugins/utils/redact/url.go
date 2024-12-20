// package redact provides function utilities to redact sensitive information.
package redact

import (
	"net/url"
)

// URL redacts the user and password from a valid URL.
// If the URL is invalid, the original URL is returned.
func URL(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return input
	}

	if u.User == nil {
		return u.String()
	}

	return u.Scheme + "://" + "****:****@" + u.Host + u.Path
}
