package autodiscovery

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// pluginCrawlerKind is the crawler kind used by WASM autodiscovery plugins.
	pluginCrawlerKind = "plugin"

	// reservedGroupName names, in pipeline IDs, the dependencies matching no group.
	reservedGroupName = "default"

	tracksKey        = "tracks"
	groupsKey        = "groups"
	nameKey          = "name"
	versionFilterKey = "versionfilter"
	ageKey           = "age"
	onlyKey          = "only"
	ignoreKey        = "ignore"
)

var (
	// variantNameRegex restricts names to characters safe in a pipeline ID, a git branch and a label value.
	// "_" separates the parts of a variant pipeline ID, so names can't contain it.
	variantNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*$`)

	// trackSettingKeys are the crawler settings a track sets, in place of the crawler ones.
	trackSettingKeys = []string{versionFilterKey, ageKey}
)

// VariantSpec holds the settings, shared by every crawler, that split one crawler into several variants.
type VariantSpec struct {
	// "tracks" defines an ordered list of tracks, each one choosing versions its own way.
	// The crawler runs once per track, and each track gets its own pipelines.
	//
	// default:
	//   empty, the crawler runs once with its own settings.
	//
	// remark:
	//   * each track requires a "name", and sets "versionfilter", "age" or both.
	//   * a setting used by a track can't also be set on the crawler,
	//     so a track without it uses the crawler default.
	//   * "age" is only accepted by crawlers supporting it.
	//   * the name is added to the generated pipeline IDs and action titles,
	//     and set as the "updatecli.io/track" label, so "--labels updatecli.io/track:major" selects them.
	//   * a track waits while an earlier track still has a pending update,
	//     so a major update is proposed once no minor update is pending.
	//     With "groupby: individual", it waits for the same dependency.
	//     With "groupby: all", or in a group, the dependencies share one pull request per crawler,
	//     so it waits for the whole pull request: one pending minor update holds back every major update of that crawler.
	//     An update is pending while its pull request is open or, without an action, while its working branch exists.
	//     Tracks are run in the order they are listed.
	//   * the version filters of the tracks should not overlap, for example "minor" then "majoronly".
	//     Tracks setting only "age" share the crawler version filter, so they overlap:
	//     with "allowdowngrade", one track may undo the updates of another.
	//   * it isn't supported by WASM plugins.
	//
	// example:
	//   ```
	//   tracks:
	//     - name: minor
	//       versionfilter:
	//         kind: semver
	//         pattern: minor
	//       age:
	//         minimum: 3d
	//     - name: major
	//       versionfilter:
	//         kind: semver
	//         pattern: majoronly
	//       age:
	//         minimum: 14d
	//   ```
	//
	Tracks []Track `yaml:",omitempty"`
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

// Track defines how one variant of a crawler chooses versions.
type Track struct {
	// "name" identifies the track in the generated pipelines.
	//
	// remark:
	//   * it must be unique, start with a letter or a digit, and only contain letters, digits, "." and "-".
	//
	Name string `yaml:",omitempty" jsonschema:"required"`
	// "versionfilter" defines the version filter used by the track, as the crawler "versionfilter" does.
	//
	// default:
	//   the crawler default version filter.
	//
	VersionFilter *version.Filter `yaml:",omitempty"`
	// "age" defines the release age accepted by the track, as the crawler "age" does.
	//
	// default:
	//   the crawler default age.
	//
	// remark:
	//   * it's only accepted by crawlers supporting "age".
	//
	Age *age.AutodiscoverySpec `yaml:",omitempty"`
}

// Group defines dependencies updated together.
type Group struct {
	// "name" identifies the group in the generated pipelines.
	//
	// remark:
	//   * it must be unique, start with a letter or a digit, and only contain letters, digits, "." and "-".
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
	// Track is the track name, empty when "tracks" isn't set.
	Track string
	// Rank is the position of the track in "tracks".
	Rank int
	// GroupBy is the grouping used for the pipelines of this variant.
	GroupBy GroupBy
}

// IsDefault reports whether the variant covers the whole crawler, as without "tracks" and "groups".
func (v Variant) IsDefault() bool {
	return v.Group == "" && v.Track == ""
}

// crawlerVariant is a crawler spec ready for one crawler run.
type crawlerVariant struct {
	spec    any
	variant Variant
}

// expandCrawlerSpec splits a crawler spec into one spec per group and track.
//
// The crawler code is untouched: "tracks" and "groups" are removed from the spec,
// and the crawler settings they map to, "versionfilter", "age", "only" and "ignore", are set instead.
// Variants are returned ordered by track rank, so earlier tracks run first.
func expandCrawlerSpec(kind string, spec any, groupBy GroupBy) ([]crawlerVariant, error) {
	defaultVariant := []crawlerVariant{{spec: spec, variant: Variant{GroupBy: groupBy}}}

	specMap, ok := asMap(spec)
	if !ok {
		return defaultVariant, nil
	}

	base := specMap
	rawTracks, hasTracks := popKey(base, tracksKey)
	rawGroups, hasGroups := popKey(base, groupsKey)

	if !hasTracks && !hasGroups {
		return defaultVariant, nil
	}

	if kind == pluginCrawlerKind || strings.HasSuffix(kind, ".wasm") {
		return nil, fmt.Errorf("%q and %q aren't supported by WASM plugins", tracksKey, groupsKey)
	}

	tracks, trackSettings, err := decodeTracks(kind, rawTracks, hasTracks)
	if err != nil {
		return nil, err
	}

	for _, key := range trackSettingKeys {
		if _, ok := lookupKey(base, key); !ok {
			continue
		}

		for i, settings := range trackSettings {
			if _, ok := settings[key]; ok {
				return nil, fmt.Errorf("%q is set both on the crawler and in track %q: set it in each track instead", key, tracks[i].Name)
			}
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

	// Without "tracks", every variant keeps the crawler's own settings.
	if len(tracks) == 0 {
		tracks = []Track{{}}
		trackSettings = []map[string]any{nil}
	}

	var variants []crawlerVariant

	for rank, track := range tracks {
		withTrack := func(m map[string]any) map[string]any {
			for key, value := range trackSettings[rank] {
				setKey(m, key, value)
			}
			return m
		}

		// Each group ignores the dependencies claimed by the groups before it.
		claimed := append([]any{}, parentIgnore...)

		for i, group := range groups {
			m := withTrack(maps.Clone(base))
			setKey(m, onlyKey, append([]any{}, groupOnly[i]...))
			setKey(m, ignoreKey, append([]any{}, claimed...))

			variants = append(variants, crawlerVariant{
				spec: m,
				variant: Variant{
					Group:   group.Name,
					Track:   track.Name,
					Rank:    rank,
					GroupBy: GROUPBYALL,
				},
			})

			claimed = append(claimed, group.Only...)
		}

		m := withTrack(maps.Clone(base))
		if len(groups) > 0 {
			setKey(m, ignoreKey, claimed)
		}

		variants = append(variants, crawlerVariant{
			spec: m,
			variant: Variant{
				Track:   track.Name,
				Rank:    rank,
				GroupBy: groupBy,
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

// decodeTracks decodes and validates "tracks".
// It also returns, for each track, the crawler settings it sets, keyed by their lowercase name.
func decodeTracks(kind string, raw any, set bool) ([]Track, []map[string]any, error) {
	if !set {
		return nil, nil, nil
	}

	rawList, ok := asList(raw)
	if !ok {
		return nil, nil, fmt.Errorf("%q must be a list", tracksKey)
	}

	if len(rawList) == 0 {
		return nil, nil, fmt.Errorf("%q must not be empty", tracksKey)
	}

	tracks := make([]Track, 0, len(rawList))
	settings := make([]map[string]any, 0, len(rawList))
	names := map[string]bool{}

	var errs []error
	for i, item := range rawList {
		track, trackSettings, err := decodeTrack(kind, item, names)
		if err != nil {
			if track.Name != "" {
				errs = append(errs, fmt.Errorf("%s[%d] %q: %w", tracksKey, i, track.Name, err))
			} else {
				errs = append(errs, fmt.Errorf("%s[%d]: %w", tracksKey, i, err))
			}
			continue
		}

		tracks = append(tracks, track)
		settings = append(settings, trackSettings)
	}

	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}

	return tracks, settings, nil
}

// decodeTrack decodes and validates one track, and records its name in names.
// The returned track holds its name, when decoded, even on error.
func decodeTrack(kind string, item any, names map[string]bool) (Track, map[string]any, error) {
	var track Track

	itemMap, ok := asMap(item)
	if !ok {
		return track, nil, errors.New("must be a map")
	}

	// Reject unknown settings, as the crawler would otherwise ignore a mistyped or unsupported one.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &track,
	})
	if err != nil {
		return track, nil, err
	}

	if err := decoder.Decode(itemMap); err != nil {
		return track, nil, err
	}

	if err := validateVariantName(track.Name, names); err != nil {
		return Track{}, nil, err
	}

	// A setting written without a value, such as "versionfilter:", sets nothing.
	settings := map[string]any{}
	for _, key := range trackSettingKeys {
		if value, ok := lookupKey(itemMap, key); ok && value != nil {
			settings[key] = value
		}
	}

	if len(settings) == 0 {
		return track, nil, fmt.Errorf("must set %q or %q", versionFilterKey, ageKey)
	}

	for key := range settings {
		if !crawlerSupports(kind, key) {
			return track, nil, fmt.Errorf("%q isn't supported by the %q crawler", key, kind)
		}
	}

	if track.VersionFilter != nil {
		// The crawler applies the defaults, such as kind "latest", so only validate them here.
		initialized, err := track.VersionFilter.Init()
		if err == nil {
			err = initialized.Validate()
		}
		if err != nil {
			return track, nil, fmt.Errorf("%s: %w", versionFilterKey, err)
		}
	}

	if track.Age != nil {
		if err := track.Age.Validate(); err != nil {
			return track, nil, fmt.Errorf("%s: %w", ageKey, err)
		}
	}

	return track, settings, nil
}

// crawlerSupports reports whether the spec of the crawler kind has a setting named key.
// Unknown kinds, and specs that aren't structs, are reported as supporting it, leaving the check to the crawler.
func crawlerSupports(kind, key string) bool {
	spec, ok := GetAutodiscoverySpecsMapping()[kind]
	if !ok {
		return true
	}

	return structHasSetting(reflect.TypeOf(spec), key)
}

// structHasSetting reports whether a struct type has a field decoded from key, as mapstructure matches them.
func structHasSetting(t reflect.Type, key string) bool {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t == nil || t.Kind() != reflect.Struct {
		return true
	}

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tagName, tagOptions, _ := strings.Cut(field.Tag.Get("mapstructure"), ",")
		if field.Anonymous && tagName == "" && strings.Contains(tagOptions, "squash") {
			if structHasSetting(field.Type, key) {
				return true
			}
			continue
		}

		name := field.Name
		if tagName != "" {
			name = tagName
		}

		if strings.EqualFold(name, key) {
			return true
		}
	}

	return false
}

// decodeGroups decodes and validates "groups".
func decodeGroups(raw any, set bool) ([]Group, error) {
	if !set {
		return nil, nil
	}

	// Reject unknown settings instead of dropping them.
	var groups []Group
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &groups,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(raw); err != nil {
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

		if strings.EqualFold(group.Name, reservedGroupName) {
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
		return fmt.Errorf("name %q must start with a letter or a digit, and only contain letters, digits, \".\" and \"-\"", name)
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
