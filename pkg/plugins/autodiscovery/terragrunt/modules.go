package terragrunt

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/sirupsen/logrus"
)

const (
	SourceTypeLocal     = "local"
	SourceTypeRegistry  = "registry"
	SourceTypeGithub    = "github"
	SourceTypeGit       = "git"
	SourceTypeHttp      = "http"
	SourceTypeS3        = "s3"
	SourceTypeGCS       = "gcs"
	SourceTypeMercurial = "mercurial"
)

type terragruntModuleSource struct {
	sourceType string
	protocol   string
	baseUrl    string
	version    string
	rawSource  string
}

type terragruntModule struct {
	registryModule *terraformRegistryAddress.Module
	source         terragruntModuleSource
	hclContext     *map[string]string
}

func (t *terragruntModule) ForDisplay() string {
	if t.source.sourceType == SourceTypeRegistry {
		return t.registryModule.ForDisplay()
	} else if t.source.sourceType == SourceTypeGit || t.source.sourceType == SourceTypeGithub {
		// Need to make sure we drop the version from it and the leading protocol info
		return t.source.baseUrl
	}
	return t.source.rawSource
}

func (t Terragrunt) discoverTerragruntManifests() ([][]byte, error) {
	var manifests [][]byte

	foundFiles, err := searchTerragruntFiles(t.rootDir)
	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(t.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		module, err := getTerragruntModule(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		if module == nil {
			logrus.Debugf("Skipping %q: no valid module found in it", foundFile)
			continue
		}

		// // Test if the ignore rule based on path is respected
		// if len(t.spec.Ignore) > 0 {
		// 	if t.spec.Ignore.isMatchingRules(t.rootDir, relativeFoundFile, provider, providerVersion) {
		// 		logrus.Debugf("Ignoring provider %q from file %q, as matching ignore rule(s)\n", provider, relativeFoundFile)
		// 		continue
		// 	}
		// }

		// // Test if the only rule based on path is respected
		// if len(t.spec.Only) > 0 {
		// 	if !t.spec.Only.isMatchingRules(t.rootDir, relativeFoundFile, provider, providerVersion) {
		// 		logrus.Debugf("Ignoring provider %q from %q, as not matching only rule(s)\n", provider, relativeFoundFile)
		// 		continue
		// 	}
		// }

		// versionPattern, err := t.versionFilter.GreaterThanPattern(providerVersion)
		// if err != nil {
		// 	logrus.Debugf("skipping provider %q due to: %s", provider, err)
		// 	continue
		// }

		moduleManifest, err := t.getTerragruntManifest(
			relativeFoundFile,
			module,
		)
		if err != nil {
			logrus.Errorf("skipping module %q due to: %s", module.source, err)
			logrus.Debugf("skipping module %q due to: %s", module.source, err)
			continue
		}

		manifests = append(manifests, moduleManifest)

	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

func (t Terragrunt) getTerragruntManifest(filename string, module *terragruntModule) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(terragruntModuleManifestTemplate)
	if err != nil {
		return nil, err
	}

	var sourceTypeKind string
	var ModuleHost string
	var ModuleNameSpace string
	var ModuleName string
	var ModuleNameTargetSystem string

	var ModuleSourceScm string
	var ModuleSourceScmUrl string

	if module.source.sourceType == SourceTypeRegistry {
		sourceTypeKind = "terraform/registry"
		ModuleHost = module.registryModule.Package.Host.String()
		ModuleNameSpace = module.registryModule.Package.Namespace
		ModuleName = module.registryModule.Package.Name
		ModuleNameTargetSystem = module.registryModule.Package.TargetSystem
	} else if module.source.sourceType == SourceTypeGit {
		sourceTypeKind = "gitag"
		ModuleSourceScm = "module"
		ModuleSourceScmUrl = fmt.Sprintf("%s://%s", module.source.protocol, module.source.baseUrl)
	} else if module.source.sourceType == SourceTypeGithub {
		sourceTypeKind = "github"
	} else {
		return nil, fmt.Errorf("Unsupported source type: %q", module.source.sourceType)
	}

	params := struct {
		TerragruntModulePath string
		SourceType           string
		SourceTypeKind       string
		Module               string
		ModuleHost           string
		ModuleNamespace      string
		ModuleName           string
		ModuleTargetSystem   string
		ModuleSourceScm      string
		ModuleSourceScmUrl   string
		ScmID                string
		TargetName           string
		TargetPath           string
	}{
		TerragruntModulePath: filename,
		Module:               module.ForDisplay(),
		SourceType:           module.source.sourceType,
		SourceTypeKind:       sourceTypeKind,
		ModuleHost:           ModuleHost,
		ModuleNamespace:      ModuleNameSpace,
		ModuleName:           ModuleName,
		ModuleTargetSystem:   ModuleNameTargetSystem,
		ModuleSourceScm:      ModuleSourceScm,
		ModuleSourceScmUrl:   ModuleSourceScmUrl,
		ScmID:                t.scmID,
		TargetPath:           "terraform.source",
		TargetName:           fmt.Sprintf("Bump %s to {{ source \"latestVersion\" }}", module.ForDisplay()),
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
