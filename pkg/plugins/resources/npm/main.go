package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"

	"gopkg.in/ini.v1"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
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
	Name    string
	Version string
	// Deprecated can either be a boolean set to false
	// or a string with a deprecating message
	Deprecated interface{}
}

type Data struct {
	orderedVersions []string
	Versions        map[string]versions `json:"versions,omitempty"`
	DistTags        distTags            `json:"dist-tags,omitempty"`
	Repository      Repository
}

type Repository struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
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
		webClient:     &http.Client{},
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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return Data{}, err
	}

	err = json.Unmarshal(data, &d)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %q", err)
		return Data{}, err
	}

	orderedVersion, err := getOrderedVersions(data)
	if err != nil {
		return Data{}, fmt.Errorf("error getting ordered versions: %q", err)
	}

	if len(orderedVersion) > 0 {
		d.orderedVersions = orderedVersion
	}

	return d, nil
}

func getOrderedVersions(data []byte) ([]string, error) {
	var rawStruct struct {
		Versions json.RawMessage `json:"versions"`
	}

	var orderedVersions []string

	err := json.Unmarshal(data, &rawStruct)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %q", err)
	}

	// Now, use a json.Decoder to extract ordered keys from "versions"
	decoder := json.NewDecoder(strings.NewReader(string(rawStruct.Versions)))

	// Read the opening '{' of the "versions" object
	_, err = decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error decoding json: %q", err)
	}

	// Read version keys in order
	for decoder.More() {
		// Read key
		t, err := decoder.Token()
		if err != nil {
			fmt.Println("Error reading version key:", err)
			return nil, err
		}

		// If we encounter '}', the object has ended
		if t == json.Delim('}') {
			break
		}

		// Ensure it's a string (version key)
		versionKey, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("error decoding json: expected string key but got: %T", t)
		}

		// Store the key in order
		orderedVersions = append(orderedVersions, versionKey)

		// Skip the value (we don't need it)
		var value json.RawMessage
		err = decoder.Decode(&value)
		if err != nil {
			return nil, fmt.Errorf("error decoding json: %q", err)
		}
	}

	return orderedVersions, nil

}

// CleanConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (n *Npm) CleanConfig() interface{} {
	return Spec{
		Name:          n.spec.Name,
		Version:       n.spec.Version,
		URL:           redact.URL(n.spec.URL),
		VersionFilter: n.spec.VersionFilter,
		NpmrcPath:     n.spec.NpmrcPath,
	}
}
