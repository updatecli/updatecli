package azuredevopssearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("uses defaults and sanitizes values", func(t *testing.T) {
		search, err := New(map[string]interface{}{
			"organization": " updatecli ",
			"project":      " updatecli-.* ",
			"repository":   " charts-.* ",
		})

		require.NoError(t, err)
		assert.Equal(t, DefaultRepositoryLimit, search.limit)
		assert.Equal(t, "^main$", search.branch)
		assert.Equal(t, "updatecli", search.spec.Organization)
		assert.Equal(t, "updatecli-.*", search.projectPattern)
		assert.Equal(t, "charts-.*", search.repositoryPattern)
		assert.Equal(t, "https://dev.azure.com", search.spec.URL)
	})

	t.Run("fails when organization is missing", func(t *testing.T) {
		_, err := New(map[string]interface{}{
			"project": "updatecli-project",
		})

		require.ErrorContains(t, err, ErrOrganizationEmpty)
	})

	t.Run("fails when project regex is invalid", func(t *testing.T) {
		_, err := New(map[string]interface{}{
			"organization": "updatecli",
			"project":      "[",
		})

		require.ErrorContains(t, err, "invalid project regex")
	})
}
