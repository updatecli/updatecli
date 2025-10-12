package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/token"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// GetBranches returns source, working and target branch
func (g *Github) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = g.Spec.Branch
	workingBranch = g.Spec.Branch
	targetBranch = g.Spec.Branch

	if len(g.pipelineID) > 0 && g.workingBranch {
		workingBranch = g.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{g.workingBranchPrefix, targetBranch, g.pipelineID}, g.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// GetURL returns a "GitHub " git URL
func (g *Github) GetURL() string {
	URL, err := url.JoinPath(g.Spec.URL, g.Spec.Owner, g.Spec.Repository+".git")
	if err != nil {
		logrus.Errorln(err)
		return ""
	}

	return URL
}

// GetDirectory returns the local git repository path.
func (g *Github) GetDirectory() (directory string) {
	return g.Spec.Directory
}

// Clean deletes github working directory.
func (g *Github) Clean() error {
	err := os.RemoveAll(g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Github) Clone() (string, error) {
	g.setDirectory()

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	err = g.nativeGitHandler.Clone(
		g.username,
		accessToken,
		g.GetURL(),
		g.GetDirectory(),
		g.Spec.Submodules,
	)
	if err != nil {
		logrus.Errorf("failed cloning GitHub repository %q", g.GetURL())
		return "", err
	}

	return g.Spec.Directory, nil
}

// Commit run `git commit`.
func (g *Github) Commit(message string) error {

	workingDir := g.GetDirectory()

	// Generate the conventional commit message
	commitMessage, err := g.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	if g.commitUsingApi {

		logrus.Debugf("Creating commit using GitHub API")
		_, workingBranch, _ := g.GetBranches()

		commitHashPrePull, err := g.nativeGitHandler.GetLatestCommitHash(workingDir)
		if err != nil {
			return err
		}

		commitHasPostApiQuery, err := g.CreateCommit(workingDir, commitMessage, 0)
		if err != nil {
			return err
		}

		accessToken, err := token.GetAccessToken(g.token)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}

		if err = g.nativeGitHandler.Pull(
			g.username,
			accessToken,
			workingDir,
			workingBranch,
			true,
			true,
		); err != nil {
			return err
		}

		commitHashPostPull, err := g.nativeGitHandler.GetLatestCommitHash(workingDir)
		if err != nil {
			return err
		}

		// Probably due to some caching, the commit is not immediately available after creation
		// We retry to pull the commit a few times until we find it or we reach the max retry
		maxRetry := 3
		for counter := 0; counter < maxRetry; counter++ {

			// Ideally we should check that the latest local commit hash
			// is an ancestor of the newly created commit hash
			// but this operation is expensive so we just check that
			// the latest local commit hash is different from the newly created one.
			// This should be enough in most of the cases considering the working branch
			// is created and maintained by Updatecli.
			if commitHashPostPull == commitHasPostApiQuery {
				break
			}

			logrus.Debugf("Latest local commit %q should have been %q", commitHashPrePull, commitHasPostApiQuery)
			logrus.Debugf("Waiting for GitHub to make the commit %q available", commitHasPostApiQuery)

			logrus.Debugf("Commit not found yet, retrying to pull it")

			if err = g.nativeGitHandler.Pull(
				g.Spec.Username,
				g.Spec.Token,
				workingDir,
				workingBranch,
				true,
				true,
			); err != nil {
				return err
			}

			commitHashPostPull, err = g.nativeGitHandler.GetLatestCommitHash(workingDir)
			if err != nil {
				return err
			}

			logrus.Debugf("Latest commit after creating a new one: %q", commitHasPostApiQuery)

			if counter == maxRetry-1 {
				logrus.Debugf("Giving up trying to pull the newly created commit")
			}
		}

		if g.Spec.CommitMessage.IsSquash() {
			logrus.Warningf("Squash commit is not supported when using GitHub API to create the commit. Ignoring the squash option.")
		}

	} else {
		logrus.Debugf("Creating commit using native git")

		if err = g.nativeGitHandler.Commit(
			g.Spec.User,
			g.Spec.Email,
			commitMessage,
			workingDir,
			g.Spec.GPG.SigningKey,
			g.Spec.GPG.Passphrase,
		); err != nil {
			return err
		}

		if g.Spec.CommitMessage.IsSquash() {
			sourceBranch, workingBranch, _ := g.GetBranches()
			if err = g.nativeGitHandler.SquashCommit(workingDir, sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
				IncludeCommitTitles: true,
				Message:             commitMessage,
				SigninKey:           g.Spec.GPG.SigningKey,
				SigninPassphrase:    g.Spec.GPG.Passphrase,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// CommitQuery defines a github v4 API mutation to create a commit on a branch
type commitQuery struct {
	CreateCommitOnBranch struct {
		Commit struct {
			URL string
			OID string
		}
	} `graphql:"createCommitOnBranch(input:$input)"`
}

// CreateCommit creates a commit on a branch using GitHub API
func (g *Github) CreateCommit(workingDir string, commitMessage string, retry int) (createdCommitHash string, err error) {
	var m commitQuery

	sourceBranch, workingBranch, _ := g.GetBranches()

	files, err := g.nativeGitHandler.GetChangedFiles(workingDir)
	if err != nil {
		return "", err
	}

	additions, err := processChangedFiles(workingDir, files)
	if err != nil {
		return "", err
	}

	if g.nativeGitHandler.IsForceReset() {
		// Ensure that locally reset branch is pushed to remote branch
		// before continuing with commit creation as that will otherwise
		// be lost.
		logrus.Debugf("local branch %q was reset, pushing to remote to ensure correct state", workingBranch)
		if _, err = g.Push(); err != nil {
			return "", fmt.Errorf("failed to push branch %q before creating commit: %w", workingBranch, err)
		}
	}

	repoRef, err := g.GetLatestCommitHash(workingBranch)
	if err != nil {
		return "", fmt.Errorf("retrieving latest commit hash for branch %q: %w", workingBranch, err)
	}

	headOid := repoRef.HeadOid
	if headOid == "" {
		sourceBranchRepoRef, err := g.GetLatestCommitHash(sourceBranch)
		if err != nil {
			return "", fmt.Errorf("retrieving latest commit hash for source branch %q: %w", sourceBranch, err)
		}

		headOid = sourceBranchRepoRef.HeadOid
		logrus.Debugf("Branch %s does not exist, creating it from commit %q", workingBranch, headOid)
		if err := g.createBranch(workingBranch, repoRef.ID, headOid, 0); err != nil {
			return "", fmt.Errorf("creating branch %q from commit %q: %w", workingBranch, headOid, err)
		}
	}

	repositoryName := fmt.Sprintf("%s/%s", g.Spec.Owner, g.Spec.Repository)
	input := githubv4.CreateCommitOnBranchInput{
		Branch: githubv4.CommittableBranch{
			RepositoryNameWithOwner: githubv4.NewString(githubv4.String(repositoryName)),
			BranchName:              githubv4.NewString(githubv4.String(fmt.Sprintf("refs/heads/%s", workingBranch))),
		},
		ExpectedHeadOid: githubv4.GitObjectID(headOid),
		Message: githubv4.CommitMessage{
			Headline: githubv4.String(commitMessage),
		},
		FileChanges: &githubv4.FileChanges{
			Additions: &additions,
		},
	}

	rateLimit, err := queryRateLimit(g.client, context.Background())
	logrus.Debugln(rateLimit)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return g.CreateCommit(workingDir, commitMessage, retry+1)
			}
			return "", errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return "", fmt.Errorf("querying GitHub API rate limit: %w", err)

	}

	if err := g.client.Mutate(context.Background(), &m, input, nil); err != nil {
		// In some occasions, GitHub API may respond with
		// "Expected branch to point to <commit hash> but was <commit hash>"
		// even if we provide the correct commit hash.
		// In that case, we retry a few times before giving up.
		// This is probably due to some caching on GitHub side.
		if retry < client.MaxRetry && strings.Contains(err.Error(), "Expected branch to point to") {
			return g.CreateCommit(workingDir, commitMessage, retry+1)
		}
		return "", fmt.Errorf("creating commit on branch %q: %w", workingBranch, err)
	}

	logrus.Debugf("commit created: %s", m.CreateCommitOnBranch.Commit.URL)

	return m.CreateCommitOnBranch.Commit.OID, nil
}

// processChangedFiles reads the content of the files and prepare them to be
func processChangedFiles(workingDir string, files []string) ([]githubv4.FileAddition, error) {
	additions := make([]githubv4.FileAddition, 0, len(files))
	for _, f := range files {
		fullPath := fmt.Sprintf("%s/%s", workingDir, f)
		enc, err := utils.Base64EncodeFile(fullPath)
		if err != nil {
			return additions, err
		}
		additions = append(additions, githubv4.FileAddition{
			Path:     githubv4.String(f),
			Contents: githubv4.Base64String(enc),
		})
	}
	return additions, nil
}

// GetLatestCommitHash returns the latest commit hash of the specified branch
func (g *Github) GetLatestCommitHash(workingBranch string) (*RepositoryRef, error) {
	repoRef, err := g.queryHeadOid(workingBranch, 0)
	if err != nil {
		return nil, err
	}
	return repoRef, nil
}

// Checkout create and then uses a temporary git branch.
func (g *Github) Checkout() error {
	sourceBranch, workingBranch, _ := g.GetBranches()

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	return g.nativeGitHandler.Checkout(
		g.username,
		accessToken,
		sourceBranch,
		workingBranch,
		g.Spec.Directory,
		g.force,
	)
}

// Add run `git add`.
func (g *Github) Add(files []string) error {
	err := g.nativeGitHandler.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the branch reference name is published on
// on the default remote
func (g *Github) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := g.GetBranches()

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return false, fmt.Errorf("failed to get access token: %w", err)
	}

	return g.nativeGitHandler.IsLocalBranchPublished(
		sourceBranch,
		workingBranch,
		g.username,
		accessToken,
		g.GetDirectory())
}

// Push run `git push` on the GitHub remote branch if not already created.
func (g *Github) Push() (bool, error) {

	// If the commit is done using the GitHub API, we don't need to push
	// the commit as it is done in the same operation.
	if g.commitUsingApi {
		logrus.Debugf("commit done using GitHub API, normally nothing need to be push but we may have left over.")
	}

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return false, fmt.Errorf("failed to get access token: %w", err)
	}

	return g.nativeGitHandler.Push(
		g.username,
		accessToken,
		g.GetDirectory(),
		g.force,
	)
}

// PushTag push tags
func (g *Github) PushTag(tag string) error {

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	err = g.nativeGitHandler.PushTag(
		tag,
		g.username,
		accessToken,
		g.GetDirectory(),
		g.force,
	)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push tags
func (g *Github) PushBranch(branch string) error {

	accessToken, err := token.GetAccessToken(g.token)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	err = g.nativeGitHandler.PushBranch(
		branch,
		g.username,
		accessToken,
		g.GetDirectory(),
		g.force)
	if err != nil {
		return err
	}

	return nil
}

// GetChangedFiles returns a list of changed files in the working directory
func (g *Github) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
