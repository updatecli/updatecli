package jsonschema

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/config"
)

// GenerateSchema generates updatecli json schema based the config struct
func GenerateSchema() string {

	r := new(jsonschema.Reflector)

	r.SetBaseSchemaID("https://www.updatecli.io/schema")

	r.PreferYAMLSchema = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	if err := r.AddGoComments("github.com/updatecli/updatecli/pkg", "./"); err != nil {
		fmt.Println(err)
		return ""
	}

	s := r.Reflect(config.Config{})

	u, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(u)
}
