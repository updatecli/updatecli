package helm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
	"helm.sh/helm/v3/pkg/action"
	helm "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	yml "sigs.k8s.io/yaml"
)

// DependencyUpdate updates the "Chart.lock" file if needed
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
// It fails if API Version isn't set (ErrNoAPIVersion) or if the "unmarshal" operation fails.
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
// It fails if API Version isn't set (ErrNoAPIVersion) or if the "unmarshal" operation fails.
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
func (c *Chart) MetadataUpdate(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	md := helm.Metadata{}

	metadataFilename := filepath.Join(c.spec.Name, "Chart.yaml")
	if scm != nil {
		metadataFilename = filepath.Join(scm.GetDirectory(), metadataFilename)
	}

	file, err := os.Open(metadataFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := yml.Unmarshal(data, &md); err != nil {
		return err
	}

	newVersion := md.Version
	oldVersion := md.Version

	// Init Chart Version if not set yet
	if len(md.Version) == 0 {
		newVersion = "0.0.0"
	}

forLoop:
	for _, inc := range strings.Split(c.spec.VersionIncrement, ",") {
		v, err := semver.NewVersion(newVersion)
		if err != nil {
			return err
		}

		switch inc {
		case MAJORVERSION:
			newVersion = v.IncMajor().String()
		case MINORVERSION:
			newVersion = v.IncMinor().String()
		case PATCHVERSION:
			newVersion = v.IncPatch().String()
		case NOINCREMENT:
			// Reset Version to its initial value
			newVersion = oldVersion
			break forLoop
		default:
			logrus.Warningf("Wrong increment rule %q, ignoring", inc)
		}
	}

	if err := c.metadataYamlPathUpdate("$.version", newVersion, scm, dryRun, resultTarget); err != nil {
		return err
	}

	if len(md.AppVersion) > 0 && c.spec.AppVersion {
		if err := c.metadataYamlPathUpdate("$.appVersion", source, scm, dryRun, resultTarget); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chart) metadataYamlPathUpdate(key string, value string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	yamlSpec := yaml.Spec{
		File: filepath.Join(c.spec.Name, "Chart.yaml"),
		Key:  key,
	}

	yamlResource, err := yaml.New(yamlSpec)
	if err != nil {
		return err
	}

	metadataResultTarget := result.Target{}
	if err := yamlResource.Target(value, scm, dryRun, &metadataResultTarget); err != nil {
		return err
	}

	if metadataResultTarget.Changed {
		resultTarget.Description = fmt.Sprintf("%s\n%s",
			resultTarget.Description,
			metadataResultTarget.Description)
	}

	return nil
}

// RequirementsUpdate test if Updatecli updated the "requirements.yaml" file
// if it's the case then Updatecli also delete and recreate the "requirements.lock" file
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
