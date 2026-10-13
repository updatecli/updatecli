package gitbranch

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// branchRefs lists the repository branches. With lsremote they come from the remote,
// sorted by name and without dates.
func (gb *GitBranch) branchRefs() ([]gitgeneric.DatedBranch, error) {
	if !gb.lsRemote {
		return gb.nativeGitHandler.BranchRefs(gb.directory)
	}

	remote := git.NewRemote(nil, &config.RemoteConfig{Name: "origin", URLs: []string{gb.spec.URL}})
	listOptions := &git.ListOptions{}
	if gb.spec.Username != "" && gb.spec.Password != "" {
		listOptions.Auth = &http.BasicAuth{Username: gb.spec.Username, Password: gb.spec.Password}
	}

	refs, err := remote.List(listOptions)
	if err != nil {
		return nil, fmt.Errorf("listing remote branches: %w", err)
	}

	branches := []gitgeneric.DatedBranch{}
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branches = append(branches, gitgeneric.DatedBranch{Name: ref.Name().Short(), Hash: ref.Hash().String()})
		}
	}
	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })

	return branches, nil
}
