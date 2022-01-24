package mavenmetadata

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// DefaultHandler is the default implementation for a maven metadata handler
type DefaultHandler struct {
	metadataURL string
	webClient   httpclient.HTTPClient
}

// New returns a newly initialized DefaultHandler object
func New(metadataURL string) *DefaultHandler {
	return &DefaultHandler{
		metadataURL: metadataURL,
		webClient:   http.DefaultClient,
	}
}

// getMetadataFile is an internal method that returns the parsed metadata object
func (d *DefaultHandler) getMetadataFile() (metadata, error) {
	req, err := http.NewRequest("GET", d.metadataURL, nil)
	if err != nil {
		return metadata{}, err
	}

	logrus.Debugf("Sending HTTP request to the maven repository at %s", d.metadataURL)
	res, err := d.webClient.Do(req)
	if err != nil {
		return metadata{}, err
	}
	defer res.Body.Close()

	// If HTTP code in 4xx or 5xx, then it's an error
	if res.StatusCode >= 400 {
		return metadata{}, fmt.Errorf("HTTP error returned from %s: %v", d.metadataURL, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return metadata{}, err
	}

	logrus.Debugf("Received the following response (HTTP status %d):\n%s", res.StatusCode, body)

	data := metadata{}

	err = xml.Unmarshal(body, &data)
	if err != nil {
		return metadata{}, err
	}

	return data, nil
}

func (d *DefaultHandler) GetLatestVersion() (string, error) {
	data, err := d.getMetadataFile()
	if err != nil {
		return "", err
	}

	if data.Versioning.Latest == "" {
		return "", fmt.Errorf("%s No latest version found at %s", result.FAILURE, d.metadataURL)
	}
	return data.Versioning.Latest, nil
}

func (d *DefaultHandler) GetVersions() ([]string, error) {
	data, err := d.getMetadataFile()
	if err != nil {
		return []string{}, err
	}

	versions := []string{}
	versions = append(versions, data.Versioning.Versions.Version...)

	return versions, nil
}

func (d *DefaultHandler) GetMetadataURL() string {
	return d.metadataURL
}
