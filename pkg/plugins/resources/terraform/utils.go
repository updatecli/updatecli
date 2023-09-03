package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func ParseHcl(content string, filePath string) (*hclwrite.File, error) {
	file, diags := hclwrite.ParseConfig([]byte(content), filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("%s failed to parse file %q: %s",
			result.FAILURE,
			filePath,
			diags)
	}

	return file, nil
}
