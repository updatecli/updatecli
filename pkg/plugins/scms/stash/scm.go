package stash

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func (s *Stash) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = s.Spec.Branch
	workingBranch = s.nativeGitHandler.SanitizeBranchName(fmt.Sprintf("updatecli_%v", s.pipelineID))
	targetBranch = s.Spec.Branch

	return sourceBranch, workingBranch, targetBranch
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
		s.Spec.User,
		s.Spec.Token,
		s.GetURL(),
		s.GetDirectory())

	if err != nil {
		logrus.Errorf("failed cloning Bitbucket repository %q", s.GetURL())
		return "", err
	}

	sourceBranch, workingBranch, _ := s.GetBranches()

	if len(workingBranch) > 0 && len(s.GetDirectory()) > 0 {
		err = s.nativeGitHandler.Checkout(
			s.Spec.Username,
			s.Spec.Token,
			sourceBranch,
			workingBranch,
			s.GetDirectory(),
			true)
	}

	if err != nil {
		logrus.Errorf("initial Bitbucket checkout failed for repository %q", s.GetURL())
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
		false)
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

// IsRemoteBranchUpToDate checks if the branch reference name is published on
// on the default remote
func (s *Stash) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := s.GetBranches()

	return s.nativeGitHandler.IsLocalBranchPublished(
		sourceBranch,
		workingBranch,
		s.Spec.Username,
		s.Spec.Token,
		s.GetDirectory())
}

// Push run `git push` to the corresponding Bitbucket remote branch if not already created.
func (s *Stash) Push() error {

	err := s.nativeGitHandler.Push(s.Spec.Username, s.Spec.Token, s.GetDirectory(), s.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (s *Stash) PushTag(tag string) error {

	err := s.nativeGitHandler.PushTag(tag, s.Spec.Username, s.Spec.Token, s.GetDirectory(), s.Spec.Force)
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
		s.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stash) GetChangedFiles(workingDir string) ([]string, error) {
	return s.nativeGitHandler.GetChangedFiles(workingDir)
}
