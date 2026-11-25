package autodiscovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandCrawlerSpec(t *testing.T) {
	minorFilter := map[string]any{"kind": "semver", "pattern": "minor"}
	majorFilter := map[string]any{"kind": "semver", "pattern": "majoronly"}
	minor := map[string]any{"name": "minor", "versionfilter": minorFilter}
	major := map[string]any{"name": "major", "versionfilter": majorFilter}
	aws := map[string]any{"modules": map[string]any{"github.com/aws/*": ""}}
	gx := map[string]any{"modules": map[string]any{"golang.org/*": ""}}
	gomod := map[string]any{"path": "go.mod"}
	awsInGomod := map[string]any{"path": "go.mod", "modules": map[string]any{"github.com/aws/*": ""}}
	gxInGomod := map[string]any{"path": "go.mod", "modules": map[string]any{"golang.org/*": ""}}

	tests := []struct {
		name     string
		kind     string
		spec     any
		groupBy  GroupBy
		expected []crawlerVariant
		wantErr  string
	}{
		{
			name:    "spec without variant settings is unchanged",
			kind:    "golang",
			spec:    map[string]any{"only": []any{gomod}},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{spec: map[string]any{"only": []any{gomod}}, variant: Variant{GroupBy: GROUPBYINDIVIDUAL}},
			},
		},
		{
			name:    "default crawler spec struct is unchanged",
			kind:    "golang",
			spec:    nil,
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{spec: nil, variant: Variant{GroupBy: GROUPBYALL}},
			},
		},
		{
			name: "one variant per track, in order, without its name",
			kind: "golang",
			spec: map[string]any{
				"only":   []any{gomod},
				"tracks": []any{minor, major},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec: map[string]any{
						"only":          []any{gomod},
						"versionfilter": minorFilter,
					},
					variant: Variant{Track: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec: map[string]any{
						"only":          []any{gomod},
						"versionfilter": majorFilter,
					},
					variant: Variant{Track: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name: "groups claim dependencies in order, the rest ignores every group",
			kind: "golang",
			spec: map[string]any{
				"only":   []any{gomod},
				"ignore": []any{gomod},
				"groups": []any{
					map[string]any{"name": "aws", "only": []any{aws}},
					map[string]any{"name": "golang.org", "only": []any{gx}},
				},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"only": []any{awsInGomod}, "ignore": []any{gomod}},
					variant: Variant{Group: "aws", GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"only": []any{gxInGomod}, "ignore": []any{gomod, aws}},
					variant: Variant{Group: "golang.org", GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"only": []any{gomod}, "ignore": []any{gomod, aws, gx}},
					variant: Variant{GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name: "groups and tracks",
			kind: "golang",
			spec: map[string]any{
				"tracks": []any{minor, major},
				"groups": []any{map[string]any{"name": "aws", "only": []any{aws}}},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec: map[string]any{
						"only":          []any{aws},
						"ignore":        []any{},
						"versionfilter": minorFilter,
					},
					variant: Variant{Group: "aws", Track: "minor", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec: map[string]any{
						"ignore":        []any{aws},
						"versionfilter": minorFilter,
					},
					variant: Variant{Track: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec: map[string]any{
						"only":          []any{aws},
						"ignore":        []any{},
						"versionfilter": majorFilter,
					},
					variant: Variant{Group: "aws", Track: "major", Rank: 1, GroupBy: GROUPBYALL},
				},
				{
					spec: map[string]any{
						"ignore":        []any{aws},
						"versionfilter": majorFilter,
					},
					variant: Variant{Track: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name: "track version filters accept any kind",
			kind: "dockerfile",
			spec: map[string]any{
				"tracks": []any{
					map[string]any{"name": "latest", "versionfilter": map[string]any{"kind": "latest"}},
					map[string]any{"name": "jdk", "versionfilter": map[string]any{"kind": "regex", "pattern": `^\d+-jdk$`}},
				},
			},
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"versionfilter": map[string]any{"kind": "latest"}},
					variant: Variant{Track: "latest", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"versionfilter": map[string]any{"kind": "regex", "pattern": `^\d+-jdk$`}},
					variant: Variant{Track: "jdk", Rank: 1, GroupBy: GROUPBYALL},
				},
			},
		},
		{
			name: "tracks keep the crawler settings they don't set",
			kind: "golang",
			spec: map[string]any{
				"age":    map[string]any{"minimum": "3d"},
				"tracks": []any{minor, major},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"age": map[string]any{"minimum": "3d"}, "versionfilter": minorFilter},
					variant: Variant{Track: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec:    map[string]any{"age": map[string]any{"minimum": "3d"}, "versionfilter": majorFilter},
					variant: Variant{Track: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name: "tracks set both version filter and age",
			kind: "npm",
			spec: map[string]any{
				"tracks": []any{
					map[string]any{"name": "minor", "versionfilter": minorFilter, "age": map[string]any{"minimum": "3d"}},
					map[string]any{"name": "major", "versionfilter": majorFilter, "Age": map[string]any{"minimum": "14d", "allowdowngrade": true}},
				},
			},
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"versionfilter": minorFilter, "age": map[string]any{"minimum": "3d"}},
					variant: Variant{Track: "minor", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"versionfilter": majorFilter, "age": map[string]any{"minimum": "14d", "allowdowngrade": true}},
					variant: Variant{Track: "major", Rank: 1, GroupBy: GROUPBYALL},
				},
			},
		},
		{
			name:    "a track setting can't also be set on the crawler",
			kind:    "golang",
			spec:    map[string]any{"versionFilter": map[string]any{"kind": "semver"}, "tracks": []any{minor}},
			wantErr: `"versionfilter" is set both on the crawler and in track "minor"`,
		},
		{
			name: "a track age can't also be set on the crawler",
			kind: "golang",
			spec: map[string]any{
				"age":    map[string]any{"minimum": "1d"},
				"tracks": []any{minor, map[string]any{"name": "slow", "age": map[string]any{"minimum": "14d"}}},
			},
			wantErr: `"age" is set both on the crawler and in track "slow"`,
		},
		{
			name:    "age is rejected for crawlers without age",
			kind:    "helm",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "slow", "age": map[string]any{"minimum": "14d"}}}},
			wantErr: `"age" isn't supported by the "helm" crawler`,
		},
		{
			name:    "unknown track settings are rejected",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "security", "vulnerability": map[string]any{}}}},
			wantErr: "vulnerability",
		},
		{
			name:    "version filter settings must be nested",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "minor", "kind": "semver", "pattern": "minor"}}},
			wantErr: "kind",
		},
		{
			name:    "a track sets a version filter or an age",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "empty"}}},
			wantErr: `tracks[0] "empty": must set "versionfilter" or "age"`,
		},
		{
			name:    "a setting without a value sets nothing",
			kind:    "golang",
			spec:    map[string]any{"versionfilter": minorFilter, "tracks": []any{map[string]any{"name": "major", "versionfilter": nil}}},
			wantErr: `tracks[0] "major": must set "versionfilter" or "age"`,
		},
		{
			name:    "track name is required",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"versionfilter": minorFilter}}},
			wantErr: `"name" is required`,
		},
		{
			name:    "track names are unique",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{minor, minor}},
			wantErr: "used more than once",
		},
		{
			name:    "track name is safe in a pipeline ID",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "a b", "versionfilter": minorFilter}}},
			wantErr: "must start with a letter or a digit",
		},
		{
			name:    "track version filter kind is validated",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "x", "versionfilter": map[string]any{"kind": "unknown"}}}},
			wantErr: `tracks[0] "x": versionfilter`,
		},
		{
			name:    "track age is validated",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "x", "age": map[string]any{"minimum": "7"}}}},
			wantErr: `tracks[0] "x": age`,
		},
		{
			name:    "track version filter kind defaults to latest",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "any", "versionfilter": map[string]any{}}}},
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"versionfilter": map[string]any{}},
					variant: Variant{Track: "any", Rank: 0, GroupBy: GROUPBYALL},
				},
			},
		},
		{
			name: "a track with wait false doesn't wait, and wait isn't a crawler setting",
			kind: "golang",
			spec: map[string]any{
				"tracks": []any{
					map[string]any{"name": "minor", "versionfilter": minorFilter, "wait": true},
					map[string]any{"name": "lts", "versionfilter": majorFilter, "Wait": false},
				},
				"groups": []any{map[string]any{"name": "aws", "only": []any{aws}}},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"only": []any{aws}, "ignore": []any{}, "versionfilter": minorFilter},
					variant: Variant{Group: "aws", Track: "minor", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"ignore": []any{aws}, "versionfilter": minorFilter},
					variant: Variant{Track: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec:    map[string]any{"only": []any{aws}, "ignore": []any{}, "versionfilter": majorFilter},
					variant: Variant{Group: "aws", Track: "lts", Rank: 1, NoWait: true, GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"ignore": []any{aws}, "versionfilter": majorFilter},
					variant: Variant{Track: "lts", Rank: 1, NoWait: true, GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name:    "a track with only wait sets nothing",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "lts", "wait": false}}},
			wantErr: `tracks[0] "lts": must set "versionfilter" or "age"`,
		},
		{
			name:    "wait must be a boolean",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "lts", "versionfilter": minorFilter, "wait": "no"}}},
			wantErr: "expected type 'bool'",
		},
		{
			name:    "tracks must not be empty",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{}},
			wantErr: "must not be empty",
		},
		{
			name:    "group only is required",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "aws"}}},
			wantErr: `"only" must not be empty`,
		},
		{
			name: "group only must combine with one crawler only rule",
			kind: "golang",
			spec: map[string]any{
				"only":   []any{gomod},
				"groups": []any{map[string]any{"name": "nested", "only": []any{map[string]any{"path": "*/go.mod"}}}},
			},
			wantErr: `group "nested"`,
		},
		{
			name:    "track names can't contain the pipeline ID separator",
			kind:    "golang",
			spec:    map[string]any{"tracks": []any{map[string]any{"name": "b_c", "versionfilter": minorFilter}}},
			wantErr: "must start with a letter or a digit",
		},
		{
			name:    "group names can't contain the pipeline ID separator",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "a_b", "only": []any{aws}}}},
			wantErr: "must start with a letter or a digit",
		},
		{
			name:    "unknown group settings are rejected",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "aws", "only": []any{aws}, "ignore": []any{gx}}}},
			wantErr: "ignore",
		},
		{
			name:    "group name default is reserved whatever its case",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "Default", "only": []any{aws}}}},
			wantErr: "reserved",
		},
		{
			name:    "group name default is reserved",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "default", "only": []any{aws}}}},
			wantErr: "reserved",
		},
		{
			name:    "WASM plugins are not supported",
			kind:    "plugin",
			spec:    map[string]any{"tracks": []any{minor}},
			wantErr: "WASM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandCrawlerSpec(tt.kind, tt.spec, tt.groupBy)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestExpandCrawlerSpecDoesNotChangeInput(t *testing.T) {
	spec := map[string]any{
		"ignore": []any{"a"},
		"tracks": []any{map[string]any{"name": "minor", "versionfilter": map[string]any{"kind": "semver", "pattern": "minor"}}},
		"groups": []any{map[string]any{"name": "g", "only": []any{"b"}}},
	}

	_, err := expandCrawlerSpec("golang", spec, GROUPBYALL)
	require.NoError(t, err)

	assert.Contains(t, spec, "tracks")
	assert.Contains(t, spec, "groups")
	assert.Equal(t, []any{"a"}, spec["ignore"])
}

func TestNewBuildsOneCrawlerPerVariant(t *testing.T) {
	ad, err := New(Config{
		GroupBy: GROUPBYINDIVIDUAL,
		Crawlers: CrawlersConfig{
			"golang": map[string]any{
				"tracks": []any{
					map[string]any{"name": "minor", "versionfilter": map[string]any{"kind": "semver", "pattern": "minor"}},
					map[string]any{"name": "major", "versionfilter": map[string]any{"kind": "semver", "pattern": "majoronly"}},
				},
				"groups": []any{map[string]any{"name": "aws", "only": []any{map[string]any{"modules": map[string]any{"github.com/aws/*": ""}}}}},
			},
		},
	}, t.TempDir())
	require.NoError(t, err)

	var got []Variant
	for _, c := range ad.crawlers {
		got = append(got, c.variant)
	}

	assert.Equal(t, []Variant{
		{Group: "aws", Track: "minor", Rank: 0, GroupBy: GROUPBYALL},
		{Track: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
		{Group: "aws", Track: "major", Rank: 1, GroupBy: GROUPBYALL},
		{Track: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
	}, got)
}

func TestNewRejectsInvalidVariants(t *testing.T) {
	_, err := New(Config{
		Crawlers: CrawlersConfig{
			"golang": map[string]any{
				"versionfilter": map[string]any{"kind": "semver"},
				"tracks":        []any{map[string]any{"name": "minor", "versionfilter": map[string]any{"kind": "semver"}}},
			},
		},
	}, t.TempDir())
	require.Error(t, err)
}

// Decoded manifests can hold crawler specs with a named map type, such as CrawlersConfig.
func TestExpandCrawlerSpecNamedMapTypes(t *testing.T) {
	spec := CrawlersConfig{
		"tracks": []any{
			CrawlersConfig{"name": "minor", "versionfilter": CrawlersConfig{"kind": "semver", "pattern": "minor"}},
			CrawlersConfig{"name": "major", "versionfilter": CrawlersConfig{"kind": "semver", "pattern": "majoronly"}, "age": CrawlersConfig{"minimum": "7d"}},
		},
	}

	got, err := expandCrawlerSpec("golang", spec, GROUPBYINDIVIDUAL)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, map[string]any{
		"versionfilter": CrawlersConfig{"kind": "semver", "pattern": "majoronly"},
		"age":           CrawlersConfig{"minimum": "7d"},
	}, got[1].spec)
	assert.Equal(t, Variant{Track: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL}, got[1].variant)
}

func TestIntersectRules(t *testing.T) {
	gomod := map[string]any{"path": "go.mod"}
	nested := map[string]any{"path": "*/go.mod"}
	aws := map[string]any{"modules": map[string]any{"github.com/aws/*": ""}}
	gx := map[string]any{"modules": map[string]any{"golang.org/*": ""}}

	got, err := intersectRules(nil, []any{aws})
	require.NoError(t, err)
	assert.Equal(t, []any{aws}, got)

	got, err = intersectRules([]any{gomod, nested}, []any{aws, gx})
	require.NoError(t, err)
	assert.Equal(t, []any{
		map[string]any{"path": "go.mod", "modules": map[string]any{"github.com/aws/*": ""}},
		map[string]any{"path": "go.mod", "modules": map[string]any{"golang.org/*": ""}},
		map[string]any{"path": "*/go.mod", "modules": map[string]any{"github.com/aws/*": ""}},
		map[string]any{"path": "*/go.mod", "modules": map[string]any{"golang.org/*": ""}},
	}, got)

	// The same field set to the same value, whatever its case, is kept once.
	got, err = intersectRules([]any{gomod}, []any{map[string]any{"Path": "go.mod"}})
	require.NoError(t, err)
	assert.Equal(t, []any{gomod}, got)

	// A pair setting the same field to different values matches nothing, so it's dropped.
	got, err = intersectRules([]any{gomod, nested}, []any{map[string]any{"path": "go.mod", "modules": map[string]any{"github.com/aws/*": ""}}})
	require.NoError(t, err)
	assert.Equal(t, []any{map[string]any{"path": "go.mod", "modules": map[string]any{"github.com/aws/*": ""}}}, got)

	_, err = intersectRules([]any{gomod}, []any{nested})
	require.ErrorContains(t, err, "no \"only\" rule can be combined")
}

func TestCrawlerSupports(t *testing.T) {
	for kind := range GetAutodiscoverySpecsMapping() {
		if kind == pluginCrawlerKind {
			continue
		}
		assert.Truef(t, crawlerSupports(kind, versionFilterKey), "crawler %q should support %q", kind, versionFilterKey)
	}

	for _, kind := range []string{"golang", "npm", "maven", "githubaction"} {
		assert.Truef(t, crawlerSupports(kind, ageKey), "crawler %q should support %q", kind, ageKey)
	}

	assert.False(t, crawlerSupports("helm", ageKey))
	assert.True(t, crawlerSupports("unknown", ageKey))
}

func TestCrawlerSchemasOfferAgeOnlyWhenSupported(t *testing.T) {
	schema := CrawlersConfig{}.JSONSchema()

	trackProperties := func(kind string) []string {
		crawler, ok := schema.Properties.Get(kind)
		require.Truef(t, ok, "crawler %q", kind)
		tracks, ok := crawler.Properties.Get(tracksKey)
		require.Truef(t, ok, "crawler %q tracks", kind)

		var keys []string
		for pair := tracks.Items.Properties.Oldest(); pair != nil; pair = pair.Next() {
			keys = append(keys, pair.Key)
		}
		return keys
	}

	assert.Contains(t, trackProperties("golang"), ageKey)
	assert.NotContains(t, trackProperties("helm"), ageKey)
	assert.Contains(t, trackProperties("helm"), versionFilterKey)
}
