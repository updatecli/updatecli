package gitgeneric

import (
	"fmt"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
)

// SquashCommitOptions defines options for the squash operation
type SquashCommitOptions struct {
	// IncludeCommitTitles indicates whether to include original commit titles in the squash message
	IncludeCommitTitles bool
	// Format defines how to format included commits
	Format string
	// Message defines the main message for the squash commit
	Message string
	// SigningKey is the GPG key used to sign the commit
	SigninKey string
	// SigninPassphrase is the passphrase for the GPG key
	SigninPassphrase string
}

// SquashCommits squashes all commits between baseBranch and featureBranch into a single commit
func (g GoGit) SquashCommit(repoDir, baseBranch, featureBranch string, options SquashCommitOptions) error {

	if err := validateSquashOperation(repoDir, baseBranch, featureBranch); err != nil {
		return fmt.Errorf("squash operation validation failed: %w", err)
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Resolve branch heads
	baseRef, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return fmt.Errorf("failed to resolve base branch %s: %w", baseBranch, err)
	}
	featureRef, err := repo.Reference(plumbing.NewBranchReferenceName(featureBranch), true)
	if err != nil {
		return fmt.Errorf("failed to resolve feature branch %s: %w", featureBranch, err)
	}

	// Get commit objects
	baseCommit, err := repo.CommitObject(baseRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get base commit: %w", err)
	}
	featureCommit, err := repo.CommitObject(featureRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get feature commit: %w", err)
	}

	// Validation: Check if feature branch has commits to squash
	if baseCommit.Hash == featureCommit.Hash {
		return fmt.Errorf("feature branch %s is at same commit as base branch %s, nothing to squash", featureBranch, baseBranch)
	}

	// Check if feature branch is ahead of base
	isAncestor, err := baseCommit.IsAncestor(featureCommit)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %w", err)
	}
	if !isAncestor {
		return fmt.Errorf("feature branch %s is not ahead of base branch %s", featureBranch, baseBranch)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get base tree: %w", err)
	}
	featureTree, err := featureCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get feature tree: %w", err)
	}

	// Check for changes
	changes, err := baseTree.Diff(featureTree)
	if err != nil {
		return fmt.Errorf("failed to compute diff: %w", err)
	}

	// If no changes, nothing to squash then in the context of Updatecli
	// We may decide to reset the branch to the base commit.
	if len(changes) == 0 {
		return fmt.Errorf("no changes found between %s and %s", baseBranch, featureBranch)
	}

	// Collect original commit information if requested
	var finalMessage string
	if options.IncludeCommitTitles {
		originalCommits, err := getOriginalCommits(baseCommit, featureCommit)
		if err != nil {
			return fmt.Errorf("failed to collect original commits: %w", err)
		}

		finalMessage = buildCommitMessage(options.Message, originalCommits, options.Format)
		logrus.Debugf("Collected %d original commits for inclusion in message\n", len(originalCommits))
	} else {
		finalMessage = options.Message
	}

	newCommit := &object.Commit{
		Author: object.Signature{
			Name:  featureCommit.Author.Name,
			Email: featureCommit.Author.Email,
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  featureCommit.Committer.Name,
			Email: featureCommit.Committer.Email,
			When:  time.Now(),
		},
		Message:      finalMessage, // Use the enhanced message
		TreeHash:     featureTree.Hash,
		ParentHashes: []plumbing.Hash{baseCommit.Hash},
	}

	// Sign the commit if signing key is provided
	if len(options.SigninKey) > 0 {

		if err := validateGPGKey(options.SigninKey, options.SigninPassphrase); err != nil {
			return fmt.Errorf("validating GPG key: %w", err)
		}

		logrus.Debugf("Signing squashed commit with provided key\n")

		// Get the signing key using your existing function
		signer, err := sign.GetCommitSignKey(options.SigninKey, options.SigninPassphrase)
		if err != nil {
			return fmt.Errorf("getting commit sign key: %w", err)
		}

		// Sign the commit using the new SignCommit function
		signedCommit, err := sign.SignCommit(newCommit, signer)
		if err != nil {
			return fmt.Errorf("failed to sign commit: %w", err)
		}

		// Use the signed commit
		newCommit = signedCommit
		logrus.Debugf("Commit successfully signed")
	} else {
		logrus.Debugln("No signing key provided, creating unsigned squashed commit")
	}

	// Store the commit object in the repository
	obj := repo.Storer.NewEncodedObject()
	if err := newCommit.Encode(obj); err != nil {
		return fmt.Errorf("failed to encode commit: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store commit: %w", err)
	}

	// Update the feature branch reference
	newRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(featureBranch), commitHash)
	if err := repo.Storer.SetReference(newRef); err != nil {
		return fmt.Errorf("failed to update branch reference: %w", err)
	}

	// Update working tree if we're currently on the feature branch
	head, err := repo.Head()
	if err == nil && head.Name().String() == plumbing.NewBranchReferenceName(featureBranch).String() {
		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}

		if err := worktree.Reset(&git.ResetOptions{
			Commit: commitHash,
			Mode:   git.HardReset,
		}); err != nil {
			return fmt.Errorf("failed to reset worktree: %w", err)
		}
	}

	return nil
}

// CommitInfo holds information about a commit being squashed
type CommitInfo struct {
	Hash      string
	Title     string
	Author    string
	Timestamp time.Time
	IsSigned  bool
}

// getOriginalCommits collects information about commits being squashed
func getOriginalCommits(baseCommit, featureCommit *object.Commit) ([]CommitInfo, error) {
	var commits []CommitInfo

	// Walk through commits from feature back to base (in reverse chronological order)
	commitIter := object.NewCommitPreorderIter(featureCommit, nil, nil)
	err := commitIter.ForEach(func(c *object.Commit) error {
		// Stop when we reach the base commit
		if c.Hash == baseCommit.Hash {
			return fmt.Errorf("reached base commit") // Use error to stop iteration
		}

		// Extract commit title (first line of message)
		title := strings.Split(strings.TrimSpace(c.Message), "\n")[0]

		commits = append(commits, CommitInfo{
			Hash:      c.Hash.String()[:8], // Short hash
			Title:     title,
			Author:    c.Author.Name,
			Timestamp: c.Author.When,
			IsSigned:  c.PGPSignature != "",
		})

		return nil
	})

	if err != nil && err.Error() != "reached base commit" {
		return nil, err
	}

	// Reverse the slice to get chronological order (oldest first)
	for i := len(commits)/2 - 1; i >= 0; i-- {
		opp := len(commits) - 1 - i
		commits[i], commits[opp] = commits[opp], commits[i]
	}

	return commits, nil
}

// buildCommitMessage creates the final commit message including original commit info
func buildCommitMessage(squashMsg string, commits []CommitInfo, format string) string {
	var builder strings.Builder

	// Start with the main squash message
	builder.WriteString(squashMsg)

	if len(commits) == 0 {
		return builder.String()
	}

	// Add separator
	builder.WriteString("\n\n")

	signedIndicatorValue := " [signed]"

	switch format {
	case "compact":
		builder.WriteString(fmt.Sprintf("Squashed %d commits:\n", len(commits)))
		for _, commit := range commits {
			signedIndicator := ""
			if commit.IsSigned {
				signedIndicator = signedIndicatorValue
			}
			builder.WriteString(fmt.Sprintf("- %s%s\n", commit.Title, signedIndicator))
		}

	case "detailed":
		builder.WriteString("Squashed commits:\n")
		for _, commit := range commits {
			signedIndicator := ""
			if commit.IsSigned {
				signedIndicator = signedIndicatorValue
			}
			builder.WriteString(fmt.Sprintf("* %s - %s (%s)%s\n",
				commit.Hash, commit.Title, commit.Author, signedIndicator))
		}

	case "list":
		builder.WriteString("Squashed commits:\n")
		for _, commit := range commits {
			signedIndicator := ""
			if commit.IsSigned {
				signedIndicator = signedIndicatorValue
			}
			builder.WriteString(fmt.Sprintf("* %s: %s%s\n", commit.Hash, commit.Title, signedIndicator))
		}

	case "":
		// Do not include original commit info

	default:
		logrus.Errorf("unknown commit message format %q", format)
	}

	return builder.String()
}

// validateSquashOperation checks if a squash operation is safe to perform
func validateSquashOperation(workingDir, baseBranch, featureBranch string) error {
	repo, err := git.PlainOpen(workingDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Check if branches exist
	if _, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true); err != nil {
		return fmt.Errorf("base branch %s does not exist: %w", baseBranch, err)
	}
	if _, err := repo.Reference(plumbing.NewBranchReferenceName(featureBranch), true); err != nil {
		return fmt.Errorf("feature branch %s does not exist: %w", featureBranch, err)
	}

	// Check for uncommitted changes
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if !status.IsClean() {
		return fmt.Errorf("working directory has uncommitted changes")
	}

	return nil
}

// validateGPGKey validates that the GPG key can be used for signing
func validateGPGKey(armoredKeyRing, passphrase string) error {
	if armoredKeyRing == "" {
		return nil
	}

	_, err := sign.GetCommitSignKey(armoredKeyRing, passphrase)
	if err != nil {
		return fmt.Errorf("failed to load GPG key: %w", err)
	}

	return nil
}
