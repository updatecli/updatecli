package maven

import (
	"encoding/xml"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
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
		logrus.Infof("\u2714 Latest version is %s on Maven Repository", data.Versioning.Latest)
		return data.Versioning.Latest, nil
	}

	logrus.Infof("\u2717 No latest version on Maven Repository")
	return "", nil
}
