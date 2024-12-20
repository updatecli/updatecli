package terragrunt

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/sirupsen/logrus"
	terraformAutoDiscovery "github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terraform"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

const (
	HclExtension string = ".hcl"
)

// searchTerragruntFiles looks, recursively, for every hcl files from a root directory.
// It will skip Terraform lock files and hcl files not containing terraform block
func searchTerragruntFiles(rootDir string) ([]string, error) {
	foundFiles := []string{}

	logrus.Debugf("Looking for Terragrunt modules in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		fileName := d.Name()

		if filepath.Ext(fileName) == HclExtension && d.Name() != terraformAutoDiscovery.TerraformLockFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}

func getTerragruntModule(filename string, allowNoVersion bool) (module *terragruntModule, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	hclfile, err := terraformUtils.ParseHcl(string(data), filename)
	if err != nil {
		return nil, err
	}

	for _, block := range hclfile.Body().Blocks() {
		if block.Type() == "terraform" {
			sourceBlock := block.Body().GetAttribute("source")
			if sourceBlock == nil {
				return nil, nil
			}
			quotedSource := strings.TrimSpace(string(sourceBlock.Expr().BuildTokens(nil).Bytes()))
			source, hclContext, err := evaluateHcl(quotedSource, data, filename)
			if err != nil {
				return nil, err
			}
			module, err = getModuleFromUrl(source, quotedSource, allowNoVersion)
			if hclContext != nil {
				module.hclContext = hclContext
			}
			if err != nil {
				return nil, err
			}
			return module, nil
		}
	}

	return module, nil
}

var moduleSourceLocalPrefixes = []string{
	"./",
	"../",
	".\\",
	"..\\",
	"/",
	"~",
}

func isModuleSourceLocal(raw string) bool {
	for _, prefix := range moduleSourceLocalPrefixes {
		if strings.HasPrefix(raw, prefix) {
			return true
		}
	}
	return false
}

func getSourceType(raw string) string {
	if isModuleSourceLocal(raw) {
		return SourceTypeLocal
	}
	if strings.HasPrefix(raw, "github.com") {
		return SourceTypeGithub
	}
	if strings.HasPrefix(raw, "git") {
		return SourceTypeGit
	}
	// Next step is to get the part before :/
	split := strings.Split(raw, "://")
	if len(split) > 1 {
		protocol := split[0]
		if protocol == "tfr" {
			return SourceTypeRegistry
		}
		if protocol == "https" || protocol == "http" {
			return SourceTypeHttp
		}
		if strings.HasPrefix(raw, "hg::http") {
			return SourceTypeMercurial
		}
		if strings.HasPrefix(raw, "s3::http") {
			return SourceTypeS3
		}
		if strings.HasPrefix(raw, "gcs::http") {
			return SourceTypeGCS
		}
	}

	return ""
}

func parseSourceUrl(evaluatedSource string, rawSource string, allowNoVersion bool) (terragruntModuleSource, error) {
	source := terragruntModuleSource{
		rawSource:       rawSource,
		evaluatedSource: evaluatedSource,
	}
	sourceType := getSourceType(evaluatedSource)
	if sourceType == "" {
		return source, fmt.Errorf("Could not get source type from source url %v", redact.URL(evaluatedSource))
	}
	source.sourceType = sourceType
	if sourceType == SourceTypeRegistry || sourceType == SourceTypeGit || sourceType == SourceTypeGithub {
		if sourceType == SourceTypeGit && strings.HasPrefix(evaluatedSource, "git") {
			evaluatedSource = strings.Replace(evaluatedSource, "git@", "git::ssh://", 1)
		}
		u, err := url.Parse(evaluatedSource)
		if err != nil {
			fmt.Printf("prevent panic by handling failure parsing url %q: %v\n", redact.URL(evaluatedSource), err)
			return source, err
		}
		params, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			fmt.Printf("prevent panic by handling failure parsing query %q: %v\n", u.RawQuery, err)
			return source, err
		}
		var param string
		if sourceType == SourceTypeRegistry {
			param = "version"
			source.baseUrl = fmt.Sprintf("%s%s", u.Host, u.Path)
		} else if sourceType == SourceTypeGit {
			// Git
			param = "ref"
			parts := strings.Split(fmt.Sprintf("%s:%s", u.Scheme, u.Opaque), "://")
			if len(parts) == 2 {
				source.protocol = parts[0]
				source.baseUrl = parts[1]
			}
		} else {
			// Github
			param = "ref"
			source.baseUrl = u.Path
		}
		version := strings.TrimPrefix(params.Get(param), "v")
		if version == "" && !allowNoVersion {
			return source, fmt.Errorf("Could not get version from source url %v", redact.URL(evaluatedSource))
		}
		source.version = version
		return source, nil

	}
	return source, nil
}

func getModuleFromUrl(evaluatedSource string, rawSource string, allowNoVersion bool) (module *terragruntModule, err error) {
	source, err := parseSourceUrl(evaluatedSource, rawSource, allowNoVersion)
	if err != nil {
		return nil, err
	}
	module = &terragruntModule{
		source: source,
	}
	if source.sourceType == SourceTypeRegistry {
		registryModule, err := terraformRegistryAddress.ParseModuleSource(source.baseUrl)
		if err != nil {
			fmt.Printf("prevent panic by handling failure parsing url %q: %v\n", redact.URL(source.baseUrl), err)
			return nil, err
		}
		module.registryModule = &registryModule
	} else if source.sourceType == SourceTypeGit {

	}
	return module, nil
}

func evaluateHcl(rawSource string, hclContent []byte, hclFileName string) (string, *map[string]string, error) {
	if (strings.HasPrefix(rawSource, `"`) && strings.HasSuffix(rawSource, `"`)) && !strings.ContainsAny(rawSource, "${}") {
		return strings.Trim(rawSource, `"`), nil, nil
	}
	// Evaluate expression
	sourceExpr, diags := hclsyntax.ParseExpression([]byte(rawSource), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("failed to construct expression from source %q: %v", rawSource, diags.Error())
	}
	// Parse the HCL Configuration
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(hclContent, hclFileName)
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("failed to parse hcl %q: %v", hclFileName, diags.Error())
	}
	// Extract locals and evaluate their values
	localsContent, _, localsDiags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "locals"},
		},
	})
	if localsDiags.HasErrors() {
		return "", nil, fmt.Errorf("failed to get locals content %q: %v", hclFileName, diags.Error())
	}
	localsValue := make(map[string]cty.Value)
	for _, block := range localsContent.Blocks {
		blockContent, diags := block.Body.JustAttributes()

		if localsDiags.HasErrors() {
			return "", nil, fmt.Errorf("failed to get locals block content %q: %v", hclFileName, diags.Error())
		}
		for attrName, attr := range blockContent {
			attrValue, diags := attr.Expr.Value(&hcl.EvalContext{})
			if diags.HasErrors() {
				logrus.Debugf("Skipping attribute %s: %v", attrName, diags.Error())
				continue
			}
			localsValue[attrName] = attrValue
		}
	}
	// Create the evaluation Context
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"local": cty.ObjectVal(localsValue),
		},
	}
	sourceValue, diags := sourceExpr.Value(ctx)
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("failed to evaluate source expression %q: %v", rawSource, diags.Error())
	}
	stringValues := make(map[string]string)
	for localName, localValue := range localsValue {
		if localValue.Type() == cty.String {
			stringValues[localName] = localValue.AsString()
		} else {
			var strVal string
			err := gocty.FromCtyValue(localValue, &strVal)
			if err == nil {
				stringValues[localName] = strVal
			}
		}
	}
	return sourceValue.AsString(), &stringValues, nil
}
