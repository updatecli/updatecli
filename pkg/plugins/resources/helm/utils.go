package helm

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	helm "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	yml "sigs.k8s.io/yaml"
)

// DependencyUpdate ensures that Chart.lock is updated if needed
func (c *Chart) DependencyUpdate(out *bytes.Buffer, chartPath string) error {

	client := action.NewDependency()
	settings := cli.New()

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(out),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)

	if err != nil {
		return err
	}

	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		Keyring:          client.Keyring,
		SkipUpdate:       client.SkipRefresh,
		Getters:          getter.All(settings),
		RegistryClient:   registryClient,
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}

	err = man.Update()

	if err != nil {
		return err
	}

	return nil
}

// GetRepoIndexFromFile loads an index file from a local file and does minimal validity checking.
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func (c *Chart) GetRepoIndexFromFile(rootDir string) (repo.IndexFile, error) {

	URL := strings.TrimPrefix(c.spec.URL, "file://")

	if rootDir != "" {
		URL = filepath.Join(rootDir, URL)
	}

	if filepath.Base(URL) != "index.yaml" {
		URL = filepath.Join(URL, "index.yaml")
	}

	rawIndexFile, err := os.Open(URL)

	if err != nil {
		return repo.IndexFile{}, err
	}

	data, err := io.ReadAll(rawIndexFile)

	if err != nil {
		return repo.IndexFile{}, err
	}

	indexFile := repo.IndexFile{}

	if err := yml.Unmarshal(data, &indexFile); err != nil {
		return indexFile, err
	}

	indexFile.SortEntries()

	if indexFile.APIVersion == "" {
		return indexFile, repo.ErrNoAPIVersion
	}

	return indexFile, nil
}

// GetRepoIndexFromUrl loads an index file and does minimal validity checking.
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func (c *Chart) GetRepoIndexFromURL() (repo.IndexFile, error) {
	var err error

	URL := c.spec.URL

	if !strings.HasSuffix(URL, "index.yaml") {
		URL, err = url.JoinPath(c.spec.URL, "index.yaml")
		if err != nil {
			return repo.IndexFile{}, err
		}
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return repo.IndexFile{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return repo.IndexFile{}, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)

		logrus.Errorf("something went wrong while contacting %q\n", URL)
		logrus.Debugf("\n%v\n", string(body))

		return *repo.NewIndexFile(), err

	}

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return repo.IndexFile{}, err
	}

	i := repo.IndexFile{}

	if err := yml.Unmarshal(data, &i); err != nil {
		return i, err
	}

	i.SortEntries()

	if i.APIVersion == "" {
		return i, repo.ErrNoAPIVersion
	}

	return i, nil
}

// MetadataUpdate updates a metadata if necessary and it bump the ChartVersion
func (c *Chart) MetadataUpdate(chartPath string, dryRun bool) error {
	md := helm.Metadata{}

	metadataFilename := filepath.Join(chartPath, "Chart.yaml")

	file, err := os.Open(metadataFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	err = yml.Unmarshal(data, &md)

	if err != nil {
		return err
	}

	// Init Chart Version if not set yet
	if len(md.Version) == 0 {
		md.Version = "0.0.0"
	}

	oldVersion := md.Version

forLoop:
	for _, inc := range strings.Split(c.spec.VersionIncrement, ",") {
		v, err := semver.NewVersion(md.Version)
		if err != nil {
			return err
		}

		switch inc {
		case MAJORVERSION:
			md.Version = v.IncMajor().String()
		case MINORVERSION:
			md.Version = v.IncMinor().String()
		case PATCHVERSION:
			md.Version = v.IncPatch().String()
		case NOINCREMENT:
			// Reset Version to its initial value
			md.Version = oldVersion
			break forLoop
		default:
			logrus.Warningf("Wrong increment rule %q, ignoring", inc)
		}
	}

	if oldVersion != md.Version {
		logrus.Infof("\tChart Version updated from %q to %q\n", oldVersion, md.Version)
	}

	if len(md.AppVersion) > 0 && c.spec.AppVersion {
		if md.AppVersion != c.spec.Value {
			logrus.Infof("\tAppVersion updated from %s to %s\n", md.AppVersion, c.spec.Value)
			md.AppVersion = c.spec.Value
		}
	}

	if err != nil {
		return err
	}

	if !dryRun {
		data, err := yml.Marshal(md)
		if err != nil {
			return err
		}

		file, err := os.Create(metadataFilename)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

// RequirementsUpdate test if we are updating the file requirements.yaml
// if it's the case then we also have to delete and recreate the file
// requirements.lock
func (c *Chart) RequirementsUpdate(chartPath string) error {
	lockFilename := filepath.Join(chartPath, "requirements.lock")

	if strings.Compare(c.spec.File, "requirements.yaml") != 0 {
		return nil
	}

	f, err := os.Stat(lockFilename)

	if os.IsExist(err) && !f.IsDir() {
		err = os.Remove(lockFilename)
		if err != nil {
			return err
		}
	}

	return nil

}
