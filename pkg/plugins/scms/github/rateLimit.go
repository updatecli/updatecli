package github

import (
	"context"
	"fmt"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

const (
	// ErrAPIRateLimitExceeded is returned when the API rate limit is exceeded
	ErrAPIRateLimitExceeded = "API rate limit already exceeded"
	// ErrAPIRateLimitExceededFinalAttempt is returned when the API rate limit is exceeded and no more retry is possible
	ErrAPIRateLimitExceededFinalAttempt = "API rate limit exceeded, final attempt failed"
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

// Pause pauses the execution if the rate limit is reached until the reset time,
// or until ctx is done. It does nothing when a is nil, as the reset time is unknown.
func (a *RateLimit) Pause(ctx context.Context) {

	if a == nil || a.isEmpty() {
		log.Debug("GitHub RateLimit is empty, skipping Pause()")
		return
	}

	if a.Remaining == 0 {
		resetAtTime, err := time.Parse(time.RFC3339, a.ResetAt)
		if err != nil {
			log.Errorf("Parsing GitHub API rate limit reset time: %s", err)
			return
		}
		sleepDuration := time.Until(resetAtTime)

		log.Warningf(
			"GitHub API rate limit reached, on hold for %d minute(s) until %s. The process will resume automatically.\n",
			int(sleepDuration.Minutes()),
			resetAtTime.UTC().Format("2006-01-02 15:04:05 UTC"))

		timer := time.NewTimer(sleepDuration)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
		}
		return
	}

	log.Debugf("GitHub API credit used %d, remaining %d (reset at %s)",
		a.Cost, a.Remaining, a.ResetAt)
}

// isEmpty returns true if the RateLimit struct is empty
func (a RateLimit) isEmpty() bool {
	return a.Cost == 0 && a.Remaining == 0 && a.ResetAt == ""
}

// queryRateLimit queries the GitHub API rate limit information
func queryRateLimit(c client.Client, ctx context.Context) (*RateLimit, error) {

	query := struct {
		RateLimit RateLimit
	}{}

	err := c.Query(ctx, &query, nil)
	if err != nil {
		// GitHub may refuse the rateLimit query too once the rate limit is
		// exceeded, so fall back to the rate limit headers of that response.
		return rateLimitFromHeaders(c), fmt.Errorf("querying GitHub API: %w", err)
	}

	return &query.RateLimit, nil
}

// rateLimitFromHeaders returns the rate limit GitHub sent in the headers of
// the last response to c, or nil if c doesn't record them or saw none yet.
func rateLimitFromHeaders(c client.Client) *RateLimit {
	reporter, ok := c.(client.RateLimitReporter)
	if !ok {
		return nil
	}

	remaining, resetAt, ok := reporter.LastRateLimit()
	if !ok {
		return nil
	}

	return &RateLimit{
		Remaining: remaining,
		ResetAt:   resetAt.UTC().Format(time.RFC3339),
	}
}
