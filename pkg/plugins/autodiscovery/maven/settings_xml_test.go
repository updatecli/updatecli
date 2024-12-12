package maven

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSettingsXML(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		expectedSettingsXML *Settings
	}{
		{
			name: "Test LoadSettingsXML",
			path: "testdata/default/settings.xml",
			expectedSettingsXML: &Settings{
				XMLName: xml.Name{
					Space: "http://maven.apache.org/SETTINGS/1.0.0",
					Local: "settings",
				},
				Mirrors: []Mirror{
					{
						ID:       "central-mirror",
						Name:     "Central Repository Mirror",
						URL:      "https://mirror.example.com/maven2",
						MirrorOf: "central",
					},
				},
				Profiles: []Profile{
					{
						ID: "default-profile",
						Repositories: []Repository{
							{
								ID:  "central",
								URL: "https://repo.maven.apache.org/maven2",
								Releases: EnabledFlag{
									Enabled: "true",
								},
								Snapshots: EnabledFlag{
									Enabled: "false",
								},
							},
						},
					},
				},
				ActiveProfiles: ActiveProfiles{
					ActiveProfile: []string{"default-profile"},
				},
			},
		},
		{
			name: "Test do not exisdt",
			path: "testdata/donotexist/settings.xml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotSettingsXML := readSettingsXML(tt.path)
			assert.Equal(t, tt.expectedSettingsXML, gotSettingsXML)
		})
	}
}

func TestUpdateEnvVariable(t *testing.T) {

	testEnvVariable := "TEST_UPDATECLI_MAVEN_USERNAME"
	testEnvVariableValue := "test"
	os.Setenv(testEnvVariable, testEnvVariableValue)

	tests := []struct {
		input          string
		expectedResult string
	}{
		{
			input:          "test",
			expectedResult: "test",
		},
		{
			input:          fmt.Sprintf("${env.%s}", testEnvVariable),
			expectedResult: testEnvVariableValue,
		},
		{
			input:          fmt.Sprintf("gcp_${env.%s}", testEnvVariable),
			expectedResult: "gcp_" + testEnvVariableValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotResult := interpolateMavenEnvVariable(tt.input)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
