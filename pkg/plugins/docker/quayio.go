package docker

import (
	"github.com/sirupsen/logrus"
)

func (d *Docker) isQuaiIO() bool {

	hostname, _, err := parseImage(d.Image)

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	if hostname == "quay.io" {
		return true
	}
	return false
}
