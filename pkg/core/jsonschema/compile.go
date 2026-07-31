package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"

	jschema "github.com/invopop/jsonschema"
	validator "github.com/santhosh-tekuri/jsonschema/v6"
)

// Compile turns a generated schema into a compiled validator identified by id,
// which must be an absolute URI.
//
// The "$schema" keyword is removed from the whole document and the draft is pinned
// explicitly. invopop stamps "$schema" from its own package global, which New()
// mutates to draft-04, so the draft a schema declares would otherwise depend on
// whether a schema was exported earlier in the same process. Every keyword the
// generator emits carries the same meaning in every draft, so pinning is safe.
func Compile(id string, schema *jschema.Schema) (*validator.Schema, error) {

	if schema == nil {
		return nil, fmt.Errorf("no schema to compile for %q", id)
	}

	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal schema %q: %w", id, err)
	}

	document, err := validator.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("unmarshal schema %q: %w", id, err)
	}

	stripSchemaKeyword(document)

	compiler := validator.NewCompiler()
	compiler.DefaultDraft(validator.Draft2020)

	if err := compiler.AddResource(id, document); err != nil {
		return nil, fmt.Errorf("add schema resource %q: %w", id, err)
	}

	compiled, err := compiler.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("compile schema %q: %w", id, err)
	}

	return compiled, nil
}

// stripSchemaKeyword recursively removes every "$schema" keyword from a schema
// document so that the declared draft cannot contradict the one pinned by Compile.
func stripSchemaKeyword(document interface{}) {

	switch node := document.(type) {
	case map[string]interface{}:
		delete(node, "$schema")
		for _, value := range node {
			stripSchemaKeyword(value)
		}
	case []interface{}:
		for _, value := range node {
			stripSchemaKeyword(value)
		}
	}
}
