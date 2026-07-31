package azuredevopssearch

import "fmt"

// Summary returns a brief description of the Azure DevOps search SCM configuration.
func (a *AzureDevOpsSearch) Summary() string {
	return fmt.Sprintf("Azure DevOps Search:\n\tOrganization: %q\n\tProject: %q\n\tRepository: %q\n\tBranch: %q\n\tLimit: %d", a.spec.Organization, a.projectPattern, a.repositoryPattern, a.branch, a.limit)
}
