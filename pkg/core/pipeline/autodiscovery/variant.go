package autodiscovery

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// pluginCrawlerKind is the crawler kind used by WASM autodiscovery plugins.
	pluginCrawlerKind = "plugin"

	// reservedGroupName names, in pipeline IDs, the dependencies matching no group.
	reservedGroupName = "default"

	versionFiltersKey = "versionfilters"
	groupsKey         = "groups"
	versionFilterKey  = "versionfilter"
	onlyKey           = "only"
	ignoreKey         = "ignore"
)

var (
	// variantNameRegex restricts names to characters safe in a pipeline ID, a git branch and a label value.
	variantNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

// VariantSpec holds the settings, shared by every crawler, that split one crawler into several variants.
type VariantSpec struct {
	// "versionfilters" defines an ordered list of version filters.
	// The crawler runs once per entry, and each entry gets its own pipelines.
	//
	// default:
	//   empty, the crawler runs once with its own "versionfilter".
	//
	// remark:
	//   * each entry accepts every "versionfilter" setting, plus a required "name".
	//   * the name is added to the generated pipeline IDs and action titles,
	//     and set as the "updatecli.io/versionfilter" label, so "--labels updatecli.io/versionfilter:major" selects them.
	//   * an entry waits while an earlier entry still has a pending update,
	//     so a major update is proposed once no minor update is pending.
	//     With "groupby: individual", it waits for the same dependency.
	//     With "groupby: all", or in a group, the dependencies share one pull request per crawler,
	//     so it waits for the whole pull request: one pending minor update holds back every major update of that crawler.
	//     An update is pending while its pull request is open or, without an action, while its working branch exists.
	//     Entries are run in the order they are listed.
	//   * the filters should not overlap, for example "minor" then "majoronly".
	//   * it can't be combined with "versionfilter".
	//   * it isn't supported by WASM plugins.
	//
	// example:
	//   ```
	//   versionfilters:
	//     - name: minor
	//       kind: semver
	//       pattern: minor
	//     - name: major
	//       kind: semver
	//       pattern: majoronly
	//   ```
	//
	VersionFilters []NamedVersionFilter `yaml:",omitempty"`
	// "groups" defines dependencies updated together, in one pipeline per group.
	//
	// default:
	//   empty, every dependency follows "groupby".
	//
	// remark:
	//   * "only" uses the same rules as the crawler "only" setting, and narrows it:
	//     a group dependency must match both the crawler "only" and the group "only".
	//     A group rule combined with a crawler rule setting the same field to another value matches nothing.
	//   * a dependency belongs to the first group it matches.
	//   * dependencies matching no group follow "groupby".
	//   * the group name is added to the generated pipeline IDs and action titles,
	//     and set as the "updatecli.io/group" label, so "--labels updatecli.io/group:aws" selects them.
	//   * the name "default" is reserved for the dependencies matching no group.
	//   * it isn't supported by WASM plugins.
	//
	// example:
	//   ```
	//   groups:
	//     - name: aws
	//       only:
	//         - modules:
	//             "github.com/aws/*": ""
	//   ```
	//
	Groups []Group `yaml:",omitempty"`
}

// NamedVersionFilter is a version filter with a name.
type NamedVersionFilter struct {
	// "name" identifies the version filter in the generated pipelines.
	//
	// remark:
	//   * it must be unique, start with a letter or a digit, and only contain letters, digits, ".", "_" and "-".
	//
	Name           string `yaml:",omitempty" jsonschema:"required"`
	version.Filter `yaml:",inline" mapstructure:",squash"`
}

// Group defines dependencies updated together.
type Group struct {
	// "name" identifies the group in the generated pipelines.
	//
	// remark:
	//   * it must be unique, start with a letter or a digit, and only contain letters, digits, ".", "_" and "-".
	//
	Name string `yaml:",omitempty" jsonschema:"required"`
	// "only" defines the rules matching the dependencies of the group.
	//
	// remark:
	//   * it accepts the same rules as the crawler "only" setting.
	//
	Only []any `yaml:",omitempty" jsonschema:"required"`
}

// Variant describes the share of a crawler's work that one crawler run covers.
type Variant struct {
	// Group is the group name, empty for dependencies matching no group.
	Group string
	// VersionFilter is the version filter name, empty when "versionfilters" isn't set.
	VersionFilter string
	// Rank is the position of the version filter in "versionfilters".
	Rank int
	// GroupBy is the grouping used for the pipelines of this variant.
	GroupBy GroupBy
}

// IsDefault reports whether the variant covers the whole crawler, as without "versionfilters" and "groups".
func (v Variant) IsDefault() bool {
	return v.Group == "" && v.VersionFilter == ""
}

// crawlerVariant is a crawler spec ready for one crawler run.
type crawlerVariant struct {
	spec    any
	variant Variant
}

// expandCrawlerSpec splits a crawler spec into one spec per group and version filter.
//
// The crawler code is untouched: "versionfilters" and "groups" are removed from the spec,
// and the keys every crawler already supports, "versionfilter", "only" and "ignore", are set instead.
// Variants are returned ordered by version filter rank, so earlier version filters run first.
func expandCrawlerSpec(kind string, spec any, groupBy GroupBy) ([]crawlerVariant, error) {
	defaultVariant := []crawlerVariant{{spec: spec, variant: Variant{GroupBy: groupBy}}}

	specMap, ok := asMap(spec)
	if !ok {
		return defaultVariant, nil
	}

	base := specMap
	rawFilters, hasFilters := popKey(base, versionFiltersKey)
	rawGroups, hasGroups := popKey(base, groupsKey)

	if !hasFilters && !hasGroups {
		return defaultVariant, nil
	}

	if kind == pluginCrawlerKind || strings.HasSuffix(kind, ".wasm") {
		return nil, fmt.Errorf("%q and %q aren't supported by WASM plugins", versionFiltersKey, groupsKey)
	}

	filters, rawFilterSpecs, err := decodeVersionFilters(rawFilters, hasFilters)
	if err != nil {
		return nil, err
	}

	if len(filters) > 0 {
		if _, ok := lookupKey(base, versionFilterKey); ok {
			return nil, fmt.Errorf("%q and %q are mutually exclusive", versionFilterKey, versionFiltersKey)
		}
	}

	groups, err := decodeGroups(rawGroups, hasGroups)
	if err != nil {
		return nil, err
	}

	parentIgnore, err := ruleList(base, ignoreKey)
	if err != nil {
		return nil, err
	}

	parentOnly, err := ruleList(base, onlyKey)
	if err != nil {
		return nil, err
	}

	groupOnly := make([][]any, len(groups))
	for i, group := range groups {
		groupOnly[i], err = intersectRules(parentOnly, group.Only)
		if err != nil {
			return nil, fmt.Errorf("group %q: %w", group.Name, err)
		}
	}

	// Without "versionfilters", every variant keeps the crawler's own "versionfilter".
	if len(filters) == 0 {
		filters = []NamedVersionFilter{{}}
		rawFilterSpecs = []map[string]any{nil}
	}

	var variants []crawlerVariant

	for rank, filter := range filters {
		withFilter := func(m map[string]any) map[string]any {
			if rawFilterSpecs[rank] != nil {
				setKey(m, versionFilterKey, rawFilterSpecs[rank])
			}
			return m
		}

		// Each group ignores the dependencies claimed by the groups before it.
		claimed := append([]any{}, parentIgnore...)

		for i, group := range groups {
			m := withFilter(maps.Clone(base))
			setKey(m, onlyKey, append([]any{}, groupOnly[i]...))
			setKey(m, ignoreKey, append([]any{}, claimed...))

			variants = append(variants, crawlerVariant{
				spec: m,
				variant: Variant{
					Group:         group.Name,
					VersionFilter: filter.Name,
					Rank:          rank,
					GroupBy:       GROUPBYALL,
				},
			})

			claimed = append(claimed, group.Only...)
		}

		m := withFilter(maps.Clone(base))
		if len(groups) > 0 {
			setKey(m, ignoreKey, claimed)
		}

		variants = append(variants, crawlerVariant{
			spec: m,
			variant: Variant{
				VersionFilter: filter.Name,
				Rank:          rank,
				GroupBy:       groupBy,
			},
		})
	}

	return variants, nil
}

// intersectRules returns the rules matching what both parent and group rules match.
//
// A rule matches when every field it sets matches, and a list matches when one of its rules matches,
// so the intersection merges each parent rule with each group rule.
// Two rules setting the same field to different values match nothing together, so the pair is dropped.
// It's an error when no pair can be merged, as the group would then match nothing.
func intersectRules(parent, group []any) ([]any, error) {
	if len(parent) == 0 {
		return group, nil
	}

	var rules []any
	for _, parentRule := range parent {
		parentMap, ok := asMap(parentRule)
		if !ok {
			return nil, fmt.Errorf("crawler %q rules must be maps", onlyKey)
		}

		for _, groupRule := range group {
			groupMap, ok := asMap(groupRule)
			if !ok {
				return nil, fmt.Errorf("%q rules must be maps", onlyKey)
			}

			merged, ok := mergeRules(parentMap, groupMap)
			if ok {
				rules = append(rules, merged)
			}
		}
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no %q rule can be combined with the crawler %q rules: each one sets a field the crawler sets to another value", onlyKey, onlyKey)
	}

	return rules, nil
}

// mergeRules returns a rule matching what both rules match, or false when they set the same field to different values.
func mergeRules(a, b map[string]any) (map[string]any, bool) {
	merged := maps.Clone(a)
	for k, v := range b {
		if existing, ok := lookupKey(merged, k); ok {
			if !reflect.DeepEqual(existing, v) {
				return nil, false
			}
			continue
		}
		merged[k] = v
	}

	return merged, true
}

// decodeVersionFilters decodes and validates "versionfilters".
// It also returns each entry without its name, as the crawler expects its "versionfilter".
func decodeVersionFilters(raw any, set bool) ([]NamedVersionFilter, []map[string]any, error) {
	if !set {
		return nil, nil, nil
	}

	rawList, ok := asList(raw)
	if !ok {
		return nil, nil, fmt.Errorf("%q must be a list", versionFiltersKey)
	}

	if len(rawList) == 0 {
		return nil, nil, fmt.Errorf("%q must not be empty", versionFiltersKey)
	}

	filters := make([]NamedVersionFilter, 0, len(rawList))
	rawSpecs := make([]map[string]any, 0, len(rawList))
	names := map[string]bool{}

	var errs []error
	for i, item := range rawList {
		itemMap, ok := asMap(item)
		if !ok {
			errs = append(errs, fmt.Errorf("%s[%d]: must be a map", versionFiltersKey, i))
			continue
		}

		var filter NamedVersionFilter
		if err := mapstructure.Decode(itemMap, &filter); err != nil {
			errs = append(errs, fmt.Errorf("%s[%d]: %w", versionFiltersKey, i, err))
			continue
		}

		if err := validateVariantName(filter.Name, names); err != nil {
			errs = append(errs, fmt.Errorf("%s[%d]: %w", versionFiltersKey, i, err))
			continue
		}

		// The crawler applies the defaults, such as kind "latest", so only validate them here.
		initialized, err := filter.Init()
		if err == nil {
			err = initialized.Validate()
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s[%d] %q: %w", versionFiltersKey, i, filter.Name, err))
			continue
		}

		rawSpec := itemMap
		popKey(rawSpec, "name")

		filters = append(filters, filter)
		rawSpecs = append(rawSpecs, rawSpec)
	}

	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}

	return filters, rawSpecs, nil
}

// decodeGroups decodes and validates "groups".
func decodeGroups(raw any, set bool) ([]Group, error) {
	if !set {
		return nil, nil
	}

	var groups []Group
	if err := mapstructure.Decode(raw, &groups); err != nil {
		return nil, fmt.Errorf("%q: %w", groupsKey, err)
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("%q must not be empty", groupsKey)
	}

	names := map[string]bool{}

	var errs []error
	for i, group := range groups {
		if err := validateVariantName(group.Name, names); err != nil {
			errs = append(errs, fmt.Errorf("%s[%d]: %w", groupsKey, i, err))
			continue
		}

		if group.Name == reservedGroupName {
			errs = append(errs, fmt.Errorf("%s[%d]: name %q is reserved for the dependencies matching no group", groupsKey, i, reservedGroupName))
			continue
		}

		if len(group.Only) == 0 {
			errs = append(errs, fmt.Errorf("%s[%d] %q: %q must not be empty", groupsKey, i, group.Name, onlyKey))
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return groups, nil
}

// validateVariantName checks a version filter or group name, and records it in names.
func validateVariantName(name string, names map[string]bool) error {
	switch {
	case name == "":
		return errors.New(`"name" is required`)
	case !variantNameRegex.MatchString(name):
		return fmt.Errorf("name %q must start with a letter or a digit, and only contain letters, digits, \".\", \"_\" and \"-\"", name)
	case names[name]:
		return fmt.Errorf("name %q is used more than once", name)
	}

	names[name] = true
	return nil
}

// ruleList returns the rules set for key, as a list.
func ruleList(m map[string]any, key string) ([]any, error) {
	value, ok := lookupKey(m, key)
	if !ok || value == nil {
		return nil, nil
	}

	rules, ok := asList(value)
	if !ok {
		return nil, fmt.Errorf("%q must be a list", key)
	}

	return rules, nil
}

// asMap returns a copy of a map with string keys, whatever its named type,
// as decoded manifests may use one, such as CrawlersConfig.
func asMap(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}

	if m, ok := value.(map[string]any); ok {
		return maps.Clone(m), true
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	m := make(map[string]any, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		m[iter.Key().String()] = iter.Value().Interface()
	}

	return m, true
}

// asList returns a list, whatever its element or named type.
func asList(value any) ([]any, bool) {
	if l, ok := value.([]any); ok {
		return l, true
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return nil, false
	}

	l := make([]any, v.Len())
	for i := range l {
		l[i] = v.Index(i).Interface()
	}

	return l, true
}

// lookupKey returns the value for key, matched without case like the crawler spec decoding does.
func lookupKey(m map[string]any, key string) (any, bool) {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

// popKey removes every key matching key without case, and returns the last value found.
func popKey(m map[string]any, key string) (any, bool) {
	var (
		value any
		found bool
	)

	for k, v := range m {
		if strings.EqualFold(k, key) {
			value, found = v, true
			delete(m, k)
		}
	}

	return value, found
}

// setKey replaces every key matching key without case with value.
func setKey(m map[string]any, key string, value any) {
	popKey(m, key)
	m[key] = value
}
