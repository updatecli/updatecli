package config

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"dario.cat/mergo"
	"github.com/Masterminds/sprig/v3"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Template contains template information used to generate updatecli configuration struct
type Template struct {
	// CfgFile is the updatecli configuration file
	CfgFile string
	// ValuesFiles contains one or multiple yaml files containing key/values
	ValuesFiles []string
	// SecretsFiles contains one or multiple sops files containing secrets
	SecretsFiles []string
	// Values contains key/value extracted from a values file
	Values map[string]interface{} `yaml:"-,inline"`
	// Secrets contains key/value extracted from a sops file
	Secrets map[string]interface{} `yaml:"-,inline"`
	// fs is a file system abstraction used to read files
	fs fs.FS
}

// Init parses a golang template then return an updatecli configuration as a struct
func (t *Template) New(content []byte) ([]byte, error) {
	err := t.readValuesFiles(t.ValuesFiles, false)
	if err != nil {
		return []byte{}, err
	}

	err = t.readValuesFiles(t.SecretsFiles, true)
	if err != nil {
		return []byte{}, err
	}

	// Merge yaml configuration and sops secrets into one configuration
	templateValues := mergeValueFile(t.Values, t.Secrets)

	tmpl, err := template.New("cfg").
		Funcs(sprig.FuncMap()).
		Funcs(helmFuncMap()).      // add helm funcMap
		Funcs(updatecliFuncMap()). // add custom funcMap last so that it takes precedence
		Parse(string(content))

	if err != nil {
		return []byte{}, err
	}

	b := bytes.Buffer{}

	if err := tmpl.Execute(&b, templateValues); err != nil {
		return []byte{}, err
	}

	return b.Bytes(), nil
}

// readValuesFiles reads one or multiple updatecli values files and merge them into one
func (t *Template) readValuesFiles(valueFiles []string, encrypted bool) error {
	// Read every files containing yaml key/values
	for _, valueFile := range valueFiles {
		var v map[string]interface{}
		err := t.readFile(valueFile, &v, encrypted)
		// Stop early, no need to lead more values files
		// if something went wrong with at least one
		if err != nil {
			return err
		}

		// Merge yaml configuration and sops secrets into different variable
		switch encrypted {
		case true:
			t.Secrets = mergeValueFile(t.Secrets, v)
		case false:
			t.Values = mergeValueFile(t.Values, v)
		}
	}
	return nil
}

// ReadFile reads an updatecli values file, it can also read encrypted sops files
func (t *Template) readFile(filename string, values *map[string]interface{}, encrypted bool) (err error) {

	baseFilename := filepath.Base(filename)
	extension := filepath.Ext(baseFilename)

	// Check if the file extension is either yaml or yml
	if strings.Compare(extension, ".yml") != 0 &&
		strings.Compare(extension, ".yaml") != 0 &&
		strings.Compare(extension, ".json") != 0 {
		err = fmt.Errorf("wrong file extension %q for file %q", extension, baseFilename)
		logrus.Errorln(err)
		return err
	}

	if filename == "" {
		fmt.Println("No filename defined, nothing else to do")
		return nil
	}

	// I am struggling to find a way to mock the file system for the unit test
	// when file is not in the current directory
	// So I am using a condition to make sure that the unit test work
	if filename != baseFilename {
		t.fs = os.DirFS(filepath.Dir(filename))
	}

	v, err := t.fs.Open((baseFilename))
	if err != nil {
		return err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return err
	}

	if encrypted {
		switch extension {
		case ".yaml", ".yml":
			content, err = decrypt.Data(content, "yaml")
			if err != nil {
				return err
			}
		case ".json":
			content, err = decrypt.Data(content, "json")
			if err != nil {
				return err
			}
		default:
			err = fmt.Errorf("wrong file extension %q for file %q", extension, baseFilename)
		}
	}

	err = yaml.Unmarshal(content, &values)

	return err
}

// mergeValueFile merges one or multiple updatecli value files content into one
func mergeValueFile(valuesFiles ...map[string]interface{}) (results map[string]interface{}) {

	results = make(map[string]interface{})

	for _, values := range valuesFiles {
		if err := mergo.Merge(&results, values, mergo.WithOverride); err != nil {
			err = fmt.Errorf("merging values files: %w", err)
			logrus.Errorln(err)
		}
	}

	return results
}
