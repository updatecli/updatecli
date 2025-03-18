package flux

import (
	"fmt"

	fluxcdv1 "github.com/fluxcd/source-controller/api/v1beta2"

	"sigs.k8s.io/yaml"
)

// https://fluxcd.io/flux/components/source/helmrepositories/#writing-a-helmrepository-spec

func loadHelmRepositoryFromBytes(data []byte) (*fluxcdv1.HelmRepository, error) {
	helmRepository := fluxcdv1.HelmRepository{}
	err := yaml.Unmarshal(data, &helmRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling HelmRepository: %s", err)
	}

	gvk := helmRepository.GroupVersionKind()
	if gvk.GroupKind().String() == "HelmRepository.source.toolkit.fluxcd.io" {
		return &helmRepository, nil
	}

	return nil, nil
}
