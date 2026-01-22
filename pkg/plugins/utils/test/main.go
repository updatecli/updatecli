package test

import (
	"bytes"
	"sort"
	"testing"

	"go.yaml.in/yaml/v3"

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

	return assert.Equal(t,
		yamlMarshalUnmarshal(t, buf.String()),
		yamlMarshalUnmarshal(t, manifest))
}

// SortConfigSpecArray allows to sort an array of config.Spec using the length of the name field as sorting
// This function is used to sort a config.Spec array before comparing it to a manifest of type array of byte
func SortConfigSpecArray(t *testing.T, configSpecs []config.Spec, byteSpecs [][]byte) {
	// We convert byteSpecs to an array of config.Spec so we can apply the same sort
	// algorithm to both byteSpecs and configSpecs
	tmpSpecs := make([]config.Spec, len(byteSpecs))
	for i := range byteSpecs {
		err := yaml.Unmarshal(byteSpecs[i], &tmpSpecs[i])
		require.NoError(t, err)
	}

	sort.Slice(tmpSpecs, func(i, j int) bool {
		return len(tmpSpecs[i].Name) < len(tmpSpecs[j].Name)
	})

	sort.Slice(configSpecs, func(i, j int) bool {
		return len(configSpecs[i].Name) < len(configSpecs[j].Name)
	})

	// We convert back byteSpecs to an array of byte
	for i := range byteSpecs {
		buf := bytes.NewBufferString("")
		yamlEncoder := yaml.NewEncoder(buf)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(tmpSpecs[i])
		require.NoError(t, err)
		byteSpecs[i] = buf.Bytes()
	}
}

// yamlMarshalUnmarshal is used to parse a manifest to ensure it's a valid yaml one.
// yamlMarshalUnmarshal is also used to trim single quotes from yaml values
func yamlMarshalUnmarshal(t *testing.T, manifest string) string {
	var spec config.Spec
	err := yaml.Unmarshal([]byte(manifest), &spec)
	require.NoError(t, err)

	buf := bytes.NewBufferString("")
	yamlEncoder := yaml.NewEncoder(buf)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(spec)
	require.NoError(t, err)

	return buf.String()
}
