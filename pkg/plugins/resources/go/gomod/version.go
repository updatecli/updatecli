package gomod

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

var (
	ErrModuleNotFound error = errors.New("GO module not found")

	majorMinorRegex      *regexp.Regexp = regexp.MustCompile(`^\d+\.\d+$`)
	majorMinorPatchRegex *regexp.Regexp = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
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
		return "", fmt.Errorf("failed reading %q: %w", filename, err)
	}

	switch g.kind {
	case kindGolang:
		return modfile.Go.Version, nil

	case kindModule:
		if g.spec.Replace {
			for _, r := range modfile.Replace {
				if g.spec.ReplaceVersion != "" && r.Old.Version != g.spec.ReplaceVersion {
					continue
				}
				if r.Old.Path == g.spec.Module {
					return r.New.Version, nil
				}
			}
		} else {
			for _, r := range modfile.Require {
				if r.Indirect != g.spec.Indirect {
					continue
				}
				if r.Mod.Path == g.spec.Module {
					return r.Mod.Version, nil
				}
			}
		}

		return "", ErrModuleNotFound
	}

	return "", errors.New("something unexpected happened in go modfile")
}

// setVersion update a go.mod file with the version specified by a GO module
func (g *GoMod) setVersion(version, filename string, dryrun bool) (oldVersion, newVersion string, changed bool, err error) {

	oldContent, err := os.ReadFile(filename)

	if err != nil {
		return "", "", false, fmt.Errorf("failed reading %q: %w", filename, err)
	}

	modFile, err := modfile.Parse(filename, oldContent, nil)
	if err != nil {
		return "", "", false, fmt.Errorf("failed reading %q: %w", filename, err)
	}

	switch g.kind {
	case kindGolang:

		oldVersion = modFile.Go.Version
		newVersion, err = getNewVersion(oldVersion, version)
		if err != nil {
			return "", "", false, fmt.Errorf("failed parsing go version %q: %w", version, err)
		}

		if oldVersion != newVersion {
			err = modFile.AddGoStmt(newVersion)
			if err != nil {
				return "", "", false, fmt.Errorf("failed updating go version %q: %w", version, err)
			}

			changed = true
		}

	case kindModule:
		moduleFound := false
		if g.spec.Replace {
		outReplace:
			for _, r := range modFile.Replace {
				if r.Old.Path == g.spec.Module {

					if g.spec.ReplaceVersion != "" && r.Old.Version != g.spec.ReplaceVersion {
						continue
					}

					moduleFound = true
					oldVersion = r.New.Version
					newVersion = version

					if newVersion != oldVersion {
						err = modFile.AddReplace(r.Old.Path, r.Old.Version, r.New.Path, version)
						if err != nil {
							return "", "", false, fmt.Errorf("failed updating go module replacer %q to %q\n%w", g.spec.Module, version, err)
						}
						changed = true
						break outReplace
					}
				}
			}
		} else {
		outRequire:
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
							return "", "", false, fmt.Errorf("failed updating go module %q to %q\n%w", g.spec.Module, version, err)
						}

						changed = true
						break outRequire
					}
				}
			}
		}

		if !moduleFound {
			err := fmt.Errorf("module %q not found in file %q", g.spec.Module, filename)
			return "", "", false, err
		}

	default:
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
		return oldVersion, newVersion, changed, fmt.Errorf("failed opening file %q: %w", filename, err)
	}
	defer f.Close()

	_, err = f.Write(newContent)
	if err != nil {
		return oldVersion, newVersion, changed, fmt.Errorf("failed writing data to %q: %w", filename, err)
	}

	logrus.Debugf("%q updated\n", filename)

	return oldVersion, newVersion, changed, nil
}

// getNewVersion returns the new version of a GO module.
// It tries to detect if the version is a major.minor or major.minor.patch
func getNewVersion(oldVersion, newVersion string) (string, error) {

	s, err := semver.NewVersion(newVersion)
	if err != nil {
		return "", fmt.Errorf("failed parsing go version %q: %w", oldVersion, err)
	}

	if majorMinorRegex.MatchString(oldVersion) {
		return fmt.Sprintf("%d.%d", s.Major(), s.Minor()), nil
	}

	if majorMinorPatchRegex.MatchString(oldVersion) {
		return fmt.Sprintf("%d.%d.%d", s.Major(), s.Minor(), s.Patch()), nil
	}

	return newVersion, nil
}
