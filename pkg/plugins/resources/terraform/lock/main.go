package lock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/minamijoyo/tfupdate/lock"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/zclconf/go-cty/cty"
)

type openTofuVersionsResponse struct {
	Versions []struct {
		Version   string `json:"version"`
		Platforms []struct {
			OS   string `json:"os"`
			Arch string `json:"arch"`
		} `json:"platforms"`
	} `json:"versions"`
}

type openTofuDownloadResponse struct {
	Packages map[string]struct {
		Hashes []string `json:"hashes"`
	} `json:"packages"`
}

type TerraformLock struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]file // map of file paths to file contents
	lockIndex        lock.Index      // index is a cached index for updating dependency lock files.
	provider         terraformRegistryAddress.Provider
	httpClient       httpclient.HTTPClient
}

type file struct {
	originalFilePath string
	filePath         string
	content          string
}

func New(spec interface{}) (*TerraformLock, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &TerraformLock{
		spec:             newSpec,
		contentRetriever: &text.Text{},
		httpClient:       httpclient.NewRetryClient(),
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	newResource.files = make(map[string]file)
	// File as unique element of newResource.files
	if len(newResource.spec.File) > 0 {
		filePath := strings.TrimPrefix(newResource.spec.File, "file://")
		newResource.files[filePath] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}
	// Files
	for _, filePath := range newResource.spec.Files {
		filePath := strings.TrimPrefix(filePath, "file://")
		newResource.files[filePath] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}

	provider, err := terraformRegistryAddress.ParseProviderSource(newResource.spec.Provider)
	if err != nil {
		return nil, err
	}

	newResource.provider = provider

	client, err := lock.NewProviderDownloaderClient(lock.TFRegistryConfig{
		BaseURL: fmt.Sprintf("https://%s/", provider.Hostname),
	})
	if err != nil {
		return nil, err
	}

	newResource.lockIndex = lock.NewIndex(client)

	return newResource, nil
}

func (t *TerraformLock) Query(resourceFile file) (string, []string, error) {
	file, err := terraformUtils.ParseHcl(resourceFile.content, resourceFile.originalFilePath)
	if err != nil {
		return "", nil, err
	}

	providerBlock, err := getProviderBlock(file, resourceFile.originalFilePath, t.provider.String())
	if err != nil {
		return "", nil, err
	}

	quotedValue := strings.TrimSpace(string(providerBlock.Body().GetAttribute("version").Expr().BuildTokens(nil).Bytes()))

	version := strings.Trim(quotedValue, `"`)

	hashesTokens := providerBlock.Body().GetAttribute("hashes").Expr().BuildTokens(nil)

	var hashes []string

	for _, t := range hashesTokens {
		if t.Type == hclsyntax.TokenQuotedLit {
			hashes = append(hashes, string(t.Bytes))
		}
	}

	return version, hashes, nil
}

func (t *TerraformLock) Apply(filePath string, versionToWrite string, hashesToWrite []string) error {
	resourceFile := t.files[filePath]

	file, err := terraformUtils.ParseHcl(resourceFile.content, resourceFile.originalFilePath)
	if err != nil {
		return err
	}

	providerBlock, err := getProviderBlock(file, resourceFile.originalFilePath, t.provider.String())
	if err != nil {
		return err
	}

	providerBlock.Body().SetAttributeValue("version", cty.StringVal(versionToWrite))

	if providerBlock.Body().GetAttribute("constraints") != nil && !t.spec.SkipConstraints {
		providerBlock.Body().SetAttributeValue("constraints", cty.StringVal(versionToWrite))
	}

	providerBlock.Body().SetAttributeRaw("hashes", tokensForListPerLine(hashesToWrite))

	resourceFile.content = string(hclwrite.Format(file.BuildTokens(nil).Bytes()))

	t.files[filePath] = resourceFile

	return nil
}

// Read puts the content of the file(s) as value of the y.files map if the file(s) exist(s) or log the non existence of the file
func (t *TerraformLock) Read() error {
	var err error

	// Retrieve files content
	for filePath := range t.files {
		f := t.files[filePath]
		if t.contentRetriever.FileExists(f.filePath) {
			f.content, err = t.contentRetriever.ReadAll(f.filePath)
			if err != nil {
				return err
			}
			t.files[filePath] = f

		} else {
			return fmt.Errorf("%s The specified file %q does not exist", result.FAILURE, f.filePath)
		}
	}
	return nil
}

func (t *TerraformLock) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range t.files {
		if workDir != "" {
			f := t.files[filePath]
			f.filePath = utils.JoinFilePathWithWorkingDirectoryPath(f.originalFilePath, workDir)
			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", f.originalFilePath, f.filePath)
			t.files[filePath] = f
		}
	}
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (t *TerraformLock) Changelog(from, to string) *result.Changelogs {
	return nil
}

func (t *TerraformLock) getProviderHashes(version string) ([]string, error) {
	if t.provider.Hostname == "registry.opentofu.org" {
		hashes, err := t.getOpenTofuProviderHashes(version)
		if err == nil {
			return hashes, nil
		}
		logrus.Warnf("Failed to fetch OpenTofu provider hashes directly: %s. Falling back to default method.", err)
	}

	pv, err := t.lockIndex.GetOrCreateProviderVersion(context.Background(), t.provider.ForDisplay(), version, t.spec.Platforms)
	if err != nil {
		return nil, fmt.Errorf("%s failed to query provider locks for provider: %q, version: %q, platforms: %q: %s",
			result.FAILURE,
			t.spec.Provider,
			version,
			t.spec.Platforms,
			err.Error(),
		)
	}

	return pv.AllHashes(), nil
}

func (t *TerraformLock) getOpenTofuProviderHashes(version string) ([]string, error) {
	scheme := "https"
	if strings.HasPrefix(t.provider.Hostname, "127.0.0.1") || strings.HasPrefix(t.provider.Hostname, "localhost") {
		scheme = "http"
	}

	urlVersions := fmt.Sprintf("%s://%s/v1/providers/%s/%s/versions", scheme, t.provider.Hostname, t.provider.Namespace, t.provider.Type)
	req, err := http.NewRequestWithContext(context.Background(), "GET", urlVersions, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d fetching versions", resp.StatusCode)
	}
	var versionsRes openTofuVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionsRes); err != nil {
		return nil, err
	}

	var targetOS, targetArch string
	for _, v := range versionsRes.Versions {
		if v.Version == version {
			if len(v.Platforms) > 0 {
				targetOS = v.Platforms[0].OS
				targetArch = v.Platforms[0].Arch
				break
			}
		}
	}
	if targetOS == "" || targetArch == "" {
		return nil, fmt.Errorf("version %s or its platforms not found in registry", version)
	}

	urlDownload := fmt.Sprintf("%s://%s/v1/providers/%s/%s/%s/download/%s/%s", scheme, t.provider.Hostname, t.provider.Namespace, t.provider.Type, version, targetOS, targetArch)
	reqDownload, err := http.NewRequestWithContext(context.Background(), "GET", urlDownload, nil)
	if err != nil {
		return nil, err
	}
	respDownload, err := t.httpClient.Do(reqDownload)
	if err != nil {
		return nil, err
	}
	defer respDownload.Body.Close()
	if respDownload.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d fetching download metadata", respDownload.StatusCode)
	}
	var downloadRes openTofuDownloadResponse
	if err := json.NewDecoder(respDownload.Body).Decode(&downloadRes); err != nil {
		return nil, err
	}

	var allHashes []string
	for _, pkg := range downloadRes.Packages {
		allHashes = append(allHashes, pkg.Hashes...)
	}

	sort.Strings(allHashes)
	var uniqueHashes []string
	for _, hash := range allHashes {
		if len(uniqueHashes) == 0 || uniqueHashes[len(uniqueHashes)-1] != hash {
			uniqueHashes = append(uniqueHashes, hash)
		}
	}

	return uniqueHashes, nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (t *TerraformLock) ReportConfig() interface{} {
	return Spec{
		File:            t.spec.File,
		Files:           t.spec.Files,
		Provider:        t.spec.Provider,
		Value:           t.spec.Value,
		Platforms:       t.spec.Platforms,
		SkipConstraints: t.spec.SkipConstraints,
	}
}
