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

	if len(g.Spec.Tag) == 0 {
		g.Spec.Tag = source
	}

	if len(g.Spec.Title) == 0 {
		g.Spec.Title = g.Spec.Tag
	}

	if len(g.Spec.Commitish) == 0 {
		logrus.Warningf("No commitish provided, fallback to branch %q\n", "main")
		g.Spec.Commitish = "main"
	}

	// Ensure that a release doesn't exist yet

	ctx := context.Background()
	// Timeout api query after 30 second
	ctx, cancelListQuery := context.WithTimeout(ctx, 30*time.Second)
	defer cancelListQuery()

	releases, resp, err := g.client.Releases.List(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
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
		if r.Tag == g.Spec.Tag {
			logrus.Infof("%s Release Tag %q already exist, nothing else to do", result.SUCCESS, g.Spec.Tag)
			return false, nil
		}
	}

	if dryRun {
		logrus.Infof("%s Release Tag %q doesn't exist, we need to create it", result.SUCCESS, g.Spec.Tag)
		return true, nil
	}

	if len(g.Spec.Token) == 0 {
		return true, fmt.Errorf("wrong configuration, missing parameter %q", "token")
	}

	// Create a new release as it doesn't exist yet

	ctx = context.Background()
	// Timeout api query after 30 second
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	release, resp, err := g.client.Releases.Create(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		&goscm.ReleaseInput{
			Title:       g.Spec.Title,
			Description: g.Spec.Drescription + "\n" + updatecliCredits,
			Tag:         g.Spec.Tag,
			Commitish:   g.Spec.Commitish,
			Draft:       g.Spec.Draft,
			Prerelease:  g.Spec.Prerelease,
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
