package dockerhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Docker contains various information to interact with a dockerhub registry
type Docker struct {
	Image        string
	Tag          string
	Architecture string
	Token        string
}

// Digest retrieve docker image tag digest from Dockerhub
func (d *Docker) Digest() (string, error) {

	URL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s/",
		d.Image,
		d.Tag)

	architecture := "amd64"

	if d.Architecture != "" {
		architecture = d.Architecture
	}

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	if len(d.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", nil
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	type response struct {
		ID     string
		Images []map[string]string
	}

	data := response{}

	json.Unmarshal(body, &data)

	for _, image := range data.Images {
		if image["architecture"] == architecture {
			digest := strings.TrimPrefix(image["digest"], "sha256:")
			return digest, nil
		}
	}

	return "", nil

}
