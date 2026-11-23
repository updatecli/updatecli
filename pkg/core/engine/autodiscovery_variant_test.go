package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
)

func TestVariantHelpers(t *testing.T) {
	individual := autodiscovery.Variant{VersionFilter: "major", Rank: 1, GroupBy: autodiscovery.GROUPBYINDIVIDUAL}
	grouped := autodiscovery.Variant{Group: "aws", VersionFilter: "minor", GroupBy: autodiscovery.GROUPBYALL}
	groupOnly := autodiscovery.Variant{Group: "aws", GroupBy: autodiscovery.GROUPBYALL}
	plain := autodiscovery.Variant{GroupBy: autodiscovery.GROUPBYALL}

	assert.Equal(t, "parent", variantPipelineID("parent", "golang", plain))
	assert.Equal(t, "parent/golang/default/major", variantPipelineID("parent", "golang", individual))
	assert.Equal(t, "parent/golang/aws/minor", variantPipelineID("parent", "golang", grouped))
	assert.Equal(t, "parent/golang/aws", variantPipelineID("parent", "golang", groupOnly))

	assert.Empty(t, variantLabels(plain))
	assert.Equal(t, map[string]string{"updatecli.io/group": "aws", "updatecli.io/versionfilter": "minor"}, variantLabels(grouped))

	assert.Empty(t, variantTitleSuffix("golang", plain))
	assert.Equal(t, "(major)", variantTitleSuffix("golang", individual))
	assert.Equal(t, "(golang, aws, minor)", variantTitleSuffix("golang", grouped))
	assert.Equal(t, "(dockerfile, minor)", variantTitleSuffix("dockerfile", autodiscovery.Variant{VersionFilter: "minor", GroupBy: autodiscovery.GROUPBYALL}))

	assert.Nil(t, variantGate("m", "golang", plain, autodiscovery.GROUPBYALL, "x"))
	assert.Nil(t, variantGate("m", "golang", groupOnly, autodiscovery.GROUPBYALL, "x"))
	assert.Equal(t,
		&pipeline.Gate{Key: "m/golang/default/deps(go): bump module x", Rank: 1, VersionFilter: "major"},
		variantGate("m", "golang", individual, autodiscovery.GROUPBYINDIVIDUAL, "deps(go): bump module x"))
	assert.Equal(t,
		&pipeline.Gate{Key: "m/golang/aws", Rank: 0, VersionFilter: "minor"},
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
			name:          "grouped variant adds the group and version filter",
			variant:       autodiscovery.Variant{Group: "aws", VersionFilter: "major", GroupBy: autodiscovery.GROUPBYALL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "Bump Go modules (golang, aws, major)",
		},
		{
			name:          "individual variant adds the version filter to the crawler title",
			variant:       autodiscovery.Variant{VersionFilter: "major", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			crawlerAction: action.Config{Title: "deps(go): bump module x"},
			expectedTitle: "deps(go): bump module x (major)",
		},
		{
			name:                "user title takes precedence and gets the variant",
			autodiscoveryAction: action.Config{Title: "Go update"},
			variant:             autodiscovery.Variant{VersionFilter: "minor", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
			crawlerAction:       action.Config{Title: "deps(go): bump module x"},
			expectedTitle:       "Go update (minor)",
		},
		{
			name:          "individual variant without any title uses the manifest name",
			variant:       autodiscovery.Variant{VersionFilter: "minor", GroupBy: autodiscovery.GROUPBYINDIVIDUAL},
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

func TestWithVariantLabels(t *testing.T) {
	variant := autodiscovery.Variant{Group: "aws", VersionFilter: "minor"}

	got := withVariantLabels(map[string]string{"updatecli.io/group": "infra", "group": "backend", "team": "core"}, variant, "Bump x")
	assert.Equal(t, map[string]string{"updatecli.io/group": "aws", "updatecli.io/versionfilter": "minor", "group": "backend", "team": "core"}, got)

	assert.Equal(t, map[string]string{"updatecli.io/group": "aws", "updatecli.io/versionfilter": "minor"}, withVariantLabels(nil, variant, "Bump x"))
	assert.Nil(t, withVariantLabels(nil, autodiscovery.Variant{}, "Bump x"))
}

// TestVariantIDsDontCollideAcrossKinds checks that a crawler kind containing "/"
// can't produce the same ID or gate key as another kind with a group and filter.
func TestVariantIDsDontCollideAcrossKinds(t *testing.T) {
	golangVariant := autodiscovery.Variant{Group: "gomod", VersionFilter: "aws", GroupBy: autodiscovery.GROUPBYALL}
	gomodVariant := autodiscovery.Variant{Group: "aws", GroupBy: autodiscovery.GROUPBYALL}

	golangID := variantPipelineID("parent", "golang", golangVariant)
	gomodID := variantPipelineID("parent", "golang/gomod", gomodVariant)

	assert.Equal(t, "parent/golang/gomod/aws", golangID)
	assert.Equal(t, "parent/golang-gomod/aws", gomodID)
	assert.NotEqual(t, golangID, gomodID)

	gomodFiltered := autodiscovery.Variant{Group: "aws", VersionFilter: "aws", GroupBy: autodiscovery.GROUPBYALL}
	assert.NotEqual(t,
		variantGate("m", "golang", golangVariant, autodiscovery.GROUPBYALL, "x").Key,
		variantGate("m", "golang/gomod", gomodFiltered, autodiscovery.GROUPBYALL, "x").Key)
}
