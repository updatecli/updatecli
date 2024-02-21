package flux

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// https://fluxcd.io/flux/components/helm/helmreleases/#writing-a-helmrelease-spec

type helmRelease struct {
	ApiVersion string            `yaml:"apiVersion,omitempty"`
	Kind       string            `yaml:"kind,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
	Spec       helmReleaseSpec   `yaml:"spec,omitempty"`
}

type helmReleaseSpec struct {
	Chart helmReleaseChart `yaml:"chart,omitempty"`
}

type helmReleaseChart struct {
	Spec helmReleaseChartSpec `yaml:"spec,omitempty"`
}

type helmReleaseChartSpec struct {
	Chart     string        `yaml:"chart,omitempty"`
	Version   string        `yaml:"version,omitempty"`
	SourceRef sourceRefSpec `yaml:"sourceRef,omitempty"`
}

type sourceRefSpec struct {
	Kind      string `yaml:"kind,omitempty"`
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

func loadHelmRelease(filename string) (*helmRelease, error) {
	var helmRelease helmRelease
	var data []byte

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s:%s", filename, err.Error())
	}

	err = yaml.Unmarshal(data, &helmRelease)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling HelmRelease file %s: %s", filename, err.Error())
	}

	apiVersion := strings.Split(helmRelease.ApiVersion, "/")[0]

	if strings.HasSuffix(apiVersion, "fluxcd.io") && helmRelease.Kind == "HelmRelease" {
		return &helmRelease, nil
	}

	return nil, nil
}
