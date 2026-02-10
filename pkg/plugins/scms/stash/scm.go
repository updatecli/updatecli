package stash

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func (s *Stash) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = s.Spec.Branch
	workingBranch = s.Spec.Branch
	targetBranch = s.Spec.Branch

	if len(s.pipelineID) > 0 && s.workingBranch {
		workingBranch = s.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{s.workingBranchPrefix, targetBranch, s.pipelineID}, s.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch
// and remove it if not.
func (s *Stash) CleanWorkingBranch() (bool, error) {
	_, workingBranch, targetBranch := s.GetBranches()

	if workingBranch == targetBranch {
		logrus.Infof("Skipping cleaning working branch %q on %q (same as target branch)\n", workingBranch, s.GetURL())
		return false, nil
	}

	isSimilarBranch, err := s.nativeGitHandler.IsSimilarBranch(workingBranch, targetBranch, s.GetDirectory())
	if err != nil {
		return false, fmt.Errorf("failed to compare working branch %q with target branch %q: %w", workingBranch, targetBranch, err)
	}

	if isSimilarBranch {
		if err = s.nativeGitHandler.DeleteBranch(workingBranch, s.GetDirectory(), s.Spec.Username, s.Spec.Token); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, s.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetURL returns a "Stash" git URL
func (s *Stash) GetURL() string {
	URL := fmt.Sprintf("%v/scm/%v/%v.git",
		s.Spec.URL,
		s.Spec.Owner,
		s.Spec.Repository)

	return URL
}

// GetDirectory returns the local git repository path.
func (s *Stash) GetDirectory() (directory string) {
	return s.Spec.Directory
}

// Clean deletes github working directory.
func (s *Stash) Clean() error {
	err := os.RemoveAll(s.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (s *Stash) Clone() (string, error) {
	s.setDirectory()

	err := s.nativeGitHandler.Clone(
		s.Spec.Username,
		s.Spec.Token,
		s.GetURL(),
		s.GetDirectory(),
		s.Spec.Submodules,
		s.Spec.Depth,
	)
	if err != nil {
		logrus.Errorf("failed cloning Bitbucket repository %q", s.GetURL())
		return "", err
	}

	return s.Spec.Directory, nil
}

// Commit run `git commit`.
func (s *Stash) Commit(message string) error {
	// Generate the conventional commit message
	commitMessage, err := s.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = s.nativeGitHandler.Commit(s.Spec.User, s.Spec.Email, commitMessage, s.GetDirectory(), s.Spec.GPG.SigningKey, s.Spec.GPG.Passphrase)
	if err != nil {
		return err
	}

	if s.Spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := s.GetBranches()
		if err = s.nativeGitHandler.SquashCommit(s.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           s.Spec.GPG.SigningKey,
			SigninPassphrase:    s.Spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Checkout create and then uses a temporary git branch.
func (s *Stash) Checkout() error {
	sourceBranch, workingBranch, _ := s.GetBranches()

	err := s.nativeGitHandler.Checkout(
		s.Spec.Username,
		s.Spec.Token,
		sourceBranch,
		workingBranch,
		s.Spec.Directory,
		s.force,
		s.Spec.Depth,
	)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (s *Stash) Add(files []string) error {
	err := s.nativeGitHandler.Add(files, s.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the local working branch is up to date with the remote branch.
func (s *Stash) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := s.GetBranches()

	return s.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory())
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (s *Stash) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := s.GetBranches()

	return s.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory())
}

// Push run `git push` to the corresponding Bitbucket Server remote branch if not already created.
func (s *Stash) Push() (bool, error) {
	return s.nativeGitHandler.Push(
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory(),
		s.force)
}

// PushTag push tags
func (s *Stash) PushTag(tag string) error {
	err := s.nativeGitHandler.PushTag(
		tag,
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory(),
		s.force,
	)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push branch
func (s *Stash) PushBranch(branch string) error {
	err := s.nativeGitHandler.PushTag(
		branch,
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory(),
		s.force)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stash) GetChangedFiles(workingDir string) ([]string, error) {
	return s.nativeGitHandler.GetChangedFiles(workingDir)
}
