package docker

import (
	"fmt"
)

func (d *Docker) isQuaiIO() bool {

	hostname, _, err := parseImage(d.Image)

	if err != nil {
		fmt.Println(err)
	}

	if hostname == "quay.io" {
		return true
	}
	return false
}
