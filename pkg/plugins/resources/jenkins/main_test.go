package jenkins

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

type Data struct {
	version     string
	err         error
	releaseType string
}

type DataSet []Data

var (
	dataset DataSet = DataSet{
		{
			version:     "2.275",
			releaseType: WEEKLY,
		},
		{
			version:     "2.249.3",
			releaseType: STABLE,
		},
		{
			version:     "2",
			err:         fmt.Errorf("Version 2 contains 1 component(s) which doesn't correspond to any valid release type"),
			releaseType: WRONG,
		},
		{
			version:     "2.249.3.4",
			err:         fmt.Errorf("Version 2.249.3.4 contains 4 component(s) which doesn't correspond to any valid release type"),
			releaseType: WRONG,
		},
		{
			version:     "2.249.3-rc",
			err:         fmt.Errorf("In version '2.249.3-rc', component '3-rc' is not a valid integer"),
			releaseType: WRONG,
		},
	}
)

func TestReleaseType(t *testing.T) {

	for _, data := range dataset {
		got, err := ReleaseType(data.version)
		if err != nil {
			if data.err == nil {
				t.Error(fmt.Errorf("Version '%v' expected the error:\n%v",
					data.version,
					err))
			}

			if strings.Compare(err.Error(), data.err.Error()) != 0 {
				t.Error(fmt.Errorf("For version '%v' expected error:\n%v\ngot\n%v",
					data.version,
					data.err,
					err))
			}
		}
		if got != data.releaseType {
			t.Error(fmt.Errorf("Version validity for %v expected to be %v but got %v ",
				data.version,
				data.releaseType,
				got))
		}

	}
}

func Test_New(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    *Jenkins
		wantErr bool
	}{
		{
			name: "Normal case with default arguments",
			spec: Spec{},
			want: &Jenkins{
				spec: Spec{
					Release: STABLE,
				},
				mavenMetaHandler: mavenmetadata.New(jenkinsDefaultMetaURL),
			},
		},
		{
			name: "Normal case with specified weekly release baseline",
			spec: Spec{
				Release: WEEKLY,
			},
			want: &Jenkins{
				spec: Spec{
					Release: WEEKLY,
				},
				mavenMetaHandler: mavenmetadata.New(jenkinsDefaultMetaURL),
			},
		},
		{
			name: "Error case wit invalid spec",
			spec: Spec{
				Release: "FOO",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_SpecValidate(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want error
	}{
		{
			name: "Normal case with default arguments",
			spec: Spec{},
		},
		{
			name: "Normal case with specified weekly release baseline",
			spec: Spec{
				Release: WEEKLY,
			},
		},
		{
			name: "Error case wit invalid release",
			spec: Spec{
				Release: "FOO",
			},
			want: fmt.Errorf("wrong Jenkins release type 'FOO', accepted values ['weekly','stable']"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.Validate()
			assert.Equal(t, tt.want, got)
		})
	}
}
