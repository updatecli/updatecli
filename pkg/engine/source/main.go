package source

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/maven"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Kind    string
	Output  string
	Prefix  string
	Postfix string
	Spec    interface{}
}

// Execute execute actions defined by the source configuration
func (s *Source) Execute() (string, error) {

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Source"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Source")+1))

	var output string

	switch s.Kind {
	case "githubRelease":
		var spec github.Github
		err := mapstructure.Decode(s.Spec, &spec)

		if err != nil {
			return "", err
		}
		output = spec.GetVersion()

	case "maven":
		var spec maven.Maven
		err := mapstructure.Decode(s.Spec, &spec)

		if err != nil {
			return "", err
		}
		output = spec.GetVersion()

	case "dockerDigest":
		var spec docker.Docker
		err := mapstructure.Decode(s.Spec, &spec)

		if err != nil {
			return "", err
		}
		output = spec.GetVersion()

	default:
		return "", fmt.Errorf("⚠ Don't support source kind: %v", s.Kind)
	}

	s.Output = s.Prefix + output + s.Postfix

	return s.Output, nil
}
