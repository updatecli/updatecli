package dockerimage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	createdAnnotation = "org.opencontainers.image.created"
)

var (
	// defaultPlatform is the platform dating an image index when none is requested, as it's the
	// one the registry client resolves by default. Without it, the first image is used.
	defaultPlatform = v1.Platform{OS: "linux", Architecture: "amd64"}
	// errNoReleaseDate reports that none of the tags checked has a creation date. The source is
	// skipped rather than failed, as when the age filter discards every tag.
	errNoReleaseDate = errors.New("no release date")
	// errTagNotDated reports a tag the registry failed to date while it answered for a preferred one,
	// or while it rate limits requests. The source is skipped rather than falling back to a lower
	// tag, which a registry failing now and then would otherwise select one run out of two.
	errTagNotDated = errors.New("tag not dated")
	// errUnsupportedImage reports an image the age filter can't date, such as a schema1 image. It is
	// the only error about the image itself rather than the registry, see isRegistryError.
	errUnsupportedImage = errors.New("unsupported image")
	// errImageNotReadable reports that none of the tags checked is supported by the age filter, such
	// as schema1 images. As for tags without a creation date, the source is skipped rather than failed.
	errImageNotReadable = errors.New("image not readable")
	// errAgeNeedsOrderedFilter reports a version filter that can't rank tags, so that the age filter
	// has no order of preference to fall back on.
	errAgeNeedsOrderedFilter = errors.New(`the age filter needs a version filter ordering tags, such as the kind "semver": the registry lists tags alphabetically, not by date`)
)

// tagAge is the result of dating a tag for every requested platform.
type tagAge struct {
	// found is false when the tag is missing, or missing for one of the platforms.
	found bool
	// dated is false when the tag has no creation date for one of the platforms.
	dated bool
	// matches is true when the tag matches the age filter for every platform.
	matches bool
	// reason describes, for messages, why a tag doesn't match the age filter.
	reason string
}

// tagAge dates a tag for every requested platform, so that a tag selected by a source is also
// accepted by a condition checking the same platforms. The tag is fetched once, then each platform
// image is dated from it. options are the registry options, see ageOptions.
func (di *DockerImage) tagAge(tag string, options []remote.Option) (tagAge, error) {
	ref, err := di.createRef(tag)
	if err != nil {
		return tagAge{}, err
	}

	descriptor, err := remote.Get(ref, options...)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return tagAge{reason: fmt.Sprintf("docker image %s not found", ref.Name())}, nil
		}
		return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
	}

	switch {
	case descriptor.MediaType.IsIndex():
		index, err := descriptor.ImageIndex()
		if err != nil {
			return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
		}
		manifest, err := index.IndexManifest()
		if err != nil {
			return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
		}

		indexDate, indexDated := annotatedDate(manifest.Annotations)

		for _, arch := range di.architectures() {
			image := ref.Name()
			if arch != "" {
				image += " (" + arch + ")"
			}

			child, found := platformManifest(manifest.Manifests, arch)
			if !found {
				return tagAge{reason: fmt.Sprintf("docker image %s not found", image)}, nil
			}

			date, dated := indexDate, indexDated
			if !dated {
				img, err := index.Image(child.Digest)
				if err == nil {
					date, dated, err = imageDate(img)
				}
				switch {
				case err != nil && isStatus(err, http.StatusNotFound):
					// The index lists an image the registry no longer holds
					return tagAge{reason: fmt.Sprintf("docker image %s not found, the index lists %s which is missing", image, child.Digest)}, nil
				case err != nil:
					return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", image, err)
				}
			}

			if got := di.checkAge(image, date, dated); !got.matches {
				return got, nil
			}
		}

		return tagAge{found: true, dated: true, matches: true}, nil

	case descriptor.MediaType == types.DockerManifestSchema1 || descriptor.MediaType == types.DockerManifestSchema1Signed:
		return tagAge{}, fmt.Errorf("%w: the age filter isn't supported with MediaType %q", errUnsupportedImage, descriptor.MediaType)

	default:
		img, err := descriptor.Image()
		if err != nil {
			return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
		}
		date, dated, err := imageDate(img)
		if err != nil {
			return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
		}

		if len(di.spec.Architectures) > 0 {
			config, err := img.ConfigFile()
			if err != nil {
				return tagAge{}, fmt.Errorf("retrieving the Docker image %s: %w", ref.Name(), err)
			}
			for _, arch := range di.spec.Architectures {
				if !configSatisfies(config, arch) {
					return tagAge{reason: fmt.Sprintf("docker image %s (%s) not found, the image is built for %s", ref.Name(), arch, config.Platform())}, nil
				}
			}
		}

		return di.checkAge(ref.Name(), date, dated), nil
	}
}

// configSatisfies returns true if a single-platform image, described by its configuration, is
// built for a platform. Only what both sides state is compared: an image whose configuration names
// no architecture or variant is accepted, and the operating system is only compared when the
// platform names one, as parsePlatform otherwise assumes linux.
func configSatisfies(config *v1.ConfigFile, arch string) bool {
	if config.Architecture == "" {
		return true
	}

	want := parsePlatform(arch)

	switch {
	case config.Architecture != want.Architecture:
		return false
	case strings.Contains(arch, "/") && config.OS != "" && config.OS != want.OS:
		return false
	case want.Variant != "" && config.Variant != "" && config.Variant != want.Variant:
		return false
	}

	return true
}

// checkAge returns the result of a found image created on date against the age filter.
func (di *DockerImage) checkAge(image string, date time.Time, dated bool) tagAge {
	switch {
	case !dated:
		return tagAge{found: true, reason: fmt.Sprintf("docker image %s has no creation date, so the age filter can't be applied", image)}
	case !di.spec.Age.Matches(date):
		return tagAge{found: true, dated: true, reason: fmt.Sprintf("docker image %s created on %s is outside the age filter", image, date.Format(time.RFC3339))}
	default:
		return tagAge{found: true, dated: true, matches: true}
	}
}

// imageDate returns the creation date of an image, from its manifest annotation or else its
// configuration. It returns false when the image has none, or one set to the Unix epoch as
// reproducible builds do.
func imageDate(img v1.Image) (time.Time, bool, error) {
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

// platformManifest returns the image of an index for a platform. Without a platform, it returns the
// linux/amd64 image, or else the first image that isn't an attestation, which registries list with
// the platform "unknown/unknown".
func platformManifest(manifests []v1.Descriptor, arch string) (v1.Descriptor, bool) {
	var first *v1.Descriptor

	for i, manifest := range manifests {
		if !manifest.MediaType.IsImage() {
			continue
		}

		if arch != "" {
			if manifest.Platform != nil && manifest.Platform.Satisfies(parsePlatform(arch)) {
				return manifest, true
			}
			continue
		}

		if manifest.Platform != nil && manifest.Platform.Satisfies(defaultPlatform) {
			return manifest, true
		}
		if first == nil && (manifest.Platform == nil || manifest.Platform.OS != "unknown") {
			first = &manifests[i]
		}
	}

	if first != nil {
		return *first, true
	}

	return v1.Descriptor{}, false
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

// isRegistryError returns true if err may come from the registry or the network, so that it may
// not happen on the next run. Only an image the age filter doesn't support is known to fail on
// every run: anything else, such as a dropped connection, a truncated manifest or a blob failing its
// digest check, is counted as a registry failure, rather than passing the tag over to a lower one.
func isRegistryError(err error) bool {
	return !errors.Is(err, errUnsupportedImage)
}

// architectures returns the platforms to check, a single empty one when none is requested.
func (di *DockerImage) architectures() []string {
	if len(di.spec.Architectures) == 0 {
		return []string{""}
	}
	return di.spec.Architectures
}

// tagMatchingAge returns the first tag, in the order of preference of the version filter, matching
// the age filter for every requested platform.
//
// A tag missing, missing for a platform, or whose image the age filter doesn't support is passed over. A registry
// failing to answer for a tag stops the search: the source is skipped with errTagNotDated when the
// registry answered for a preferred tag or rate limits requests, and fails otherwise, as the
// registry then fails for every tag.
func (di *DockerImage) tagMatchingAge(ctx context.Context, tags []string) (version.Version, error) {
	candidates, err := di.versionFilter.Rank(tags, age.MaxChecks)
	if err != nil {
		return version.Version{}, err
	}

	options, err := di.ageOptions(ctx)
	if err != nil {
		return version.Version{}, err
	}

	// answered is set once the registry answers for a tag, even to report it missing. found is set
	// once a tag is found, discarded once the age filter discards a dated one. unreadable holds the
	// last error of an image that couldn't be read.
	answered, found, discarded := false, false, false
	var unreadable error

	// Tags are dated one at a time, in order of preference: the first one is usually the one
	// selected, and registries such as Docker Hub count every manifest request against a rate limit.
	for _, candidate := range candidates {
		tag := candidate.GetVersion()

		got, err := di.tagAge(tag, options)

		switch {
		case err != nil && !isRegistryError(err):
			// The age filter doesn't support this image, the next tag may be supported
			log.Warningf("ignoring tag %q of the Docker image %s: %s", tag, di.spec.Image, err)
			answered = true
			unreadable = err
			continue
		case err != nil && (answered || isStatus(err, http.StatusTooManyRequests)):
			return version.Version{}, fmt.Errorf("%w: the tag %q of the Docker image %s couldn't be dated: %s", errTagNotDated, tag, di.spec.Image, err)
		case err != nil:
			return version.Version{}, fmt.Errorf("dating the tag %q: %w", tag, err)
		case got.matches:
			return candidate, nil
		}

		log.Debugf("ignoring tag %q: %s", tag, got.reason)
		answered = true
		found = found || got.found
		discarded = discarded || got.dated
	}

	switch {
	case discarded:
		return version.Version{}, fmt.Errorf("%w for the Docker image %s", age.ErrNoVersionMatchingAge, di.spec.Image)
	case found:
		return version.Version{}, fmt.Errorf("%w: none of the %d tags checked of the Docker image %s has one",
			errNoReleaseDate, len(candidates), di.spec.Image)
	case unreadable != nil:
		return version.Version{}, fmt.Errorf("%w: none of the %d tags checked of the Docker image %s is supported, the last error being: %s",
			errImageNotReadable, len(candidates), di.spec.Image, unreadable)
	default:
		return version.Version{}, fmt.Errorf("none of the %d tags checked of the Docker image %s exists for %s",
			len(candidates), di.spec.Image, describeArchitectures(di.spec.Architectures))
	}
}

// ageOptions returns the registry options to date tags, sharing one puller so that the registry
// is pinged and authenticated once rather than for every tag. The requests stop once ctx is done.
func (di *DockerImage) ageOptions(ctx context.Context) ([]remote.Option, error) {
	options := append(slices.Clone(di.options), remote.WithContext(ctx))

	puller, err := remote.NewPuller(options...)
	if err != nil {
		return nil, fmt.Errorf("connecting to the registry of the Docker image %s: %w", di.spec.Image, err)
	}

	return append(options, remote.Reuse(puller)), nil
}

// checkAgeFilter returns an error if the version filter can't rank tags for the age filter. It only
// applies to a source, as a condition checks a given tag.
func (di *DockerImage) checkAgeFilter() error {
	if di.versionFilter.Kind == version.LATESTVERSIONKIND && di.versionFilter.Pattern == version.LATESTVERSIONKIND {
		return errAgeNeedsOrderedFilter
	}
	return nil
}

// describeArchitectures returns the requested platforms, for messages.
func describeArchitectures(architectures []string) string {
	if len(architectures) == 0 {
		return "any platform"
	}
	return "the platforms " + strings.Join(architectures, ", ")
}
