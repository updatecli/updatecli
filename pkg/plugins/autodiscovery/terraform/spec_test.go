package terraform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - File",
			spec: Spec{
				Platforms: []string{"linux_amd64"},
			},
			wantErr: false,
		},
		{
			name:             "Failure - No platforms",
			spec:             Spec{},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
