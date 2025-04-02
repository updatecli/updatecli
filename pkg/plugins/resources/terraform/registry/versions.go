package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sort"

	"github.com/sirupsen/logrus"
)

type versionItem struct {
	Version string `json:"version"`
}

type moduleVersionListingItem struct {
	Source   string        `json:"source"`
	Versions []versionItem `json:"versions"`
}

type moduleVersionListing struct {
	Modules []moduleVersionListingItem `json:"modules"`
}

func (m *moduleVersionListing) versions() (versions []versionItem, err error) {
	if len(m.Modules) != 1 {
		return nil, fmt.Errorf("did not received exactly one module from the api")
	}
	return m.Modules[0].Versions, nil
}

func (m *moduleVersionListing) scm() (scm string, err error) {
	if len(m.Modules) != 1 {
		return "", fmt.Errorf("did not received exactly one module from the api")
	}
	return m.Modules[0].Source, nil
}

type providerVersionListing struct {
	Versions []versionItem `json:"versions"`
}

func (m *providerVersionListing) versions() (versions []versionItem, err error) {
	return m.Versions, nil
}

func (m *providerVersionListing) scm() (scm string, err error) {
	return "", nil
}

type versionListing interface {
	versions() (versions []versionItem, err error)
	scm() (scm string, err error)
}

func (t *TerraformRegistry) versions() (versions []string, err error) {
	req, err := http.NewRequest("GET", t.registryAddress.API(), nil)
	if err != nil {
		return nil, err
	}

	res, err := t.webClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Debugf("\n%v\n", string(body))
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var versionListing versionListing
	switch t.registryAddress.registryType {
	case TypeProvider:
		versionListing = &providerVersionListing{}
	case TypeModule:
		versionListing = &moduleVersionListing{}
	default:
		logrus.Errorf("unknown registry type %q", t.registryAddress.registryType)
		return []string{}, err
	}

	err = json.Unmarshal(data, versionListing)
	if err != nil {
		return nil, err
	}

	var rawVersions []versionItem

	rawVersions, err = versionListing.versions()
	if err != nil {
		return nil, err
	}

	versions = make([]string, len(rawVersions))
	for i, version := range rawVersions {
		versions[i] = version.Version
	}

	sort.Strings(versions)
	t.Version, err = t.versionFilter.Search(versions)
	if err != nil {
		return nil, err
	}

	t.scm, err = versionListing.scm()
	if err != nil {
		return nil, err
	}

	return versions, nil
}
