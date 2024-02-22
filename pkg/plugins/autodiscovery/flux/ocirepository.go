package flux

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// https://fluxcd.io/flux/components/source/ocirepositories/#writing-an-ocirepository-spec

type ociRepository struct {
	ApiVersion string            `yaml:"apiVersion,omitempty"`
	Kind       string            `yaml:"kind,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
	Spec       ociRepositorySpec `yaml:"spec,omitempty"`
}

type ociRepositorySpec struct {
	URL string               `yaml:"url,omitempty"`
	Ref ociRepositorySpecRef `yaml:"ref,omitempty"`
}

type ociRepositorySpecRef struct {
	Tag string `yaml:"tag,omitempty"`
}

func loadOCIRepository(filename string) (*ociRepository, error) {
	var ociRepository ociRepository

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	err = yaml.Unmarshal(data, &ociRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling OCIRepository file %s: %s", filename, err)
	}

	apiVersion := strings.Split(ociRepository.ApiVersion, "/")[0]
	if strings.HasSuffix(apiVersion, "fluxcd.io") && ociRepository.Kind == "OCIRepository" {
		return &ociRepository, nil
	}
	return nil, nil
}
