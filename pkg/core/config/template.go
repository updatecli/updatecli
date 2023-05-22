package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// Template contains template information used to generate updatecli configuration struct
type Template struct {
	CfgFile      string                 // Specify updatecli configuration file
	ValuesFiles  []string               // Specify value filename
	SecretsFiles []string               // Specify sops secrets filename
	Values       map[string]interface{} `yaml:"-,inline"` // Contains key/value extracted from a yaml file
	Secrets      map[string]interface{} `yaml:"-,inline"` // Contains mozilla/sops information using yaml format
	fs           fs.FS
}

// Init parses a golang template then return an updatecli configuration as a struct
func (t *Template) New(content []byte) ([]byte, error) {
	funcMap := template.FuncMap{
		// Retrieve value from environment variable, return error if not found
		"requiredEnv": func(s string) (string, error) {
			value := os.Getenv(s)
			if value == "" {
				return "", errors.New("no value found for environment variable " + s)
			}
			return value, nil
		},
		"pipeline": func(s string) (string, error) {
			return fmt.Sprintf(`{{ pipeline %q }}`, s), nil
		},
		"source": func(s string) (string, error) {
			return fmt.Sprintf(`{{ source %q }}`, s), nil
		},
	}

	err := t.readValuesFiles()

	if err != nil {
		return []byte{}, err
	}

	err = t.readSecretsFiles()

	if err != nil {
		return []byte{}, err
	}

	// Merge yaml configuration and sops secrets into one configuration
	// Deepmerge is not supported so a secrets override unencrypted values
	templateValues := Merge(t.Values, t.Secrets)

	tmpl, err := template.New("cfg").
		Funcs(sprig.FuncMap()).
		Funcs(funcMap). // add custom funcMap last so that it takes precedence
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

func (t *Template) readValuesFiles() error {
	// Read every files containing yaml key/values
	for _, valuesFile := range t.ValuesFiles {
		err := t.readFile(valuesFile, &t.Values, false)

		// Stop early, no need to lead more values files
		// if something went wrong with at least one
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) readSecretsFiles() error {
	// Read every files containing sops secrets using the yaml format
	// Order matter, last element always override
	for _, secretsFile := range t.SecretsFiles {
		err := t.readFile(secretsFile, &t.Secrets, true)

		// Stop early, no need to lead more secrets files
		// if something went wrong with at least one.
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadFile reads an updatecli values file, it can also read encrypted sops files
func (t *Template) readFile(filename string, values *map[string]interface{}, encrypted bool) (err error) {

	baseFilename := filepath.Base(filename)

	if extension := filepath.Ext(baseFilename); strings.Compare(extension, ".yml") != 0 &&
		strings.Compare(extension, ".yaml") != 0 {
		err = fmt.Errorf("wrong file extension %q for file %q", extension, baseFilename)
		logrus.Errorln(err)
		return err
	}

	if filename == "" {
		fmt.Println("No filename defined, nothing else to do")
		return nil
	}

	v, err := t.fs.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return err
	}

	if encrypted {
		content, err = decrypt.Data(content, "yaml")
		if err != nil {
			return err
		}
	}

	err = yaml.Unmarshal(content, &values)

	return err
}

// Merge merges one are multiple updatecli value files content into one
func Merge(valuesFiles ...map[string]interface{}) (results map[string]interface{}) {

	results = make(map[string]interface{})

	for _, values := range valuesFiles {
		for k, v := range values {
			results[k] = v
		}
	}

	return results
}
