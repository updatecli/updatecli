package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/olblak/updateCli/pkg/core/engine/source"
)

type Data struct {
	Config         Config
	ExpectedConfig Config
	ExpectedErr    error
}
type DataSet []Data

var (
	dataSet DataSet = DataSet{
		{
			Config: Config{
				Name: "{{ pipeline \"Source.Kind\" }}",
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "jenkins",
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedErr: nil,
		},
		{
			Config: Config{
				Name: `{{ pipeline "Source.Output" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: `"{{ pipeline "Source.Output" }}"`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedErr: nil,
		},
		{
			Config: Config{
				Name: `{{ pipeline "Source.kind" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{},
			ExpectedErr:    ErrNoKeyDefined,
		},
		{
			Config: Config{
				Name: `{{ pipeline Source.kind }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{},
			ExpectedErr:    fmt.Errorf(`function "Source" not defined`),
		},
		{
			Config: Config{
				Name: `{{ requiredEnv "XXX" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{},
			ExpectedErr:    ErrNoEnvironmentVariableSet,
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, data := range dataSet {
		err := data.Config.Update()
		if err != nil && !strings.Contains(err.Error(), data.ExpectedErr.Error()) {
			t.Errorf("Expected error:\n%v\ngot\n%v\n", data.ExpectedErr, err)
		} else if err != nil && data.ExpectedErr == err {
			continue
		} else if err == nil {
			if strings.Compare(data.Config.Name, data.ExpectedConfig.Name) != 0 {
				t.Errorf("\n\nWrong output, expect: \n`%v`\n\ngot\n\n`%v`\n", data.ExpectedConfig.Name, data.Config.Name)
			}
		}
	}
}

func TestReset(t *testing.T) {
	for _, data := range dataSet {
		data.Config.Reset()

		c := Config{}

		if !reflect.DeepEqual(data.Config, c) {

			t.Errorf("\n\nWrong output, expect: \n`%v`\n\ngot\n\n`%v`\n", c, data.Config)
		}
	}
}
