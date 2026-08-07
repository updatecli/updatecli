package jsonschema

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
)

const (
	schemaVersionDraft04 string = "http://json-schema.org/draft-04/schema"
)

var (
	// defaultSchemaDir defines the schema root directory
	defaultSchemaDir string = "../../../../schema"
	// Set base schema ID
	defaultBaseSchemaID string = "https://www.updatecli.io/schema"
	// Set base package name
	updatecliPackageName string = "github.com/updatecli/updatecli/"

	// commentDir defines the temporary directory where updatecli git repository is cloned to retrieve code comments
	// it is used to populate jsonschema field "description"
	commentDir string = path.Join(os.TempDir(), "updatecli", "_comments")
	// commentURL defines the updatecli git url
	commentURL string = "https://github.com/updatecli/updatecli.git"

	// commentMapMutex guards cachedCommentMap
	commentMapMutex sync.Mutex
	// cachedCommentMap holds the code comments once they have been successfully retrieved
	cachedCommentMap map[string]string
)

// getCommentMap returns the updatecli code comments used to populate the jsonschema
// "description" fields.
//
// Comments are a documentation nicety: they require the updatecli git repository to be
// cloned locally, which only makes sense when generating the published schema. When they
// cannot be retrieved, an empty map is returned so that a schema can still be built, for
// example to validate a manifest at runtime.
//
// Only a successful lookup is cached, so a call made before CloneCommentDirectory does
// not prevent a later call from picking the comments up.
func getCommentMap() map[string]string {
	commentMapMutex.Lock()
	defer commentMapMutex.Unlock()

	if cachedCommentMap != nil {
		return cachedCommentMap
	}

	commentMap, err := GetPackageComments(commentDir)
	if err != nil {
		logrus.Debugf("retrieve code comments: %s", err.Error())
		return nil
	}

	cachedCommentMap = commentMap

	return cachedCommentMap
}

// setCommentMap caches a comment map that was retrieved by a caller.
func setCommentMap(commentMap map[string]string) {
	commentMapMutex.Lock()
	defer commentMapMutex.Unlock()

	if len(commentMap) > 0 {
		cachedCommentMap = commentMap
	}
}

type Schema struct {
	SchemaDir    string
	BaseSchemaID string
	JsonSchema   jschema.Schema
}

func New(baseSchemaID, schemaDirectory string) *Schema {

	jschema.Version = schemaVersionDraft04

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

	r := new(jschema.Reflector)

	r.SetBaseSchemaID(s.BaseSchemaID)

	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	r.KeyNamer = strings.ToLower

	r.CommentMap, err = GetPackageComments(commentDir)

	if err != nil {
		return err
	}

	setCommentMap(r.CommentMap)

	s.JsonSchema = *r.Reflect(object)

	return nil
}

// Save export a jsonschema to a local file
func (s *Schema) Save() error {
	err := os.WriteFile(filepath.Join(s.SchemaDir, "config.json"), []byte(s.String()), 0600)
	if err != nil {
		return err
	}
	return nil
}

// String implements the string interface
func (s *Schema) String() string {
	indentedJsonSchema, err := json.MarshalIndent(s.JsonSchema, "", "    ")
	if err != nil {
		logrus.Errorf("marshal indent the jsonschema: %s", err.Error())
	}

	return string(indentedJsonSchema)
}

// GetPackageComments retrieves all updatecli code comments
func GetPackageComments(rootPackagePath string) (map[string]string, error) {

	r := new(jschema.Reflector)

	r.Anonymous = true
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

// BuildKindSchemas generates one jsonschema per kind, where each schema is the base
// resource configuration with "kind" pinned to that single value and "spec" replaced by
// the specification of that kind. A kind mapped to a nil specification keeps the base
// configuration without any "spec" constraint.
//
// Validating a resource against the schema of its declared kind gives a far more precise
// error than validating it against the "oneOf" of every kind, which can only report that
// the resource matched none of them.
func BuildKindSchemas(baseConfig interface{}, kinds map[string]interface{}) map[string]*jschema.Schema {

	if baseConfig == nil {
		return nil
	}

	// Retrieve Updatecli code comments, they only populate the "description" fields so
	// an empty map still produces a usable schema.
	commentMap := getCommentMap()

	schemas := make(map[string]*jschema.Schema, len(kinds))

	for id := range kinds {
		r := new(jschema.Reflector)

		r.Anonymous = true
		r.DoNotReference = true
		r.RequiredFromJSONSchemaTags = true
		r.CommentMap = commentMap
		r.KeyNamer = strings.ToLower

		resourceConfig := r.Reflect(baseConfig)

		if kinds[id] != nil {
			resourceConfig.Properties.Set("spec", r.Reflect(kinds[id]))
		}

		resourceConfig.Properties.Set("kind", &jschema.Schema{
			Enum: []interface{}{id}})

		schemas[id] = resourceConfig
	}

	return schemas
}

// AppendOneOfToJsonschema generates a jsonschema based on a baseConfig and then append a oneOf based on the mapConfig.
func AppendOneOfToJsonSchema(baseConfig interface{}, anyOf map[string]interface{}) *jschema.Schema {

	kindSchemas := BuildKindSchemas(baseConfig, anyOf)

	if kindSchemas == nil {
		return nil
	}

	keys := slices.Collect(maps.Keys(kindSchemas))

	sort.Strings(keys)

	resourceSchema := jschema.Schema{}

	for _, id := range keys {
		resourceSchema.OneOf = append(resourceSchema.OneOf, kindSchemas[id])
	}

	return &resourceSchema

}

// AppendMapToJsonSchema generates a jsonschema based on a baseConfig and then append a map of properties using the mapConfig.
func AppendMapToJsonSchema(baseConfig interface{}, mapConfig map[string]interface{}) *jschema.Schema {

	if baseConfig == nil {
		return nil
	}

	// Retrieve Updatecli code comments, they only populate the "description" fields so
	// an empty map still produces a usable schema.
	commentMap := getCommentMap()

	r := new(jschema.Reflector)

	r.Anonymous = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true
	r.CommentMap = commentMap
	r.KeyNamer = strings.ToLower

	resourceConfig := r.Reflect(baseConfig)

	if resourceConfig.Properties == nil {
		resourceConfig.Properties = jschema.NewProperties()
	}

	keys := slices.Collect(maps.Keys(mapConfig))

	sort.Strings(keys)

	for _, key := range keys {
		if mapConfig[key] == nil {
			resourceConfig.Properties.Set(key, &jschema.Schema{})
			continue
		}

		spec := r.Reflect(mapConfig[key])
		resourceConfig.Properties.Set(key, spec)
	}

	return resourceConfig
}
