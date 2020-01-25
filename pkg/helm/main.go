package helm

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/git"
	"gopkg.in/yaml.v3"
)

var (
	yamlIdent int = 2
)

// Helm stores configuration about the file and the key value that needs to be updated
type Helm struct {
	File    string
	Key     string
	Message string
	Git     git.Git
}

func searchAndUpdateVersion(entry *yaml.Node, keys []string, version string, columnRef int) (found bool, oldVersion string, column int) {

	valueFound := false
	column = columnRef
	nextLevel := false
	for _, content := range entry.Content {
		if content.Column < column {
			break
		}
		if nextLevel {
			column = content.Column
			nextLevel = false
		}

		if content.Value == keys[0] && (content.Column == columnRef) {
			column = content.Column
			nextLevel = true

			if len(keys) > 1 {
				keys = keys[1:]
			} else if len(keys) == 1 {
				valueFound = true
				continue
			}
		}

		if content.Kind == yaml.ScalarNode && valueFound == true {
			column = content.Column

			if version != content.Value {
				log.Printf("Version mismatched between %s (old) and %s (new)", content.Value, version)
				oldVersion = content.Value
				content.SetString(version)
			} else if version == content.Value {
				log.Printf("Version already set to %v", content.Value)
				oldVersion = content.Value
				content.SetString(version)
			} else {
				log.Printf("Something weird happened while comparing old and new version")
			}
			break
		} else if content.Kind == yaml.MappingNode {
			valueFound, oldVersion, column = searchAndUpdateVersion(content, keys, version, column)
		}
	}
	return valueFound, oldVersion, column
}

// UpdateChart reads and updates helm chart value
func (helm *Helm) UpdateChart(version string) {

	helm.Git.Init()

	path := filepath.Join(helm.Git.Directory, helm.File)

	helm.Git.Clone()

	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := searchAndUpdateVersion(&out, strings.Split(helm.Key, "."), version, 1)

	if valueFound != true {
		log.Printf("cannot find key '%v' in file %v", helm.Key, path)
		return
	}

	if oldVersion == version {
		log.Printf("Value %v at %v already up to date", helm.Key, path)
		return
	}

	message := fmt.Sprintf("Update key '%v' to %s",
		helm.Key,
		version)

	log.Printf("%s\n", message)

	newFile, err := os.Create(path)
	defer newFile.Close()

	encoder := yaml.NewEncoder(newFile)
	defer encoder.Close()
	encoder.SetIndent(yamlIdent)
	err = encoder.Encode(&out)

	if err != nil {
		log.Fatalf("Something went wrong while encoding %v", err)
	}

	helm.Git.Add(helm.File)
	helm.Git.Commit(helm.File, message)
	helm.Git.Push()
	helm.Git.Clean()
}
