package cargopackage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	httputils "github.com/updatecli/updatecli/pkg/plugins/utils/http"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// URL of the default Crates index api
	cratesDefaultIndexApiUrl string = "https://crates.io/api/v1/crates"
)

// CargoPackage defines a resource of type "cargopackage"
type CargoPackage struct {
	spec Spec
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	foundVersion  version.Version
	packageData   PackageData
	registry      cargo.Registry
	isSCM         bool
	webClient     httpclient.HTTPClient
}

type PackageVersion struct {
	Num     string `json:"num,omitempty"`
	Version string `json:"vers,omitempty"`
	Yanked  bool   `json:"yanked"`
}

type PackageCrate struct {
	Name string `json:"name"`
}

type PackageData struct {
	Crate    PackageCrate     `json:"crate"`
	Versions []PackageVersion `json:"versions"`
}

// New returns a reference to a newly initialized CargoPackage object from a cargopackage.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}, isSCM bool) (*CargoPackage, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if newSpec.IndexUrl != "" {
		logrus.Infof("IndexURL IS SET, but not used")
		switch newSpec.Registry.URL != "" {
		case true:
			logrus.Warningf("Registry.URL and IndexUrl are mutually exclusive, unset indexurl")
			newSpec.IndexUrl = ""
		case false:
			logrus.Warningf("indexurl is deprecated in favor of registry.url")
			newSpec.Registry.URL = newSpec.IndexUrl
			newSpec.IndexUrl = ""
		}
	}

	if !newSpec.Registry.Validate() {
		return nil, fmt.Errorf("invalid registry configuration")
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	webClient := httpclient.NewThrottledClient(1*time.Second, 1, http.DefaultTransport)
	newResource := &CargoPackage{
		spec:          newSpec,
		versionFilter: newFilter,
		isSCM:         isSCM,
		registry:      newSpec.Registry,
		webClient:     webClient,
	}

	if !newResource.isSCM && newSpec.Registry.RootDir == "" && newSpec.Registry.URL == "" {
		newResource.registry.URL = cratesDefaultIndexApiUrl
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (cp *CargoPackage) Changelog(from, to string) *result.Changelogs {
	return nil
}

// GetVersions fetch all versions of the Cargo package
func (cp *CargoPackage) getVersions() (v string, versions []string, err error) {
	cp.packageData, err = cp.getPackageData()

	if err != nil {
		return "", nil, err
	}

	for _, value := range cp.packageData.Versions {
		if !value.Yanked {
			versions = append(versions, value.Num)
		}
	}

	if len(versions) == 0 {
		// No versions found
		return "", versions, nil
	}
	sort.Strings(versions)
	cp.foundVersion, err = cp.versionFilter.Search(versions)
	if err != nil {
		return "", nil, err
	}

	return cp.foundVersion.GetVersion(), versions, nil
}

func getPackageFileDir(packageName string) (string, error) {
	if packageName == "" {
		err := errors.New("got empty package name")
		logrus.Errorf("%q\n", err)
		return "", err
	}
	switch packageNameLen := len(packageName); packageNameLen {
	case 1:
		return fmt.Sprintf("%d", packageNameLen), nil
	case 2:
		return fmt.Sprintf("%d", packageNameLen), nil
	case 3:
		return fmt.Sprintf("%d/%s", packageNameLen, string(packageName[0])), nil
	default:
		return fmt.Sprintf("%s/%s", packageName[0:2], packageName[2:4]), nil
	}
}

func (cp *CargoPackage) getPackageDataFromApi(name string, indexUrl string) (PackageData, error) {
	packageUrl := fmt.Sprintf("%s/%s", indexUrl, name)

	req, err := http.NewRequest("GET", packageUrl, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting cargo api data %q\n", err)
		return PackageData{}, err
	}

	req.Header.Set("User-Agent", httputils.UserAgent)

	if cp.registry.Auth.Token != "" {
		format := "Bearer %s"
		if cp.registry.Auth.HeaderFormat != "" {
			format = cp.registry.Auth.HeaderFormat
		}
		req.Header.Set("Authorization", fmt.Sprintf(format, cp.registry.Auth.Token))
	}

	res, err := cp.webClient.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting cargo api data %q\n", err)
		return PackageData{}, err
	}
	defer res.Body.Close()

	var d PackageData
	err = json.NewDecoder(res.Body).Decode(&d)
	if err != nil && err != io.EOF {
		logrus.Errorf("something went wrong while reading cargo api data%q\n", err)
		return PackageData{}, err
	}
	return d, nil
}

func (cp *CargoPackage) getPackageDataFromFS(name string, indexDir string) (PackageData, error) {
	var pd PackageData
	pd.Crate.Name = name
	packageDir, err := getPackageFileDir(name)
	if err != nil {
		logrus.Errorf("something went wrong while getting the package directory from its name %q\n", err)
		return pd, err
	}
	packageFilePath := filepath.Join(indexDir, packageDir, name)
	packageInfoFile, err := os.Open(packageFilePath)
	if err != nil {
		return pd, nil
	}
	defer func(packageInfoFile *os.File) {
		err := packageInfoFile.Close()
		if err != nil {
			logrus.Errorf("something went wrong while cleaning the package file %q\n", err)
		}
	}(packageInfoFile)

	scanner := bufio.NewScanner(packageInfoFile)
	for scanner.Scan() {
		var packageVersion PackageVersion
		err = json.Unmarshal(scanner.Bytes(), &packageVersion)
		if err != nil {
			logrus.Errorf("something went wrong while parsing the version %q\n", err)
		}
		if packageVersion.Yanked {
			continue
		}
		// File index store version info in Version Field
		packageVersion.Num = packageVersion.Version
		pd.Versions = append(pd.Versions, packageVersion)
	}
	return pd, nil
}

// Get package data from Json API
func (cp *CargoPackage) getPackageData() (PackageData, error) {
	if cp.registry.RootDir != "" {
		return cp.getPackageDataFromFS(cp.spec.Package, cp.registry.RootDir)
	}
	return cp.getPackageDataFromApi(cp.spec.Package, cp.registry.URL)
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// to identify the resource without any sensitive information or context specific data.
func (cp *CargoPackage) ReportConfig() interface{} {
	return Spec{
		IndexUrl: cp.spec.IndexUrl,
		Registry: cp.spec.Registry,
		Package:  cp.spec.Package,
		Version:  cp.spec.Version,
	}
}
