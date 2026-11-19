package dockerimage

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// maxAgeChecks bounds the number of tags dated to find one matching the age filter,
	// as every tag costs requests to the registry.
	maxAgeChecks      = 20
	createdAnnotation = "org.opencontainers.image.created"
)

// errNoReleaseDate reports that none of the tags checked has a creation date. The source is
// skipped rather than failed, as when the age filter discards every tag.
var errNoReleaseDate = errors.New("no release date")

// releaseDate returns the creation date of a tag for a platform. It returns false when the image
// has none, or one set to the Unix epoch as reproducible builds do.
func (di *DockerImage) releaseDate(tag, arch string) (time.Time, bool, error) {
	ref, err := di.createRef(tag)
	if err != nil {
		return time.Time{}, false, err
	}

	descriptor, err := remote.Get(ref, di.remoteOptions(arch)...)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
	}

	if descriptor.MediaType.IsIndex() {
		index, err := descriptor.ImageIndex()
		if err != nil {
			return time.Time{}, false, err
		}
		manifest, err := index.IndexManifest()
		if err != nil {
			return time.Time{}, false, err
		}
		if date, ok := annotatedDate(manifest.Annotations); ok {
			return date, true, nil
		}
	}

	img, err := descriptor.Image()
	if err != nil {
		return time.Time{}, false, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return time.Time{}, false, err
	}
	if date, ok := annotatedDate(manifest.Annotations); ok {
		return date, true, nil
	}

	config, err := img.ConfigFile()
	if err != nil {
		return time.Time{}, false, err
	}

	return config.Created.Time, isDated(config.Created.Time), nil
}

func annotatedDate(annotations map[string]string) (time.Time, bool) {
	value, found := annotations[createdAnnotation]
	if !found {
		return time.Time{}, false
	}

	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		log.Debugf("ignoring the annotation %s=%q: %s", createdAnnotation, value, err)
		return time.Time{}, false
	}

	return date, isDated(date)
}

func isDated(date time.Time) bool {
	return date.Unix() > 0
}

// tagMatchingAge returns the tag selected by the version filter among the ones matching the age filter.
func (di *DockerImage) tagMatchingAge(tags []string, arch string) (version.Version, error) {
	candidates, err := ageCandidates(&di.versionFilter, tags)
	if err != nil {
		return version.Version{}, err
	}

	var dated, kept []string
	for _, candidate := range candidates {
		date, ok, err := di.releaseDate(candidate, arch)
		if err != nil {
			return version.Version{}, err
		}

		if !ok {
			log.Debugf("ignoring tag %q without a creation date", candidate)
			continue
		}

		dated = append(dated, candidate)

		if di.spec.Age.Matches(date) {
			kept = append(kept, candidate)
			break
		}

		log.Debugf("ignoring tag %q created on %s, outside the age filter", candidate, date)
	}

	if len(dated) == 0 {
		return version.Version{}, fmt.Errorf("%w: none of the %d tags checked of the Docker image %s has one",
			errNoReleaseDate, len(candidates), di.spec.Image)
	}

	return age.Search(&di.versionFilter, dated, kept)
}

// ageCandidates returns the tags accepted by the version filter in order of preference,
// at most maxAgeChecks as each one costs requests.
func ageCandidates(filter *version.Filter, tags []string) ([]string, error) {
	tags = slices.Clone(tags)

	var candidates []string
	for len(candidates) < maxAgeChecks {
		found, err := filter.Search(tags)
		if err != nil {
			// A version filter matching no tag at all is an error, running out of candidates isn't
			if len(candidates) == 0 {
				return nil, err
			}
			break
		}

		candidate := found.GetVersion()
		if candidate == "" {
			if len(candidates) == 0 {
				return nil, fmt.Errorf("no version found matching pattern %q", filter.Pattern)
			}
			break
		}
		candidates = append(candidates, candidate)
		tags = slices.DeleteFunc(tags, func(tag string) bool { return tag == candidate })
	}

	return candidates, nil
}
