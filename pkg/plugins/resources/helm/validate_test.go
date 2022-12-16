package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateTarget is an integration tests that updating a target if working
func TestValidateTarget(t *testing.T) {

	dataset := []struct {
		chart         Chart
		expectedError string
		wantErr       bool
	}{
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "none",
				},
			},
		},
		{
			chart: Chart{
				spec: Spec{
					Key:              "container.image",
					VersionIncrement: "none",
				},
			},
			wantErr:       true,
			expectedError: ErrWrongConfig.Error(),
		},
		{
			chart: Chart{
				spec: Spec{
					VersionIncrement: "none",
				},
			},
			wantErr:       true,
			expectedError: ErrWrongConfig.Error(),
		},
	}

	for _, d := range dataset {
		t.Run("x", func(t *testing.T) {
			gotErr := d.chart.ValidateTarget()
			if d.wantErr {
				require.Error(t, gotErr)
				assert.EqualError(t, gotErr, d.expectedError)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

// TestValidateVersionInc tests that increment a version effectively works
func TestValidateVersionInc(t *testing.T) {

	dataset := []struct {
		chart         Chart
		expectedError string
		wantErr       bool
	}{
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "none",
				},
			},
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "nonei",
				},
			},
			expectedError: "unrecognized increment rule \"nonei\", accepted values are a comma separated list of [major,minor,patch], or 'none' to disable version increment",
			wantErr:       true,
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "none,patch",
				},
			},
			expectedError: "rule \"none\" is not compatible with others from \"none,patch\"",
			wantErr:       true,
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "patch,patch",
				},
			},
			expectedError: "rule \"patch\" appears multiple time in patch,patch",
			wantErr:       true,
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "minor,minor",
				},
			},
			expectedError: "rule \"minor\" appears multiple time in minor,minor",
			wantErr:       true,
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "major,major",
				},
			},
			expectedError: "rule \"major\" appears multiple time in major,major",
			wantErr:       true,
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "patch,major",
				},
			},
		},
		{
			chart: Chart{
				spec: Spec{
					Name:             "My chart",
					Key:              "container.image",
					VersionIncrement: "major,minor,patch",
				},
			},
		},
	}

	for _, d := range dataset {
		t.Run("x", func(t *testing.T) {
			gotErrs := d.chart.validateVersionInc()
			if d.wantErr {
				for _, gotErr := range gotErrs {
					require.Error(t, gotErr)
					assert.EqualError(t, gotErr, d.expectedError)
				}
				return
			}
			for _, gotErr := range gotErrs {
				require.NoError(t, gotErr)
			}
		})
	}
}
