package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/olblak/updateCli/pkg/core/engine/source"
)

type Data struct {
	ID                  string
	Config              Config
	ExpectedConfig      Config
	ExpectedUpdateErr   error
	ExpectedValidateErr error
}
type DataSet []Data

var (
	dataSet DataSet = DataSet{
		{
			ID: "1",
			Config: Config{
				Name: "{{ pipeline \"Sources.default.Kind\" }}",
				Sources: map[string]source.Source{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedConfig: Config{
				Name: "jenkins",
				Sources: map[string]source.Source{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		{
			ID: "1.1",
			Config: Config{
				Name:       "{{ context \"pipelineID\" }}",
				PipelineID: "xyz",
				Sources: map[string]source.Source{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedConfig: Config{
				Name: "xyz",
				Sources: map[string]source.Source{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		{
			ID: "2",
			Config: Config{
				Name: `{{ pipeline "Source.Name" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: `Get Version`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedUpdateErr: nil,
		},
		{
			ID: "2.1",
			Config: Config{
				Name: `{{ pipeline "source.name" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: `Get Version`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedUpdateErr: nil,
		},
		{
			ID: "2.2",
			Config: Config{
				Name: `{{ pipeline "soUrce.name" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: `Get Version`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedUpdateErr: nil,
		},
		{
			ID: "3",
			Config: Config{
				Name: `{{ pipeline "Source.kindd" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "jenkins",
			},
			ExpectedUpdateErr: ErrNoKeyDefined,
		},
		{
			ID: "3.1",
			Config: Config{
				Name: `{{ context "Source.kindd" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "",
			},
			ExpectedUpdateErr: ErrNoKeyDefined,
		},
		{
			ID: "4",
			Config: Config{
				Name: `{{ pipeline Source.kind }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig:    Config{},
			ExpectedUpdateErr: fmt.Errorf(`function "Source" not defined`),
		},
		{
			ID: "5",
			Config: Config{
				Name: `{{ pipeline Source.Kind }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "jenkins",
			},
			ExpectedUpdateErr: fmt.Errorf(`function "Source" not defined`),
		},
		{
			ID: "6",
			Config: Config{
				Name: `lts-{{ pipeline "Source.kind" }}-jdk11`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "lts-jenkins-jdk11",
			},
		},
		{
			ID: "wrongSourceKeyName",
			Config: Config{
				Name: `lts-jenkins-jdk11`,
				Sources: map[string]source.Source{
					`{{ pipeline "Source.Name" }}`: {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedConfig: Config{
				Name: "lts-jenkins-jdk11",
			},
			ExpectedUpdateErr:   ErrNotAllowedTemplatedKey,
			ExpectedValidateErr: ErrNotAllowedTemplatedKey,
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, data := range dataSet {
		err := data.Config.Update(data.Config)
		if err != nil && data.ExpectedUpdateErr != nil {
			if !strings.Contains(err.Error(), data.ExpectedUpdateErr.Error()) {
				t.Errorf("Wrong error expected for dataset ID %q:\n\tExpected:\t\t%q\n\tGot\t\t%q\n",
					data.ID,
					data.ExpectedUpdateErr.Error(),
					err.Error())
				continue
			}
		} else if err != nil && data.ExpectedUpdateErr == nil {
			t.Errorf("Wrong error expected for dataset ID %q:\n\tExpected:\t\t%q\n\tGot\t\t%q\n",
				data.ID,
				err.Error(),
				"nil")
		} else if err == nil && data.ExpectedUpdateErr != nil {
			t.Errorf("Wrong error expected for dataset ID %q:\n\tExpected:\t\t%q\n\tGot\t\t%q\n",
				data.ID,
				data.ExpectedUpdateErr.Error(),
				"nil")
		} else if err == nil && data.ExpectedUpdateErr == nil {
			if strings.Compare(data.Config.Name, data.ExpectedConfig.Name) != 0 {
				t.Errorf("\n\nWrong output expected for dataset ID %q:\n\tExpected:\t\t`%q`\n\tGot:\t\t\t`%q`\n",
					data.ID,
					data.ExpectedConfig.Name,
					data.Config.Name)
			}
		}
	}
}

func TestChecksum(t *testing.T) {
	got, err := Checksum("./checksum.example")
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if err != nil {
		t.Errorf("Got an unexpected error: %q", err.Error())
	}

	if got != expected {
		t.Errorf("Got %q, expected %q", got, expected)
	}
}

func TestValidate(t *testing.T) {
	for _, data := range dataSet {
		err := data.Config.Validate()
		if err != nil && data.ExpectedValidateErr == nil {
			t.Errorf("Unexpected Validate Error result for data %q\n\tExpected:\t\t'nil'\n\tGot:\t\t\t%q",
				data.ID,
				err.Error())

		} else if err == nil && data.ExpectedValidateErr != nil {
			t.Errorf("Unexpected Validate Error result for data %q\n\tExpected:\t\t%q\n\tGot:\t\t\t\"nil\"",
				data.ID,
				data.ExpectedValidateErr.Error())

		} else if err != nil && data.ExpectedValidateErr != nil {
			if strings.Compare(err.Error(), data.ExpectedValidateErr.Error()) != 0 {
				t.Errorf("Unexpected Validate Error for data %q",
					data.ID)
			}
		}
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
		if got != data.ExpectedResult {
			t.Errorf("Expected '%v' for key %q but got '%v' ", data.ExpectedResult, data.Key, got)
		}
	}
}
