package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	fluxcdv1 "github.com/fluxcd/source-controller/api/v1beta2"
	goyaml "go.yaml.in/yaml/v4"
)

// https://fluxcd.io/flux/components/source/helmrepositories/#writing-a-helmrepository-spec

func loadHelmRepositoryData(filename string) ([]fluxcdv1.HelmRepository, error) {

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filename, err)
	}

	return loadHelmRepositoryDataFromBytes(filename, content)
}

func loadHelmRepositoryDataFromBytes(filename string, content []byte) ([]fluxcdv1.HelmRepository, error) {

	loader, err := goyaml.NewLoader(bytes.NewReader(content), goyaml.V4)
	if err != nil {
		return nil, fmt.Errorf("creating yaml loader: %w", err)
	}

	docNum := 1
	result := make([]fluxcdv1.HelmRepository, 0)

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

		jsonContent, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshaling yaml document #%d from %q: %w", docNum, filename, err)
		}

		data := fluxcdv1.HelmRepository{}
		if err := json.Unmarshal(jsonContent, &data); err != nil {
			return nil, fmt.Errorf("decoding yaml document #%d from %q: %w", docNum, filename, err)
		}

		gvk := data.GroupVersionKind()
		if gvk.GroupKind().String() == "HelmRepository.source.toolkit.fluxcd.io" {
			result = append(result, data)
		}

		docNum++
	}
	return result, nil
}
