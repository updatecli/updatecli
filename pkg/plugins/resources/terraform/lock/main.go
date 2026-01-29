package lock

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/minamijoyo/tfupdate/lock"
	"github.com/minamijoyo/tfupdate/tfregistry"
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

	lockIndex, err := lock.NewIndexFromConfig(tfregistry.Config{
		BaseURL: fmt.Sprintf("https://%s/", provider.Hostname),
	})
	if err != nil {
		return nil, err
	}

	newResource.lockIndex = lockIndex

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
