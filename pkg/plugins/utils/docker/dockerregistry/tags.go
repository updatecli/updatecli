package dockerregistry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/link"
)

// Tags retrieves all tags of the provided docker image from the registry
func (dgr DockerGenericRegistry) Tags(image dockerimage.Image) ([]string, error) {
	var tags []string

	endpoints := registryEndpoints(image)

	URL := fmt.Sprintf("https://%s/v2/%s/%s/tags/list",
		endpoints.ApiService,
		image.Namespace,
		image.Repository,
	)

	for URL != "" {

		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			return []string{}, err
		}

		// Retrieve a bearer token to authenticate further requests
		token, err := dgr.login(image)
		if err != nil {
			logrus.Error(err)
			return []string{}, err
		}
		if token == "" {
			return []string{}, fmt.Errorf("could not retrieve a bearer token (empty value)")
		}

		logrus.Debugf("Bearer token retrieved successfully.")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		logrus.Debugf("Emitting a request to: %v", URL)

		res, err := dgr.WebClient.Do(req)
		if err != nil {
			return []string{}, err
		}

		logrus.Debugf("Received the following response from the registry %q: %v", image.Registry, res)

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return []string{}, err
		}

		if res.StatusCode != 200 {
			if res.StatusCode == 404 {
				// Return an empty string but no error (it's the caller responsibility to handle the error)
				return []string{}, nil
			}
			err = fmt.Errorf("unexpected error from the registry %s for image %s:%s",
				image.Registry,
				image.Repository,
				image.Tag,
			)
			logrus.Error(err)
			return []string{}, err
		}

		// OCI registry relies on the header "Link" to handle pagination as defined in RFC 5988
		// https://tools.ietf.org/html/rfc5988
		headerLink := res.Header.Get("Link")

		switch headerLink {
		case "":
			URL = ""

		default:
			linkGroup := link.Parse(headerLink)

			for _, l := range linkGroup {
				URL = fmt.Sprintf("https://%s%s", endpoints.ApiService, l.URI)
			}
		}

		switch res.Header.Get("content-type") {
		case "application/json":
			type response struct {
				Name string
				Tags []string
			}

			data := response{}

			err = json.Unmarshal(body, &data)
			if err != nil {
				return data.Tags, err
			}

			tags = append(tags, data.Tags...)

		default:
			logrus.Debugf("Returned response body:\n%v", string(body)) // Shows answer in debug mode to help diagnostics
		}
	}

	return tags, nil
}
