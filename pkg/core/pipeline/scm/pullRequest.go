package scm

// PullRequest interface defines required funcions to be an pullRequest
type PullRequest interface {
	CreatePullRequest(title, changelog, pipelineReport string) error
}
