package autodiscovery

import (
	jschema "github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
)

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (Config) JSONSchema() *jschema.Schema {
	type configAlias Config
	return jsonschema.GenerateJsonSchema(configAlias{}, AutodiscoverySpecsMapping)
}
