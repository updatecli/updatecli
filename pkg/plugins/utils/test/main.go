package test

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
)

// AssertConfigSpecEqualByteArray is a testing function used to compare two Updatecli manifest.
// One use the config.Spec struct and the second one is described as array of byte
func AssertConfigSpecEqualByteArray(t *testing.T, spec *config.Spec, manifest string) bool {
	buf := bytes.NewBufferString("")
	yamlEncoder := yaml.NewEncoder(buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(spec)
	require.NoError(t, err)

	return assert.Equal(t, buf.String(), yamlMarshalUnmarshal(t, manifest))
}

// yamlMarshalUmarshal is used to parse a manifest to ensure it's a valid yaml one.
// yamlMarshalUmarshal is also used to trim single quotes from yaml values
func yamlMarshalUnmarshal(t *testing.T, manifest string) string {

	var spec config.Spec
	yaml.Unmarshal([]byte(manifest), &spec)

	buf := bytes.NewBufferString("")
	yamlEncoder := yaml.NewEncoder(buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(spec)
	require.NoError(t, err)

	return buf.String()
}
