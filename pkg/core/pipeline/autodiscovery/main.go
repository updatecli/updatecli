package autodiscovery

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/argocd"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terragrunt"

	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/updatecli"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockercompose"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/fleet"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/flux"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/githubaction"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/golang"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helmfile"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/ko"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/kubernetes"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/maven"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/npm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/precommit"
)

var (
	// DefaultGenericsSpecs defines the default builder that we want to run
	DefaultCrawlerSpecs = Config{
		Crawlers: CrawlersConfig{
			"argocd":        argocd.Spec{},
			"cargo":         cargo.Spec{},
			"dockercompose": dockercompose.Spec{},
			"dockerfile":    dockerfile.Spec{},
			"flux":          flux.Spec{},
			// gitea/action share the same behavior as github/action
			// so we use the last one.
			// The day we have a specific behavior for gitea/action
			// then we will add it to the default autodiscovery.
			"github/action": githubaction.Spec{},
			"golang":        golang.Spec{},
			"helm":          helm.Spec{},
			"helmfile":      helmfile.Spec{},
			"ko":            ko.Spec{},
			"kubernetes":    kubernetes.Spec{},
			"maven":         maven.Spec{},
			"npm":           npm.Spec{},
			"precommit":     precommit.Spec{},
			"prow":          kubernetes.Spec{},
			"rancher/fleet": fleet.Spec{},
			"terraform":     &terraform.Spec{},
			"terragrunt":    &terragrunt.Spec{},
			"updatecli":     updatecli.Spec{},
		},
	}
	// AutodiscoverySpecs is a map of all Autodiscovery specification
	AutodiscoverySpecsMapping = map[string]interface{}{
		"argocd":        argocd.Spec{},
		"cargo":         &cargo.Spec{},
		"dockercompose": &dockercompose.Spec{},
		"dockerfile":    &dockerfile.Spec{},
		"flux":          &flux.Spec{},
		"github/action": &githubaction.Spec{},
		"gitea/action":  &githubaction.Spec{},
		"golang":        &golang.Spec{},
		"helm":          &helm.Spec{},
		"helmfile":      &helmfile.Spec{},
		"ko":            &ko.Spec{},
		"kubernetes":    &kubernetes.Spec{},
		"maven":         &maven.Spec{},
		"npm":           &npm.Spec{},
		"precommit":     &precommit.Spec{},
		"prow":          &kubernetes.Spec{},
		"rancher/fleet": &fleet.Spec{},
		"terraform":     &terraform.Spec{},
		"updatecli":     &updatecli.Spec{},
	}
)

type Crawler interface {
	DiscoverManifests() ([][]byte, error)
}

type AutoDiscovery struct {
	spec     Config
	crawlers []Crawler
}

// New returns an initiated autodiscovery object
//
//nolint:funlen // This function is responsible to create all the crawlers
func New(spec Config, workDir string) (*AutoDiscovery, error) {
	var errs []error
	var s Config

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return &AutoDiscovery{}, err
	}

	g := AutoDiscovery{
		spec: s,
	}

	for kind := range g.spec.Crawlers {
		if workDir == "" {
			logrus.Errorf("skipping crawler %q due to: %s", kind, err)
			continue
		}

		switch kind {
		case "argocd":
			argocdCrawler, err := argocd.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, argocdCrawler)
		case "cargo":
			cargoCrawler, err := cargo.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, cargoCrawler)

		case "dockercompose":
			crawler, err := dockercompose.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "dockerfile":
			crawler, err := dockerfile.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "flux":
			crawler, err := flux.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "github/action", "gitea/action":
			crawler, err := githubaction.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "golang", "go", "golang/gomod":
			crawler, err := golang.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "helm":
			crawler, err := helm.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "helmfile":
			crawler, err := helmfile.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "ko":
			crawler, err := ko.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "kubernetes":
			crawler, err := kubernetes.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				kubernetes.FlavorKubernetes)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "terraform":
			crawler, err := terraform.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "terragrunt":
			crawler, err := terragrunt.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "maven":
			crawler, err := maven.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "npm":
			crawler, err := npm.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)
		case "prow":
			crawler, err := kubernetes.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				kubernetes.FlavorProw)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "precommit":
			crawler, err := precommit.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "rancher/fleet":
			crawler, err := fleet.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		case "updatecli":
			crawler, err := updatecli.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, crawler)

		default:
			logrus.Infof("Crawler of type %q is not supported", kind)
		}
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Info(errs[i])
		}
	}

	return &g, nil
}

// Run execute each Autodiscovery crawlers to generate Updatecli manifests
func (g *AutoDiscovery) Run() ([][]byte, error) {
	var totalDiscoveredManifests [][]byte

	for _, crawler := range g.crawlers {

		discoveredManifests, err := crawler.DiscoverManifests()
		if err != nil {
			logrus.Errorln(err)
		}

		logrus.Printf("Manifest detected: %d\n", len(discoveredManifests))
		if len(discoveredManifests) > 0 {
			totalDiscoveredManifests = append(totalDiscoveredManifests, discoveredManifests...)
		}
	}

	logrus.Printf("\n\n---\n\n=> Total manifest detected: %d\n\n", len(totalDiscoveredManifests))

	return totalDiscoveredManifests, nil
}
