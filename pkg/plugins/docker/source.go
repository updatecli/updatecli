package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Source retrieve docker image tag digest from a registry
func (d *Docker) Source() (string, error) {

	if ok, err := d.Check(); !ok {
		return "", err
	}

	// https://hub.docker.com/v2/repositories/olblak/updatecli/tags/latest
	URL := ""

	if d.isDockerHub() {
		URL = fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s/",
			d.URL,
			d.Image,
			d.Tag)

	} else {
		if ok, err := d.IsDockerRegistry(); !ok {
			return "", err
		}
		URL = fmt.Sprintf("https://%s/v2/%s/manifests/%s",
			d.URL,
			d.Image,
			d.Tag)
	}

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	if ok, err := d.IsDockerRegistry(); ok && err == nil {
		// Retrieve v2 manifest
		// application/vnd.docker.distribution.manifest.v1+prettyjws v1 manifest
		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if d.isDockerHub() {

		type respond struct {
			ID     string
			Images []map[string]string
		}

		data := respond{}

		json.Unmarshal(body, &data)

		for _, image := range data.Images {
			if image["architecture"] == d.Architecture {
				digest := strings.TrimPrefix(image["digest"], "sha256:")
				fmt.Printf("\u2714 Digest '%v' found for docker image %s:%s available from Docker Registry\n", digest, d.Image, d.Tag)
				fmt.Printf("\nRemark: Do not forget to add @sha256 after your the docker image name\n")
				fmt.Printf("Example: %v@sha256:%v\n", d.Image, digest)
				return digest, nil
			}
		}

		fmt.Printf("\u2717 No Digest found for docker image %s:%s on the Docker Registry \n", d.Image, d.Tag)

		return "", nil
	}

	digest := res.Header.Get("Docker-Content-Digest")
	digest = strings.TrimPrefix(digest, "sha256:")

	fmt.Printf("\u2714 Digest '%v' found for docker image %s:%s available from Docker Registry\n", digest, d.Image, d.Tag)
	fmt.Printf("\nRemark: Do not forget to add @sha256 after your the docker image name\n")
	fmt.Printf("Example: %v/%v@sha256:%v\n", d.URL, d.Image, digest)

	return digest, nil

}
