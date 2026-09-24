package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	jschema "github.com/invopop/jsonschema"
	validator "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// section identifies a group of manifest entries sharing the same kind dispatching.
type section string

const (
	sectionSources    section = "sources"
	sectionConditions section = "conditions"
	sectionTargets    section = "targets"
	sectionSCMs       section = "scms"
	sectionActions    section = "actions"

	// nodeRoot names the manifest top level when reporting a problem.
	nodeRoot string = "root"

	// schemaBaseID prefixes the identifier given to each compiled schema. Compiled
	// schemas are never fetched, the identifier only has to be a unique absolute URI.
	schemaBaseID string = "https://www.updatecli.io/schema/validate"
)

// resourceSections lists the sections holding kind dispatched entries, in the order they
// are reported.
var resourceSections = []section{
	sectionSources,
	sectionConditions,
	sectionTargets,
	sectionSCMs,
	sectionActions,
}

// Aliases of the manifest types. An alias does not carry the methods of the type it is
// derived from, which is what keeps the JSONSchema() implementations, and therefore the
// "oneOf" listing every supported kind, out of the reflected schema.
type (
	sourceConfigAlias    source.Config
	conditionConfigAlias condition.Config
	targetConfigAlias    target.Config
	scmConfigAlias       scm.Config
	actionConfigAlias    action.Config
)

// schemaRegistry holds the compiled schemas used to validate a manifest.
//
// Validating a whole manifest against the generated schema would go through a "oneOf"
// listing every supported kind, which can only report that a resource matched none of
// them. A manifest is instead validated in layers: its top level, then each resource,
// then the specification of the kind that resource declares.
type schemaRegistry struct {
	// root validates the manifest top level, with the resource maps left unconstrained.
	root *validator.Schema

	// resources validates a resource entry of a section, minus its specification.
	resources map[section]*validator.Schema

	// kindSpecs holds, per section, the specification of every supported kind.
	kindSpecs map[section]map[string]interface{}

	// deprecated holds, per node, the keys still accepted but hidden from the schema.
	deprecated map[string]map[string]string

	mutex sync.Mutex
	// compiled caches the specification schemas built so far, keyed by node.
	compiled map[string]*validator.Schema
}

// getSchemaRegistry builds the registry once per process, as reflecting the manifest
// types costs significantly more than validating against the result.
var getSchemaRegistry = sync.OnceValues(newSchemaRegistry)

func newSchemaRegistry() (*schemaRegistry, error) {

	root, err := jsonschema.Compile(schemaBaseID+"/"+nodeRoot, rootSchema())
	if err != nil {
		return nil, fmt.Errorf("build the manifest schema: %w", err)
	}

	registry := &schemaRegistry{
		root:      root,
		resources: map[section]*validator.Schema{},
		kindSpecs: map[section]map[string]interface{}{
			sectionSources:    resource.GetResourceMapping(),
			sectionConditions: resource.GetResourceMapping(),
			sectionTargets:    resource.GetResourceMapping(),
			sectionSCMs:       scm.GetScmMapping(),
			sectionActions:    action.GetActionMapping(),
		},
		deprecated: map[string]map[string]string{
			// Only the manifest top level is reflected here. A key deprecated inside a
			// nested section, such as ManifestOptions, has to add its type to this list
			// to be reported as deprecated rather than as unknown.
			nodeRoot: deprecatedFieldKeys([]reflect.Type{reflect.TypeOf(Spec{})}),
		},
		compiled: map[string]*validator.Schema{},
	}

	resourceAliases := map[section]interface{}{
		sectionSources:    sourceConfigAlias{},
		sectionConditions: conditionConfigAlias{},
		sectionTargets:    targetConfigAlias{},
		sectionSCMs:       scmConfigAlias{},
		sectionActions:    actionConfigAlias{},
	}

	for _, s := range resourceSections {
		alias := resourceAliases[s]

		compiled, err := jsonschema.Compile(schemaBaseID+"/"+string(s), specReflector().Reflect(alias))
		if err != nil {
			return nil, fmt.Errorf("build the %q schema: %w", s, err)
		}

		registry.resources[s] = compiled
		registry.deprecated[string(s)] = deprecatedFieldKeys([]reflect.Type{reflect.TypeOf(alias)})
	}

	return registry, nil
}

// specReflector returns the reflector used for everything but the manifest top level.
func specReflector() *jschema.Reflector {

	reflector := new(jschema.Reflector)
	reflector.Anonymous = true
	reflector.DoNotReference = true
	reflector.RequiredFromJSONSchemaTags = true
	reflector.KeyNamer = strings.ToLower

	return reflector
}

// rootSchema reflects the manifest top level while replacing every resource map by an
// unconstrained object.
//
// The Mapper is consulted before a type's own JSONSchema() implementation, which is what
// keeps the "oneOf" of every supported kind out of the root schema.
func rootSchema() *jschema.Schema {

	reflector := new(jschema.Reflector)
	reflector.DoNotReference = true
	reflector.RequiredFromJSONSchemaTags = true
	reflector.KeyNamer = strings.ToLower

	opaque := map[reflect.Type]bool{
		reflect.TypeOf(source.Config{}):                true,
		reflect.TypeOf(condition.Config{}):             true,
		reflect.TypeOf(target.Config{}):                true,
		reflect.TypeOf(scm.Config{}):                   true,
		reflect.TypeOf(action.Config{}):                true,
		reflect.TypeOf(autodiscovery.CrawlersConfig{}): true,
	}

	reflector.Mapper = func(t reflect.Type) *jschema.Schema {
		if opaque[t] {
			return &jschema.Schema{Type: "object"}
		}
		return nil
	}

	return reflector.Reflect(Spec{})
}

// kinds returns the kinds supported by a section, used to suggest a replacement for an
// unknown one.
func (r *schemaRegistry) kinds(s section) []string {

	names := make([]string, 0, len(r.kindSpecs[s]))
	for name := range r.kindSpecs[s] {
		names = append(names, name)
	}

	return names
}

// deprecatedKey reports whether a key is deprecated but still accepted at a node, along
// with the key replacing it.
func (r *schemaRegistry) deprecatedKey(node string, key string) (string, bool) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	replacement, ok := r.deprecated[node][key]

	return replacement, ok
}

// specNode names the node holding the specification of a kind.
func specNode(s section, kind string) string {
	return string(s) + "/" + kind
}

// specSchema returns the compiled schema validating the specification of a kind.
//
// It returns a nil schema when the kind takes no specification, and an error when the
// kind is not supported at all.
func (r *schemaRegistry) specSchema(s section, kind string) (*validator.Schema, error) {

	specs, ok := r.kindSpecs[s]
	if !ok {
		return nil, fmt.Errorf("unknown section %q", s)
	}

	spec, ok := specs[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported kind %q", kind)
	}

	if spec == nil {
		return nil, nil
	}

	node := specNode(s, kind)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if compiled, ok := r.compiled[node]; ok {
		return compiled, nil
	}

	compiled, err := jsonschema.Compile(schemaBaseID+"/"+node, specReflector().Reflect(spec))
	if err != nil {
		return nil, err
	}

	r.compiled[node] = compiled
	r.deprecated[node] = deprecatedFieldKeys([]reflect.Type{reflect.TypeOf(spec)})

	return compiled, nil
}
