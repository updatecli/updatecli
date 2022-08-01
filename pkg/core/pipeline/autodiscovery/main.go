package autodiscovery

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helm"
)

var (
	// DefaultGenericsSpecs defines the default builder that we want to run
	DefaultCrawlerSpecs = discoveryConfig.Config{
		Crawlers: map[string]interface{}{
			"helm": helm.Spec{},
		},
	}
)

type Options struct {
	Disabled bool
}
type Crawler interface {
	DiscoverManifests(scmSpec *scm.Config) ([]config.Spec, error)
	Enabled() bool
}

type AutoDiscovery struct {
	scmConfig *scm.Config
	spec      discoveryConfig.Config
	crawlers  []Crawler
}

//
func New(spec discoveryConfig.Config,
	scmHandler scm.ScmHandler,
	scmConfig *scm.Config) (*AutoDiscovery, error) {

	var errs []error
	var s discoveryConfig.Config

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return &AutoDiscovery{}, err
	}

	tmpCrawlers := DefaultCrawlerSpecs.Crawlers

	for id, crawlerSpec := range s.Crawlers {
		tmpCrawlers[id] = crawlerSpec
	}

	s.Crawlers = tmpCrawlers

	g := AutoDiscovery{
		spec: s,
	}

	// Init scm configuration if one is specified
	if len(scmConfig.Kind) > 0 {
		g.scmConfig = scmConfig
	}

	for kind := range DefaultCrawlerSpecs.Crawlers {

		// Init workDir based on process running directory
		workDir, err := os.Getwd()
		if err != nil {
			logrus.Errorf("skipping crawler %q due to: %s", kind, err)
			continue
		}

		// Retrieve the scm workdir if it exist
		// As long as the autodiscovery specifies one
		if _, ok := g.spec.Crawlers[kind]; ok && scmHandler != nil {
			workDir = scmHandler.GetDirectory()
		}

		switch kind {
		case "helm":

			helmCrawler, err := helm.New(g.spec.Crawlers[kind], workDir)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, helmCrawler)

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

func (g *AutoDiscovery) Run() ([]config.Spec, error) {
	var totalDiscoveredManifests []config.Spec

	for id, crawler := range g.crawlers {
		if !crawler.Enabled() {
			logrus.Infof("Manifest autodiscovering is disabled for %q", id)
			continue
		}

		discoveredManifests, err := crawler.DiscoverManifests(g.scmConfig)

		if err != nil {
			logrus.Errorln(err)
		}

		if len(discoveredManifests) > 0 {
			for i := range discoveredManifests {
				totalDiscoveredManifests = append(totalDiscoveredManifests, discoveredManifests[i])
			}
		}
	}

	return totalDiscoveredManifests, nil

}
