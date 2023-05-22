package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	goscm "github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target ensure that a specific release exist on GitLab, otherwise creates it
func (g Gitlab) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	if len(g.spec.Tag) == 0 {
		g.spec.Tag = source
	}

	resultTarget.NewInformation = g.spec.Tag

	if len(g.spec.Title) == 0 {
		g.spec.Title = g.spec.Tag
	}

	if len(g.spec.Commitish) == 0 {
		logrus.Warningf("No commitish provided, fallback to branch %q\n", "main")
		g.spec.Commitish = "main"
	}

	// Ensure that a release doesn't exist yet

	ctx := context.Background()
	// Timeout api query after 30 second
	ctx, cancelListQuery := context.WithTimeout(ctx, 30*time.Second)
	defer cancelListQuery()

	releases, resp, err := g.client.Releases.List(
		ctx,
		strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
		goscm.ReleaseListOptions{
			Page:   1,
			Size:   30,
			Open:   true,
			Closed: true,
		},
	)

	if err != nil {
		logrus.Debugf("GitLab Api Response:\nReturn Code: %q\nBody:\n%s", resp.Status, resp.Body)
		return err
	}

	for _, r := range releases {
		if r.Tag == g.spec.Tag {
			resultTarget.Result = result.SUCCESS
			resultTarget.OldInformation = g.spec.Tag
			resultTarget.Description = fmt.Sprintf("GitLab release tag %q already exist", g.spec.Tag)
			return nil
		}
	}

	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true

	if dryRun {
		resultTarget.Description = fmt.Sprintf("GitLab release tag %q doesn't exist, we need to create it", g.spec.Tag)
		return nil
	}

	if len(g.spec.Token) == 0 {
		return fmt.Errorf("wrong configuration, missing parameter %q", "token")
	}

	// Create a new release as it doesn't exist yet

	ctx = context.Background()
	// Timeout api query after 30 second
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	release, resp, err := g.client.Releases.Create(
		ctx,
		strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
		&goscm.ReleaseInput{
			Title:       g.spec.Title,
			Description: g.spec.Description + "\n" + updatecliCredits,
			Tag:         g.spec.Tag,
			Commitish:   g.spec.Commitish,
			Draft:       g.spec.Draft,
			Prerelease:  g.spec.Prerelease,
		},
	)

	if err != nil {
		return err
	}

	if resp.Status >= 400 {
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		return fmt.Errorf("error from GitLab api: %v", resp.Status)
	}

	resultTarget.Description = fmt.Sprintf("GitLab Release %q successfully opened on %q", release.Title, release.Link)

	return nil
}
