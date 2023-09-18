package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// push_test is a test for the push function
func TestPush(t *testing.T) {
	err := Push(
		[]string{
			"testdata/venom.yaml",
		},
		[]string{
			"testdata/values.yaml",
		},
		[]string{
			"testdata/secrets.yaml",
		},
		"localhost:5000/myrepo:latest",
		true)
	require.NoError(t, err)
}
