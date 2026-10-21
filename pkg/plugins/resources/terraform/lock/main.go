package lock

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/minamijoyo/tfupdate/lock"
	"github.com/minamijoyo/tfupdate/tfregistry"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/zclconf/go-cty/cty"
)

type TerraformLock struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]file          // map of file paths to file contents
	lockIndex        lock.Index               // index is a cached index for updating dependency lock files.
	registry         tfregistry.ProviderV1API // registry retrieves the hashes of every platform published for a provider version.
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

	registryConfig := tfregistry.Config{
		BaseURL:    fmt.Sprintf("https://%s/", provider.Hostname),
		HTTPClient: httpclient.NewRetryClient(),
	}

	newResource.lockIndex, err = lock.NewIndexFromConfig(registryConfig)
	if err != nil {
		return nil, err
	}

	newResource.registry, err = tfregistry.NewClient(registryConfig)
	if err != nil {
		return nil, err
	}

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

// UpdateAbsoluteFilePath resolves every file of the t.files map against the path resolver,
// leaving the path the user wrote available as originalFilePath for reporting.
func (t *TerraformLock) UpdateAbsoluteFilePath(pathResolver pathresolver.Resolver) error {
	for filePath := range t.files {
		f := t.files[filePath]

		// Join, not Resolve: absolute and parent directory paths are allowed, even with an scm.
		resolvedPath := pathResolver.Join(f.originalFilePath)

		if resolvedPath != f.filePath {
			log.Debugf("Relative path detected: changing from %q to %q", f.originalFilePath, resolvedPath)
		}

		f.filePath = resolvedPath
		t.files[filePath] = f
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (t *TerraformLock) Changelog(from, to string) *result.Changelogs {
	return nil
}

// openTofuRegistryHostname is the public OpenTofu registry, which publishes
// both the h1 and zh hashes of every platform in its provider package metadata.
const openTofuRegistryHostname = "registry.opentofu.org"

func (t *TerraformLock) getProviderHashes(version string) ([]string, error) {
	// Without platforms, record the h1 and zh hashes of every platform published
	// in the registry metadata, which is what "tofu init" writes for the OpenTofu registry.
	// It fails if the registry does not publish h1 hashes, such as the Terraform registry.
	if len(t.spec.Platforms) == 0 || t.provider.Hostname == openTofuRegistryHostname {
		hashes, err := t.getAllPlatformsHashes(context.Background(), version)
		if err == nil {
			return hashes, nil
		}

		if len(t.spec.Platforms) == 0 {
			return nil, fmt.Errorf("%s failed to query provider locks for provider: %q, version: %q, for all platforms: %s",
				result.FAILURE,
				t.spec.Provider,
				version,
				err.Error(),
			)
		}

		log.Warningf("failed to query provider locks for provider %q, version %q, for all platforms, falling back to platforms %q: %s",
			t.spec.Provider,
			version,
			t.spec.Platforms,
			err.Error(),
		)
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

// getAllPlatformsHashes returns the h1 and zh hashes of every platform published for a provider version.
// The package metadata of any platform lists the hashes of all of them, so it requests a platform
// published for this version rather than the runner's own, which the provider may not publish.
func (t *TerraformLock) getAllPlatformsHashes(ctx context.Context, version string) ([]string, error) {
	versions, err := t.registry.ListProviderVersions(ctx, &tfregistry.ListProviderVersionsRequest{
		Namespace: t.provider.Namespace,
		Type:      t.provider.Type,
	})
	if err != nil {
		return nil, err
	}

	var platform *tfregistry.ProviderPlatform
	for _, v := range versions.Versions {
		if v.Version == version && len(v.Platforms) > 0 {
			platform = &v.Platforms[0]
			break
		}
	}

	if platform == nil {
		return nil, fmt.Errorf("the registry publishes no platform for version %q", version)
	}

	metadata, err := t.registry.ProviderPackageMetadata(ctx, &tfregistry.ProviderPackageMetadataRequest{
		Namespace: t.provider.Namespace,
		Type:      t.provider.Type,
		Version:   version,
		OS:        platform.OS,
		Arch:      platform.Arch,
	})
	if err != nil {
		return nil, err
	}

	var hashes []string
	hasH1 := false
	for _, pkg := range metadata.Packages {
		for _, hash := range pkg.Hashes {
			if strings.HasPrefix(hash, "h1:") {
				hasH1 = true
			}
			hashes = append(hashes, hash)
		}
	}

	if !hasH1 {
		return nil, fmt.Errorf("the registry does not publish h1 hashes for version %q", version)
	}

	slices.Sort(hashes)

	return slices.Compact(hashes), nil
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
