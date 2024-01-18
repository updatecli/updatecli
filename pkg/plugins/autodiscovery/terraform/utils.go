package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
)

const (
	TerraformLockFile string = ".terraform.lock.hcl"
)

// searchTerraformLockFiles looks, recursively, for every files named .terraform.lock.hcl from a root directory.
func searchTerraformLockFiles(rootDir string) ([]string, error) {
	foundFiles := []string{}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if d.Name() == TerraformLockFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}

func getTerraformLockContent(filename string) (providers map[string]string, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return providers, err
	}

	lockfile, err := terraformUtils.ParseHcl(string(data), filename)
	if err != nil {
		return nil, err
	}

	providers = make(map[string]string)
	for _, block := range lockfile.Body().Blocks() {
		if block.Type() == "provider" {
			name := block.Labels()[0]
			quotedValue := strings.TrimSpace(string(block.Body().GetAttribute("version").Expr().BuildTokens(nil).Bytes()))
			version := strings.Trim(quotedValue, `"`)
			providers[name] = version
		}
	}

	return providers, nil
}
