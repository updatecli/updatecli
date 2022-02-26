package jsonschema

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jschema "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"gopkg.in/yaml.v2"
)

// Config specify updatecli configuration for validating its manifest
type Config struct {
	// MainJsonSchema defines the principal Json schema file
	// all relative schema will use the mainschema dirname
	MainJsonSchema string
	// UpdatecliConfiguration specifies an updatecli file, which could either
	// a regular file or a directory
	UpdatecliConfiguration string
}

// getFilesWithSuffix searches for every files with a matching suffix,
// from a specified root directory. This function is both used to collection and json schema
// and updatecli configuration file
func getFilesWithSuffix(root, suffix string) ([]string, error) {

	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
			return err
		}
		if info.Mode().IsRegular() {
			if strings.HasSuffix(path, suffix) {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, nil
	}

	return files, err
}

// readFile return a file content.
func readFile(file string) ([]byte, error) {
	c, err := os.Open(file)

	if err != nil {
		return nil, err
	}

	defer c.Close()

	content, err := ioutil.ReadAll(c)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// Validate tests that updatecli manifest respect updatecli JSON schema
func (c *Config) Validate() (bool, error) {

	updatecliManifests := []string{}
	// Look for updatecli manifest matching suffix
	for _, suffix := range []string{".yaml", ".yml", ".tpl"} {
		m, err := getFilesWithSuffix(c.UpdatecliConfiguration, suffix)
		if err != nil {
			return false, err
		}
		updatecliManifests = append(updatecliManifests, m...)
	}

	for _, manifest := range updatecliManifests {
		logrus.Infof("Validating file: %q", manifest)
		m, err := readFile(manifest)
		if err != nil {
			return false, err
		}
		err = c.validateYAML(m)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return false, err
		}
	}

	return true, nil
}

func (c *Config) validateYAML(manifest []byte) error {
	var m interface{}
	err := yaml.Unmarshal(manifest, &m)
	if err != nil {
		return err
	}
	m, err = toStringKeys(m)
	if err != nil {
		return err
	}
	s, err := c.loadJsonSchema()
	if err != nil {
		return err
	}

	if err := s.Validate(m); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadJsonSchema() (*jschema.Schema, error) {

	jsonSchemaRootDir := filepath.Dir(c.MainJsonSchema)
	configSchema := filepath.Base(c.MainJsonSchema)

	compiler := jschema.NewCompiler()

	jsonSchemaFiles, err := getFilesWithSuffix(jsonSchemaRootDir, ".json")
	if err != nil {
		logrus.Errorf("%s", err)
		return nil, err
	}

	for _, jsFile := range jsonSchemaFiles {

		// Removing
		jsID := strings.TrimPrefix(jsFile, jsonSchemaRootDir)
		jsID = strings.TrimSuffix(jsID, ".json")

		file, err := os.Open(jsFile)

		if err != nil {
			logrus.Errorf("%q - %s", jsID, err)
			continue
		}

		defer file.Close()

		if err := compiler.AddResource(jsID, bufio.NewReader(file)); err != nil {
			logrus.Errorf("%q - %s", jsID, err)
			continue
		}

	}

	return compiler.Compile(filepath.Join(jsonSchemaRootDir, configSchema))

}

// since yaml supports non-string keys, such yaml documents are rendered as invalid json documents.
// yaml parser returns map[interface{}]interface{} for object, whereas json parser returns map[string]interafce{}.
// this package accepts only map[string]interface{}, so we need to manually convert them to map[string]interface{}
func toStringKeys(val interface{}) (interface{}, error) {
	var err error
	switch val := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			k, ok := k.(string)
			if !ok {
				return nil, errors.New("found non-string key")
			}
			m[k], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case []interface{}:
		var l = make([]interface{}, len(val))
		for i, v := range val {
			l[i], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return l, nil
	default:
		return val, nil
	}
}
