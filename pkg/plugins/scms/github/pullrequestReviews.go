package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// addPullrequestReviewers adds reviewers to a pull request
func (p *PullRequest) addPullrequestReviewers(prID string, retry int) error {

	rateLimit, err := queryRateLimit(p.gh.client, context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "API rate limit exceeded") {
			logrus.Debugln(rateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				rateLimit.Pause()
				return p.addPullrequestReviewers(prID, retry+1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}
	logrus.Debugln(rateLimit)

	if len(p.spec.Reviewers) == 0 {
		return nil
	}

	var mutation struct {
		RequestReviews struct {
			PullRequest struct {
				ID string
			}
		} `graphql:"requestReviews(input: $input)"`
	}

	input := githubv4.RequestReviewsInput{
		PullRequestID: githubv4.ID(prID),
		Union:         githubv4.NewBoolean(true), // Keep existing reviewers
	}

	var userIDs []githubv4.ID
	var teamIDs []githubv4.ID

	for _, reviewer := range p.spec.Reviewers {
		a := strings.Split(reviewer, "/")

		switch len(a) {
		case 2:
			teamID, err := getTeamID(p.gh.client, a[0], a[1], 0)
			logrus.Debugf("Team ID: %q found for %q", teamID, reviewer)
			if err != nil {
				logrus.Warningf("Failed to get team id for %s/%s: %v", a[0], a[1], err)
			} else {
				teamIDs = append(teamIDs, githubv4.NewID(teamID))
			}
		case 1:
			user, err := getUserInfo(p.gh.client, a[0], 0)
			logrus.Debugf("User ID: %q found for %q", user.ID, reviewer)
			if err != nil {
				logrus.Warningf("Failed to get user id for %s/%s: %v", a[0], a[1], err)
			} else {
				userIDs = append(userIDs, githubv4.NewID(user.ID))
			}
		case 0:
			// Ignore empty reviewer
		default:
			logrus.Warningf("Invalid reviewer format: %s", reviewer)
		}
	}

	if len(userIDs) > 0 {
		input.UserIDs = &userIDs
	}

	if len(teamIDs) > 0 {
		input.TeamIDs = &teamIDs
	}

	if len(userIDs) == 0 && len(teamIDs) == 0 {
		return fmt.Errorf("no valid reviewers found among %v", p.spec.Reviewers)
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return fmt.Errorf("adding pullrequest reviewers: %w", err)
	}

	return nil
}
