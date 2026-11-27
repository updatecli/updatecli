package docker

import (
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"regexp"
	"slices"
	"strings"
)

var (
	// repositoryPathComponent matches a path component of an image name after the registry host,
	// as the OCI distribution specification allows it: lowercase only.
	repositoryPathComponent = regexp.MustCompile(`^[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*$`)
)

// RegistryMappings maps an image name prefix, as written in the discovered files,
// to the image name prefix where the new versions are looked up.
//
// It is meant for images pulled through an internal mirror, such as
// "my-registry.com/docker-io/nginx", whose tags are checked on the upstream
// registry, such as "docker.io/nginx", while the files keep the mirror name.
type RegistryMappings map[string]string

// Validate returns an error if a mapping has an empty prefix on either side,
// if two mirror prefixes only differ by a trailing "/", or if an upstream prefix
// can't start an image name.
func (m RegistryMappings) Validate() error {
	var validationErrors []string

	// Sorted, so that the error message is the same on every run.
	mirrorPrefixes := slices.Sorted(maps.Keys(m))
	// seen maps a mirror prefix, as Resolve compares it, to the key it comes from.
	seen := make(map[string]string, len(m))

	for _, mirrorPrefix := range mirrorPrefixes {
		to := m[mirrorPrefix]

		if strings.Trim(mirrorPrefix, "/") == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("registry mapping %q => %q: the mirror prefix is empty", mirrorPrefix, to))
		} else {
			prefix := strings.TrimSuffix(mirrorPrefix, "/")
			if other, found := seen[prefix]; found {
				validationErrors = append(validationErrors, fmt.Sprintf("registry mapping %q => %q: the mirror prefix is the same as %q", mirrorPrefix, to, other))
			}
			seen[prefix] = mirrorPrefix
		}

		if err := validateUpstreamPrefix(to); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("registry mapping %q => %q: %s", mirrorPrefix, to, err))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("invalid registry mappings:\n%s", strings.Join(validationErrors, "\n"))
	}

	return nil
}

// validateUpstreamPrefix returns an error if to can't start an image name,
// such as a prefix holding a tag, a digest, an empty path segment or an invalid path component.
func validateUpstreamPrefix(to string) error {
	if strings.Trim(to, "/") == "" {
		return errors.New("the upstream prefix is empty")
	}

	if strings.ContainsAny(to, " \t\r\n@") {
		return errors.New("the upstream prefix holds whitespace or a digest")
	}

	segments := strings.Split(strings.TrimSuffix(to, "/"), "/")
	for i, segment := range segments {
		if segment == "" {
			return errors.New("the upstream prefix has an empty path segment")
		}

		// The first segment is the registry host, which may be followed by a port.
		if i == 0 {
			if err := validateRegistryHost(segment); err != nil {
				return err
			}
			continue
		}

		if strings.Contains(segment, ":") {
			return errors.New("the upstream prefix holds a tag")
		}

		if !repositoryPathComponent.MatchString(segment) {
			return fmt.Errorf("the upstream prefix has an invalid repository path component %q, only lowercase letters, digits and separators are allowed", segment)
		}
	}

	return nil
}

// validateRegistryHost returns an error if host, the first segment of an upstream prefix,
// holds a ":" not followed by a port. A bracketed IPv6 address, such as "[::1]" or "[::1]:5000", is accepted.
func validateRegistryHost(host string) error {
	port := ""
	found := false

	if strings.HasPrefix(host, "[") {
		address, rest, closed := strings.Cut(host[1:], "]")
		if !closed {
			return errors.New("the upstream prefix has an unclosed IPv6 address")
		}
		if ip, err := netip.ParseAddr(address); err != nil || !ip.Is6() {
			return fmt.Errorf("the upstream prefix has an invalid IPv6 address %q", address)
		}
		if rest == "" {
			return nil
		}
		port, found = strings.CutPrefix(rest, ":")
		if !found {
			return errors.New("the upstream prefix has an invalid registry host")
		}
	} else {
		var name string
		name, port, found = strings.Cut(host, ":")
		if !found {
			return nil
		}
		if name == "" {
			return errors.New("the upstream prefix holds a tag")
		}
	}

	if port == "" || strings.Trim(port, "0123456789") != "" {
		return errors.New("the upstream prefix holds a tag")
	}

	return nil
}

// Resolve returns the image name to look up for image, an image name without tag or digest.
//
// A mapping applies when its prefix equals the image name or is followed by "/" in it,
// so "my-registry.com/docker-io" matches "my-registry.com/docker-io/nginx"
// but not "my-registry.com/docker-iox/nginx". When several mappings apply,
// the longest prefix wins. Without any match, image is returned unchanged.
func (m RegistryMappings) Resolve(image string) string {
	bestFrom := ""
	bestTo := ""

	for from, to := range m {
		from = strings.TrimSuffix(from, "/")
		if from == "" {
			continue
		}

		if image != from && !strings.HasPrefix(image, from+"/") {
			continue
		}

		if len(from) > len(bestFrom) {
			bestFrom = from
			bestTo = strings.TrimSuffix(to, "/")
		}
	}

	if bestFrom == "" {
		return image
	}

	return bestTo + strings.TrimPrefix(image, bestFrom)
}
