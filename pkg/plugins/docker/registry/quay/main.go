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
	Architecture string
	Image        string
	Tag          string
	Token        string
}

// Digest retrieve docker image tag digest from a registry
func (d *Docker) Digest() (string, error) {

	URL := fmt.Sprintf("https://quay.io/api/v1/repository/%s", d.Image)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	if len(d.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 && res.StatusCode < 500 {

		return "", fmt.Errorf("quay.io/%s:%s - doesn't exist on quay.io", d.Image, d.Tag)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	type tagMetadata struct {
		ImageID        string `json:"image_id"`
		LastModified   string `json:"last_modified"`
		Name           string `json:"name"`
		ManifestDigest string `json:"manifest_digest"`
		Size           int    `json:"size"`
	}

	type response struct {
		Title        string
		Description  string
		Name         string
		Namespace    string
		Tags         map[string]tagMetadata
		ErrorMessage string `json:"error_message"`
	}

	data := response{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	if tag, ok := data.Tags[d.Tag]; ok {
		digest := strings.TrimLeft(tag.ManifestDigest, "sha256:")

		return digest, nil
	}
	err = fmt.Errorf("tag doesn't exist for quay.io/%s:%s", d.Image, d.Tag)
	return "", err
}
