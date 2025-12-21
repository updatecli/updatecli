package updatecli

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/compose"
)

var (
	// DefaultFiles specifies accepted Updatecli compose filename
	DefaultFiles []string = []string{
		compose.DeprecatedDefaultComposeFilename,
		compose.DefaultComposeFilename,
	}
)

// discoverUpdatecliPolicyManifests search recursively from a root directory for Updatecli compose file
func (u Updatecli) discoverUpdatecliPolicyManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := u.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if u.spec.RootDir != "" && !path.IsAbs(u.spec.RootDir) {
		searchFromDir = filepath.Join(u.rootDir, u.spec.RootDir)
	}

	foundUpdateComposeFiles, err := searchUpdatecliComposeFiles(
		searchFromDir,
		u.files[:])

	if err != nil {
		return nil, err
	}

	for _, foundUpdateComposeFile := range foundUpdateComposeFiles {
		logrus.Debugf("parsing file %q", foundUpdateComposeFile)

		relativeUpdateComposeFile, err := filepath.Rel(u.rootDir, foundUpdateComposeFile)
		if err != nil {
			// Jump to the next Update compose file if current failed
			logrus.Debugln(err)
			continue
		}

		updateComposeRelativeMetadataPath := filepath.Dir(relativeUpdateComposeFile)
		composeFilename := filepath.Base(updateComposeRelativeMetadataPath)

		metadata, err := getComposeFileMetadata(foundUpdateComposeFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if metadata == nil {
			continue
		}

		if len(metadata.Policies) == 0 {
			continue
		}

		for i, policy := range metadata.Policies {
			policyName, policyVersion, err := getPolicyName(policy.Policy)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			logrus.Debugf("policy name: %q detected", policyName)

			switch policyVersion {
			case "": // No version specified
				logrus.Debug("No version detected")
			case "latest":
				logrus.Debug("latest version detected")
			default:
				logrus.Debugf("version %q detected", policyVersion)
			}

			if len(u.spec.Ignore) > 0 {
				if u.spec.Ignore.isMatchingRules(u.rootDir, relativeUpdateComposeFile, policyName, policyVersion) {
					logrus.Debugf("Ignoring Updatecli policy %q from %q, as matching ignore rule(s)\n", policyName, composeFilename)
					continue
				}
			}

			if len(u.spec.Only) > 0 {
				if !u.spec.Only.isMatchingRules(u.rootDir, relativeUpdateComposeFile, policyName, policyVersion) {
					logrus.Debugf("Ignoring Updatecli policy %q from %q, as not matching only rule(s)\n", policyName, composeFilename)
					continue
				}
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			sourceVersionFilterKind := "semver"
			sourceVersionFilterPattern := "*"
			sourceVersionFilterRegex := "*"
			if !u.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = u.versionFilter.Kind
				if policyVersion != "latest" {
					sourceVersionFilterPattern, err = u.versionFilter.GreaterThanPattern(policyVersion)
					sourceVersionFilterRegex = u.versionFilter.Regex
					if err != nil {
						logrus.Debugf("building version filter pattern: %s", err)
						sourceVersionFilterPattern = "*"
					}
				}
			}

			params := struct {
				ActionID                   string
				ManifestName               string
				PolicyName                 string
				SourceVersionID            string
				SourceVersionName          string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				SourceVersionFilterRegex   string
				SourceDigestID             string
				SourceDigestName           string
				SourceDigestTag            string
				TargetName                 string
				TargetKey                  string
				File                       string
				ScmID                      string
			}{
				ActionID:                   u.actionID,
				ManifestName:               fmt.Sprintf("deps(updatecli/policy): bump %q Updatecli policy version", policyName),
				PolicyName:                 policyName,
				SourceVersionID:            "version",
				SourceVersionName:          fmt.Sprintf("Get latest %q Updatecli policy version", policyName),
				SourceDigestID:             "digest",
				SourceDigestName:           fmt.Sprintf("Get latest %q Updatecli policy digest", policyName),
				SourceDigestTag:            "{{ source \"version\" }}",
				SourceVersionFilterKind:    sourceVersionFilterKind,
				SourceVersionFilterPattern: sourceVersionFilterPattern,
				SourceVersionFilterRegex:   sourceVersionFilterRegex,
				TargetName:                 fmt.Sprintf("deps(updatecli): bump %q policy to {{ source %q}}", policyName, "version"),
				TargetKey:                  fmt.Sprintf("$.policies[%d].policy", i),
				File:                       foundUpdateComposeFile,
				ScmID:                      u.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
