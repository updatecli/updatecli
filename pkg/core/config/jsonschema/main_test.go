package jsonschema

import "testing"

func TestGenerateSchema(t *testing.T) {
	schema := GenerateSchema()

	t.Errorf("Error: %v", schema)
}
