package gomodule

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetVersions fetch all versions of a Golang module
func (g *GoModule) versions() (v string, versions []string, err error) {

	var GOPROXY string
	if g.Spec.Proxy != "" {
		GOPROXY = g.Spec.Proxy
	} else if os.Getenv("GOPROXY") != "" {
		GOPROXY = os.Getenv("GOPROXY")
	} else {
		GOPROXY = goModuleDefaultProxy
	}

	for _, proxy := range strings.Split(GOPROXY, ",") {
		if !isSupportedGoProxy(proxy) {
			continue
		}

		URL, err := url.JoinPath(
			sanitizeGoProxy(proxy),
			sanitizeGoModuleNameForProxy(g.Spec.Module),
			"@v", "list")
		if err != nil {
			logrus.Errorf("something went wrong while getting go module api data %q\n", err)
			return "", []string{}, err
		}

		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			logrus.Errorf("something went wrong while getting go module api data %q\n", err)
			return "", []string{}, err
		}

		res, err := g.webClient.Do(req)
		if err != nil {
			logrus.Errorf("something went wrong while getting go module api data %q\n", err)
			return "", []string{}, err
		}

		defer res.Body.Close()
		if res.StatusCode >= 400 {
			body, err := httputil.DumpResponse(res, false)
			logrus.Errorf("something went wrong while getting golang module data %q\n", err)
			logrus.Debugf("skipping proxy %q due to %q\n", proxy, err)
			logrus.Debugf("\n%v\n", string(body))
			continue
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			logrus.Errorf("something went wrong while getting npm api data%q\n", err)
			return "", []string{}, err
		}

		/*
			The response should be a list of version separated by \n
			as explained on https://go.dev/ref/mod#goproxy-protocol
		*/
		versions = append(versions, strings.Split(string(data), "\n")...)

		sort.Strings(versions)
		g.Version, err = g.versionFilter.Search(versions)
		if err != nil {
			return "", nil, err
		}

		return g.Version.GetVersion(), versions, nil

	}

	return "", nil, fmt.Errorf("GO module %q not found on proxy %q", g.Spec.Module, GOPROXY)
}
