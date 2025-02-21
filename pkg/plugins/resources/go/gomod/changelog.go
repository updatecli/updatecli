package gomod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/go/language"
	gomodule "github.com/updatecli/updatecli/pkg/plugins/resources/go/module"
)

// Changelog returns a link to the Golang version
func (g *GoMod) Changelog(from, to string) *result.Changelogs {

	switch g.kind {
	case kindGolang:
		l, err := language.New(g.spec)
		if err != nil {
			logrus.Debugf("failed to init golang language: %s", err)
		}

		return l.Changelog(from, to)

	case kindModule:
		m, err := gomodule.New(g.spec)
		if err != nil {
			logrus.Debugf("failed to init golang module: %s", err)
		}

		return m.Changelog(from, to)

	default:
		fmt.Printf("Golang changelog of kind %q is not supported\n", g.kind)
	}

	return nil
}
