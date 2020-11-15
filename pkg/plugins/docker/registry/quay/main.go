package quay

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
}

// Digest retrieve docker image tag digest from a registry
func (d *Docker) Digest() (string, error) {

	URL := fmt.Sprintf("https://quay.io/api/v1/repository/%s", d.Image)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	if d.Architecture != "" {
		fmt.Printf("Quay.io doesn't support arch '%v', fallback to amd64\n", d.Architecture)
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

	type tagMetadata struct {
		ImageID        string `json:"image_id"`
		LastModified   string `json:"last_modified"`
		Name           string `json:"name"`
		ManifestDigest string `json:"manifest_digest"`
		Size           string `json:"size"`
	}

	type response struct {
		Description string
		Name        string
		Namespace   string
		Tags        map[string]tagMetadata
	}

	data := response{}

	json.Unmarshal(body, &data)

	if tag, ok := data.Tags[d.Tag]; ok {
		digest := strings.TrimLeft(tag.ManifestDigest, "sha256:")

		return digest, nil
	}
	return "", nil

}
