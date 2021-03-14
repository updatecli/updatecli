package config

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Template contains template information
type Template struct {
	Values     map[string]interface{} `yaml:"-,inline"`
	ValuesFile string
	CfgFile    string
}

// Init parse golang templates then return its config struct
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
	}

	c, err := os.Open(t.CfgFile)

	defer c.Close()
	if err != nil {
		return err
	}

	content := []byte{}

	if _, err := os.Stat(t.ValuesFile); err == nil && t.ValuesFile != "" {

		if t.ValuesFile != "" {
			v, err := os.Open(t.ValuesFile)
			defer v.Close()
			if err != nil {
				return err
			}

			content, err = ioutil.ReadAll(v)
			if err != nil {
				return err
			}
		}
	} else if err != nil && t.ValuesFile != "" {
		return err
	}

	err = yaml.Unmarshal(content, &t.Values)
	if err != nil {
		return err
	}

	content, err = ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	tmpl, err := template.New("cfg").Funcs(funcMap).Parse(string(content))

	if err != nil {
		return err
	}

	b := bytes.Buffer{}

	if err := tmpl.Execute(&b, t.Values); err != nil {
		return err
	}

	err = yaml.Unmarshal(b.Bytes(), &config)
	if err != nil {
		return err
	}

	return nil
}
