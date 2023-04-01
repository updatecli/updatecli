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

	data, err := os.ReadFile(g.filename)

	if err != nil {
		return "", fmt.Errorf("failed reading %q", g.filename)
	}

	modfile, err := modfile.Parse(g.filename, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed reading %q", g.filename)
	}

	for _, r := range modfile.Require {
		if r.Indirect != g.spec.Indirect {
			continue
		}
		if r.Mod.Path == g.spec.Module {
			return r.Mod.Version, nil
		}
	}

	logrus.Errorf("GO module %q not found in %q", g.spec.Module, g.filename)
	return "", ErrModuleNotFound

}

//// updateVersion update a version if needed
//func (g *GoMod) updateVersion(version string) (bool, error) {
//
//	data, err := os.ReadFile(g.filename)
//
//	if err != nil {
//		return false, fmt.Errorf("failed reading %q", g.filename)
//	}
//
//	modfile, err := modfile.Parse(g.filename, data, nil)
//	if err != nil {
//		return false, fmt.Errorf("failed reading %q", g.filename)
//	}
//
//	foundVersion := ""
//	foundModule := false
//	for _, r := range modfile.Require {
//		if r.Indirect != g.spec.Indirect {
//			continue
//		}
//		if r.Mod.Path == g.spec.ModulePath {
//			foundVersion = r.Mod.Version
//			foundModule = true
//			break
//		}
//	}
//
//	if foundVersion == version {
//		logrus.Infof("%s Module %q already set to %q",
//			result.SUCCESS, g.spec.ModulePath, version)
//		return false, nil
//	}
//
//	if !foundModule {
//		logrus.Infof("%s Module %q not found, skipping",
//			result.FAILURE, g.spec.ModulePath)
//		return false, ErrModuleNotFound
//	}
//
//	modfile.AddRequire(g.spec.ModulePath, version)
//	modfile.Cleanup()
//	newData, err := modfile.Format()
//
//	if err != nil {
//		return false, fmt.Errorf("something went wrong while formatting %q - %w", g.filename, err)
//	}
//	os.WriteFile(g.filename, newData, 0644)
//
//	return true, nil
//}
//
