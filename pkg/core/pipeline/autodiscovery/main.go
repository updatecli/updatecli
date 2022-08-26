package autodiscovery

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/fleet"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helm"
)

var (
	// DefaultGenericsSpecs defines the default builder that we want to run
	DefaultCrawlerSpecs = discoveryConfig.Config{
		Crawlers: map[string]interface{}{
			"helm":          helm.Spec{},
			"rancher/fleet": fleet.Spec{},
		},
	}
)

type Options struct {
	Enabled bool
}
type Crawler interface {
	DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error)
	Enabled() bool
}

type AutoDiscovery struct {
	scmConfig         *scm.Config
	pullrequestConfig *pullrequest.Config
	spec              discoveryConfig.Config
	crawlers          []Crawler
}

//
func New(spec discoveryConfig.Config,
	scmHandler scm.ScmHandler,
	scmConfig *scm.Config,
	pullrequestConfig *pullrequest.Config) (*AutoDiscovery, error) {

	var errs []error
	var s discoveryConfig.Config

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return &AutoDiscovery{}, err
	}

	if len(spec.GroupBy) == 0 {
		spec.GroupBy.Validate()
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

	if len(pullrequestConfig.Kind) > 0 {
		g.pullrequestConfig = pullrequestConfig
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

		case "rancher/fleet":
			fleetCrawler, err := fleet.New(g.spec.Crawlers[kind], workDir)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, fleetCrawler)

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

		discoveredManifests, err := crawler.DiscoverManifests(
			discoveryConfig.Input{
				ScmSpec:         g.scmConfig,
				ScmID:           g.spec.ScmId,
				PullRequestSpec: g.pullrequestConfig,
				PullrequestID:   g.spec.PullrequestId,
			})

		if err != nil {
			logrus.Errorln(err)
		}

		logrus.Printf("Manifest detected: %d\n", len(discoveredManifests))

		if len(discoveredManifests) > 0 {

			for i := range discoveredManifests {
				totalDiscoveredManifests = append(totalDiscoveredManifests, discoveredManifests[i])
			}
		}
	}

	logrus.Printf("\n\n---\n\n=> Total manifest detected: %d\n\n", len(totalDiscoveredManifests))

	if g.spec.GroupBy == "" {
		logrus.Warningf("Autodiscovery settings %q is undefined, fallback to %q",
			"groupby",
			discoveryConfig.GROUPEBYINDIVIDUAL)
	}

	// Set pipelineId for each manifest by on the autodiscovery groupby rule

	// We use a sha256 hash to avoid colusion between pipelineID
	hash := sha256.New()
	batchPipelineID := fmt.Sprintf("%x", hash.Sum([]byte("updatecli/autodiscovery/batch")))

	for i := range totalDiscoveredManifests {
		switch g.spec.GroupBy {
		case discoveryConfig.GROUPEBYALL:
			totalDiscoveredManifests[i].PipelineID = batchPipelineID[0:32]

		case discoveryConfig.GROUPEBYINDIVIDUAL, "":
			pipelineID := fmt.Sprintf("%x", hash.Sum([]byte(totalDiscoveredManifests[i].Name)))

			totalDiscoveredManifests[i].PipelineID = pipelineID[0:32]

		default:
			logrus.Errorln("something unexpected happened while specifying pipelineid to generated Updatecli manifest")
		}
	}

	return totalDiscoveredManifests, nil

}
