package gomodule

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a go module with a specific version is published
func (g *GoModule) Condition(ctx context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	versionToCheck := g.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, "", fmt.Errorf("no version defined")
	}

	var GOPROXY string
	if g.Spec.Proxy != "" {
		GOPROXY = g.Spec.Proxy
	} else if os.Getenv("GOPROXY") != "" {
		GOPROXY = os.Getenv("GOPROXY")
	} else {
		GOPROXY = goModuleDefaultProxy
	}

	for _, proxy := range strings.Split(GOPROXY, ",") {
		proxy = strings.TrimSpace(proxy)
		if !isSupportedGoProxy(proxy) {
			continue
		}
		version, err := getVersionInfoFromProxy(ctx, g.webClient, proxy, g.Spec.Module, versionToCheck)
		if err != nil {
			logrus.Debugf("skipping proxy %q due to %q\n", proxy, err)
		}

		if version != "" {
			return true, fmt.Sprintf("version %q available", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("version %q doesn't exist", versionToCheck), nil
}
