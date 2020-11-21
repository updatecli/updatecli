package docker

import (
	"fmt"

	"github.com/olblak/updateCli/pkg/plugins/docker/registry/dockerhub"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/dockerregistry"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/ghcr"
	"github.com/olblak/updateCli/pkg/plugins/docker/registry/quay"
)

// Source retrieve docker image tag digest from a registry
func (d *Docker) Source() (string, error) {

	hostname, image, err := parseImage(d.Image)

	if err != nil {
		return "", err
	}

	if ok, err := d.Check(); !ok {
		return "", err
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
			return "", err
		}

		dr := dockerregistry.Docker{
			Image:    image,
			Tag:      d.Tag,
			Hostname: hostname,
			Token:    d.Token,
		}

		r = &dr

	} else {
		return "", fmt.Errorf("Unknown Docker Registry API")
	}

	digest, err := r.Digest()

	if err != nil {
		return "", err
	}

	if digest == "" {
		fmt.Printf("\u2717 No Digest found for docker image %s:%s on the Docker Registry \n", d.Image, d.Tag)
	} else {
		fmt.Printf("\u2714 Digest '%v' found for docker image %s:%s available from Docker Registry\n", digest, d.Image, d.Tag)
		fmt.Printf("\nRemark: Do not forget to add @sha256 after your the docker image name\n")
		fmt.Printf("Example: %v@sha256:%v\n", d.Image, digest)
	}

	return digest, nil
}
