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
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestHelmFuncs(t *testing.T) {
	//TODO write tests for failure cases
	tests := []struct {
		tpl, expect string
		vars        interface{}
	}{
		{
			tpl:    `{{ toYaml . }}`,
			expect: `foo: bar`,
			vars:   map[string]interface{}{"foo": "bar"},
		},
		{
			tpl:    `{{ fromYaml . }}`,
			expect: "map[hello:world]",
			vars:   `hello: world`,
		},
		{
			tpl:    `{{ fromYamlArray . }}`,
			expect: "[one 2 map[name:helm]]",
			vars:   "- one\n- 2\n- name: helm\n",
		},
		{
			tpl:    `{{ fromYamlArray . }}`,
			expect: "[one 2 map[name:helm]]",
			vars:   `["one", 2, { "name": "helm" }]`,
		},
		{
			tpl:    `{{ fromYaml . }}`,
			expect: "map[Error:yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into map[string]interface {}]",
			vars:   "- one\n- two\n",
		},
		{
			tpl:    `{{ fromYaml . }}`,
			expect: "map[Error:yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into map[string]interface {}]",
			vars:   `["one", "two"]`,
		},
		{
			tpl:    `{{ fromYamlArray . }}`,
			expect: "[yaml: unmarshal errors:\n  line 1: cannot unmarshal !!map into []interface {}]",
			vars:   `hello: world`,
		},
	}

	for _, tt := range tests {
		var b strings.Builder
		err := template.Must(template.New("test").Funcs(helmFuncMap()).Parse(tt.tpl)).Execute(&b, tt.vars)
		assert.NoError(t, err)
		assert.Equal(t, tt.expect, b.String(), tt.tpl)
	}
}
