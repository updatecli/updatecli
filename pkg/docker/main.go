package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image string
	Tag   string
	URL   string
}

// IsTagPublished checks if a docker image with a specific tag is published
func (d *Docker) IsTagPublished() bool {

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

// GetVersion retrieve docker image tag digest from a registry
func (d *Docker) GetVersion() string {

	// https://hub.docker.com/v2/repositories/olblak/updatecli/tags/latest
	URL := fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s",
		d.URL,
		d.Image,
		d.Tag)

	req, err := http.NewRequest("GET", URL, nil)

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

	type respond struct {
		ID     string
		Images []map[string]string
	}

	data := respond{}

	json.Unmarshal(body, &data)

	log.Printf("Data: %v", data)

	for _, image := range data.Images {
		if image["architecture"] == "amd64" {
			digest := image["digest"]
			fmt.Printf("%s:%s digest found is %v", d.Image, d.Tag, digest)
			return digest
		}
	}
	return ""
}
