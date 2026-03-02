package gitlabsearch

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	gitlabscm "github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ScmsGenerator discovers GitLab repositories within the configured group
// and returns a list of gitlab.Spec, one per matching branch.
func (g *GitLabSearch) ScmsGenerator(ctx context.Context) ([]gitlabscm.Spec, error) {
	results := make([]gitlabscm.Spec, 0)

	re, err := regexp.Compile(g.branch)
	if err != nil {
		return nil, fmt.Errorf("invalid branch regex %q: %w", g.branch, err)
	}

	glClient := (*gitlab.Client)(g.client)

	includeSubGroups := g.includeSubgroups
	listOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions:      gitlab.ListOptions{Page: 1, PerPage: 100},
		IncludeSubGroups: &includeSubGroups,
	}

	if g.search != "" {
		listOpt.Search = &g.search
	}

	for {
		projects, resp, err := glClient.Groups.ListGroupProjects(g.group, listOpt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list projects in group %q: %w", g.group, err)
		}

		for _, project := range projects {
			logrus.Debugf("Processing GitLab project: %s", project.PathWithNamespace)

			// Split "group/subgroup/repo" into owner="group/subgroup" and repository="repo"
			lastSlash := strings.LastIndex(project.PathWithNamespace, "/")
			if lastSlash < 0 {
				logrus.Warningf("Skipping project with unexpected path format: %s", project.PathWithNamespace)
				continue
			}

			owner := project.PathWithNamespace[:lastSlash]
			repository := project.PathWithNamespace[lastSlash+1:]

			branches, err := g.listBranches(ctx, glClient, project.PathWithNamespace)
			if err != nil {
				return nil, fmt.Errorf("failed to list branches for %s: %w", project.PathWithNamespace, err)
			}

			for _, branchName := range branches {
				if !re.MatchString(branchName) {
					continue
				}

				spec := gitlabscm.Spec{
					Branch:                 branchName,
					CommitMessage:          g.spec.CommitMessage,
					Directory:              g.spec.Directory,
					Email:                  g.spec.Email,
					Force:                  g.spec.Force,
					GPG:                    g.spec.GPG,
					Owner:                  owner,
					Repository:             repository,
					Submodules:             g.spec.Submodules,
					User:                   g.spec.User,
					WorkingBranch:          g.spec.WorkingBranch,
					WorkingBranchPrefix:    g.spec.WorkingBranchPrefix,
					WorkingBranchSeparator: g.spec.WorkingBranchSeparator,
				}
				spec.Spec = g.spec.Spec

				results = append(results, spec)

				if g.limit > 0 && len(results) >= g.limit {
					return results, nil
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		listOpt.Page = resp.NextPage
	}

	return results, nil
}

// listBranches returns all branch names for the given project path.
func (g *GitLabSearch) listBranches(ctx context.Context, glClient *gitlab.Client, projectPath string) ([]string, error) {
	var branches []string

	branchOpt := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{Page: 1, PerPage: 100},
	}

	for {
		result, resp, err := glClient.Branches.ListBranches(projectPath, branchOpt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for _, b := range result {
			branches = append(branches, b.Name)
		}

		if resp.NextPage == 0 {
			break
		}
		branchOpt.Page = resp.NextPage
	}

	return branches, nil
}
