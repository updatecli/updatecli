package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
)

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(source string) (bool, error) {
	if m.Version != "" {
		logrus.Infof("Version %v, already defined from configuration file", m.Version)
	} else {
		m.Version = source
	}
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)
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
		if version == m.Version {
			logrus.Infof("\u2713 Version %s is available on Maven Repository", m.Version)
			return true, nil
		}

	}

	logrus.Infof("\u2716 Version %s is not available on Maven Repository", m.Version)
	return false, nil
}

// ConditionFromSCM returns an error because it's not supported
func (m *Maven) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for maven condition, aborting")
}
