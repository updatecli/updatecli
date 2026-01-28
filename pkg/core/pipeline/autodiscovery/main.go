package autodiscovery

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
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
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/plugin"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/precommit"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terraform"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terragrunt"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/updatecli"
)

// GetDefaultCrawlerSpecs return config that defines the default builder that we want to run
var GetDefaultCrawlerSpecs = sync.OnceValue(func() Config {
	ret := Config{
		Crawlers: make(CrawlersConfig, len(crawlerMap)),
	}
	for k, v := range crawlerMap {
		if v.ignoreDefault {
			continue
		}
		ret.Crawlers[k] = v.spec
	}
	return ret
})

// GetAutodiscoverySpecs return a map of all Autodiscovery specification
var GetAutodiscoverySpecsMapping = sync.OnceValue(func() CrawlersConfig {
	ret := make(CrawlersConfig, len(crawlerMap))
	for k, v := range crawlerMap {
		ret[k] = v.spec
	}
	return ret
})

type Crawler interface {
	DiscoverManifests() ([][]byte, error)
}

type AutoDiscovery struct {
	spec     Config
	crawlers []Crawler
}

// ensure alias not conflict with existing key
func init() {
	aliases := make(map[string]string)
	for k, v := range crawlerMap {
		for _, alias := range v.alias {
			if _, ok := crawlerMap[alias]; ok {
				panic(fmt.Sprintf("alias(%q) for %q conflicts with existing crawler", alias, k))
			}

			if old, ok := aliases[alias]; ok {
				panic(fmt.Sprintf("alias(%q) for %q already exists on %q ", alias, k, old))
			}
			aliases[alias] = k
		}
	}
}

// return crawlerFuncMap map with alias
var crawlerFuncMap = sync.OnceValue(func() map[string]crawlerFunc {
	ret := make(map[string]crawlerFunc, len(crawlerMap))
	for k, v := range crawlerMap {
		ret[k] = v.newFunc
		// add alias
		for _, alias := range v.alias {
			ret[alias] = v.newFunc
		}
	}
	return ret
})

type crawlerFunc = func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error)

// crawlerMap is a map of all the crawlers and spec
var crawlerMap = map[string]struct {
	newFunc crawlerFunc
	spec    any
	// ignore in GetDefaultCrawlerSpecs
	ignoreDefault bool
	// alias is a list than can be used to refer to the crawler
	// like `go` `golang` `golang/gomod`
	// alias name should not be used as key
	alias []string
}{
	"argocd": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return argocd.New(spec, rootDir, scmID, actionID)
		},
		spec: argocd.Spec{},
	},
	"cargo": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return cargo.New(spec, rootDir, scmID, actionID)
		},
		spec: cargo.Spec{},
	},
	"dockercompose": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return dockercompose.New(spec, rootDir, scmID, actionID)
		},
		spec: dockercompose.Spec{},
	},
	"dockerfile": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return dockerfile.New(spec, rootDir, scmID, actionID)
		},
		spec: dockerfile.Spec{},
	},
	"flux": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return flux.New(spec, rootDir, scmID, actionID)
		},
		spec: flux.Spec{},
	},

	"github/action": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return githubaction.New(spec, rootDir, scmID, actionID)
		},
		spec: githubaction.Spec{},
	},
	"gitea/action": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return githubaction.New(spec, rootDir, scmID, actionID)
		},
		spec: githubaction.Spec{},
		// gitea/action share the same behavior as github/action
		// so we use the last one.
		// The day we have a specific behavior for gitea/action
		// then we will add it to the default autodiscovery.
		ignoreDefault: true,
	},
	"golang": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return golang.New(spec, rootDir, scmID, actionID)
		},
		spec:  golang.Spec{},
		alias: []string{"go", "golang/gomod"},
	},
	"helm": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return helm.New(spec, rootDir, scmID, actionID)
		},
		spec: helm.Spec{},
	},
	"helmfile": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return helmfile.New(spec, rootDir, scmID, actionID)
		},
		spec: helmfile.Spec{},
	},
	"ko": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return ko.New(spec, rootDir, scmID, actionID)
		},
		spec: ko.Spec{},
	},
	"kubernetes": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return kubernetes.New(spec, rootDir, scmID, actionID, kubernetes.FlavorKubernetes)
		},
		spec: kubernetes.Spec{},
	},
	"maven": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return maven.New(spec, rootDir, scmID, actionID)
		},
		spec: maven.Spec{},
	},
	"nomad": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return nomad.New(spec, rootDir, scmID, actionID)
		},
		spec: nomad.Spec{},
	},

	"npm": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return npm.New(spec, rootDir, scmID, actionID)
		},
		spec: npm.Spec{},
	},
	"precommit": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return precommit.New(spec, rootDir, scmID, actionID)
		},
		spec: precommit.Spec{},
	},
	"prow": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return kubernetes.New(spec, rootDir, scmID, actionID, kubernetes.FlavorProw)
		},
		spec: kubernetes.Spec{},
	},
	"plugin": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return plugin.New(spec, rootDir, scmID, actionID, pluginName)
		},
		ignoreDefault: true,
		spec:          plugin.Spec{},
	},

	"rancher/fleet": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return fleet.New(spec, rootDir, scmID, actionID)
		},
		spec: fleet.Spec{},
	},
	"terraform": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return terraform.New(spec, rootDir, scmID, actionID)
		},
		spec: terraform.Spec{},
	},
	"terragrunt": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return terragrunt.New(spec, rootDir, scmID, actionID)
		},
		spec: terragrunt.Spec{},
	},
	"updatecli": {
		newFunc: func(spec any, rootDir string, scmID string, actionID, pluginName string) (Crawler, error) {
			return updatecli.New(spec, rootDir, scmID, actionID)
		},
		spec: updatecli.Spec{},
	},
}

// New returns an initiated autodiscovery object
//
// This function is responsible to create all the crawlers
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

		pluginName := kind
		if strings.HasSuffix(kind, ".wasm") {
			if !cmdoptions.Experimental {
				logrus.Warningf("Skipping autodiscovery plugin %q because experimental features are disabled", kind)
				continue
			}
			kind = "plugin"
		}

		if workDir == "" {
			logrus.Errorf("skipping crawler %q due to: %s", kind, err)
			continue
		}
		if f, ok := crawlerFuncMap()[kind]; ok {
			crawler, err := f(
				g.spec.Crawlers[pluginName],
				workDir,
				g.spec.ScmId,
				g.spec.ActionId,
				pluginName,
			)
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
			logrus.Errorln(errs[i])
		}
		return &g, fmt.Errorf("autodiscovery failed with %d error(s)", len(errs))
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
