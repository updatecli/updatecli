package jsonschema

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/config/jsonschema/commentmap"
)

var (
	// SchemaDir defines the schema root directory
	SchemaDir string = "../../../../schema"
	// Set base schema ID
	BaseSchemaID string = "https://www.updatecli.io/schema"
)

// initSchemaDir create required directories so we can generate json schema
func initSchemaDir() error {
	if _, err := os.Stat(SchemaDir); os.IsNotExist(err) {

		err := os.MkdirAll(SchemaDir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateSchema generates updatecli json schema based the config struct
func GenerateSchema() string {

	var err error

	r := new(jsonschema.Reflector)

	r.SetBaseSchemaID(BaseSchemaID)

	r.PreferYAMLSchema = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true
	r.ExpandedStruct = true

	r.CommentMap, err = commentmap.Get("../../../../pkg")

	if err != nil {
		logrus.Errorf(err.Error())
		return ""
	}

	s := r.Reflect(&config.Config{})

	u, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		logrus.Errorf(err.Error())
		return ""
	}

	err = ioutil.WriteFile(filepath.Join(SchemaDir, "config.json"), u, 0644)
	if err != nil {
		return ""
	}

	return string(u)
}
