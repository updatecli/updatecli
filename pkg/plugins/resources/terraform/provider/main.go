package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

type TerraformProvider struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]file // map of file paths to file contents
}

type file struct {
	originalFilePath string
	filePath         string
	content          string
}

func New(spec interface{}) (*TerraformProvider, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &TerraformProvider{
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

	return newResource, nil
}

func (t *TerraformProvider) Query(resourceFile file) (string, error) {
	file, err := terraformUtils.ParseHcl(resourceFile.content, resourceFile.originalFilePath)
	if err != nil {
		return "", err
	}

	var version string

	for _, tf := range file.Body().Blocks() {
		if tf.Type() == "terraform" {
			providers := tf.Body().FirstMatchingBlock("required_providers", []string{})
			if providers == nil {
				continue
			}

			attr := providers.Body().GetAttribute(t.spec.Provider)
			if attr == nil {
				continue
			}

			// Here we iterate of tokens, when we find version as TokenIdent we get the next TokenQuotedLit as value
			next := false
			for _, tokens := range attr.Expr().BuildTokens(nil) {
				if tokens.Type == hclsyntax.TokenIdent {
					if string(tokens.Bytes) == "version" {
						next = true
						continue
					}
				}
				if tokens.Type == hclsyntax.TokenQuotedLit && next {
					version = string(tokens.Bytes)
					break
				}
			}
		}
	}

	if version == "" {
		err := fmt.Errorf("%s cannot find value for %q from file %q",
			result.FAILURE,
			t.spec.Provider,
			resourceFile.originalFilePath)
		return "", err
	}

	return version, nil
}

func (t *TerraformProvider) Apply(filePath string, versionToWrite string) error {
	resourceFile := t.files[filePath]

	file, err := terraformUtils.ParseHcl(resourceFile.content, resourceFile.originalFilePath)
	if err != nil {
		return err
	}

	updater, err := tfupdate.NewProviderUpdater(t.spec.Provider, versionToWrite)
	if err != nil {
		return err
	}

	// Second arguments not used downstream
	if err := updater.Update(context.Background(), nil, resourceFile.originalFilePath, file); err != nil {
		return err
	}

	resourceFile.content = string(hclwrite.Format(file.BuildTokens(nil).Bytes()))

	t.files[filePath] = resourceFile

	return nil
}

// Read puts the content of the file(s) as value of the y.files map if the file(s) exist(s) or log the non existence of the file
func (t *TerraformProvider) Read() error {
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

func (t *TerraformProvider) UpdateAbsoluteFilePath(workDir string) {
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
func (t *TerraformProvider) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new resource configuration with only the necessary configuration fields without any sensitive information
// or context specific data.
func (t *TerraformProvider) ReportConfig() interface{} {
	return Spec{
		Provider: t.spec.Provider,
		File:     t.spec.File,
		Files:    t.spec.Files,
		Value:    t.spec.Value,
	}
}
