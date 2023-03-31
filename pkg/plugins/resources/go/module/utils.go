package gomodule

import (
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
)

// sanitizeGoModuleNameForProxy is used to lowercase any uppercase character with a ! prefix as explained on https://go.dev/ref/mod#goproxy-protocol
func sanitizeGoModuleNameForProxy(module string) string {
	var result string
	for _, r := range module {
		switch !unicode.IsLower(r) && unicode.IsLetter(r) {
		case true:
			result += "!" + strings.ToLower(string(r))
		case false:
			result += string(r)
		}
	}
	return result
}

func isSupportedGoProxy(proxy string) bool {
	if proxy == "direct" || proxy == "off" {
		logrus.Debugf("proxy %q has no meaning from an Updatecli stand point", proxy)
		return false
	}
	if strings.HasPrefix(proxy, "file://") {
		logrus.Debugln("updatecli do not support proxy using file protocol at this time. Feel free to open a pullrequest")
		return false
	}

	return true
}

func sanitizeGoProxy(proxy string) string {
	if strings.HasPrefix(proxy, "https://") || strings.HasPrefix(proxy, "http://") {
		return proxy
	}
	return "https://" + proxy
}
