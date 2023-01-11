package cargopackage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	"os"
	"path/filepath"
	"sort"
)

// CargoPackage defines a resource of type "cargopackage"
type CargoPackage struct {
	spec Spec
	//options []remote.Option
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	foundVersion  version.Version
	packageData   PackageData
}

type PackageVersion struct {
	Name     string `json:"name"`
	Version  string `json:"vers"`
	Yanked   bool   `json:"yanked"`
	Checksum string `json:"chsum"`
}

type PackageData struct {
	Name     string
	Versions []PackageVersion
}

const (
	// URL of the default Npm api data
	cargoDefaultIndexURL string = "https://github.com/rust-lang/crates.io-index.git"
)

// New returns a reference to a newly initialized CargoPackage object from a cargopackage.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*CargoPackage, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if newSpec.IndexUrl == "" {
		// Default to `crates.io` url
		newSpec.IndexUrl = cargoDefaultIndexURL
	}

	err = newSpec.Validate()
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
		versions = append(versions, value.Version)
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

// Get package data from Json API
func (cp *CargoPackage) getPackageData() (PackageData, error) {
	var pd PackageData

	pd.Name = cp.spec.Package

	dir, err := os.MkdirTemp("", "*-cargo-index")
	if err != nil {
		logrus.Errorf("something went wrong while creating a temp directory for the cargo index %q\n", err)
		return pd, err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			logrus.Errorf("something went wrong while cleaning the temp directory for the cargo index %q\n", err)
		}
	}(dir) // clean up

	var auth transport.AuthMethod
	if cp.spec.Username != "" {
		auth = &http.BasicAuth{
			Username: cp.spec.Username,
			Password: cp.spec.Password,
		}
	} else if cp.spec.PrivateKey != "" {
		privateKeyUser := "git"
		if cp.spec.PrivateKeyUser != "" {
			privateKeyUser = cp.spec.PrivateKeyUser
		}
		publicKeys, err := ssh.NewPublicKeys(privateKeyUser, []byte(cp.spec.PrivateKey), cp.spec.PrivateKeyPassword)
		if err != nil {
			logrus.Errorf("something went wrong while parsing the private key for the cargo index %q\n", err)
			return pd, err
		}
		auth = publicKeys
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:          cp.spec.IndexUrl,
		Auth:         auth,
		Depth:        1,
		SingleBranch: true,
	})

	if err != nil {
		logrus.Errorf("something went wrong while cloning the cargo index %q\n", err)
		return pd, err
	}

	packageDir, err := getPackageFileDir(cp.spec.Package)
	packageFilePath := filepath.Join(dir, packageDir, cp.spec.Package)
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
		pd.Versions = append(pd.Versions, packageVersion)
	}
	return pd, nil
}
