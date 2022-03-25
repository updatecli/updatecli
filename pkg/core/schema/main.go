package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
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

	// commentDir defines the temporary directory where
	commentDir string = path.Join(os.TempDir(), "updatecli/")
	// commentURL defines the updatecli git url
	commentURL string = "https://github.com/updatecli/updatecli.git"
)

type Schema struct {
	SchemaDir    string
	BaseSchemaID string
	JsonSchema   jsonschema.Schema
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

	r.CommentMap, err = GetPackageComments(commentDir)

	if err != nil {
		return err
	}

	s.JsonSchema = *r.Reflect(object)

	return nil
}

func GenerateJsonSchema(resourceConfigSchema interface{}, anyOf map[string]interface{}) *jsonschema.Schema {

	var err error
	var commentMap map[string]string

	// Retrieve Updatecli code comments
	commentMap, err = GetPackageComments(commentDir)

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

		s := r.Reflect(spec)
		s.ContentSchema = nil

		var spec jsonschema.Schema

		spec.Type = "object"
		spec.Properties = orderedmap.New()
		spec.Properties.Set("spec", s.Properties)
		spec.Properties.Set("kind", jsonschema.Schema{
			Enum: []interface{}{id}})

		resourceSchema.OneOf = append(resourceSchema.OneOf, &spec)

	}
	return resourceSchema

}

// Save export a jsonschema to a local file
func (s *Schema) Save() error {
	err := ioutil.WriteFile(filepath.Join(s.SchemaDir, "config.json"), []byte(s.String()), 0600)
	if err != nil {
		return err
	}
	return nil
}

// String implements the string interface
func (s *Schema) String() string {
	indentedJsonSchema, err := json.MarshalIndent(s.JsonSchema, "", "    ")
	if err != nil {
		logrus.Errorf(err.Error())
	}

	return string(indentedJsonSchema)
}

// GetPackageComments retrieves all updatecli code comments
func GetPackageComments(rootPackagePath string) (map[string]string, error) {

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

		newkey := key
		for _, element := range strings.Split(newkey, "/") {
			if element == "pkg" {
				break
			}
			newkey = strings.TrimLeft(newkey, "./")
			newkey = strings.TrimPrefix(newkey, element)
		}

		if !strings.HasPrefix(newkey, updatecliPackageName) {
			newkey = filepath.Join(updatecliPackageName, newkey)
			if newkey != key {
				r.CommentMap[newkey] = value
				delete(r.CommentMap, key)
			}
		}
	}

	if len(r.CommentMap) == 0 {
		return nil, fmt.Errorf("no comments retrieved for package %s", updatecliPackageName)
	}

	return r.CommentMap, nil
}

// CloneCommentDirectory clones the updatecli git repository in a
// temporary location so we can parse comments
func CloneCommentDirectory() error {

	// Clone the given repository to the given directory
	logrus.Debugf("git clone %s %s --recursive", commentURL, commentDir)

	_, err := git.PlainClone(commentDir, false, &git.CloneOptions{
		URL:               commentURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	switch err {
	case git.ErrRepositoryAlreadyExists:
		return nil
	case nil:
		return nil
	default:
		return err
	}
}

// CleanCommentDirectory will remove the main temporary directory used by updatecli.
func CleanCommentDirectory() error {
	err := os.RemoveAll(commentDir)

	if err != nil {
		return err
	}

	return nil
}
