package gomod

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/go/language"
	gomodule "github.com/updatecli/updatecli/pkg/plugins/resources/go/module"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Changelog returns a link to the Golang version
func (g *GoMod) Changelog() string {

	switch g.kind {
	case kindGolang:
		l := language.Language{
			Version: version.Version{
				OriginalVersion: g.foundVersion,
				ParsedVersion:   g.foundVersion,
			},
		}
		return l.Changelog()
	case kindModule:
		gomodule := gomodule.GoModule{
			Spec: gomodule.Spec{
				Module: g.spec.Module,
			},
			Version: version.Version{
				OriginalVersion: g.foundVersion,
				ParsedVersion:   g.foundVersion,
			},
		}
		return gomodule.Changelog()
	}

	return ""
}
