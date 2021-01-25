package dockerregistry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image    string
	Tag      string
	Hostname string
	Token    string
}

// Digest retrieve docker image tag digest from a registry
func (d *Docker) Digest() (string, error) {
	type error struct {
		Code    string
		Message string
		Detail  string
	}

	type response struct {
		MediaType     string
		SchemaVersion int
		Errors        []error
	}

	URL := fmt.Sprintf("https://%s/v2/%s/manifests/%s",
		d.Hostname,
		d.Image,
		d.Tag)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	if len(d.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	data := response{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	if len(data.Errors) > 0 {
		e := fmt.Errorf("%s:%s", d.Image, d.Tag)
		for _, err := range data.Errors {
			e = fmt.Errorf("%s - %s", e, err.Message)
		}
		return "", e
	}

	if len(data.Errors) > 0 {
		e := fmt.Errorf("%s:%s", d.Image, d.Tag)
		for _, err := range data.Errors {
			e = fmt.Errorf("%s - %s", e, err.Message)
		}
		return "", e
	}

	digest := res.Header.Get("Docker-Content-Digest")
	digest = strings.TrimPrefix(digest, "sha256:")

	return digest, nil
}
