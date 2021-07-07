package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"go.mozilla.org/sops/decrypt"
	"gopkg.in/yaml.v3"
)

// Template contains template information used to generate updatecli configuration struct
type Template struct {
	CfgFile      string                 // Specify updatecli configuration file
	ValuesFiles  []string               // Specify value filename
	SecretsFiles []string               // Specify sops secrets filename
	Values       map[string]interface{} `yaml:"-,inline"` // Contains key/value extracted from a yaml file
	Secrets      map[string]interface{} `yaml:"-,inline"` // Contains mozilla/sops information using yaml format
}

// Init parses a golang template then return an updatecli configuration as a struct
func (t *Template) Init(config *Config) error {
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
	}

	c, err := os.Open(t.CfgFile)

	defer c.Close()
	if err != nil {
		return err
	}

	// Read every files containing yaml key/values
	for _, valuesFile := range t.ValuesFiles {
		err = ReadFile(valuesFile, &t.Values, false)

		if err != nil {
			return err
		}
	}

	// Read every files containing sops secrets using the yaml format
	// Order matter, last element always override
	for _, secretsFile := range t.SecretsFiles {
		err = ReadFile(secretsFile, &t.Secrets, true)

		if err != nil {
			return err
		}
	}

	// Merge yaml configuration and sops secrets into one configuration
	// Deepmerge is not supported so a secrets override unencrypted values
	templateValues := Merge(t.Values, t.Secrets)

	content, err := ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	tmpl, err := template.New("cfg").Funcs(funcMap).Parse(string(content))

	if err != nil {
		return err
	}

	b := bytes.Buffer{}

	if err := tmpl.Execute(&b, templateValues); err != nil {
		return err
	}

	err = yaml.Unmarshal(b.Bytes(), &config)
	if err != nil {
		return err
	}

	return nil
}

// ReadFile reads an udpatecli values file, it can also read encrypted sops files
func ReadFile(filename string, values *map[string]interface{}, encrypted bool) (err error) {

	baseFilename := filepath.Base(filename)

	if extension := filepath.Ext(baseFilename); strings.Compare(extension, ".yaml") != 0 ||
		strings.Compare(extension, ".yml") != 0 &&
			strings.Compare(extension, ".yaml") != 0 {
		err = fmt.Errorf("wrong file extension %q for file %q", extension, baseFilename)
		logrus.Errorln(err)
		return err
	}
	if filename == "" {
		fmt.Println("No filename defined, nothing else to do")
		return nil
	}

	if _, err := os.Stat(filename); err != nil {
		return err
	}

	v, err := os.Open(filename)
	defer v.Close()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(v)
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
