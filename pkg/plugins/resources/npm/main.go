package npm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"sort"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for an Npm package
// parsed from an updatecli manifest file
type Spec struct {
	// Defines the specific npm package name
	Name string `yaml:",omitempty"`
	// Defines a specific package version
	Version string `yaml:",omitempty"`
	// Defines registry url
	URL string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

type distTags struct {
	Latest string
	Next   string
}

type versions struct {
	Name       string
	Version    string
	Deprecated string
}

type Data struct {
	Versions map[string]versions
	DistTags distTags `json:"dist-tags,omitempty"`
}

// Npm defines a resource of kind "npm"
type Npm struct {
	spec          Spec
	versionFilter version.Filter // Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	foundVersion  version.Version
	data          Data
}

const (
	// URL of the default Npm api data
	npmDefaultApiURL string = "https://registry.npmjs.org/"
)

// New returns a new valid Npm package object.
func New(spec interface{}) (*Npm, error) {
	var newSpec Spec

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Npm{}, err
	}

	if newSpec.URL == "" {
		newSpec.URL = npmDefaultApiURL
	}

	err = newSpec.Validate()
	if err != nil {
		return &Npm{}, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &Npm{}, err
	}

	return &Npm{
		spec:          newSpec,
		versionFilter: newFilter,
	}, nil
}

// Validate run some validation on the Npm struct
func (s *Spec) Validate() (err error) {
	if len(s.Name) == 0 {
		logrus.Errorf("npm package name not defined")
		return errors.New("npm package name not defined")
	}
	return nil
}

// GetVersions fetch all versions of the Npm package
func (n *Npm) getVersions() (v string, versions []string, err error) {
	n.data, err = getPackageData(n.spec.URL + n.spec.Name)

	if err != nil {
		return "", nil, err
	}

	for _, value := range n.data.Versions {
		versions = append(versions, value.Version)
	}

	if n.versionFilter.Kind == version.LATESTVERSIONKIND {
		return n.data.DistTags.Latest, versions, nil
	}

	sort.Strings(versions)
	n.foundVersion, err = n.versionFilter.Search(versions)
	if err != nil {
		return "", nil, err
	}

	return n.foundVersion.GetVersion(), versions, nil
}

// Get package data from Json API
func getPackageData(URL string) (Data, error) {
	var d Data

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data %q\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data %q\n", err)
		return Data{}, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Errorf("something went wrong while getting npm api data %q\n", URL)
		logrus.Errorf("Error %q\n", err)
		logrus.Debugf("\n%v\n", string(body))
		return Data{}, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return Data{}, err
	}

	err = json.Unmarshal(data, &d)
	if err != nil {
		logrus.Errorf("error unmarshaling json: %q", err)
		return Data{}, err
	}

	return d, nil
}
