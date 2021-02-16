package scm

// PullRequest interface defines required funcions to be an pullRequest
type PullRequest interface {
	UpdatePullRequest(ID string) error
	OpenPullRequest() error
	IsPullRequest() (ID string, err error)
}
