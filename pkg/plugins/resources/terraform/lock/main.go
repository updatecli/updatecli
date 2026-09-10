package lock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/minamijoyo/tfupdate/lock"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/zclconf/go-cty/cty"
)

type TerraformLock struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]file // map of file paths to file contents
	lockIndex        lock.Index      // index is a cached index for updating dependency lock files.
	provider         terraformRegistryAddress.Provider
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
	// For OpenTofu registry, use bounded timeout to avoid hanging on stalled registry
	if t.provider.Hostname == "registry.opentofu.org" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if hashes, err := t.getOpenTofuAllHashes(ctx, version); err == nil && len(hashes) > 0 {
			return hashes, nil
		} else if err != nil {
			logrus.Debugf("opentofu all-hashes fetch failed, falling back to per-platform: %v", err)
		}
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

// getOpenTofuAllHashes fetches provider metadata from registry.opentofu.org
// and returns hashes for all published platforms. It uses the `packages` map
// returned by the provider package metadata endpoint, which contains both
// zh and h1 hashes per platform without requiring zip downloads.
// Uses a timeout-bounded client to avoid stalling.
func (t *TerraformLock) getOpenTofuAllHashes(ctx context.Context, version string) ([]string, error) {
	if len(t.spec.Platforms) == 0 {
		return nil, fmt.Errorf("platforms required")
	}
	platform := t.spec.Platforms[0]
	parts := strings.Split(platform, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid platform %q", platform)
	}
	osName, arch := parts[0], parts[1]
	baseURL := fmt.Sprintf("https://%s/", t.provider.Hostname)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", t.provider.Namespace, t.provider.Type, version, osName, arch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("opentofu registry unexpected status %s for %s", resp.Status, u.String())
	}
	var body struct {
		Packages map[string]struct {
			Hashes []string `json:"hashes"`
		} `json:"packages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if len(body.Packages) == 0 {
		return nil, fmt.Errorf("no packages in opentofu response")
	}
	var hashes []string
	for _, pkg := range body.Packages {
		hashes = append(hashes, pkg.Hashes...)
	}
	slices.Sort(hashes)
	hashes = slices.Compact(hashes)
	return hashes, nil
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
