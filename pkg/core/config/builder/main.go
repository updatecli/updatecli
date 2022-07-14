package builder

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/generics/helm"
)

var (
	//// defaultGenericsSpecs defines the default builder that we want to run
	//defaultGenericSpecs = map[string]interface{}{
	//	"helm": helm.Spec{},
	//}

	defaultSpecs = Spec{
		genericSpecs: map[string]interface{}{
			"helm": helm.Spec{},
		},
		Scm: &scm.Config{},
	}
)

type Builder interface {
	Manifests(scmSpec *scm.Config) ([]config.Spec, error)
}

type Spec struct {
	genericSpecs map[string]interface{}
	Scm          *scm.Config
}

type Generator struct {
	spec Spec
	//helm     helm.Helm
	builders []Builder
}

//
func New(spec Spec) (*Generator, error) {

	var errs []error
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return &Generator{}, err
	}

	logrus.Info("SCM: %v", s.Scm)

	if len(s.genericSpecs) == 0 {
		s.genericSpecs = defaultSpecs.genericSpecs
	}

	//if *s.Scm == (scm.Config{}) {
	//	s.Scm = nil
	//}

	g := Generator{}

	for kind, genericSpec := range s.genericSpecs {
		logrus.Infof("%s", kind)

		switch kind {
		case "helm":
			// Init Helm generator
			helm, err := helm.New(genericSpec)

			if err != nil {
				errs = append(errs, fmt.Errorf("%s - %s", kind, err))
				continue
			}

			g.builders = append(g.builders, helm)

		default:
			logrus.Info("Builder of type %q not supported", kind)
		}
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Info(errs[i])
		}
	}

	return &g, nil
}

func (g *Generator) Run() ([]config.Spec, error) {
	var totalManifests []config.Spec

	for i := range g.builders {
		builderManifests, err := g.builders[i].Manifests(g.spec.Scm)

		if err != nil {
			logrus.Errorln(err)
		}

		if len(builderManifests) > 0 {
			for i := range builderManifests {
				totalManifests = append(totalManifests, builderManifests[i])
			}
		}
	}

	return totalManifests, nil

}
