package autodiscovery

import (
	"fmt"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/cargo"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockercompose"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/fleet"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helm"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helmfile"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/maven"
)

var (
	// DefaultGenericsSpecs defines the default builder that we want to run
	DefaultCrawlerSpecs = Config{
		Crawlers: map[string]interface{}{
			"cargo":         cargo.Spec{},
			"dockercompose": dockercompose.Spec{},
			"dockerfile":    dockerfile.Spec{},
			"helm":          helm.Spec{},
			"helmfile":      helmfile.Spec{},
			"maven":         maven.Spec{},
			"rancher/fleet": fleet.Spec{},
		},
	}
	// AutodiscoverySpecs is a map of all Autodiscovery specification
	AutodiscoverySpecsMapping = map[string]interface{}{
		"cargo":         &cargo.Spec{},
		"dockercompose": &dockercompose.Spec{},
		"dockerfile":    &dockerfile.Spec{},
		"helm":          &helm.Spec{},
		"helmfile":      &helmfile.Spec{},
		"maven":         &maven.Spec{},
		"rancher/fleet": &fleet.Spec{},
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

	for kind := range s.Crawlers {

		if workDir == "" {
			logrus.Errorf("skipping crawler %q due to: %s", kind, err)
			continue
		}

		// Commenting for now while refactoring
		switch kind {
		case "cargo":

			cargoCrawler, err := cargo.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, cargoCrawler)

		case "dockercompose":

			dockerComposeCrawler, err := dockercompose.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, dockerComposeCrawler)

		case "dockerfile":

			dockerfileCrawler, err := dockerfile.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, dockerfileCrawler)

		case "helm":

			helmCrawler, err := helm.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, helmCrawler)

		case "helmfile":

			helmfileCrawler, err := helmfile.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, helmfileCrawler)

		case "maven":
			mavenCrawler, err := maven.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.crawlers = append(g.crawlers, mavenCrawler)

		case "rancher/fleet":
			fleetCrawler, err := fleet.New(
				g.spec.Crawlers[kind],
				workDir,
				g.spec.ScmId)

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

			for i := range discoveredManifests {
				totalDiscoveredManifests = append(totalDiscoveredManifests, discoveredManifests[i])
			}
		}
	}

	logrus.Printf("\n\n---\n\n=> Total manifest detected: %d\n\n", len(totalDiscoveredManifests))

	return totalDiscoveredManifests, nil
}
