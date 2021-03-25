package docker

import (
	"fmt"
	"github.com/olblak/updateCli/pkg/core/helpers"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image string
	Tag   string
	Token string
	client helpers.HttpClient
}

// Registry is an interface for every docker registry api
type Registry interface {
	Digest() (string, error)
}

func parseImage(name string) (hostname string, image string, err error) {
	URL, err := url.ParseRequestURI("https://" + name)
	if err != nil {
		return "", "", err
	}

	// if hostname doesn't contains valid url with at least a dot
	if s := strings.Split(URL.Hostname(), "."); len(s) <= 1 {
		URL, err = url.Parse("https://hub.docker.com/" + name)
	}

	hostname = URL.Hostname()
	image = strings.TrimPrefix(URL.EscapedPath(), "/")

	if len(strings.Split(image, "/")) == 1 && URL.Hostname() == "hub.docker.com" {
		image = "library" + URL.EscapedPath()
	}

	return hostname, image, nil
}

// Check verify if Docker parameters are correctly set
func (d *Docker) Check() (ok bool, err error) {

	if d.Image == "" {
		err = fmt.Errorf("Docker Image is required")
		return false, err
	}

	if d.Tag == "" {
		d.Tag = "latest"
	}

	return true, nil
}

func (d *Docker) isDockerHub() bool {

	hostname, _, err := parseImage(d.Image)

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	if hostname == "hub.docker.com" || hostname == "docker.io" {
		return true
	}
	return false
}

func (d *Docker) isGHCR() bool {

	hostname, _, err := parseImage(d.Image)

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	if hostname == "ghcr.io" {
		return true
	}
	return false
}

// IsDockerRegistry validates that we are on docker registry api
// https://docs.docker.com/registry/spec/api/#api-version-check
func (d *Docker) IsDockerRegistry() (bool, error) {

	hostname, _, err := parseImage(d.Image)

	if err != nil {
		return false, err
	}

	if ok, err := d.Check(); !ok {
		return false, err
	}

	if d.isDockerHub() {
		return false, fmt.Errorf("DockerHub Api is not docker registry API compliant")
	}

	URL := fmt.Sprintf("https://%s/v2/", hostname)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return false, err
	}

	if len(d.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	}

	res, err := d.client.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}
