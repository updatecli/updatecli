package dockerhub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/olblak/updateCli/pkg/core/helpers"
)

// Docker contains various information to interact with a dockerhub registry
type Docker struct {
	Image        string
	Tag          string
	Architecture string
	Token        string
	Client       helpers.HttpClient
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

	URL := "https://hub.docker.com/v2/users/login/"

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer([]byte(authentication)))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := d.Client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode == 403 {
		return "", fmt.Errorf("incorrect authentication credentials")
	} else if res.StatusCode != 200 {
		return "", fmt.Errorf("something went wrong while login to dockerhub")
	}

	type response struct {
		Token string
	}

	data := response{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

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

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
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

	type images struct {
		Architecture string
		Digest       string
	}

	type response struct {
		ID     int
		Images []images
	}

	data := response{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	for _, image := range data.Images {
		if image.Architecture == architecture {
			digest := strings.TrimPrefix(image.Digest, "sha256:")
			return digest, nil
		}
	}

	return "", nil
}
