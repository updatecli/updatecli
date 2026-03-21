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

// https://fluxcd.io/flux/components/source/ocirepositories/#writing-an-ocirepository-spec

func loadOCIRepository(filename string) (map[int]fluxcdv1.OCIRepository, error) {

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %s", filename, err)
	}

	loader, err := goyaml.NewLoader(bytes.NewReader(content), goyaml.V4)
	if err != nil {
		return nil, fmt.Errorf("creating yaml loader: %w", err)
	}

	docNum := 1
	result := make(map[int]fluxcdv1.OCIRepository)

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

		// First convert the YAML document to JSON, then unmarshal it into the OCIRepository struct
		// This approach allows us to handle YAML documents that may contain fields not defined in the OCIRepository struct without causing unmarshalling errors
		jsonContent, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshaling yaml document #%d from %q: %w", docNum, filename, err)
		}

		data := fluxcdv1.OCIRepository{}
		if err := json.Unmarshal(jsonContent, &data); err != nil {
			return nil, fmt.Errorf("decoding yaml document #%d from %q: %w", docNum, filename, err)
		}

		gvk := data.GroupVersionKind()
		if gvk.GroupKind().String() == "OCIRepository.source.toolkit.fluxcd.io" {
			result[docNum-1] = data
		}

		docNum++
	}

	return result, nil
}
