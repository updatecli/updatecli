package language

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	newMinorVersion = regexp.MustCompile(`^(\d)*.(\d)*$`)
)

type Info struct {
	Stable  bool
	Version string
}

// versions fetch all stable Golang version
func (l *Language) versions() (versions []string, err error) {

	if err != nil {
		logrus.Errorf("something went wrong while generating the go url to retrieve versions %q\n", err)
		return []string{}, err
	}

	req, err := http.NewRequest("GET", "https://go.dev/dl/?mode=json&include=all", nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting go version data %q\n", err)
		return []string{}, err
	}

	res, err := l.webClient.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting go version data %q\n", err)
		return []string{}, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Errorf("something went wrong while getting golang version data %q\n", err)
		logrus.Debugf("\n%v\n", string(body))
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return []string{}, err
	}

	versionsInfo := []Info{}

	err = json.Unmarshal(data, &versionsInfo)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %q", err)
		return []string{}, err
	}

	for _, v := range versionsInfo {
		if v.Stable {
			version := strings.TrimPrefix(v.Version, "go")
			if newMinorVersion.MatchString(version) {
				version = version + ".0"
			}

			versions = append(versions, version)
		}
	}

	sort.Strings(versions)
	l.foundVersion, err = l.versionFilter.Search(versions)
	if err != nil {
		return nil, err
	}

	return versions, nil

}
