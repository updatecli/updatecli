package generic

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Add run `git add`.
func Add(files []string, workingDir string) error {

	for _, file := range files {
		fmt.Printf("Adding file: %s\n", file)

		r, err := git.PlainOpen(workingDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		_, err = w.Add(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func Checkout(branch, remoteBranch, workingDir string) error {

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// If remoteBranch already exist, use it
	// otherwise use the one define in the spec

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", remoteBranch),
		Create: false,
	})

	if err == plumbing.ErrReferenceNotFound {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewRemoteReferenceName("origin", branch),
			Create: false,
		})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(remoteBranch),
		Create: false,
	})

	if err == plumbing.ErrReferenceNotFound {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(remoteBranch),
			Create: true,
		})
		if err != nil {
			return err
		}
	} else if err != plumbing.ErrReferenceNotFound {
		return err
	} else {
		// Means that a local branch named remoteBranch already exist
		// so we want to be sure that the local branch is
		// aligned with the remote one.
		remoteBranchRef := fmt.Sprintf("refs/remotes/origin/%s", remoteBranch)

		fmt.Println(remoteBranchRef)

		remoteRef, err := r.Reference(
			plumbing.ReferenceName(
				remoteBranchRef), true)

		err = w.Reset(&git.ResetOptions{
			Commit: remoteRef.Hash(),
			Mode:   git.HardReset,
		})
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n")

	return nil
}

// Commit run `git commit`.
func Commit(user, email, message, workingDir string) error {

	fmt.Printf("Commit changes \n\n")

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}
	fmt.Println(status)

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  user,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	obj, err := r.CommitObject(commit)
	if err != nil {
		return err
	}
	fmt.Println(obj)

	return nil

}

// Clone run `git clone`.
func Clone(username, password, URL, workingDir string) error {

	var repo *git.Repository

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Cloning git repository: %s in %s\n", URL, workingDir))
	repo, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:      URL,
		Auth:     &auth,
		Progress: &b,
	})

	if err == git.ErrRepositoryAlreadyExists {
		b.Reset()

		repo, err = git.PlainOpen(workingDir)
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		status, err := w.Status()
		if err != nil {
			return err
		}
		b.WriteString(status.String())

		err = w.Pull(&git.PullOptions{
			Auth:     &auth,
			Force:    true,
			Progress: &b,
		})

		fmt.Println(b.String())
		b.Reset()
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	remotes, err := repo.Remotes()

	if err != nil {
		return err
	}

	b.WriteString("Fetching remote branches")
	for _, r := range remotes {

		err := r.Fetch(&git.FetchOptions{Progress: &b})
		fmt.Println(b.String())
		b.Reset()
		if err != nil {
			return err
		}
	}

	return err
}

// Push run `git push`.
func Push(username, password, workingDir string) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything excepted an empty string
		Password: password,
	}

	fmt.Printf("Push changes\n\n")

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	// Retrieve local branch
	head, err := r.Head()
	if err != nil {
		return err
	}

	if !head.Name().IsBranch() {
		return fmt.Errorf("Not pushing from a branch")
	}

	localBranch := strings.TrimPrefix(head.Name().String(), "refs/heads/")
	localRefSpec := head.Name().String()

	refspec := config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s",
		localRefSpec,
		localBranch))

	if err := refspec.Validate(); err != nil {
		return err
	}

	// Only push one branch at a time
	err = r.Push(&git.PushOptions{
		Auth:     &auth,
		Progress: os.Stdout,
		RefSpecs: []config.RefSpec{
			refspec,
		},
	})

	if err != nil {
		return err
	}

	fmt.Printf("\n")

	return nil
}

// SanitizeBranchName replace wrong character in the branch name
func SanitizeBranchName(branch string) string {
	branch = strings.ReplaceAll(branch, " ", "")
	branch = strings.ReplaceAll(branch, ":", "")
	branch = strings.ReplaceAll(branch, "*", "")
	branch = strings.ReplaceAll(branch, "+", "")

	if len(branch) > 255 {
		return branch[0:255]
	}
	return branch
}
