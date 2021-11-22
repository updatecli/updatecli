package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest version
func (m *Maven) Source(workingDir string) (string, error) {
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	data := Metadata{}

	err = xml.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	if data.Versioning.Latest != "" {
		logrus.Infof("%s Latest version is %s on Maven Repository", result.SUCCESS, data.Versioning.Latest)
		return data.Versioning.Latest, nil
	}

	logrus.Infof("%s No latest version on Maven Repository", result.FAILURE)
	return "", nil
}
