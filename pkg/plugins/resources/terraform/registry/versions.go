package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"sort"

	"github.com/sirupsen/logrus"
)

type versionResponse struct {
	Source   string   `json:"source"`
	Versions []string `json:"versions"`
}

func (t *TerraformRegistry) versions() (versions []string, err error) {
	req, err := http.NewRequest("GET", t.registryAddress.API(), nil)
	if err != nil {
		return nil, err
	}

	res, err := t.webClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Debugf("\n%v\n", string(body))
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	versionsInfo := versionResponse{}

	err = json.Unmarshal(data, &versionsInfo)
	if err != nil {
		return nil, err
	}

	versions = versionsInfo.Versions

	sort.Strings(versions)
	t.Version, err = t.versionFilter.Search(versions)
	if err != nil {
		return nil, err
	}

	t.scm = versionsInfo.Source

	return versions, nil
}
