package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(source string) (bool, error) {
	if m.spec.Version != "" {
		logrus.Infof("Version %v, already defined from configuration file", m.spec.Version)
	} else {
		m.spec.Version = source
	}

	req, err := http.NewRequest("GET", m.RepositoryURL, nil)
	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	data := Metadata{}

	err = xml.Unmarshal(body, &data)
	if err != nil {
		return false, err
	}

	for _, version := range data.Versioning.Versions.Version {
		if version == m.spec.Version {
			logrus.Infof("%s Version %s is available on Maven Repository", result.SUCCESS, m.spec.Version)
			return true, nil
		}

	}

	logrus.Infof("%s Version %s is not available on Maven Repository", result.FAILURE, m.spec.Version)
	return false, nil
}

// ConditionFromSCM returns an error because it's not supported
func (m *Maven) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for maven condition, aborting")
}
