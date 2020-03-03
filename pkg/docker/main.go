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
	Image string
	Tag   string
	URL   string
}

// IsTagPublished checks if a docker image with a specific tag is published
func (docker *Docker) IsTagPublished() bool {

	url := fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s",
		docker.URL,
		docker.Image,
		docker.Tag)

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
		fmt.Printf("\t\t\u2717\t%s:%s doesn't exist on the Docker Registry \n", docker.Image, docker.Tag)
		return false
	}

	if val, ok := data["name"]; ok && val == docker.Tag {
		fmt.Printf("\t\t\u2714\t%s:%s available on the Docker Registry\n", docker.Image, docker.Tag)
		return true
	}

	fmt.Printf("\t\t\u2717Something went wrong, no field 'name' founded from %s\n", url)

	return false
}
