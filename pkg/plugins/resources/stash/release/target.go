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

func (g Stash) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget result.Target) error {
	if len(g.spec.Tag) == 0 {
		g.spec.Tag = source
	}

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
		logrus.Debugf("Bitbucket Api Response:\nReturn Code: %d\nBody:\n%s", resp.Status, resp.Body)
		return err
	}

	resultTarget.NewInformation = g.spec.Tag
	for _, r := range releases {
		if r.Tag == g.spec.Tag {
			resultTarget.Information = resultTarget.NewInformation
			resultTarget.Result = result.SUCCESS
			resultTarget.Description = fmt.Sprintf("Stash release tag %q already exist", g.spec.Tag)
			return nil
		}
	}

	resultTarget.Information = ""
	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true

	if dryRun {
		resultTarget.Description = fmt.Sprintf("Stash release tag %q should be created", resultTarget.NewInformation)
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
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		return fmt.Errorf("error from Bitbucket api: %v", resp.Status)
	}

	resultTarget.Description = fmt.Sprintf("Stash release %q successfully open on %q", release.Title, release.Link)
	return nil
}
