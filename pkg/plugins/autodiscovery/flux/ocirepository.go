package flux

import (
	"fmt"
	"os"

	fluxcdv1 "github.com/fluxcd/source-controller/api/v1beta2"

	"sigs.k8s.io/yaml"
)

// https://fluxcd.io/flux/components/source/ocirepositories/#writing-an-ocirepository-spec

func loadOCIRepository(filename string) (*fluxcdv1.OCIRepository, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	ociRepository := fluxcdv1.OCIRepository{}
	err = yaml.Unmarshal(data, &ociRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling OCIRepository file %s: %s", filename, err)
	}

	gvk := ociRepository.GroupVersionKind()
	if gvk.GroupKind().String() == "OCIRepository.source.toolkit.fluxcd.io" {
		return &ociRepository, nil
	}

	return nil, nil
}

func loadOCIRepositoryFromBytes(data []byte) (*fluxcdv1.OCIRepository, error) {
	ociRepository := fluxcdv1.OCIRepository{}
	err := yaml.Unmarshal(data, &ociRepository)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling OCIRepository: %s", err)
	}

	gvk := ociRepository.GroupVersionKind()
	if gvk.GroupKind().String() == "OCIRepository.source.toolkit.fluxcd.io" {
		return &ociRepository, nil
	}

	return nil, nil
}
