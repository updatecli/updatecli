package dockerimage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (di *DockerImage) Source(workingDir string) (string, error) {

	logrus.Debugf(
		"Searching tags for the image %q",
		di.spec.Image,
	)
	tags, err := di.registry.Tags(di.image)
	if err != nil {
		return "", err
	}

	logrus.Debugf("%d Docker image tag(s) found", len(tags))

	for i, t := range tags {
		logrus.Printf("%d\t%q\n", i, t)
	}

	found := false
	tag := ""

	// registry.Tags doesn't filter tags based on architecture
	// which means that we need another api call to show tag information
searchTag:
	for !found {
		di.foundVersion, err = di.versionFilter.Search(tags)
		if err != nil {
			return "", err
		}

		// Todo: validate that result is valid for architecture

		tag = di.foundVersion.ParsedVersion

		digest, err := di.registry.Digest(di.image)
		if err != nil {
			return "", err
		}

		switch digest {
		case "":
			logrus.Debugf("Docker image tag %q, doesn't support architecture %q, looking for another one\n", tag, di.image.Architecture)
			removeTag(tags, tag)
			if len(tags) == 0 {
				tag = ""
				break searchTag
			}
		default:
			found = true
		}

	}

	if len(tag) == 0 {
		logrus.Infof("%s No Docker Image Tag found matching pattern %q", result.FAILURE, di.versionFilter.Pattern)
		return tag, fmt.Errorf("no Docker Image Tag tag found matching pattern %q", di.versionFilter.Pattern)
	} else if len(tag) > 0 {
		logrus.Infof("%s Docker Image Tag %q found matching pattern %q", result.SUCCESS, tag, di.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in Github source")
	}

	return tag, nil
}

func removeTag(tags []string, tag string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	index := 0
	for i, t := range tags {
		if t == tag {
			index = i
			break
		}
	}

	l := make([]string, 0)
	l = append(l, tags[:index]...)

	return append(l, tags[index+1:]...)
}
