package flux

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// https://fluxcd.io/flux/components/source/helmrepositories/#writing-a-helmrepository-spec

// helmRepository is the structure of a HelmRelease file
type helmRepository struct {
	ApiVersion string             `yaml:"apiVersion,omitempty"`
	Kind       string             `yaml:"kind,omitempty"`
	Metadata   map[string]string  `yaml:"metadata,omitempty"`
	Spec       helmRepositorySpec `yaml:"spec,omitempty"`
}

type helmRepositorySpec struct {
	Url  string `yaml:"url,omitempty"`
	Type string `yaml:"type,omitempty"`
}

func isHelmRepository(filename string) (*helmRepository, error) {
	var helmRepository helmRepository
	var data []byte

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	err = yaml.Unmarshal(data, &helmRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling HelmRepository file %s: %s", filename, err)
	}

	if strings.Contains(helmRepository.ApiVersion, "source.toolkit.fluxcd.io") && helmRepository.Kind == "HelmRepository" {
		return &helmRepository, nil
	}

	return nil, nil
}

// getHelmRepositoryURL returns the URL of a HelmRepository
func (f Flux) getHelmRepositoryURL(ref sourceRefSpec) (string, bool) {
	for _, helmRepository := range f.helmRepositories {
		helmRepositoryName, ok := helmRepository.Metadata["name"]
		// I don't understand why Flux is using the parameter OCI if the url
		// already has the oci schema specified, but the code is using it
		// so I'm checking its existence here
		oci := helmRepository.Spec.Type == "oci"

		repositoryNamespace := ""
		if namespace, found := helmRepository.Metadata["namespace"]; found {
			repositoryNamespace = namespace
		}

		if ok && helmRepositoryName == ref.Name && repositoryNamespace == ref.Namespace {
			return helmRepository.Spec.Url, oci
		}
	}
	return "", false
}
