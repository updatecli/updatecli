package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Mocking the context package
type mockSourceContext struct {
	Output string
	Result result.Source
}

// Mocking the context package
type context struct {
	Sources map[string]mockSourceContext
}

type Data struct {
	ID                  string
	Config              Config
	Context             context
	ExpectedConfig      Config
	ExpectedUpdateErr   error
	ExpectedValidateErr error
}
type DataSet []Data

var (
	dataSet DataSet = DataSet{
		// Testing that we get the correct values
		{
			ID: "1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ pipeline \"Sources.default.Output\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "jenkins - 2.289.2",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		{
			ID: "1.1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
						Result: result.Source{
							Result: result.SUCCESS,
						},
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "jenkins - 2.289.2",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		// Test a failed source
		{
			ID: "1.2",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Result: result.Source{
							Result: result.FAILURE,
						},
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("template: cfg:1:19: executing \"cfg\" at <source \"default\">: error calling source: parent source \"default\" failed"),
			ExpectedValidateErr: nil,
		},
		// Testing key case sensitive
		{
			ID: "2",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ pipeline \"sources.default.output\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "jenkins - 2.289.2",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		{
			ID: "2.1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"Default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
						Result: result.Source{
							Result: result.SUCCESS,
						},
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "jenkins - 2.289.2",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		// Testing wrong key returning error message
		{
			ID: "3",
			Config: Config{
				Spec: Spec{
					Name: `{{ pipeline "Source.kind_" }}`,
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "jenkins",
				},
			},
			ExpectedUpdateErr: fmt.Errorf(`template: cfg:1:10: executing "cfg" at <pipeline "Source.kind_">: error calling pipeline: key not defined in configuration`),
		},
		// Testing wrong function name
		{
			ID: "4",
			Config: Config{
				Spec: Spec{
					Name: `{{ pipeline Source.kind }}`,
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedConfig:    Config{},
			ExpectedUpdateErr: fmt.Errorf(`template: cfg:1: function "Source" not defined`),
		},
		{
			ID: "5",
			Config: Config{
				Spec: Spec{
					Name: `{{ source "default" }}-jdk11`,
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
						Result: result.Source{
							Result: result.SUCCESS,
						},
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "2.289.2-jdk11",
				},
			},
		},
		{
			ID: "6",
			Config: Config{
				Spec: Spec{
					Name: `lts-jenkins-jdk11`,
					Sources: map[string]source.Config{
						`{{ pipeline "Sources.default.output" }}`: {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedConfig: Config{
				Spec: Spec{
					Name: "lts-jenkins-jdk11",
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("sources validation error:\n%s", ErrNotAllowedTemplatedKey),
			ExpectedValidateErr: fmt.Errorf("sources validation error:\n%s", ErrNotAllowedTemplatedKey),
		},
		{
			ID: "7",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Conditions: map[string]condition.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
							SourceID: "default",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
		},
		{
			ID: "7.1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Conditions: map[string]condition.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
							SourceID: "ShouldNotExist",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("conditions validation error:\n%s", ErrBadConfig),
			ExpectedValidateErr: fmt.Errorf("conditions validation error:\n%s", ErrBadConfig),
		},
		{
			ID: "7.2",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Conditions: map[string]condition.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
		},
		{
			ID: "7.3",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"one": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
						"two": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get text from shell",
								Kind: "shell",
							},
						},
					},
					Conditions: map[string]condition.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("conditions validation error:\n%s", ErrBadConfig),
			ExpectedValidateErr: fmt.Errorf("conditions validation error:\n%s", ErrBadConfig),
		},
		{
			ID: "8",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
							SourceID: "default",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
		},
		{
			ID: "8.1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
							SourceID: "ShouldNotExist",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("targets validation error:\n%s", ErrBadConfig),
			ExpectedValidateErr: fmt.Errorf("targets validation error:\n%s", ErrBadConfig),
		},
		{
			ID: "8.2",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
		},
		{
			ID: "8.3",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ source \"default\" }}",
					Sources: map[string]source.Config{
						"one": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
						"two": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Test SourceID",
								Kind: "shell",
							},
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("targets validation error:\n%s", ErrBadConfig),
			ExpectedValidateErr: fmt.Errorf("targets validation error:\n%s", ErrBadConfig),
		},
		{
			ID: "9",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ pipeline \"Sources.default.Output\" }}",
					SCMs: map[string]scm.Config{
						"default": {
							Kind: "github",
							Spec: map[string]string{
								"user":       "updatecli",
								"email":      "me@olblak.com",
								"owner":      "updatecli",
								"repository": "updatecli",
								"token":      "SuperSecret",
								"username":   "olblak",
								"branch":     "main",
							},
						},
					},
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"updateDefault": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Update Default Version",
								Kind: "shell",
							},
						},
					},
					Actions: map[string]action.Config{
						"default": {
							Title: "default Pull Request",
							Kind:  "github/pullrequest",
							ScmID: "default",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
		},
		{
			ID: "9.1",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ pipeline \"Sources.default.Output\" }}",
					SCMs: map[string]scm.Config{
						"default": {
							Kind: "github",
							Spec: map[string]string{
								"user":       "updatecli",
								"email":      "me@olblak.com",
								"owner":      "updatecli",
								"repository": "updatecli",
								"token":      "SuperSecret",
								"username":   "olblak",
								"branch":     "main",
							},
						},
					},
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"updateDefault": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Update Default Version",
								Kind: "shell",
							},
						},
					},
					Actions: map[string]action.Config{
						"default": {},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("actions validation error:\nmissing value for parameter(s) [\"kind,scmid\"]"),
			ExpectedValidateErr: fmt.Errorf("actions validation error:\nmissing value for parameter(s) [\"kind,scmid\"]"),
		},
		{
			ID: "9.2",
			Config: Config{
				Spec: Spec{
					Name: "jenkins - {{ pipeline \"Sources.default.Output\" }}",
					SCMs: map[string]scm.Config{
						"default": {
							Kind: "github",
							Spec: map[string]string{
								"user":       "updatecli",
								"email":      "me@olblak.com",
								"owner":      "updatecli",
								"repository": "updatecli",
								"token":      "SuperSecret",
								"username":   "olblak",
								"branch":     "main",
							},
						},
					},
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get Version",
								Kind: "jenkins",
							},
						},
					},
					Targets: map[string]target.Config{
						"updateDefault": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Update Default Version",
								Kind: "shell",
							},
						},
					},
					Actions: map[string]action.Config{
						"default": {
							Title: "default Pull Request",
							Kind:  "github/pullrequest",
							ScmID: "not_existing",
						},
					},
				},
			},
			Context: context{
				Sources: map[string]mockSourceContext{
					"default": {
						Output: "2.289.2",
					},
				},
			},
			ExpectedUpdateErr:   fmt.Errorf("actions validation error:\n%s", ErrBadConfig),
			ExpectedValidateErr: fmt.Errorf("actions validation error:\n%s", ErrBadConfig),
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, data := range dataSet {
		t.Run(data.ID, func(t *testing.T) {
			err := data.Config.Update(data.Context)

			if data.ExpectedUpdateErr != nil {
				require.Error(t, err)
				assert.Equal(t, data.ExpectedUpdateErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestChecksum(t *testing.T) {
	got, err := FileChecksum("./checksum.example")
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestValidate(t *testing.T) {
	for _, data := range dataSet {
		t.Run(data.ID, func(t *testing.T) {
			err := data.Config.Validate()

			if data.ExpectedValidateErr != nil {
				require.Error(t, err)
				assert.Equal(t, data.ExpectedValidateErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestIsTemplatedString(t *testing.T) {
	type templatedStringData struct {
		Key            string
		ExpectedResult bool
	}
	dataset := []templatedStringData{
		{Key: "bob", ExpectedResult: false},
		{Key: "", ExpectedResult: false},
		{Key: "{{ bob }}", ExpectedResult: true},
		{Key: "{{ {{ bob }}", ExpectedResult: true},
		{Key: "{{ bob }} }}", ExpectedResult: true},
		{Key: "{{ bob", ExpectedResult: false},
		{Key: "bob }}", ExpectedResult: false},
		{Key: "}} bob {{", ExpectedResult: false},
		{Key: "alpha-{{ version }}-jdk11", ExpectedResult: true},
	}

	for _, data := range dataset {
		got := IsTemplatedString(data.Key)
		assert.Equal(t, data.ExpectedResult, got)
	}
}

func TestNew(t *testing.T) {
	dataset := []struct {
		id             string
		option         Option
		pipelineIDs    []string
		labels         map[string]string
		expectedResult []Config
		expectedError  error
	}{
		{
			id: "Test with a valid manifest containing one config",
			option: Option{
				ManifestFile: "testdata/updatecli.d/jenkins.yaml",
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Get latest Jenkins version",
					},
				},
			},
		},
		{
			id: "Test with a valid manifest containing one json config",
			option: Option{
				ManifestFile: "testdata/updatecli.d/jenkins.json",
				ValuesFiles:  []string{"testdata/values.json"},
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Get lts Jenkins version",
					},
				},
			},
		},
		{
			id: "Test with a valid manifest containing two configs",
			option: Option{
				ManifestFile: "testdata/updatecli.d/multiJenkins.yaml",
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Get latest stable Jenkins version",
					},
				},
				{
					Spec: Spec{
						Name: "Get latest weekly Jenkins version",
					},
				},
			},
		},
		{
			id: "Test with a valid templated manifest containing two configs",
			option: Option{
				ManifestFile: "testdata/updatecli.d/multiJenkins.yaml",
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Get latest stable Jenkins version",
					},
				},
				{
					Spec: Spec{
						Name: "Get latest weekly Jenkins version",
					},
				},
			},
		},
		{
			id: "Test with a valid templated manifest containing two configs",
			option: Option{
				ManifestFile: "testdata/updatecli.d/multiTemplatedJenkins.tpl",
				ValuesFiles:  []string{"testdata/values.yaml"},
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Get latest lts Jenkins version",
					},
				},
				{
					Spec: Spec{
						Name: "Get latest weekly Jenkins version",
					},
				},
			},
		},
		{
			id: "Test with a bad manifest containing one config",
			option: Option{
				ManifestFile: "testdata/updatecli.d/badJenkins.yaml",
			},
			expectedResult: []Config{},
			expectedError:  errors.New("yaml: unmarshal errors:\n  line 3: cannot unmarshal !!seq into map[string]source.Config"),
		},
		{
			id: "Test with matching label",
			option: Option{
				ManifestFile: "testdata/labels/alpine.yaml",
			},
			labels: map[string]string{
				"id": "alpine",
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Update Alpine Docker Image",
					},
				},
			},
		},
		{
			id: "Test with matching label key only (empty value)",
			option: Option{
				ManifestFile: "testdata/labels/alpine.yaml",
			},
			labels: map[string]string{
				"id": "",
			},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Update Alpine Docker Image",
					},
				},
			},
		},
		{
			id: "Test with none matching label debian",
			option: Option{
				ManifestFile: "testdata/labels/alpine.yaml",
			},
			labels: map[string]string{
				"id": "debian",
			},
			expectedResult: []Config{},
		},
		{
			id: "Test with matching pipelineid",
			option: Option{
				ManifestFile: "testdata/labels/alpine.yaml",
			},
			pipelineIDs: []string{"alpine"},
			expectedResult: []Config{
				{
					Spec: Spec{
						Name: "Update Alpine Docker Image",
					},
				},
			},
		},
		{
			id: "Test with none matching pipelineid debian",
			option: Option{
				ManifestFile: "testdata/labels/alpine.yaml",
			},
			pipelineIDs:    []string{"debian"},
			expectedResult: []Config{},
		},
	}

	for _, data := range dataset {
		got, err := New(data.option, data.pipelineIDs, data.labels)

		switch data.expectedError {
		case nil:
			require.NoError(t, err)
		default:
			require.ErrorContains(t, err, data.expectedError.Error())
		}

		require.EqualValues(t, len(data.expectedResult), len(got))
		for i := range data.expectedResult {
			require.Equal(t, data.expectedResult[i].Spec.Name, got[i].Spec.Name)
		}
	}
}
