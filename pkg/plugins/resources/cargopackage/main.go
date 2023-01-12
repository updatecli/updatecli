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

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
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
	// indexDir holds the location of the indexDir
	indexDir string
	indexUrl string
	isSCM    bool
}

type CargoUser struct {
	Avatar string `json:"avatar"`
	Id     int    `json:"id"`
	Login  string `json:"login"`
	Name   string `json:"name"`
	Url    string `json:"url"`
}

type PackageVersion struct {
	AuditActions []struct {
		Action string    `json:"action"`
		Time   string    `json:"time"`
		User   CargoUser `json:"user"`
	} `json:"audit_actions"`
	Checksum  string              `json:"checksum"`
	Crate     string              `json:"crate"`
	CrateSize int                 `json:"crate_size"`
	CreatedAt string              `json:"created_at"`
	DlPath    string              `json:"dl_path"`
	Downloads int                 `json:"downloads"`
	Features  map[string][]string `json:"features"`
	Id        int                 `json:"id"`
	License   string              `json:"license"`
	Links     struct {
		Authors         string `json:"authors"`
		Dependencies    string `json:"dependencies"`
		VersionDownload string `json:"version_downloads"`
	} `json:"links"`
	Num         string    `json:"num,omitempty"`
	PublishedBy CargoUser `json:"published_by"`
	ReadmePath  string    `json:"readme_path"`
	UpdatedAt   string    `json:"updated_at"`
	Version     string    `json:"vers,omitempty"`
	Yanked      bool      `json:"yanked"`
}

type PackageCategory struct {
	Category    string `json:"category"`
	CratesCnt   int    `json:"crates_cnt"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Id          string `json:"id"`
	Slug        string `json:"slug"`
}

type PackageKeyword struct {
	CratesCnt int    `json:"crates_cnt"`
	CreatedAt string `json:"created_at"`
	Keyword   string `json:"keyword"`
	Id        string `json:"id"`
}

type PackageCrate struct {
	Categories    []string `json:"categories"`
	CreatedAt     string   `json:"created_at"`
	Description   string   `json:"description"`
	Documentation string   `json:"documentation"`
	Downloads     int      `json:"downloads"`
	ExactMatch    bool     `json:"exact_match"`
	Homepage      string   `json:"homepage"`
	Id            string   `json:"id"`
	Keywords      []string `json:"keywords"`
	Links         struct {
		OwnerTeam           string `json:"owner_team"`
		OwnerUser           string `json:"owner_user"`
		Owners              string `json:"owners"`
		ReverseDependencies string `json:"reverse_dependencies"`
		VersionDownloads    string `json:"version_downloads"`
		Versions            string `json:"versions"`
	} `json:"links"`
	MaxStableVersion string `json:"max_stable_version"`
	MaxVersion       string `json:"max_version"`
	Name             string `json:"name"`
	NewestVersion    string `json:"newest_version"`
	RecentDownloads  int    `json:"recent_downloads"`
	Repository       string `json:"repository"`
	UpdatedAt        string `json:"updated_at"`
	Versions         []int  `json:"versions"`
}

type PackageData struct {
	Categories []PackageCategory `json:"categories"`
	Keywords   []PackageKeyword  `json:"keywords"`
	Crate      PackageCrate      `json:"crate"`
	Versions   []PackageVersion  `json:"versions"`
}

// New returns a reference to a newly initialized CargoPackage object from a cargopackage.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}, isSCM bool) (*CargoPackage, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	newResource := &CargoPackage{
		spec:          newSpec,
		versionFilter: newFilter,
		isSCM:         isSCM,
		indexDir:      newSpec.IndexDir,
		indexUrl:      newSpec.IndexUrl,
	}

	if !newResource.isSCM && newSpec.IndexDir == "" && newSpec.IndexUrl == "" {
		newResource.indexUrl = cratesDefaultIndexApiUrl
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (cp *CargoPackage) Changelog() string {
	return ""
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

func getPackageDataFromApi(name string, indexUrl string) (PackageData, error) {
	packageUrl := fmt.Sprintf("%s/%s", indexUrl, name)

	req, err := http.NewRequest("GET", packageUrl, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting cargo api data %q\n", err)
		return PackageData{}, err
	}

	res, err := http.DefaultClient.Do(req)
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

func getPackageDataFromFS(name string, indexDir string) (PackageData, error) {
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
		logrus.Errorf("something went wrong while opening the package file %q\n", err)
		return pd, err
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
	if cp.indexDir != "" {
		return getPackageDataFromFS(cp.spec.Package, cp.indexDir)
	}
	return getPackageDataFromApi(cp.spec.Package, cp.indexUrl)
}
