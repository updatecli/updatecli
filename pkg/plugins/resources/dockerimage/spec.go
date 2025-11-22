package dockerimage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// architectures specifies a list of architectures to check container images for (conditions only)
	//
	// compatible:
	//   * condition
	//   * source
	//
	// example: windows/amd64, linux/arm64, linux/arm64/v8
	//
	// default: linux/amd64
	//
	// remark:
	//   If an architecture is undefined, Updatecli retrieves the digest of the image index
	//   which can be used regardless of the architecture.
	//   But if an architecture is specified then Updatecli retrieves a specific image digest.
	//   More information on https://github.com/updatecli/updatecli/issues/1603
	Architectures []string `yaml:",omitempty"`
	// architecture specifies the container image architecture such as `amd64`
	//
	// compatible:
	//   * condition
	//   * source
	//
	// example: windows/amd64, linux/arm64, linux/arm64/v8
	//
	// default: linux/amd64
	//
	// remark:
	//   If an architecture is undefined, Updatecli retrieves the digest of the image index
	//   which can be used regardless of the architecture.
	//   But if an architecture is specified then Updatecli retrieves a specific image digest.
	//   More information on https://github.com/updatecli/updatecli/issues/1603
	Architecture string `yaml:",omitempty"`
	// image specifies the container image such as `updatecli/updatecli`
	//
	// compatible:
	//   * condition
	//   * source
	Image string `yaml:",omitempty"`
	// tag specifies the container image tag such as `latest`
	//
	// compatible:
	//   * condition
	//
	// default: latest
	Tag                   string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// versionfilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	VersionFilter version.Filter `yaml:",omitempty"`
	// tagfilter allows to restrict tags retrieved from a remote registry by using a regular expression.
	//
	// compatible:
	//   * source
	//
	// example: ^v\d*(\.\d*){2}-alpine$
	//
	// default: none
	TagFilter string `yaml:",omitempty"`
}

func sanitizeRegistryEndpoint(repository string) string {
	if repository == "" {
		return ""
	}
	ref, err := name.ParseReference(repository)
	if err != nil {
		logrus.Debugf("Unable to parse repository %q: %v", repository, err)
	}
	return ref.Context().RegistryStr()
}

// NewDockerImageSpecFromImage return a new docker image specification using an image provided as parameter
func NewDockerImageSpecFromImage(image, tag string, auths map[string]docker.InlineKeyChain) *Spec {
	tagFilter, err := getTagFilterFromValue(tag)
	if err != nil {
		// We couldn't identify a good versionFilter so we do not return any dockerimage spec
		// At the time of writing, semantic versioning is the only way to have reliable results
		// across the different registries.
		// More information on https://github.com/updatecli/updatecli/issues/977
		logrus.Debugf("analyzing OCI image %q: %s", image, err)
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

	credential, found := auths[registry]
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
			logrus.Printf("Endpoint:\t%q\n", endpoint)
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

		logrus.Debug(warningMessage)
	}

	return &dockerimagespec
}

// getTagFilterFromValue tries to identify the closest tagFilter based on an existing tag
func getTagFilterFromValue(tag string) (string, error) {
	logrus.Debugf("Trying the identify the best versionFilter for %q", tag)

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

			logrus.Debugf("=> closest regex %q identify for value %q", newRule, tag)
			return newRule, nil
		}
	}

	logrus.Debugf("=> no matching rule identified for Docker image tag %q, feel free to ignore this image with a manifest or to suggest a new rule on https://github.com/updatecli/updatecli/issues/new/choose", tag)
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
