package commentmap

import (
	"strings"

	"github.com/invopop/jsonschema"
)

var (
	// Set base package name
	UpdatecliPackageName string = "github.com/updatecli/updatecli/"
)

// Get retrieves all updatecli code comments
func Get(rootPackagePath string) (map[string]string, error) {

	// It appears that the path change based on where we call the function
	// so we need to parametrize it
	if rootPackagePath == "" {
		rootPackagePath = "../../../../../pkg"
	}
	r := new(jsonschema.Reflector)

	r.Anonymous = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	if err := r.AddGoComments(
		UpdatecliPackageName,
		rootPackagePath); err != nil {
		return nil, err
	}

	for key, value := range r.CommentMap {
		newkey := strings.Replace(key, "../", UpdatecliPackageName, -1)
		r.CommentMap[newkey] = value
	}

	return r.CommentMap, nil
}
