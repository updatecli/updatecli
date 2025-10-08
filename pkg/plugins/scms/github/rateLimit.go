package github

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimit is a struct that contains GitHub Api limit information
type RateLimit struct {
	Cost      int
	Remaining int
	ResetAt   string
}

// String returns a string representation of the RateLimit struct
func (a RateLimit) String() string {
	if a.isEmpty() {
		return "GitHub RateLimit is empty"
	}

	return fmt.Sprintf("GitHub API credit used %d, remaining %d (reset at %s)",
		a.Cost, a.Remaining, a.ResetAt)
}

// Pause pauses the execution if the rate limit is reached until the reset time
func (a RateLimit) Pause() {

	if a.isEmpty() {
		logrus.Debug("GitHub RateLimit is empty, skipping Pause()")
		return
	}

	if a.Remaining == 0 {
		resetAtTime, err := time.Parse(time.RFC3339, a.ResetAt)
		if err != nil {
			logrus.Errorf("Parsing GitHub API rate limit reset time: %s", err)
			return
		}
		sleepDuration := time.Until(resetAtTime)

		logrus.Warningf(
			"GitHub API rate limit reached, on hold for %d minute(s) until %s. The process will resume automatically.\n",
			int(sleepDuration.Minutes()),
			resetAtTime.UTC().Format("2006-01-02 15:04:05 UTC"))

		time.Sleep(sleepDuration)
		return
	}

	logrus.Debugf("GitHub API credit used %d, remaining %d (reset at %s)",
		a.Cost, a.Remaining, a.ResetAt)
}

// isEmpty returns true if the RateLimit struct is empty
func (a RateLimit) isEmpty() bool {
	return a.Cost == 0 && a.Remaining == 0 && a.ResetAt == ""
}

// queryRateLimit queries the GitHub API rate limit information
func queryRateLimit(client GitHubClient, ctx context.Context) (*RateLimit, error) {

	query := struct {
		RateLimit RateLimit
	}{}

	err := client.Query(ctx, &query, nil)
	if err != nil {
		return nil, fmt.Errorf("querying GitHub API: %w", err)
	}

	return &query.RateLimit, nil
}
