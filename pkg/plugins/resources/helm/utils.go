package helm

import (
	"bytes"
	"encoding/base64"
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
	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
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

	if c.spec.Username != "" && c.spec.Password != "" {
		userPass := []byte(fmt.Sprintf("%s:%s", c.spec.Username, c.spec.Password))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(userPass)))
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
	var err error

	/*
		currentChartMetadata is the metadata of the chart we are working on.
		It can be modified by different Updatecli execution.
	*/
	var currentChartMetadata helm.Metadata
	/*
		originChartMetadata is the metadata of the chart from the base from.
		we want to be sure that we are not modifying the chart version more than needed
	*/
	var originChartMetadata helm.Metadata

	metadataFilename := filepath.Join(c.spec.Name, "Chart.yaml")
	if scm != nil {
		// We need to retrieve the source branch to know where to look for the original Chart.yaml file
		sourceBranch, _, _ := scm.GetBranches()

		data, err := git.ReadFileFromRevision(
			scm.GetDirectory(),
			sourceBranch,
			metadataFilename)
		if err != nil {
			return fmt.Errorf("reading %q from branch %q: %w", metadataFilename, sourceBranch, err)
		}

		// Unmarshal the source yaml file into a struct
		if err := yml.Unmarshal(data, &originChartMetadata); err != nil {
			return fmt.Errorf("unmarshalling %q: %w", metadataFilename, err)
		}

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

	// Unmarshal the working yaml file into a struct
	if err := yml.Unmarshal(data, &currentChartMetadata); err != nil {
		return err
	}

	if len(currentChartMetadata.AppVersion) > 0 && c.spec.AppVersion {
		if err := c.metadataYamlPathUpdate("$.appVersion", source, scm, dryRun, resultTarget); err != nil {
			return err
		}
	}

	/*
	  We only want to update the Chart metadata if the chart has been modified during the current target execution.
	  To make this process more idempotent in the context of a scm, we could also check if the helm chart
	  has been modified during one of the previous target execution by comparing the current chart versus the one defined
	  on the source branch. But the code complexity induced by this check is probably not worth the effort today.
	*/
	if !resultTarget.Changed {
		return nil
	}

	// Handle the situation where the version is not set yet
	if currentChartMetadata.Version == "" {
		currentChartMetadata.Version = "0.0.1"
	}

	computedVersion := ""
	switch scm {
	// If scm is not defined then we want to update the chart version from the current working chart
	case nil:
		computedVersion = currentChartMetadata.Version
	default:
		computedVersion = originChartMetadata.Version

	}

	// Init Chart Version if not set yet
	if computedVersion == "" {
		computedVersion = "0.0.0"
	}
	initialComputedVersion := computedVersion

forLoop:
	for _, inc := range strings.Split(c.spec.VersionIncrement, ",") {
		v, err := semver.NewVersion(computedVersion)
		if err != nil {
			return err
		}

		switch inc {
		case MAJORVERSION:
			computedVersion = v.IncMajor().String()
		case MINORVERSION:
			computedVersion = v.IncMinor().String()
		case PATCHVERSION:
			computedVersion = v.IncPatch().String()
		case NOINCREMENT:
			// Reset Version to its initial value
			computedVersion = initialComputedVersion
			break forLoop
		case AUTO:
			origVer, oErr := semver.NewVersion(strings.Trim(resultTarget.Information, "~><=^"))
			newVer, nErr := semver.NewVersion(strings.Trim(resultTarget.NewInformation, "~><=^"))

			if oErr != nil || nErr != nil {
				computedVersion = v.IncMinor().String()
				continue
			}

			if newVer.Major() != origVer.Major() {
				computedVersion = v.IncMajor().String()
			} else if newVer.Minor() != origVer.Minor() {
				computedVersion = v.IncMinor().String()
			} else if newVer.Patch() != origVer.Patch() {
				computedVersion = v.IncPatch().String()
			} else {
				computedVersion = v.IncMinor().String()
			}
		default:
			logrus.Warningf("Wrong increment rule %q, ignoring", inc)
		}
	}

	workingChartVersion, err := semver.NewVersion(currentChartMetadata.Version)
	if err != nil {
		return err
	}

	computedChartVersion, err := semver.NewVersion(computedVersion)
	if err != nil {
		return err
	}

	if scm == nil && !resultTarget.Changed {
		return nil
	}

	// If the work chart version is greater than the newly calculated version
	// then we assume that a previous Updatecli execution already updated the chart version
	// to a greater value than the one we computed in this execution.
	if workingChartVersion.GreaterThan(computedChartVersion) {
		computedVersion = currentChartMetadata.Version
	}

	if computedVersion != currentChartMetadata.Version && scm != nil {
		// If the chart version is updated then we need to update the resultTarget if it wasn't already updated
		if !resultTarget.Changed {
			resultTarget.Description = fmt.Sprintf("Chart version updated to %s", computedVersion)
			resultTarget.Changed = true
			resultTarget.Result = result.ATTENTION
		}

		logrus.Debugf("Updating chart version from %q to %q", currentChartMetadata.Version, computedVersion)
	}

	if err := c.metadataYamlPathUpdate("$.version", computedVersion, scm, dryRun, resultTarget); err != nil {
		return err
	}

	return nil
}

// metadataYamlPathUpdate updates the Chart.yaml
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
		resultTarget.Result = result.ATTENTION
		resultTarget.Changed = true
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
