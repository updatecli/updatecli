package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/agext/levenshtein"
)

// lowercaseKeys returns a copy of a resource specification with every mapping key
// lowercased, mirroring how mapstructure matches a specification field regardless of its
// case. On a collision the first key wins, as validating twice the same field would only
// duplicate problems.
func lowercaseKeys(value interface{}) interface{} {

	switch typed := value.(type) {

	case map[string]interface{}:
		lowercased := make(map[string]interface{}, len(typed))
		for _, key := range sortedKeys(typed) {
			normalized := strings.ToLower(key)
			if _, ok := lowercased[normalized]; ok {
				continue
			}
			lowercased[normalized] = lowercaseKeys(typed[key])
		}
		return lowercased

	case []interface{}:
		lowercased := make([]interface{}, len(typed))
		for i, item := range typed {
			lowercased[i] = lowercaseKeys(item)
		}
		return lowercased

	default:
		return value
	}
}

// deprecatedReplacements names the key superseding a deprecated one, when the
// replacement cannot be derived from the field name itself.
var deprecatedReplacements = map[string]string{
	"title":        "name",
	"pullrequests": "actions",
	"conditionids": "dependson",
	"depends_on":   "dependson",
}

// deprecatedFieldKeys collects the YAML keys Updatecli still accepts but hides from its
// schema, which it marks with a `jsonschema:"-"` struct tag.
//
// Deriving the list by reflection rather than maintaining it by hand keeps it correct as
// fields are deprecated: a key hidden from the schema would otherwise be reported as
// unknown even though the manifest still works.
func deprecatedFieldKeys(types []reflect.Type) map[string]string {

	keys := map[string]string{}

	var collect func(t reflect.Type, depth int)
	collect = func(t reflect.Type, depth int) {
		for t != nil {
			switch t.Kind() {
			case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Map:
				t = t.Elem()
				continue
			}
			break
		}

		// Manifest types nest shallowly, the bound only guards against a cycle.
		if t == nil || t.Kind() != reflect.Struct || depth > 6 {
			return
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// An inlined struct contributes its own keys to the mapping holding it.
			if field.Anonymous && field.Tag.Get("jsonschema") != "-" {
				collect(field.Type, depth)
				continue
			}

			key := yamlKey(field)
			if key == "" {
				continue
			}

			if field.Tag.Get("jsonschema") == "-" {
				replacement, ok := deprecatedReplacements[strings.ToLower(key)]
				if !ok {
					replacement = strings.ToLower(key)
				}
				keys[key] = replacement
				continue
			}

			collect(field.Type, depth+1)
		}
	}

	for _, t := range types {
		collect(t, 0)
	}

	return keys
}

// yamlKey returns the mapping key a struct field is decoded from, mirroring how the YAML
// library derives it: the name in the tag when present, the lowercased field name
// otherwise. It returns an empty string for a field that is never decoded.
func yamlKey(field reflect.StructField) string {

	if field.PkgPath != "" {
		return ""
	}

	tag := field.Tag.Get("yaml")
	if tag == "-" {
		return ""
	}

	name, _, _ := strings.Cut(tag, ",")
	if name != "" {
		return name
	}

	if field.Anonymous {
		return ""
	}

	return strings.ToLower(field.Name)
}

// suggest returns a hint naming the closest candidate to an unknown value, or an empty
// string when no candidate is close enough. A wrong suggestion is worse than none, so a
// candidate is only proposed when it is both close and unambiguous.
func suggest(value string, candidates []string) string {

	const maxDistance int = 2

	best := ""
	bestDistance := maxDistance + 1
	ambiguous := false

	for _, candidate := range candidates {
		distance := levenshtein.Distance(value, candidate, nil)

		switch {
		case distance < bestDistance:
			best, bestDistance, ambiguous = candidate, distance, false
		case distance == bestDistance:
			ambiguous = true
		}
	}

	if best == "" || ambiguous || bestDistance > maxDistance || bestDistance*3 > len(value) {
		return ""
	}

	return fmt.Sprintf(", did you mean %q?", best)
}
