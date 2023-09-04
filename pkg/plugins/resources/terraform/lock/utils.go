package lock

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/zclconf/go-cty/cty"
)

func getProviderBlock(file *hclwrite.File, filePath string, provider string) (*hclwrite.Block, error) {
	providerBlock := file.Body().FirstMatchingBlock("provider", []string{provider})

	if providerBlock == nil {
		err := fmt.Errorf("%s cannot find value for %q from file %q",
			result.FAILURE,
			provider,
			filePath)
		return nil, err
	}

	return providerBlock, nil
}

func tokensForListPerLine(list []string) hclwrite.Tokens {
	tokens := hclwrite.Tokens{}
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}})
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}})

	for _, i := range list {
		ts := hclwrite.TokensForValue(cty.StringVal(i))
		tokens = append(tokens, ts...)
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}})
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}})
	}

	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}})

	return tokens
}
