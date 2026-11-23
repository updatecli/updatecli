package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
)

func TestVariantHelpers(t *testing.T) {
	individual := autodiscovery.Variant{Track: "major", Rank: 1, GroupBy: autodiscovery.GROUPBYINDIVIDUAL}
	grouped := autodiscovery.Variant{Group: "aws", Track: "minor", GroupBy: autodiscovery.GROUPBYALL}
	groupOnly := autodiscovery.Variant{Group: "aws", GroupBy: autodiscovery.GROUPBYALL}
	plain := autodiscovery.Variant{GroupBy: autodiscovery.GROUPBYALL}

	assert.Equal(t, "parent", variantPipelineID("parent", "golang", plain))
	assert.Equal(t, "parent/golang/default/major", variantPipelineID("parent", "golang", individual))
	assert.Equal(t, "parent/golang/aws/minor", variantPipelineID("parent", "golang", grouped))
	assert.Equal(t, "parent/golang/aws", variantPipelineID("parent", "golang", groupOnly))

	assert.Equal(t, map[string]string{"updatecli.io/autodiscovery": "golang", "updatecli.io/parent": "parent"}, autodiscoveryLabels("parent", "golang", plain))
	assert.Equal(t, map[string]string{
		"updatecli.io/autodiscovery": "golang",
		"updatecli.io/parent":        "parent",
		"updatecli.io/group":         "aws",
		"updatecli.io/track":         "minor",
	}, autodiscoveryLabels("parent", "golang", grouped))
	assert.Empty(t, autodiscoveryLabels("", "", plain))

	assert.Empty(t, variantTitleSuffix("golang", plain))
	assert.Equal(t, "(major)", variantTitleSuffix("golang", individual))
	assert.Equal(t, "(golang, aws, minor)", variantTitleSuffix("golang", grouped))
	assert.Equal(t, "(dockerfile, minor)", variantTitleSuffix("dockerfile", autodiscovery.Variant{Track: "minor", GroupBy: autodiscovery.GROUPBYALL}))

	assert.Nil(t, variantGate("m", "golang", plain, autodiscovery.GROUPBYALL, "x"))
	assert.Nil(t, variantGate("m", "golang", groupOnly, autodiscovery.GROUPBYALL, "x"))
	assert.Equal(t,
		&pipeline.Gate{Key: "m/golang/default/deps(go): bump module x", Rank: 1, Track: "major"},
		variantGate("m", "golang", individual, autodiscovery.GROUPBYINDIVIDUAL, "deps(go): bump module x"))
	assert.Equal(t,
		&pipeline.Gate{Key: "m/golang/aws", Rank: 0, Track: "minor"},
		variantGate("m", "golang", grouped, autodiscovery.GROUPBYALL, "deps(go): bump module x"))
}

func TestGeneratedActionConfig(t *testing.T) {
	groupedTitle := func() string { return "Bump Go modules" }

	tests := []struct {
		name                string
		autodiscoveryAction action.Config
		crawlerAction       action.Config
		variant             autodiscovery.Variant
		expectedTitle       string
	}{
		{
			name:          "grouped without variant uses the autodiscovery title",
			variant:       autodiscovery.Variant{GroupBy: autodiscovery.GROUPBYALL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "Bump Go modules",
		},
		{
			name:          "individual without variant uses the crawler title",
			variant:       autodiscovery.Variant{GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "deps(go): bump module x",
		},
		{
			name:          "grouped variant adds the group and track",
			variant:       autodiscovery.Variant{Group: "aws", Track: "major", GroupBy: autodiscovery.GROUPBYALL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "Bump Go modules (golang, aws, major)",
		},
		{
			name:          "individual variant adds the track to the crawler title",
			variant:       autodiscovery.Variant{Track: "major", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "deps(go): bump module x (major)",
		},
		{
			name:                "user title takes precedence and gets the variant",
			autodiscoveryAction: action.Config{Title: "Go update"},
			variant:             autodiscovery.Variant{Track: "minor", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			crawlerAction:       action.Config{Title: "deps(go): bump module x"},
			expectedTitle:       "Go update (minor)",
		},
		{
			name:          "individual variant without any title uses the manifest name",
			variant:       autodiscovery.Variant{Track: "minor", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			expectedTitle: "Bump module x (minor)",
		},
		{
			name:          "empty title stays empty without variant",
			variant:       autodiscovery.Variant{GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			expectedTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatedActionConfig(tt.autodiscoveryAction, tt.crawlerAction, "golang", tt.variant, groupedTitle, "Bump module x")
			assert.Equal(t, tt.expectedTitle, got.Title)
		})
	}
}

func TestWithAutodiscoveryLabels(t *testing.T) {
	variant := autodiscovery.Variant{Group: "aws", Track: "minor"}
	expected := map[string]string{
		"updatecli.io/autodiscovery": "golang",
		"updatecli.io/parent":        "parent",
		"updatecli.io/group":         "aws",
		"updatecli.io/track":         "minor",
	}

	got := withAutodiscoveryLabels(map[string]string{"updatecli.io/group": "infra", "updatecli.io/autodiscovery": "npm", "group": "backend", "team": "core"}, "parent", "golang", variant, "Bump x")
	assert.Equal(t, map[string]string{
		"updatecli.io/autodiscovery": "golang",
		"updatecli.io/parent":        "parent",
		"updatecli.io/group":         "aws",
		"updatecli.io/track":         "minor",
		"group":                      "backend",
		"team":                       "core",
	}, got)

	assert.Equal(t, expected, withAutodiscoveryLabels(nil, "parent", "golang", variant, "Bump x"))
	assert.Nil(t, withAutodiscoveryLabels(nil, "", "", autodiscovery.Variant{}, "Bump x"))
}

// TestVariantIDsDontCollideAcrossKinds checks that a crawler kind containing "/"
// can't produce the same ID or gate key as another kind with a group and filter.
func TestVariantIDsDontCollideAcrossKinds(t *testing.T) {
	golangVariant := autodiscovery.Variant{Group: "gomod", Track: "aws", GroupBy: autodiscovery.GROUPBYALL}
	gomodVariant := autodiscovery.Variant{Group: "aws", GroupBy: autodiscovery.GROUPBYALL}

	golangID := variantPipelineID("parent", "golang", golangVariant)
	gomodID := variantPipelineID("parent", "golang/gomod", gomodVariant)

	assert.Equal(t, "parent/golang/gomod/aws", golangID)
	assert.Equal(t, "parent/golang-gomod/aws", gomodID)
	assert.NotEqual(t, golangID, gomodID)

	gomodFiltered := autodiscovery.Variant{Group: "aws", Track: "aws", GroupBy: autodiscovery.GROUPBYALL}
	assert.NotEqual(t,
		variantGate("m", "golang", golangVariant, autodiscovery.GROUPBYALL, "x").Key,
		variantGate("m", "golang/gomod", gomodFiltered, autodiscovery.GROUPBYALL, "x").Key)
}
