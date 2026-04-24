package azuredevops

import (
	"fmt"
	"net/url"
	"strings"

	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
)

// Summary returns a brief description of the Azure DevOps SCM configuration.
func (a *AzureDevOps) Summary() string {
	URL, err := url.Parse(azdoclient.GitURL(a.Spec.URL, a.Spec.Project, a.Spec.Repository))
	if err != nil || URL == nil {
		return ""
	}

	URL.User = nil
	URL.Scheme = ""

	return fmt.Sprintf("%s@%s", strings.TrimPrefix(URL.String(), "//"), a.Spec.Branch)
}
