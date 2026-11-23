package autodiscovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandCrawlerSpec(t *testing.T) {
	minor := map[string]any{"name": "minor", "kind": "semver", "pattern": "minor"}
	major := map[string]any{"name": "major", "kind": "semver", "pattern": "majoronly"}
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
			name: "one variant per version filter, in order, without its name",
			kind: "golang",
			spec: map[string]any{
				"only":           []any{gomod},
				"versionfilters": []any{minor, major},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec: map[string]any{
						"only":          []any{gomod},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "minor"},
					},
					variant: Variant{VersionFilter: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec: map[string]any{
						"only":          []any{gomod},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "majoronly"},
					},
					variant: Variant{VersionFilter: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
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
			name: "groups and version filters",
			kind: "golang",
			spec: map[string]any{
				"versionfilters": []any{minor, major},
				"groups":         []any{map[string]any{"name": "aws", "only": []any{aws}}},
			},
			groupBy: GROUPBYINDIVIDUAL,
			expected: []crawlerVariant{
				{
					spec: map[string]any{
						"only":          []any{aws},
						"ignore":        []any{},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "minor"},
					},
					variant: Variant{Group: "aws", VersionFilter: "minor", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec: map[string]any{
						"ignore":        []any{aws},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "minor"},
					},
					variant: Variant{VersionFilter: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
				},
				{
					spec: map[string]any{
						"only":          []any{aws},
						"ignore":        []any{},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "majoronly"},
					},
					variant: Variant{Group: "aws", VersionFilter: "major", Rank: 1, GroupBy: GROUPBYALL},
				},
				{
					spec: map[string]any{
						"ignore":        []any{aws},
						"versionfilter": map[string]any{"kind": "semver", "pattern": "majoronly"},
					},
					variant: Variant{VersionFilter: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
				},
			},
		},
		{
			name: "version filters accept any kind",
			kind: "dockerfile",
			spec: map[string]any{
				"versionfilters": []any{
					map[string]any{"name": "latest", "kind": "latest"},
					map[string]any{"name": "jdk", "kind": "regex", "pattern": `^\d+-jdk$`},
				},
			},
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"versionfilter": map[string]any{"kind": "latest"}},
					variant: Variant{VersionFilter: "latest", Rank: 0, GroupBy: GROUPBYALL},
				},
				{
					spec:    map[string]any{"versionfilter": map[string]any{"kind": "regex", "pattern": `^\d+-jdk$`}},
					variant: Variant{VersionFilter: "jdk", Rank: 1, GroupBy: GROUPBYALL},
				},
			},
		},
		{
			name:    "versionfilter and versionfilters are mutually exclusive",
			kind:    "golang",
			spec:    map[string]any{"versionFilter": map[string]any{"kind": "semver"}, "versionfilters": []any{minor}},
			wantErr: "mutually exclusive",
		},
		{
			name:    "version filter name is required",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{map[string]any{"kind": "semver"}}},
			wantErr: `"name" is required`,
		},
		{
			name:    "version filter names are unique",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{minor, minor}},
			wantErr: "used more than once",
		},
		{
			name:    "version filter name is safe in a pipeline ID",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{map[string]any{"name": "a b", "kind": "semver"}}},
			wantErr: "must start with a letter or a digit",
		},
		{
			name:    "version filter kind is validated",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{map[string]any{"name": "x", "kind": "unknown"}}},
			wantErr: `"x"`,
		},
		{
			name:    "version filter kind defaults to latest",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{map[string]any{"name": "any"}}},
			groupBy: GROUPBYALL,
			expected: []crawlerVariant{
				{
					spec:    map[string]any{"versionfilter": map[string]any{}},
					variant: Variant{VersionFilter: "any", Rank: 0, GroupBy: GROUPBYALL},
				},
			},
		},
		{
			name:    "version filters must not be empty",
			kind:    "golang",
			spec:    map[string]any{"versionfilters": []any{}},
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
			name:    "group name default is reserved",
			kind:    "golang",
			spec:    map[string]any{"groups": []any{map[string]any{"name": "default", "only": []any{aws}}}},
			wantErr: "reserved",
		},
		{
			name:    "WASM plugins are not supported",
			kind:    "plugin",
			spec:    map[string]any{"versionfilters": []any{minor}},
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
		"ignore":         []any{"a"},
		"versionfilters": []any{map[string]any{"name": "minor", "kind": "semver", "pattern": "minor"}},
		"groups":         []any{map[string]any{"name": "g", "only": []any{"b"}}},
	}

	_, err := expandCrawlerSpec("golang", spec, GROUPBYALL)
	require.NoError(t, err)

	assert.Contains(t, spec, "versionfilters")
	assert.Contains(t, spec, "groups")
	assert.Equal(t, []any{"a"}, spec["ignore"])
}

func TestNewBuildsOneCrawlerPerVariant(t *testing.T) {
	ad, err := New(Config{
		GroupBy: GROUPBYINDIVIDUAL,
		Crawlers: CrawlersConfig{
			"golang": map[string]any{
				"versionfilters": []any{
					map[string]any{"name": "minor", "kind": "semver", "pattern": "minor"},
					map[string]any{"name": "major", "kind": "semver", "pattern": "majoronly"},
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
		{Group: "aws", VersionFilter: "minor", Rank: 0, GroupBy: GROUPBYALL},
		{VersionFilter: "minor", Rank: 0, GroupBy: GROUPBYINDIVIDUAL},
		{Group: "aws", VersionFilter: "major", Rank: 1, GroupBy: GROUPBYALL},
		{VersionFilter: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL},
	}, got)
}

func TestNewRejectsInvalidVariants(t *testing.T) {
	_, err := New(Config{
		Crawlers: CrawlersConfig{
			"golang": map[string]any{
				"versionfilter":  map[string]any{"kind": "semver"},
				"versionfilters": []any{map[string]any{"name": "minor", "kind": "semver"}},
			},
		},
	}, t.TempDir())
	require.Error(t, err)
}

// Decoded manifests can hold crawler specs with a named map type, such as CrawlersConfig.
func TestExpandCrawlerSpecNamedMapTypes(t *testing.T) {
	spec := CrawlersConfig{
		"versionfilters": []any{
			CrawlersConfig{"name": "minor", "kind": "semver", "pattern": "minor"},
			CrawlersConfig{"name": "major", "kind": "semver", "pattern": "majoronly"},
		},
	}

	got, err := expandCrawlerSpec("golang", spec, GROUPBYINDIVIDUAL)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, map[string]any{"versionfilter": map[string]any{"kind": "semver", "pattern": "majoronly"}}, got[1].spec)
	assert.Equal(t, Variant{VersionFilter: "major", Rank: 1, GroupBy: GROUPBYINDIVIDUAL}, got[1].variant)
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
