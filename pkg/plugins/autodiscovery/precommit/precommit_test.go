package precommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrecommitData_MapsRevCommentsBySequenceOrder(t *testing.T) {
	data, err := loadPrecommitData("testdata/duplicaterepo/.pre-commit-config.yaml")
	require.NoError(t, err)
	require.Len(t, data.Repos, 2)

	assert.Equal(t, "https://github.com/example/tool", data.Repos[0].Repo)
	assert.Equal(t, "v1.2.3", data.Repos[0].RevComment)

	assert.Equal(t, "https://github.com/example/tool", data.Repos[1].Repo)
	assert.Equal(t, "v2.0.0", data.Repos[1].RevComment)
}
