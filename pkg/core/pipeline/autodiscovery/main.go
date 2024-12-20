package autodiscovery

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/argocd"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/cargo"
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
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/nomad"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/npm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/precommit"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terragrunt"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/updatecli"
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
			"nomad":         nomad.Spec{},
			"npm":           npm.Spec{},
			"precommit":     precommit.Spec{},
			"prow":          kubernetes.Spec{},
			"rancher/fleet": fleet.Spec{},
			"terraform":     terraform.Spec{},
			"terragrunt":    terragrunt.Spec{},
			"updatecli":     updatecli.Spec{},
		},
	}
	// AutodiscoverySpecs is a map of all Autodiscovery specification
	AutodiscoverySpecsMapping = map[string]any{
		"argocd":        &argocd.Spec{},
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
		"nomad":         &nomad.Spec{},
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

// todo fix all missing actionID
var discoveryNewFuncMap = map[string]discoveryFunc{
	"argocd": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return argocd.New(spec, rootDir, scmID, actionID)
	},
	"cargo": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return cargo.New(spec, rootDir, scmID, actionID)
	},
	"dockercompose": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return dockercompose.New(spec, rootDir, scmID, actionID)
	},
	"dockerfile": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return dockerfile.New(spec, rootDir, scmID, actionID)
	},
	"flux": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return flux.New(spec, rootDir, scmID, actionID)
	},
	"github/action": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return githubaction.New(spec, rootDir, scmID, actionID)
	},
	"gitea/action": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return githubaction.New(spec, rootDir, scmID, actionID)
	},
	"golang/gomod": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return golang.New(spec, rootDir, scmID, actionID)
	},
	"helm": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return helm.New(spec, rootDir, scmID, actionID)
	},
	"helmfile": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return helmfile.New(spec, rootDir, scmID, actionID)
	},
	"ko": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return ko.New(spec, rootDir, scmID, actionID)
	},
	"kubernetes": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return kubernetes.New(spec, rootDir, scmID, actionID, kubernetes.FlavorKubernetes)
	},
	"maven": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return maven.New(spec, rootDir, scmID, actionID)
	},
	"npm": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return npm.New(spec, rootDir, scmID, actionID)
	},
	"precommit": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return precommit.New(spec, rootDir, scmID, actionID)
	},
	"prow": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return kubernetes.New(spec, rootDir, scmID, actionID, kubernetes.FlavorProw)
	},
	"rancher/fleet": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return fleet.New(spec, rootDir, scmID, actionID)
	},
	"terraform": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return terraform.New(spec, rootDir, scmID, actionID)
	},
	"terragrunt": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return terragrunt.New(spec, rootDir, scmID, actionID)
	},
	"updatecli": func(spec any, rootDir string, scmID string, actionID string) (Crawler, error) {
		return updatecli.New(spec, rootDir, scmID, actionID)
	},
}

type discoveryFunc func(spec any, rootDir string, scmID string, actionID string) (Crawler, error)

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
		if f, ok := discoveryNewFuncMap[kind]; ok {
			crawler, err := f(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}
			g.crawlers = append(g.crawlers, crawler)
		} else {
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
