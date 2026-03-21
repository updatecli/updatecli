package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta2"
	goyaml "go.yaml.in/yaml/v4"
)

// https://fluxcd.io/flux/components/helm/helmreleases/#writing-a-helmrelease-spec

func loadHelmRelease(filename string) (map[int]helmv2.HelmRelease, error) {

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	loader, err := goyaml.NewLoader(bytes.NewReader(content), goyaml.V4)
	if err != nil {
		return nil, fmt.Errorf("creating yaml loader: %w", err)
	}

	docNum := 1
	result := make(map[int]helmv2.HelmRelease)

	for {
		var raw any

		err := loader.Load(&raw)
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("loading yaml document #%d from %q: %w", docNum, filename, err)
		}

		if raw == nil {
			docNum++
			continue
		}

		// First convert the YAML document to JSON, then unmarshal it into the HelmRelease struct
		// This approach allows us to handle YAML documents that may contain fields not defined in the HelmRelease struct without causing unmarshalling errors
		jsonContent, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshaling yaml document #%d from %q: %w", docNum, filename, err)
		}

		data := helmv2.HelmRelease{}
		if err := json.Unmarshal(jsonContent, &data); err != nil {
			return nil, fmt.Errorf("decoding yaml document #%d from %q: %w", docNum, filename, err)
		}

		gvk := data.GroupVersionKind()
		if gvk.GroupKind().String() == "HelmRelease.helm.toolkit.fluxcd.io" {
			result[docNum-1] = data
		}

		docNum++
	}

	return result, nil
}
