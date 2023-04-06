package gomod

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

var (
	ErrModuleNotFound error = errors.New("GO module not found")
)

// version retrieve the version specified by a GO module
func (g *GoMod) version(filename string) (string, error) {

	data, err := os.ReadFile(filename)

	if err != nil {
		return "", fmt.Errorf("failed reading %q", filename)
	}

	modfile, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed reading %q", filename)
	}

	switch g.kind {
	case kindGolang:
		return modfile.Go.Version, nil

	case kindModule:
		for _, r := range modfile.Require {
			if r.Indirect != g.spec.Indirect {
				continue
			}
			if r.Mod.Path == g.spec.Module {
				return r.Mod.Version, nil
			}
		}
		logrus.Errorf("GO module %q not found in %q", g.spec.Module, filename)
		return "", ErrModuleNotFound
	}

	return "", errors.New("something unexpected happened in go modfile")
}
