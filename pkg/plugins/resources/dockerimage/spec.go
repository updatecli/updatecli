package dockerimage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"dockerimage" defines the specification for retrieving container image tags from a registry.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "architectures" defines the list of platforms the container image must exist for.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "architecture" and "architectures" are mutually exclusive.
	//   * a platform is written "<architecture>" or "<os>/<architecture>[/<variant>]".
	//     The os defaults to "linux".
	//   * a source only checks the first platform of the list.
	//   * a condition checks every platform of the list.
	//   * when unset, Updatecli checks the image index, which applies to every platform.
	//     When set, Updatecli checks the image for that specific platform.
	//     More information on https://github.com/updatecli/updatecli/issues/1603
	//
	// example:
	// ```
	//   architectures:
	//     - windows/amd64
	//     - linux/arm64
	//     - linux/arm64/v8
	// ```
	//
	Architectures []string `yaml:",omitempty"`
	// "architecture" defines the platform the container image must exist for.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "architecture" and "architectures" are mutually exclusive.
	//   * a platform is written "<architecture>" or "<os>/<architecture>[/<variant>]".
	//     The os defaults to "linux".
	//   * when unset, Updatecli checks the image index, which applies to every platform.
	//     When set, Updatecli checks the image for that specific platform.
	//     More information on https://github.com/updatecli/updatecli/issues/1603
	//
	// example:
	//   * architecture: amd64
	//   * architecture: windows/amd64
	//   * architecture: linux/arm64/v8
	//
	Architecture string `yaml:",omitempty"`
	// "image" defines the container image name.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * image: updatecli/updatecli
	//   * image: ghcr.io/updatecli/updatecli
	//
	Image string `yaml:",omitempty"`
	// "tag" defines the container image tag to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	// example:
	//   * tag: latest
	//   * tag: v0.1.0
	//
	Tag                   string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// "versionfilter" defines the version pattern and its kind, such as regex, semver or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "tagfilter" defines a regular expression restricting the tags retrieved from the registry.
	//
	// compatible:
	//   * source
	//
	// remark:
	//   * it is applied before "versionfilter".
	//
	// example:
	//   * tagfilter: ^v\d*(\.\d*){2}-alpine$
	//
	TagFilter string `yaml:",omitempty"`
	// "age" defines how old a tag must be to be selected by a source or accepted by a condition.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * a tag is dated by its "org.opencontainers.image.created" annotation, from the image index
	//     then the image manifest, or else by the "created" field of the image configuration.
	//   * this date is recorded by the image publisher when the image is built, not by the registry
	//     when the tag is pushed, as registries don't report push dates. An image built long before
	//     its release, such as one promoted or retagged, or built with SOURCE_DATE_EPOCH set to a
	//     commit date, looks older than it is and can pass a "minimum" age on its first day.
	//   * as the publisher sets this date, the age filter can't protect against a malicious image
	//     claiming an older one, unlike with registries recording publish dates such as npm or PyPI.
	//   * a tag is dated for every platform of "architectures", and must exist and match the age
	//     filter for all of them. Without one, an image index is dated by its linux/amd64 image,
	//     or else its first image.
	//   * a tag without a creation date, or one set to 1970-01-01 by a reproducible build,
	//     is never selected. Such images, like gcr.io/distroless ones, can't be filtered by age.
	//   * a source needs a version filter ordering tags, such as the kind "semver", as the
	//     registry lists tags alphabetically: the kind "latest" with the pattern "latest" is refused.
	//   * dating a tag costs requests to the registry, so a source checks at most 20 tags
	//     in order of preference. A tag missing, missing for a platform, or whose image the age
	//     filter doesn't support, such as a schema1 image, is passed over.
	//   * a source is skipped when no tag matches the age filter yet, when none of the tags
	//     checked has a creation date or is supported, or when the registry rate limits requests
	//     or fails to answer for a tag after answering for a preferred one. It fails when the
	//     registry fails to answer for the first tag.
	//
	// example:
	// ```
	//   versionfilter:
	//     kind: semver
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
}

func sanitizeRegistryEndpoint(repository string) string {
	if repository == "" {
		return ""
	}
	ref, err := name.ParseReference(repository)
	if err != nil {
		log.Debugf("Unable to parse repository %q: %v", repository, err)
		return ""
	}
	return ref.Context().RegistryStr()
}

// RegistryCredentials returns the credentials from auths for the registry hosting image,
// and whether some were found.
func RegistryCredentials(image string, auths map[string]docker.InlineKeyChain) (docker.InlineKeyChain, bool) {
	credential, found := auths[sanitizeRegistryEndpoint(image)]
	return credential, found
}

// NewDockerImageSpecFromImage return a new docker image specification using an image provided as parameter
func NewDockerImageSpecFromImage(image, tag string, auths map[string]docker.InlineKeyChain) *Spec {
	tagFilter, err := getTagFilterFromValue(tag)
	if err != nil {
		// We couldn't identify a good versionFilter so we do not return any dockerimage spec
		// At the time of writing, semantic versioning is the only way to have reliable results
		// across the different registries.
		// More information on https://github.com/updatecli/updatecli/issues/977
		log.Debugf("analyzing OCI image %q: %s", image, err)
		return nil
	}

	dockerimagespec := Spec{
		Image:     image,
		TagFilter: tagFilter,
		VersionFilter: version.Filter{
			Kind:    version.SEMVERVERSIONKIND,
			Pattern: ">=" + tag,
		},
	}

	registry := sanitizeRegistryEndpoint(image)

	credential, found := RegistryCredentials(image, auths)
	switch found {
	case true:
		if credential.Password != "" {
			dockerimagespec.Password = credential.Password
		}
		if credential.Token != "" {
			dockerimagespec.Token = credential.Token
		}
		if credential.Username != "" {
			dockerimagespec.Username = credential.Username
		}
	default:

		registryAuths := []string{}

		for endpoint := range auths {
			log.Printf("Endpoint:\t%q\n", endpoint)
			registryAuths = append(registryAuths, endpoint)
		}

		warningMessage := fmt.Sprintf(
			"relying on default docker configuration file as no credentials have been found for docker registry %q hosting image %q, among %q",
			registry,
			image,
			strings.Join(registryAuths, ","))

		if len(registryAuths) == 0 {
			warningMessage = fmt.Sprintf("relying on default docker configuration file as no credentials have been found for docker registry %q hosting image %q",
				registry,
				image)
		}

		log.Debug(warningMessage)
	}

	return &dockerimagespec
}

// getTagFilterFromValue tries to identify the closest tagFilter based on an existing tag
func getTagFilterFromValue(tag string) (string, error) {
	log.Debugf("Trying the identify the best versionFilter for %q", tag)

	switch tag {
	case "latest":
		return "", fmt.Errorf("tag latest means nothing to me")
	case "":
		return "", fmt.Errorf("no tag specified")
	}

	patterns := []struct {
		rule    string
		newRule string
	}{
		{
			rule: `^v\d*(\.\d*){2}$`,
		},
		{
			rule: `^\d*(\.\d*){2}$`,
		},
		{
			rule: `^v\d*(\.\d*){1}$`,
		},
		{
			rule: `^\d*(\.\d*){1}$`,
		},
		{
			rule: `^v\d*$`,
		},
		{
			rule: `^\d*$`,
		},
		{
			rule:    `^v(\d*){1}(\.\d*){2}([+-].*){1}$`,
			newRule: `^v\d*(\.\d*){2}`,
		},
		{
			rule:    `^(\d*){1}(\.\d*){2}([+-].*){1}$`,
			newRule: `^\d*(\.\d*){2}`,
		},
		{
			rule:    `^v(\d*){1}(\.\d*){1}([+-].*){1}$`,
			newRule: `^v\d*(\.\d*){1}`,
		},
		{
			rule:    `^(\d*){1}(\.\d*){1}([+-].*){1}$`,
			newRule: `^\d*(\.\d*){1}`,
		},
		{
			rule:    `^v(\d*){1}([+-].*)$`,
			newRule: `^v\d*`,
		},
		{
			rule:    `^(\d*){1}([+-].*)$`,
			newRule: `^\d*`,
		},
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern.rule)
		if err != nil {
			return "", fmt.Errorf("something went wrong with tag regex - %s", err)
		}

		if re.MatchString(tag) {
			submatch := re.FindStringSubmatch(tag)

			newRule := pattern.rule
			if pattern.newRule != "" {
				newRule = pattern.newRule + submatch[len(submatch)-1] + "$"
			}

			log.Debugf("=> closest regex %q identify for value %q", newRule, tag)
			return newRule, nil
		}
	}

	log.Debugf("=> no matching rule identified for Docker image tag %q, feel free to ignore this image with a manifest or to suggest a new rule on https://github.com/updatecli/updatecli/issues/new/choose", tag)
	return "", fmt.Errorf("no tag pattern identify")
}

// ParseOCIReferenceInfo returns the OCI name, tag and digest from an OCI reference
func ParseOCIReferenceInfo(reference string) (ociName, ociTag, ociDigest string, err error) {

	iArray := strings.Split(reference, "@")
	switch len(iArray) {
	case 2:
		ociName = iArray[0]
		ociDigest = "@" + iArray[1]
	case 1:
		ociName = iArray[0]
	}

	imageArray := strings.Split(ociName, ":")
	// Get container image name and tag
	switch len(imageArray) {
	case 2:
		ociName = imageArray[0]
		ociTag = imageArray[1]
	case 1:
		ociName = imageArray[0]
	default:
		ociTag = imageArray[len(imageArray)-1]
		ociName = strings.Join(imageArray[0:len(imageArray)-1], ":")
	}

	if ociDigest == "" && ociTag == "" {
		ociTag = "latest"
	}

	return ociName, ociTag, ociDigest, nil
}
