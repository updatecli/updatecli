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
	// [S][C] Architecture specifies the container image architecture such as `amd64`
	Architecture string `yaml:",omitempty"`
	// [S][C] Image specifies the container image such as `updatecli/updatecli`
	Image string `yaml:",omitempty"`
	// [C] Tag specifies the container image tag such as `latest`
	Tag                   string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] TagFiter allows to restrict tags retrieved from a remote registry by using a regular expression.
	TagFilter string
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
		logrus.Warningln(err)
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
			"no credentials found for docker registry %q hosting image %q, among %q",
			registry,
			image,
			strings.Join(registryAuths, ","))

		if len(registryAuths) == 0 {
			warningMessage = fmt.Sprintf("no credentials found for docker registry %q hosting image %q",
				registry,
				image)
		}

		logrus.Debug(warningMessage)
	}

	return &dockerimagespec
}

// NewFrilterFromValue tries to identify the closest tagFilter based on an existing tag
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
			rule: `^v?\d*(\.\d*){1}$`,
		},
		{
			rule: `^v?\d*$`,
		},
		{
			rule:    `^v?(\d*){1}(\.\d*){2}([+-].*){1}$`,
			newRule: `^v?\d*(\.\d*){2}`,
		},
		{
			rule:    `^v?(\d*){1}(\.\d*){1}([+-].*){1}$`,
			newRule: `^v?\d*(\.\d*){1}`,
		},
		{
			rule:    `^v?(\d*){1}([+-].*)$`,
			newRule: `^v?\d*`,
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

	logrus.Warningf("=> No matching rule identified for Docker image tag %q, feel free to ignore this image with a manifest or to suggest a new rule on https://github.com/updatecli/updatecli/issues/new/choose", tag)
	return "", fmt.Errorf("no tag pattern identify")
}
