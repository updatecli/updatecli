package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	validator "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"go.yaml.in/yaml/v3"
)

// SchemaValidationOptions tunes how strictly a manifest is checked.
type SchemaValidationOptions struct {
	// Strict reports as errors what is otherwise only a warning, and disables the
	// filters hiding the checks Updatecli cannot perform reliably.
	Strict bool
	// AllowDeprecated reports a deprecated but still accepted key as a warning rather
	// than as an unknown key.
	AllowDeprecated bool
	// SkipTemplatedValues ignores the value of a field still holding a go template, as
	// those are only resolved once the pipeline runs.
	SkipTemplatedValues bool
}

// DefaultSchemaValidationOptions returns the options used unless told otherwise.
func DefaultSchemaValidationOptions() SchemaValidationOptions {
	return SchemaValidationOptions{
		AllowDeprecated:     true,
		SkipTemplatedValues: true,
	}
}

// ValidateSchema checks an already templated manifest against the Updatecli schema.
//
// A manifest problem is reported through the returned report, an error is only returned
// when the manifest cannot be read at all.
func ValidateSchema(filename string, content []byte, options SchemaValidationOptions) (SchemaReport, error) {

	report := SchemaReport{File: filename}

	registry, err := getSchemaRegistry()
	if err != nil {
		return report, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(content))

	for document := 0; ; document++ {
		var raw interface{}

		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return report, fmt.Errorf("parsing %q: %w", filename, err)
		}

		// An empty document carries nothing to validate.
		if raw == nil {
			continue
		}

		validation := &schemaValidator{
			registry: registry,
			options:  options,
			file:     filename,
			document: document,
		}

		validation.validateDocument(jsonschema.NormalizeValue(raw))

		report.Problems = append(report.Problems, validation.problems...)
	}

	return report, nil
}

// reportSchemaProblems hands the schema problems found in a manifest to onProblem, or
// logs them as warnings when onProblem is not set.
//
// It never fails a manifest: the schema has only ever been used to assist editors, so a
// manifest Updatecli runs today must keep running even when the schema disagrees.
func reportSchemaProblems(filename string, content []byte, onProblem func(SchemaProblem)) {

	report, err := ValidateSchema(filename, content, DefaultSchemaValidationOptions())
	if err != nil {
		logrus.Debugf("skipping the schema validation of %q: %s", filename, err)
		return
	}

	for _, problem := range report.Problems {
		if onProblem != nil {
			onProblem(problem)
			continue
		}

		logrus.Warningf("%s", problem.String())
	}
}

// schemaValidator walks a single manifest document.
type schemaValidator struct {
	registry *schemaRegistry
	options  SchemaValidationOptions
	file     string
	document int
	problems []SchemaProblem
}

func (v *schemaValidator) validateDocument(raw interface{}) {

	document, ok := raw.(map[string]interface{})
	if !ok {
		v.report(SeverityError, "", "expected a mapping, got %s", typeOf(raw))
		return
	}

	v.check(v.registry.root, opaqueSections(document), "", nodeRoot)

	for _, s := range resourceSections {
		v.validateResources(document, s)
	}
}

// validateResources checks every entry of a section against the schema of the kind it
// declares, which yields a precise problem instead of reporting that the entry matched
// none of the supported kinds.
func (v *schemaValidator) validateResources(document map[string]interface{}, s section) {

	raw, ok := document[string(s)]
	if !ok || raw == nil {
		return
	}

	entries, ok := raw.(map[string]interface{})
	if !ok {
		// Already reported by the root schema as a type problem.
		return
	}

	for _, id := range sortedKeys(entries) {
		path := string(s) + "." + id

		entry, ok := entries[id].(map[string]interface{})
		if !ok {
			v.report(SeverityError, path, "expected a mapping, got %s", typeOf(entries[id]))
			continue
		}

		v.validateResource(entry, s, path)
	}
}

func (v *schemaValidator) validateResource(entry map[string]interface{}, s section, path string) {

	// Everything but the specification is checked against the section schema, the
	// specification itself depends on the kind and is checked further down.
	v.check(v.registry.resources[s], withoutKey(entry, "spec"), path, string(s))

	rawKind, hasKind := entry["kind"]

	// An scm may be disabled, in which case it carries neither kind nor spec.
	if s == sectionSCMs && !hasKind {
		if disabled, ok := entry["disabled"].(bool); ok && disabled {
			return
		}
	}

	if !hasKind {
		v.report(SeverityError, path, "missing required key %q", "kind")
		return
	}

	declaredKind, ok := rawKind.(string)
	if !ok {
		v.report(SeverityError, path+".kind", "expected a string, got %s", typeOf(rawKind))
		return
	}

	// Updatecli lowercases a kind before dispatching on it, so a kind differing only by
	// its case is accepted and merely warned about.
	normalizedKind := strings.ToLower(declaredKind)
	if declaredKind != normalizedKind {
		v.report(SeverityWarning, path+".kind", "kind value %q must be lowercase", declaredKind)
	}

	specSchema, err := v.registry.specSchema(s, normalizedKind)
	if err != nil {
		v.report(SeverityError, path+".kind",
			"unsupported kind %q%s", declaredKind, suggest(normalizedKind, v.registry.kinds(s)))
		return
	}

	if specSchema == nil {
		return
	}

	spec, ok := entry["spec"]
	if !ok || spec == nil {
		return
	}

	// A spec is decoded with mapstructure, which matches field names regardless of
	// their case, unlike every key above it.
	v.check(specSchema, lowercaseKeys(spec), path+".spec", specNode(s, normalizedKind))
}

// check validates one node and turns each failure into a problem rooted at path.
func (v *schemaValidator) check(schema *validator.Schema, instance interface{}, path string, node string) {

	err := schema.Validate(instance)
	if err == nil {
		return
	}

	var validationErr *validator.ValidationError
	if !errors.As(err, &validationErr) {
		v.report(SeverityError, path, "%s", err)
		return
	}

	v.flatten(validationErr, path, node, schema)
}

// flatten walks down to the leaf causes, which are the ones carrying both a precise
// location and a typed reason.
func (v *schemaValidator) flatten(err *validator.ValidationError, path string, node string, schema *validator.Schema) {

	if len(err.Causes) > 0 {
		for _, cause := range err.Causes {
			v.flatten(cause, path, node, schema)
		}
		return
	}

	location := joinPath(path, err.InstanceLocation)

	severity, message := v.describe(err, location, node, schema)
	if message == "" {
		return
	}

	v.report(severity, location, "%s", message)
}

// describe turns a validation failure into an Updatecli flavoured message, and drops the
// ones Updatecli cannot check reliably.
func (v *schemaValidator) describe(err *validator.ValidationError, location string, node string, schema *validator.Schema) (Severity, string) {

	switch errorKind := err.ErrorKind.(type) {

	case *kind.AdditionalProperties:
		messages := []string{}
		for _, property := range errorKind.Properties {
			if replacement, ok := v.registry.deprecatedKey(node, property); ok && v.options.AllowDeprecated {
				v.report(v.deprecationSeverity(), joinKey(location, property),
					"%q is deprecated in favor of %q", property, replacement)
				continue
			}
			messages = append(messages, fmt.Sprintf("unknown key %q%s",
				property, suggest(property, propertyNames(schema, err.InstanceLocation))))
		}

		return SeverityError, strings.Join(messages, ", ")

	case *kind.Required:
		// A missing kind is reported by the walker, which resolves it before dispatching
		// on it and knows the cases where it may legitimately be absent.
		missing := remove(errorKind.Missing, "kind")
		if len(missing) == 0 {
			return SeverityWarning, ""
		}

		return v.requiredSeverity(node),
			fmt.Sprintf("missing required key(s) %s", quoteAll(missing))

	case *kind.Type:
		// A value still holding a go template is only resolved once the pipeline runs,
		// so its type cannot be checked yet.
		if v.skipTemplated(errorKind.Got) {
			return SeverityWarning, ""
		}
		return SeverityError, fmt.Sprintf("expected %s, got %s",
			strings.Join(errorKind.Want, " or "), errorKind.Got)

	case *kind.Enum:
		if value, ok := errorKind.Got.(string); ok && v.skipTemplatedString(value) {
			return SeverityWarning, ""
		}
		return SeverityError, fmt.Sprintf("value %v is not one of %v", errorKind.Got, errorKind.Want)

	default:
		return SeverityError, err.ErrorKind.LocalizedString(nil)
	}
}

func (v *schemaValidator) deprecationSeverity() Severity {

	if v.options.Strict {
		return SeverityError
	}

	return SeverityWarning
}

// requiredSeverity decides how much a missing key matters.
//
// Whether a specification key is required often depends on the rest of the manifest: a
// repository may come from "path" or from a "scmid" rather than from "url". Those
// dependencies are enforced by each resource at run time but cannot be expressed by the
// schema, which has never been checked against them. A missing specification key is
// therefore only reported as a warning, unless asked to be strict, while the manifest
// structure itself remains an error.
func (v *schemaValidator) requiredSeverity(node string) Severity {

	if v.options.Strict || !strings.Contains(node, "/") {
		return SeverityError
	}

	return SeverityWarning
}

// propertyNames returns the keys accepted at the location an unknown one was found, so
// that a replacement can be suggested.
func propertyNames(schema *validator.Schema, location []string) []string {

	schema = resolveSchema(schema, location)
	if schema == nil {
		return nil
	}

	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}

	return names
}

// resolveSchema walks a schema down to the one describing an instance location.
func resolveSchema(schema *validator.Schema, location []string) *validator.Schema {

	for _, element := range location {
		if schema == nil {
			return nil
		}

		if next, ok := schema.Properties[element]; ok {
			schema = next
			continue
		}

		if schema.Items2020 != nil {
			schema = schema.Items2020
			continue
		}

		if additional, ok := schema.AdditionalProperties.(*validator.Schema); ok {
			schema = additional
			continue
		}

		return nil
	}

	return schema
}

// remove returns the values but the excluded one.
func remove(values []string, excluded string) []string {

	kept := make([]string, 0, len(values))

	for _, value := range values {
		if value == excluded {
			continue
		}
		kept = append(kept, value)
	}

	return kept
}

// withoutKey returns a copy of a mapping with one key removed.
func withoutKey(values map[string]interface{}, key string) map[string]interface{} {

	filtered := make(map[string]interface{}, len(values))

	for name, value := range values {
		if name == key {
			continue
		}
		filtered[name] = value
	}

	return filtered
}

func (v *schemaValidator) skipTemplated(got string) bool {
	// A templated value always reaches this point as a string.
	return v.options.SkipTemplatedValues && !v.options.Strict && got == "string"
}

func (v *schemaValidator) skipTemplatedString(value string) bool {
	return v.options.SkipTemplatedValues && !v.options.Strict && IsTemplatedString(value)
}

func (v *schemaValidator) report(severity Severity, path string, format string, args ...interface{}) {

	v.problems = append(v.problems, SchemaProblem{
		File:     v.file,
		Document: v.document,
		Path:     path,
		Severity: severity,
		Message:  fmt.Sprintf(format, args...),
	})
}

// opaqueSections replaces every resource map by an empty one, so that the root schema
// only reports problems on the manifest top level. Each resource is checked separately
// against the schema of the kind it declares.
func opaqueSections(document map[string]interface{}) map[string]interface{} {

	opaque := make(map[string]interface{}, len(document))

	for key, value := range document {
		switch section(key) {
		case sectionSources, sectionConditions, sectionTargets, sectionSCMs, sectionActions:
			if _, ok := value.(map[string]interface{}); ok {
				opaque[key] = map[string]interface{}{}
				continue
			}
		}
		opaque[key] = value
	}

	return opaque
}

func joinPath(path string, location []string) string {

	for _, element := range location {
		path = joinKey(path, element)
	}

	return path
}

func joinKey(path string, key string) string {

	if path == "" {
		return key
	}

	return path + "." + key
}

func sortedKeys(values map[string]interface{}) []string {

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func quoteAll(values []string) string {

	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = fmt.Sprintf("%q", value)
	}

	return strings.Join(quoted, ", ")
}

func typeOf(value interface{}) string {

	switch value.(type) {
	case nil:
		return "null"
	case map[string]interface{}:
		return "a mapping"
	case []interface{}:
		return "a sequence"
	default:
		return fmt.Sprintf("%T", value)
	}
}
