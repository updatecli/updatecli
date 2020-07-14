package generic

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
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

	// If remoteBranch already exist then use it
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
		// Means that a local branch named remoteBranch alreayd exist so we want to be sure
		// that the local branch is aligned with remote.
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
func Commit(file, user, email, message, workingDir string) error {

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

	fmt.Printf("Cloning git repository: %s\n", URL)
	fmt.Printf("Downloaded in %s\n\n", workingDir)

	repo, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:      URL,
		Auth:     &auth,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		fmt.Println(err)
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
		fmt.Println(status)

		err = w.Pull(&git.PullOptions{
			Auth:     &auth,
			Force:    true,
			Progress: os.Stdout,
		})
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

	for _, r := range remotes {

		err := r.Fetch(&git.FetchOptions{})
		if err != nil {
			return err
		}
	}

	return err
}

// Push run `git push` then open a pull request on Github if not already created.
func Push(username, password, workingDir string) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	fmt.Printf("Push changes\n\n")

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
		Auth:     &auth,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n")

	return nil
}
