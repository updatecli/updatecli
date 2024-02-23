package flux

import (
	"fmt"
	"os"

	fluxcdv1 "github.com/fluxcd/source-controller/api/v1beta2"

	"sigs.k8s.io/yaml"
)

// https://fluxcd.io/flux/components/source/helmrepositories/#writing-a-helmrepository-spec

func isHelmRepository(filename string) (*fluxcdv1.HelmRepository, error) {
	var helmRepository fluxcdv1.HelmRepository
	var data []byte

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	err = yaml.Unmarshal(data, &helmRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling HelmRepository file %s: %s", filename, err)
	}

	gvk := helmRepository.GroupVersionKind()
	if gvk.GroupKind().String() == "HelmRepository.source.toolkit.fluxcd.io" {
		return &helmRepository, nil
	}

	return nil, nil
}
