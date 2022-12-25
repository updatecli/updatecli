package test

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/config"
)

// AssertConfigSpecEqualByteArray is a testing function used to compare two Updatecli manifest.
// One use the config.Spec struct and the second one is described as array of byte
func AssertConfigSpecEqualByteArray(t *testing.T, spec *config.Spec, manifest string) bool {
	buf := bytes.NewBufferString("")
	yamlEncoder := yaml.NewEncoder(buf)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(spec)
	expectedPipeline := buf.String()

	return assert.Equal(t, string(expectedPipeline), manifest)
}
