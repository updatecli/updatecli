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

// Target ensure that a specific release exist on gitea, otherwise creates it
func (g *Gitea) Target(source string, dryRun bool) (bool, error) {

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
		logrus.Debugf("Gitea Api Response:\nReturn Code: %q\nBody:\n%s", resp.Status, resp.Body)
		return false, err
	}

	for _, r := range releases {
		if r.Tag == g.spec.Tag {
			logrus.Infof("%s Release Tag %q already exist, nothing else to do", result.SUCCESS, g.spec.Tag)
			return false, nil
		}
	}

	if dryRun {
		logrus.Infof("%s Release Tag %q doesn't exist, we need to create it", result.SUCCESS, g.spec.Tag)
		return true, nil
	}

	if len(g.spec.Token) == 0 {
		return true, fmt.Errorf("wrong configuration, missing parameter %q", "token")
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
			Description: g.spec.Drescription + "\n" + updatecliCredits,
			Tag:         g.spec.Tag,
			Commitish:   g.spec.Commitish,
			Draft:       g.spec.Draft,
			Prerelease:  g.spec.Prerelease,
		},
	)

	if err != nil {
		return false, err
	}

	if resp.Status >= 400 {
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		return false, fmt.Errorf("error from Gitea api: %v", resp.Status)
	}

	logrus.Infof("Gitea Release %q successfully open on %q", release.Title, release.Link)

	return true, nil
}

func (g Gitea) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("target not supported for the plugin Gitea Release")
}
