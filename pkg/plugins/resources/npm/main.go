package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"

	"gopkg.in/ini.v1"

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
	// URL defines the registry url (defaults to `https://registry.npmjs.org/`)
	URL string `yaml:",omitempty"`
	// RegistryToken defines the token to use when connection to the registry
	RegistryToken string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// NpmrcPath defines the path to the .npmrc file
	NpmrcPath string `yaml:"npmrcpath,omitempty"`
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
	webClient     httpclient.HTTPClient
	rcConfig      RcConfig
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

	// parse the .npmrc files
	rcConfig, err := getNpmrcConfig(newSpec.NpmrcPath, newSpec.URL, newSpec.RegistryToken)
	if err != nil {
		return &Npm{}, err
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
		rcConfig:      rcConfig,
		webClient:     http.DefaultClient,
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

type Registry struct {
	AuthToken string
	Url       string
}

type RcConfig struct {
	Registries map[string]Registry
	Scopes     map[string]string
}

func defaultNpmConfig(defaultUrl string, defaultToken string) RcConfig {
	var config RcConfig
	config.Registries = make(map[string]Registry)
	var url string
	if defaultUrl == "" {
		url = npmDefaultApiURL
	} else {
		url = defaultUrl
	}
	config.Registries["default"] = Registry{
		Url:       url,
		AuthToken: defaultToken,
	}
	config.Scopes = make(map[string]string)
	return config
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func getNpmrcConfig(path string, defaultUrl string, defaultToken string) (RcConfig, error) {
	config := defaultNpmConfig(defaultUrl, defaultToken)
	if path == "" || !fileExists(path) {
		path = fmt.Sprintf("%s/.npmrc", os.Getenv("HOME"))
	}
	if !fileExists(path) {
		return config, nil
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return config, err
	}
	for _, section := range cfg.Section("").Keys() {
		sectionName := section.Name()
		if strings.HasPrefix(sectionName, "//") {
			// Registry section
			authTokenValue := strings.Split(section.Value(), "_authToken=")
			if len(authTokenValue) == 2 {
				config.Registries[sectionName[2:]] = Registry{
					AuthToken: authTokenValue[1],
					Url:       fmt.Sprintf("https:%s", sectionName),
				}
			}
		} else if strings.HasPrefix(sectionName, "@") {
			// Scope section
			registryValue := strings.Split(section.Value(), "registry=https://")
			if len(registryValue) == 2 {
				config.Scopes[sectionName] = registryValue[1]
			}
		}

	}
	return config, nil
}

// GetVersions fetch all versions of the Npm package
func (n *Npm) getVersions() (v string, versions []string, err error) {
	n.data, err = n.getPackageData(n.spec.Name)

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
func (n *Npm) getPackageData(packageName string) (Data, error) {
	var d Data
	var registry Registry
	// We need to find the registry URL to use for the package
	if strings.HasPrefix(packageName, "@") {
		// Scoped package
		if scope, ok := n.rcConfig.Scopes[packageName[:strings.Index(packageName, "/")]]; ok {
			// We found a scope registry for the package
			registry = n.rcConfig.Registries[scope]
		} else {
			// We didn't find a scope for the package, we use the default registry
			registry = n.rcConfig.Registries["default"]
		}
	} else {
		// Not a scoped package, using default registry
		registry = n.rcConfig.Registries["default"]
	}

	URL := fmt.Sprintf("%s%s", registry.Url, packageName)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data %q\n", err)
	}

	if registry.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", registry.AuthToken))
	}

	res, err := n.webClient.Do(req)
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
