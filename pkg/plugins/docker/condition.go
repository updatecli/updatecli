package docker

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/plugins/docker/registry/dockerhub"
	"github.com/updatecli/updatecli/pkg/plugins/docker/registry/dockerregistry"
	"github.com/updatecli/updatecli/pkg/plugins/docker/registry/quay"
)

// ConditionFromSCM returns an error because it's not supported
func (d *Docker) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for dockerRegistry condition, aborting")
}

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (d *Docker) Condition(source string) (bool, error) {
	// Init tag based on source information
	if d.Tag != "" {
		logrus.Infof("Tag %v, defined from configuration file which override the source value '%v'", d.Tag, source)
	} else {
		d.Tag = source
	}

	// Validate parameters
	errs := d.Validate()
	if len(errs) > 0 {
		for _, err := range errs {
			logrus.Errorln(err)
		}
		return false, errors.New("error found in docker parameters")
	}

	hostname, image, err := parseImage(d.Image)

	if err != nil {
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
		logrus.Infof("%s %s:%s doesn't exist on the Docker Registry", result.FAILURE, d.Image, d.Tag)
		return false, nil
	}

	logrus.Infof("%s %s:%s available on the Docker Registry", result.SUCCESS, d.Image, d.Tag)

	return true, nil

}
