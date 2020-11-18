package dockerhub

import (
	"bytes"
	"encoding/base64"
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

// Login authenticate with Dockerhub then return a valid bearer token
func (d *Docker) Login() (string, error) {

	decoded, err := base64.StdEncoding.DecodeString(d.Token)
	if err != nil {
		return "", err
	}

	value := strings.SplitAfterN(string(decoded), ":", 2)

	username := strings.TrimSuffix(value[0], ":")
	password := value[1]

	authentication := fmt.Sprintf("{\"username\": \"%s\", \"password\": \"%s\"}", username, password)

	URL := fmt.Sprintf("https://hub.docker.com/v2/users/login/")

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer([]byte(authentication)))

	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if res.StatusCode == 403 {
		return "", fmt.Errorf("Incorrect authentication credentials")
	} else if res.StatusCode != 200 {
		return "", fmt.Errorf("Something went wrong while login to Dockerhub")
	}

	type response struct {
		Token string
	}

	data := response{}

	json.Unmarshal(body, &data)

	return data.Token, nil

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

		token, err := d.Login()

		if err != nil {
			return "", err
		}

		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return "", nil
	}

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
