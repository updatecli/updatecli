package jsonschema

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jschema "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"gopkg.in/yaml.v2"
)

// jsconSchemaRootDir contains the path to access json schema from updatecli repository
const jsonSchemaRootDir string = "../../../../schema"

// getFilesWithSuffix search for every json schema from a root  directory
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

func Validate(manifest []byte) (bool, error) {
	var m interface{}
	err := yaml.Unmarshal(manifest, &m)
	if err != nil {
		return false, err
	}
	m, err = toStringKeys(m)
	if err != nil {
		return false, err
	}
	schema, err := loadJsonSchema()
	if err != nil {
		return false, err
	}

	if err := schema.Validate(m); err != nil {
		fmt.Printf("%#v\n", err)
		return false, err
	}

	return true, nil
}

func loadJsonSchema() (*jschema.Schema, error) {
	compiler := jschema.NewCompiler()
	//compiler.Draft = jschema.Draft2020

	jsonSchemaFiles, err := getFilesWithSuffix(jsonSchemaRootDir, ".json")
	if err != nil {
		logrus.Errorf("%s", err)
		return nil, err
	}

	for _, jsFile := range jsonSchemaFiles {

		// Removing
		jsID := strings.TrimPrefix(jsFile, jsonSchemaRootDir)
		jsID = strings.TrimSuffix(jsID, ".json")

		logrus.Debugf("Loading json schema %q", jsID)
		fmt.Printf("Loading json schema %q from %q\n", jsID, jsFile)

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

	return compiler.Compile(filepath.Join(jsonSchemaRootDir, "pipeline.json"))

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
