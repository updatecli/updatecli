package dockerfile

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

func (df *Dockerfile) Source(_ context.Context, resolver utils.Resolver, resultSource *result.Source) error {
	switch len(df.files) {
	case 1:
		//
	case 0:
		return fmt.Errorf("no dockerfile specified")
	default:
		return fmt.Errorf("validation error in sources of type 'dockerfile': the attributes `spec.files` can't contain more than one element for sources")
	}

	// loop over the only file
	for _, specFile := range df.files {
		file, err := resolver.Resolve(specFile)
		if err != nil {
			return fmt.Errorf("invalid file path %q: %w", specFile, err)
		}

		if !df.contentRetriever.FileExists(file) {
			return fmt.Errorf("the file %s does not exist", file)
		}

		dockerfileContent, err := df.contentRetriever.ReadAll(file)
		if err != nil {
			return fmt.Errorf("reading dockerfile: %w", err)
		}

		logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", file)

		value := df.parser.GetInstruction([]byte(dockerfileContent), df.spec.Stage)
		stageInfo := "last stage"
		if df.spec.Stage != "" {
			stageInfo = fmt.Sprintf("stage %q", df.spec.Stage)
		}
		resultSource.Result = result.SUCCESS
		resultSource.Information = value
		resultSource.Description = fmt.Sprintf("value %q found for %s in the dockerfile file %q",
			value,
			stageInfo,
			file,
		)
	}

	return nil
}
