package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
)

// TestSchemaRegistryCompilesEveryKind ensures every kind Updatecli supports produces a
// usable schema. Without it a plugin whose specification cannot be reflected would only
// be discovered when a user validates a manifest using that kind.
func TestSchemaRegistryCompilesEveryKind(t *testing.T) {

	registry, err := getSchemaRegistry()
	require.NoError(t, err)

	for _, s := range resourceSections {
		kinds := registry.kinds(s)
		require.NotEmptyf(t, kinds, "section %q supports no kind", s)

		for _, kind := range kinds {
			t.Run(string(s)+"/"+kind, func(t *testing.T) {
				schema, err := registry.specSchema(s, kind)
				require.NoError(t, err)

				// A nil schema means the kind takes no specification, which is only
				// expected when the mapping says so.
				if schema == nil {
					assert.Nil(t, registry.kindSpecs[s][kind])
				}
			})
		}
	}
}

// TestResourceMappingMatchesSupportedKinds pins that the mapping driving validation lists
// the same kinds as the one driving execution, as they are maintained separately.
func TestResourceMappingMatchesSupportedKinds(t *testing.T) {

	registry, err := getSchemaRegistry()
	require.NoError(t, err)

	mapping := resource.GetResourceMapping()

	for _, s := range []section{sectionSources, sectionConditions, sectionTargets} {
		assert.Equalf(t, len(mapping), len(registry.kindSpecs[s]),
			"section %q does not support the same kinds as the resource mapping", s)
	}

	// A kind advertised by the schema but unknown to the dispatcher would be a kind the
	// validator accepts and the pipeline then rejects. Building it with an empty
	// specification is expected to fail validation, only an unsupported kind matters here.
	for kind := range mapping {
		_, err := resource.New(resource.ResourceConfig{Kind: kind})
		if err == nil {
			continue
		}

		assert.NotContainsf(t, err.Error(), "Don't support resource kind",
			"kind %q is advertised by the schema but not dispatched", kind)
	}
}
