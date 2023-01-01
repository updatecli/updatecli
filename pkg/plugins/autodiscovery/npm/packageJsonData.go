package npm

import (
	"encoding/json"
	"os"
)

// packageJsonData represent the struct content of package.json
type packageJsonData struct {
	Name            string
	Version         string
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

// loadPackageJsonData read a file an return its content
func loadPackageJsonData(filename string) (*packageJsonData, error) {

	rawFileContent, _ := os.ReadFile(filename)
	var data packageJsonData

	err := json.Unmarshal(rawFileContent, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
