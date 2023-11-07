package engine

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

// Prepare run every actions needed before going further.
func (e *Engine) Prepare() (err error) {

	PrintTitle("Prepare")

	var defaultCrawlersEnabled bool

	err = tmp.Create()
	if err != nil {
		return err
	}

	err = e.LoadConfigurations()
	if !errors.Is(err, ErrNoManifestDetected) && err != nil {
		logrus.Errorln(err)
		logrus.Infof("\n%d pipeline(s) successfully loaded\n", len(e.Pipelines))
	}

	if errors.Is(err, ErrNoManifestDetected) {
		defaultCrawlersEnabled = true
	}

	// If one git clone fails then Updatecli exits
	// scm initialization must be done before autodiscovery as we need to identify
	// in advance git repository directories to analyze them for possible common update scenarii
	err = e.InitSCM()
	if err != nil {
		return err
	}

	err = e.LoadAutoDiscovery(defaultCrawlersEnabled)
	if err != nil {
		return err
	}

	if len(e.Pipelines) == 0 {
		return fmt.Errorf("no valid pipeline found")
	}

	return nil
}
