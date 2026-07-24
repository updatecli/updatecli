package toml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecValidate(t *testing.T) {
	testData := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "valid default spec",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			wantErr: false,
		},
		{
			name: "createmissingkey allowed with default engine",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              ".owner.age",
				CreateMissingKey: true,
			},
			wantErr: false,
		},
		{
			name: "createmissingkey rejected with dasel v3",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              "owner.age",
				CreateMissingKey: true,
				Engine:           strPtr(ENGINEDASEL_V3),
			},
			wantErr: true,
		},
		{
			name: "createmissingkey rejected with dasel alias (latest=v3)",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              "owner.age",
				CreateMissingKey: true,
				Engine:           strPtr(ENGINEDASEL_LATEST),
			},
			wantErr: true,
		},
		{
			name: "createmissingkey allowed with dasel v2",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              ".owner.age",
				CreateMissingKey: true,
				Engine:           strPtr(ENGINEDASEL_V2),
			},
			wantErr: false,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
