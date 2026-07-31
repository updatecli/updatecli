package compose

import (
	"strings"
	"sync"

	jschema "github.com/invopop/jsonschema"
	validator "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"go.yaml.in/yaml/v3"
)

// composeSchemaID identifies the compiled compose schema. It is never fetched, the
// identifier only has to be an absolute URI.
const composeSchemaID string = "https://www.updatecli.io/schema/validate/compose"

// ValidateSchema reports the compose file keys not matching the Updatecli schema.
// It is disabled by default so that loading a compose file keeps its current output.
var ValidateSchema bool

// getComposeSchema builds the compose schema once per process.
//
// Unlike a manifest, a compose file dispatches on no kind, so a single schema describes
// it entirely.
var getComposeSchema = sync.OnceValues(func() (*validator.Schema, error) {

	reflector := new(jschema.Reflector)
	reflector.DoNotReference = true
	reflector.RequiredFromJSONSchemaTags = true
	reflector.KeyNamer = strings.ToLower

	return jsonschema.Compile(composeSchemaID, reflector.Reflect(Spec{}))
})

// validateSchema reports the compose file keys not matching the Updatecli schema, such as
// a misspelled one, which would otherwise be silently ignored.
//
// Problems are only reported: a compose file Updatecli runs today must keep running even
// when the schema disagrees.
func validateSchema(filename string, content []byte) {

	schema, err := getComposeSchema()
	if err != nil {
		logrus.Debugf("building the compose schema: %s", err)
		return
	}

	var document interface{}
	if err := yaml.Unmarshal(content, &document); err != nil {
		// The caller fails on it already.
		return
	}

	for _, problem := range jsonschema.Validate(schema, document) {
		logrus.Warningf("%s %s", filename, problem.String())
	}
}
