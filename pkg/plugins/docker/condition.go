package docker

import (
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/dockerhub"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/dockerregistry"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/ghcr"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/quay"
)

// ConditionFromSCM returns an error because it's not supported
func (d *Docker) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for dockerRegistry condition, aborting")
}

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (d *Docker) Condition(source string) (bool, error) {

	hostname, image, err := parseImage(d.Image)

	if err != nil {
		return false, err
	}

	if d.Tag != "" {
		logrus.Infof("Tag %v, defined from configuration file which override the source value '%v'", d.Tag, source)
	} else {
		d.Tag = source
	}

	if ok, err := d.Check(); !ok {
		return false, err
	}

	var r Registry

	if d.isDockerHub() {
		dh := dockerhub.Docker{
			Image: image,
			Tag:   d.Tag,
			Token: d.Token,
		}

		r = &dh

	} else if d.isQuaiIO() {

		q := quay.Docker{
			Image: image,
			Tag:   d.Tag,
			Token: d.Token,
		}

		r = &q

	} else if d.isGHCR() {

		g := ghcr.Docker{
			Image: image,
			Tag:   d.Tag,
			Token: d.Token,
		}

		r = &g

	} else if ok, err := d.IsDockerRegistry(); ok {
		if err != nil {
			return false, err
		}

		dr := dockerregistry.Docker{
			Image:    image,
			Tag:      d.Tag,
			Hostname: hostname,
			Token:    d.Token,
		}

		r = &dr

	} else {
		return false, fmt.Errorf("unknown docker registry api")
	}

	digest, err := r.Digest()

	if err != nil {
		return false, err
	}

	if digest == "" {
		logrus.Infof("\u2717 %s:%s doesn't exist on the Docker Registry", d.Image, d.Tag)
		return false, nil
	}

	logrus.Infof("\u2714 %s:%s available on the Docker Registry", d.Image, d.Tag)

	return true, nil

}
