package terragrunt

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
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
	sourceType      string
	protocol        string
	baseUrl         string
	version         string
	rawSource       string
	evaluatedSource string
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

type terragruntTransformer struct {
	Kind  string
	Value string
}

func (t Terragrunt) discoverTerragruntManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := t.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if t.spec.RootDir != "" && !path.IsAbs(t.spec.RootDir) {
		searchFromDir = filepath.Join(t.rootDir, t.spec.RootDir)
	}

	foundFiles, err := searchTerragruntFiles(searchFromDir)
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
		module, err := getTerragruntModule(foundFile, false)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		if module == nil {
			logrus.Debugf("Skipping %q: no valid module found in it", foundFile)
			continue
		}

		// Test if the ignore rule based on path is respected
		if len(t.spec.Ignore) > 0 {
			ignored, err := t.spec.Ignore.isMatchingRules(t.rootDir, relativeFoundFile, module)
			if err != nil {
				logrus.Debugf("skipping module %q due to: %s", module.source, err)
				continue
			}

			if ignored {
				logrus.Debugf("Ignoring file %q, as matching ignore rule(s)\n", relativeFoundFile)
				continue
			}
		}

		// Test if the only rule based on path is respected
		if len(t.spec.Only) > 0 {
			ignored, err := t.spec.Only.isMatchingRules(t.rootDir, relativeFoundFile, module)
			if err != nil {
				logrus.Debugf("skipping module %q due to: %s", module.source, err)
				continue
			}

			if !ignored {
				logrus.Debugf("Ignoring provider file %q, as not matching only rule(s)\n", relativeFoundFile)
				continue
			}
		}

		versionPattern, err := t.versionFilter.GreaterThanPattern(module.source.version)
		if err != nil {
			logrus.Debugf("skipping file %q due to: %s", relativeFoundFile, err)
			continue
		}

		moduleManifest, err := t.getTerragruntManifest(
			relativeFoundFile,
			module,
			versionPattern,
		)
		if err != nil {
			logrus.Debugf("skipping module %q due to: %s", module.source, err)
			continue
		}

		manifests = append(manifests, moduleManifest)

	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

func mapkey(m map[string]string, value string) (key string, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

func (t Terragrunt) getTerragruntManifest(filename string, module *terragruntModule, versionFilterPattern string) ([]byte, error) {
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

	var transformers []terragruntTransformer
	targetPath := "terraform.source"

	if module.source.sourceType == SourceTypeRegistry {
		sourceTypeKind = "terraform/registry"
		ModuleHost = module.registryModule.Package.Host.String()
		ModuleNameSpace = module.registryModule.Package.Namespace
		ModuleName = module.registryModule.Package.Name
		ModuleNameTargetSystem = module.registryModule.Package.TargetSystem

	} else if module.source.sourceType == SourceTypeGit {
		sourceTypeKind = "gittag"
		ModuleSourceScm = "module"
		ModuleSourceScmUrl = strings.Replace(
			fmt.Sprintf("%s://%s", module.source.protocol, module.source.baseUrl),
			"git::",
			"",
			1)
	} else {
		return nil, fmt.Errorf("Unsupported source type: %q", module.source.sourceType)
	}
	prefix := ""
	if module.hclContext == nil || len(*module.hclContext) == 0 {
		// The source is either inlined  we'll need to add a prefix to the version
		prefix = strings.Replace(strings.Trim(module.source.rawSource, `"`), module.source.version, "", 1)
	} else if module.hclContext != nil {
		// This means the source contains parameter(s)
		// source could be:
		// - local.base_url
		// - tfr://${local.module}?version=${local.module_version}
		// - tfr://someModule?version=${local.module_version}
		// - tfr://${local.module}?version=1.2.3
		if !strings.ContainsAny(module.source.rawSource, "${}") {
			// local.base_url
			// Simple case of adding a prefix and changing the base path
			key := strings.TrimPrefix(module.source.rawSource, "local.")
			value, ok := (*module.hclContext)[key]
			if !ok {
				return nil, fmt.Errorf("Could not infer which local to update base on rawSource %q", module.source.rawSource)
			}
			// Tf defines at locals but access at local
			targetPath = strings.Replace(module.source.rawSource, "local", "locals", 1)
			prefix = strings.Replace(value, module.source.version, "", 1)
		} else {
			// Let's find if one of the key olds the version
			key, ok := mapkey(*module.hclContext, module.source.version)
			if !ok {
				// we are in the tfr://${local.module}?version=1.2.3 case
				prefix = strings.Replace(strings.Trim(module.source.rawSource, `"`), module.source.version, "", 1)
			} else {
				// we just need to update the version store in this key
				targetPath = fmt.Sprintf("locals.%s", key)
			}
		}

	} else {
		return nil, fmt.Errorf("Non inline source but no locals found %q", module.source.rawSource)

	}
	if prefix != "" {
		if module.source.sourceType == SourceTypeGit {
			// let's remove trailing v, tf doesn't care for it and git tag source will give it back
			// if there is none, it's a no op
			prefix = strings.TrimSuffix(strings.Trim(prefix, `"`), "v")
		}

		transformers = append(transformers, terragruntTransformer{
			Kind:  "addprefix",
			Value: prefix,
		})

	}

	params := struct {
		ActionID             string
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
		Transformers         []terragruntTransformer
		VersionFilterKind    string
		VersionFilterPattern string
	}{
		ActionID:             t.actionID,
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
		TargetPath:           targetPath,
		TargetName:           fmt.Sprintf("deps: bump %s to {{ source \"latestVersion\" }}", module.ForDisplay()),
		Transformers:         transformers,
		VersionFilterKind:    t.versionFilter.Kind,
		VersionFilterPattern: versionFilterPattern,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
