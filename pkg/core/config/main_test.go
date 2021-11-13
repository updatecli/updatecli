package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
)

// Mocking the context package
type mockSourceContext struct {
	Output string
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
		// Testing that we get the correct value
		{
			ID: "1",
			Config: Config{
				Name: "jenkins - {{ pipeline \"Sources.default.Output\" }}",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "jenkins - 2.289.2",
				Sources: map[string]source.Config{
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
				Name: "jenkins - {{ source \"default\" }}",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "jenkins - 2.289.2",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
					},
				},
			},
			ExpectedUpdateErr:   nil,
			ExpectedValidateErr: nil,
		},
		// Testing key case sensitive
		{
			ID: "2",
			Config: Config{
				Name: "jenkins - {{ pipeline \"sources.default.output\" }}",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "jenkins - 2.289.2",
				Sources: map[string]source.Config{
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
			ID: "2.1",
			Config: Config{
				Name: "jenkins - {{ source \"Default\" }}",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "jenkins - 2.289.2",
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: `{{ pipeline "Source.kindd" }}`,
				Source: source.Config{
					Name: "Get Version",
					Kind: "jenkins",
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
				Name: "jenkins",
			},
			ExpectedUpdateErr: ErrNoKeyDefined,
		},
		{
			ID: "3.1",
			Config: Config{
				Name: `{{ pipeline "Source.kindd" }}`,
				Source: source.Config{
					Name: "Get Version",
					Kind: "jenkins",
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
				Name: "",
			},
			ExpectedUpdateErr: ErrNoKeyDefined,
		},
		// Testing wrong function name
		{
			ID: "4",
			Config: Config{
				Name: `{{ pipeline Source.kind }}`,
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
			ExpectedUpdateErr: fmt.Errorf(`function "Source" not defined`),
		},
		{
			ID: "4.1",
			Config: Config{
				Name: `{{ source default }}`,
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
			ExpectedUpdateErr: fmt.Errorf(`function "default" not defined`),
		},
		{
			ID: "6",
			Config: Config{
				Name: `{{ source "default" }}-jdk11`,
				Sources: map[string]source.Config{
					"default": {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "2.289.2-jdk11",
			},
		},
		{
			ID: "7",
			Config: Config{
				Name: `lts-jenkins-jdk11`,
				Sources: map[string]source.Config{
					`{{ pipeline "Sources.default.output" }}`: {
						Name: "Get Version",
						Kind: "jenkins",
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
				Name: "lts-jenkins-jdk11",
			},
			ExpectedUpdateErr:   ErrNotAllowedTemplatedKey,
			ExpectedValidateErr: ErrNotAllowedTemplatedKey,
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, data := range dataSet {
		err := data.Config.Update(data.Context)
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
				"nil",
				err.Error())
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
