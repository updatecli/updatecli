package docker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image    string
	Tag      string
	Token    string
	Username string
	Password string
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

// IsDockerRegistry validates that we are on docker registry api
// https://docs.docker.com/registry/spec/api/#api-version-check
func (d *Docker) IsDockerRegistry() (bool, error) {

	errs := d.Validate()

	if len(errs) > 0 {
		for _, err := range errs {
			logrus.Errorln(err)
		}
		return false, errors.New("error found in docker parameters")
	}

	hostname, _, err := parseImage(d.Image)

	if err != nil {
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

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

// Validate ensure parameters are set
func (d *Docker) Validate() (errs []error) {

	if len(d.Username) > 0 && len(d.Password) > 0 {
		token := base64.StdEncoding.EncodeToString([]byte(d.Username + ":" + d.Password))
		if len(d.Token) > 0 {
			logrus.Warningf("Token overridden by the new one generated from username/password")
		}
		d.Token = token
	}

	if len(d.Username) > 0 && len(d.Password) == 0 {
		errs = append(errs, errors.New("Docker registry username provided but not the password"))
	} else if len(d.Username) == 0 && len(d.Password) > 0 {
		errs = append(errs, errors.New("Docker registry password provided but not the username"))
	}

	if len(d.Image) == 0 {
		errs = append(errs, errors.New("Docker image name required"))
	}
	if len(d.Tag) == 0 {
		d.Tag = "latest"
	}

	return errs
}
