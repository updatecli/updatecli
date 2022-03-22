package schema

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
)

var (
	// defaultSchemaDir defines the schema root directory
	defaultSchemaDir string = "../../../../schema"
	// Set base schema ID
	defaultBaseSchemaID string = "https://www.updatecli.io/schema"
	// Set base package name
	updatecliPackageName string = "github.com/updatecli/updatecli/"
)

type Schema struct {
	SchemaDir    string
	BaseSchemaID string
}

func New(baseSchemaID, schemaDirectory string) *Schema {
	s := Schema{
		SchemaDir:    defaultSchemaDir,
		BaseSchemaID: defaultBaseSchemaID,
	}

	if len(baseSchemaID) > 0 {
		s.BaseSchemaID = baseSchemaID
	}

	if len(schemaDirectory) > 0 {
		s.SchemaDir = schemaDirectory
	}
	return &s
}

// initSchemaDirectory create required directories so we can generate json schema
func (s *Schema) initSchemaDirectory() error {
	if _, err := os.Stat(s.SchemaDir); os.IsNotExist(err) {

		err := os.MkdirAll(s.SchemaDir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateSchema generates updatecli json schema based the config struct
func (s *Schema) GenerateSchema(object interface{}) error {

	err := s.initSchemaDirectory()

	if err != nil {
		return err
	}

	r := new(jsonschema.Reflector)

	r.SetBaseSchemaID(s.BaseSchemaID)

	r.PreferYAMLSchema = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	r.KeyNamer = strings.ToLower

	r.CommentMap, err = GetPackageComments("../../../pkg")

	if err != nil {
		return err
	}

	u, err := json.MarshalIndent(r.Reflect(object), "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(s.SchemaDir, "config.json"), u, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetPackageComments retrieves all updatecli code comments
func GetPackageComments(rootPackagePath string) (map[string]string, error) {

	// It appears that the path change based on where we call the function
	// so we need to parametrize it
	if rootPackagePath == "" {
		rootPackagePath = "../../../pkg"
	}
	r := new(jsonschema.Reflector)

	r.Anonymous = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	if err := r.AddGoComments(
		updatecliPackageName,
		rootPackagePath); err != nil {
		return nil, err
	}

	for key, value := range r.CommentMap {
		newkey := strings.TrimLeft(key, "./")
		newkey = updatecliPackageName + newkey
		r.CommentMap[newkey] = value
	}

	return r.CommentMap, nil
}

//type ResourceConfigSchema Config

func GenerateJsonSchema(resourceConfigSchema interface{}, anyOf map[string]interface{}) *jsonschema.Schema {

	// schemaResources contains list of every resource mapping

	var err error
	var commentMap map[string]string

	// Retrieve Updatecli code comments
	commentMap, err = GetPackageComments("../../../pkg")

	if err != nil {
		logrus.Errorf(err.Error())
		return nil
	}

	r := new(jsonschema.Reflector)

	r.Anonymous = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true
	r.KeyNamer = strings.ToLower
	r.CommentMap = commentMap

	if len(commentMap) == 0 {
		logrus.Errorf("no comments retrieved for package %s", updatecliPackageName)
		return nil
	}

	resourceSchema := r.Reflect(resourceConfigSchema)

	for id, spec := range anyOf {
		r := new(jsonschema.Reflector)

		r.Anonymous = true
		r.PreferYAMLSchema = true
		r.YAMLEmbeddedStructs = true
		r.DoNotReference = true
		r.RequiredFromJSONSchemaTags = true
		r.CommentMap = commentMap
		r.KeyNamer = strings.ToLower

		// schema we need a way to remove schema

		// Main resource type schema such as source or condition
		s := r.Reflect(spec)
		s.ContentSchema = nil

		var spec jsonschema.Schema

		spec.Type = "object"
		spec.Required = []string{"kind"}
		spec.Properties = orderedmap.New()
		spec.Properties.Set("spec", s.Properties)
		spec.Properties.Set("kind", jsonschema.Schema{
			Enum: []interface{}{id}})

		resourceSchema.OneOf = append(resourceSchema.OneOf, &spec)

	}
	return resourceSchema

}
