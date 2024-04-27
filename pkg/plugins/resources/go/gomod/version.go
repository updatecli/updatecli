package gomod

import (
	"errors"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

var (
	ErrModuleNotFound error = errors.New("GO module not found")
)

// version retrieve the version specified by a GO module
func (g *GoMod) version(filename string) (string, error) {

	// Test at runtime if a file exist
	if !g.contentRetriever.FileExists(filename) {
		return "", fmt.Errorf("file %q does not exist", filename)
	}

	if err := g.Read(filename); err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	data := g.currentContent

	modfile, err := modfile.Parse(filename, []byte(data), nil)
	if err != nil {
		logrus.Errorln(err)
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

// setVersion update a go.mod file with the version specified by a GO module
func (g *GoMod) setVersion(version, filename string, dryrun bool) (oldVersion, newVersion string, changed bool, err error) {

	oldContent, err := os.ReadFile(filename)

	if err != nil {
		logrus.Errorln(err)
		return "", "", false, fmt.Errorf("failed reading %q", filename)
	}

	modFile, err := modfile.Parse(filename, oldContent, nil)
	if err != nil {
		logrus.Errorln(err)
		return "", "", false, fmt.Errorf("failed reading %q", filename)
	}

	switch g.kind {
	case kindGolang:
		s, err := semver.NewVersion(version)
		if err != nil {
			logrus.Errorln(err)
			return "", "", false, fmt.Errorf("failed parsing go version %q", version)
		}
		oldVersion = modFile.Go.Version
		newVersion = fmt.Sprintf("%d.%d", s.Major(), s.Minor())

		if oldVersion != newVersion {
			err = modFile.AddGoStmt(newVersion)
			if err != nil {
				logrus.Errorln(err)
				return "", "", false, fmt.Errorf("failed updating go version %q\n%w", version, err)
			}

			changed = true
		}

	case kindModule:
		moduleFound := false
	out:
		for _, r := range modFile.Require {
			if r.Indirect != g.spec.Indirect {
				continue
			}
			if r.Mod.Path == g.spec.Module {
				moduleFound = true
				oldVersion = r.Mod.Version
				newVersion = version
				if newVersion != oldVersion {

					err = modFile.AddRequire(r.Mod.Path, version)
					if err != nil {
						logrus.Errorln(err)
						return "", "", false, fmt.Errorf("failed updating go module %q to %q\n%w", g.spec.Module, version, err)
					}

					changed = true
					break out
				}
			}
		}
		if !moduleFound {
			err := fmt.Errorf("module %q not found in file %q", g.spec.Module, filename)
			return "", "", false, err

		}
	default:
		logrus.Errorf("kind %q is not supported", g.kind)
		return "", "", false, fmt.Errorf("something unexpected happened, kind %q not supported", g.kind)
	}

	modFile.Cleanup()
	newContent, err := modFile.Format()

	if err != nil {
		logrus.Errorln(err)
		return oldVersion, newVersion, changed, fmt.Errorf("failed formatting %q", filename)
	}

	edits := myers.ComputeEdits(span.URIFromPath(filename), string(oldContent), string(newContent))
	logrus.Debugf("\n---\n%v\n---\n", gotextdiff.ToUnified("old", "new", string(oldContent), edits))

	if !changed || dryrun {
		return oldVersion, newVersion, changed, nil
	}

	f, err := os.Create(filename)
	if err != nil {
		logrus.Errorln(err)
		return oldVersion, newVersion, changed, fmt.Errorf("failed opening file %q", filename)
	}
	defer f.Close()

	_, err = f.Write(newContent)
	if err != nil {
		logrus.Errorln(err)
		return oldVersion, newVersion, changed, fmt.Errorf("failed writing data to %q", filename)
	}

	logrus.Debugf("%q updated\n", filename)

	return oldVersion, newVersion, changed, nil
}
