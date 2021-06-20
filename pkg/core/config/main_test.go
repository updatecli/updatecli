package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/olblak/updateCli/pkg/core/engine/source"
)

type Data struct {
	ID             string
	Config         Config
	ExpectedConfig Config
	ExpectedErr    error
}
type DataSet []Data

var (
	dataSet DataSet = DataSet{
		{
			ID: "1",
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
			ID: "2",
			Config: Config{
				Name: `{{ pipeline "Source.Output" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: `{{ pipeline "Source.Output" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedErr: nil,
		},
		{
			ID: "3",
			Config: Config{
				Name: `{{ pipeline "Source.kind" }}`,
				Source: source.Source{
					Name: "Get Version",
					Kind: "jenkins",
				},
			},
			ExpectedConfig: Config{
				Name: "jenkins",
			},
			ExpectedErr: ErrNoKeyDefined,
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
			ExpectedConfig: Config{},
			ExpectedErr:    fmt.Errorf(`function "Source" not defined`),
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
			ExpectedErr: fmt.Errorf(`function "Source" not defined`),
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
			ExpectedErr: ErrNoKeyDefined,
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, data := range dataSet {
		err := data.Config.Update()
		if err != nil && !strings.Contains(err.Error(), data.ExpectedErr.Error()) {
			t.Errorf("Wrong error expected for dataset ID %q:\n\tExpected:\t\t%q\nGot\t\t\t%q\n",
				data.ID,
				data.ExpectedErr,
				err)
			continue
		} else if err == nil {
			if strings.Compare(data.Config.Name, data.ExpectedConfig.Name) != 0 {
				t.Errorf("\n\nWrong output expected for dataset ID %q:\n\texpected:\t\t`%q`\n\tgot:\t\t\t`%q`\n",
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
