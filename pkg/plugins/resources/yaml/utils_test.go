package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeYamlPathKey(t *testing.T) {

	testData := []struct {
		key            string
		expectedResult string
	}{
		{key: `image\.tag`, expectedResult: `$.'image.tag'`},
		{key: `$.image\.tag`, expectedResult: `$.'image.tag'`},
		{key: `image\.`, expectedResult: `$.'image.'`},
		{key: `$.image\.`, expectedResult: `$.'image.'`},
		{key: `image*`, expectedResult: `$.image*`},
		{key: `image`, expectedResult: `$.image`},
		{key: `image.tag`, expectedResult: `$.image.tag`},
		{key: `image\`, expectedResult: `$.image\`},
		{key: `image\\`, expectedResult: `$.image\\`},
	}

	for _, tt := range testData {
		t.Run(tt.key, func(t *testing.T) {
			gotResult := sanitizeYamlPathKey(tt.key)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
