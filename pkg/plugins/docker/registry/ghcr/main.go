package ghcr

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
	Architecture string
	Token        string
}

// Digest retrieve docker image tag digest from a registry
func (d *Docker) Digest() (string, error) {

	URL := fmt.Sprintf("https://ghcr.io/v2/%s/manifests/%s", d.Image, d.Tag)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	if len(d.Architecture) == 0 {
		d.Architecture = "amd64"
	}

	if len(d.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", d.Token))
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

	type platform struct {
		Architecture string
		Os           string
	}

	type manifest struct {
		MediaType string
		Digest    string
		Platform  platform
	}

	type response struct {
		MediaType     string
		SchemaVersion string
		Manifests     []manifest
	}

	data := response{}

	json.Unmarshal(body, &data)

	for _, manifest := range data.Manifests {
		if manifest.Platform.Architecture == d.Architecture {
			digest := strings.TrimLeft(manifest.Digest, "sha256:")
			fmt.Println(digest)
			return digest, nil
		}
	}
	return "", nil

}
