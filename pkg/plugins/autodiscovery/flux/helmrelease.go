package flux

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta2"

	"sigs.k8s.io/yaml"
)

// https://fluxcd.io/flux/components/helm/helmreleases/#writing-a-helmrelease-spec

func loadHelmReleaseFromBytes(data []byte) (*helmv2.HelmRelease, error) {
	helmRelease := helmv2.HelmRelease{}
	err := yaml.Unmarshal(data, &helmRelease)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling HelmRelease: %s", err)
	}

	gvk := helmRelease.GroupVersionKind()
	if gvk.GroupKind().String() == "HelmRelease.helm.toolkit.fluxcd.io" {
		return &helmRelease, nil
	}

	return nil, nil
}
