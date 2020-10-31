package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Source return the latest version
func (m *Maven) Source() (string, error) {
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

	xml.Unmarshal(body, &data)

	if data.Versioning.Latest != "" {
		fmt.Printf("\u2714 Latest version is %s on Maven Repository\n", data.Versioning.Latest)
		return data.Versioning.Latest, nil
	}

	fmt.Printf("\u2717 No latest version on Maven Repository\n")
	return "", nil
}
