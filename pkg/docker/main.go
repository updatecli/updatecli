package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image        string
	Tag          string
	URL          string
	Architecture string
}

// Check verify if Docker parameters are correctly set
func (d *Docker) Check() (bool, error) {
	if d.Image == "" {
		err := fmt.Errorf("Docker Image is required")
		return false, err
	}

	if d.URL == "" {
		d.URL = "hub.docker.com"
	}

	if d.Tag == "" {
		d.Tag = "latest"
	}

	if d.Architecture == "" {
		d.Architecture = "amd64"
	}

	if image := strings.Split(d.Image, "/"); len(image) == 1 {
		d.Image = "library/" + d.Image
	}

	return true, nil
}

// IsTagPublished checks if a docker image with a specific tag is published
func (d *Docker) IsTagPublished() bool {

	if ok, err := d.Check(); !ok {
		fmt.Println(err)
		return ok
	}

	url := fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s",
		d.URL,
		d.Image,
		d.Tag)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}

	data := map[string]string{}

	json.Unmarshal(body, &data)

	if val, ok := data["message"]; ok && strings.Contains(val, "not found") {
		fmt.Printf("\u2717 %s:%s doesn't exist on the Docker Registry \n", d.Image, d.Tag)
		return false
	}

	if val, ok := data["name"]; ok && val == d.Tag {
		fmt.Printf("\u2714 %s:%s available on the Docker Registry\n", d.Image, d.Tag)
		return true
	}

	fmt.Printf("\t\t\u2717Something went wrong, no field 'name' founded from %s\n", url)

	return false
}

// Source retrieve docker image tag digest from a registry
func (d *Docker) Source() (string, error) {

	if ok, err := d.Check(); !ok {
		return "", err
	}

	// https://hub.docker.com/v2/repositories/olblak/updatecli/tags/latest
	URL := fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s",
		d.URL,
		d.Image,
		d.Tag)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
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
			fmt.Printf("Example: %v@sha256%v\n", d.Image, digest)
			return digest, nil
		}
	}

	fmt.Printf("\u2717 No Digest found for docker image %s:%s on the Docker Registry \n", d.Image, d.Tag)

	return "", nil
}
