package config

/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

/*
	functions from this file are from the helm project which IMHO should be in the
	masterminds/sprig library. cfr https://github.com/Masterminds/sprig/pull/360
	Once they are available from the sprig library, we can remove them from here.
*/

/*
toYAML takes an interface, marshals it to yaml, and returns a string. It will
always return a string, even on marshal error (empty string).

This is designed to be called from a template.
*/
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

/*
fromYAML converts a YAML document into a map[string]interface{}.

This is not a general-purpose YAML parser, and will not parse all valid
YAML documents. Additionally, because its intended use is within templates
it tolerates errors. It will insert the returned error message string into
m["Error"] in the returned map.
*/
func fromYAML(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

/*
fromYAMLArray converts a YAML array into a []interface{}.

This is not a general-purpose YAML parser, and will not parse all valid
YAML documents. Additionally, because its intended use is within templates
it tolerates errors. It will insert the returned error message string as
the first and only item in the returned array.
*/
func fromYAMLArray(str string) []interface{} {
	a := []interface{}{}

	if err := yaml.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}

/*
toTOML takes an interface, marshals it to toml, and returns a string. It will
always return a string, even on marshal error (empty string).

This is designed to be called from a template.
*/
func toTOML(v interface{}) string {
	b := bytes.NewBuffer(nil)
	e := toml.NewEncoder(b)
	err := e.Encode(v)
	if err != nil {
		return err.Error()
	}
	return b.String()
}

/*
toJSON takes an interface, marshals it to json, and returns a string. It will
always return a string, even on marshal error (empty string).

This is designed to be called from a template.
*/
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

/*
fromJSON converts a JSON document into a map[string]interface{}.

This is not a general-purpose JSON parser, and will not parse all valid
JSON documents. Additionally, because its intended use is within templates
it tolerates errors. It will insert the returned error message string into
m["Error"] in the returned map.
*/
func fromJSON(str string) map[string]interface{} {
	m := make(map[string]interface{})

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

/*
fromJSONArray converts a JSON array into a []interface{}.

This is not a general-purpose JSON parser, and will not parse all valid
JSON documents. Additionally, because its intended use is within templates
it tolerates errors. It will insert the returned error message string as
the first and only item in the returned array.
*/
func fromJSONArray(str string) []interface{} {
	a := []interface{}{}

	if err := json.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}

// These are late-bound in Engine.Render().  The
// version included in the FuncMap is a placeholder.
func helmFuncMap() template.FuncMap {
	return template.FuncMap{
		"toToml":        toTOML,
		"toYaml":        toYAML,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,
	}

}
