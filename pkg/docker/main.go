package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Docker contains various information to check if a docker image exists
type Docker struct {
	Image string
	Tag   string
	URL   string
}

// IsTagPublished check if a docker image with a specific tag is published
func (docker *Docker) IsTagPublished() bool {

	url := fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s",
		docker.URL,
		docker.Image,
		docker.Tag)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
	}

	data := map[string]string{}

	json.Unmarshal(body, &data)

	if val, ok := data["message"]; ok && strings.Contains(val, "not found") {
		log.Printf("\u2717 %s:%s doesn't exist on the Docker Registry \n", docker.Image, docker.Tag)
		return false
	}

	if val, ok := data["name"]; ok && val == docker.Tag {
		log.Printf("\u2714 %s:%s available on the Docker Registry\n", docker.Image, docker.Tag)
		return true
	}

	log.Printf("Something went wrong, no field 'name' founded from %s\n", url)

	return false
}
