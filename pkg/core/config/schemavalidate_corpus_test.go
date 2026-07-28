package config

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// corpusRoots lists the manifest collections shipped with Updatecli. They are the
// closest thing available to a representative sample of the manifests found in the wild,
// so a schema change making one of them invalid is a change that would break users.
var corpusRoots = []string{
	filepath.Join("..", "..", "..", "e2e", "updatecli.d"),
	filepath.Join("..", "..", "..", "updatecli", "updatecli.d"),
}

func corpusManifests(t *testing.T) []string {
	t.Helper()

	manifests := []string{}
	tracked := trackedFiles(t)

	for _, root := range corpusRoots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				return nil
			}

			switch filepath.Ext(path) {
			case EXTENSIONYAML, EXTENSIONYML:
			default:
				return nil
			}

			// A partial is a manifest fragment, it is only meaningful once concatenated
			// with the manifest including it.
			if strings.HasPrefix(entry.Name(), "_") {
				return nil
			}

			// Only the committed corpus is checked, a manifest being worked on locally
			// is not expected to be valid yet.
			if tracked != nil && !tracked[filepath.Clean(path)] {
				return nil
			}

			manifests = append(manifests, path)

			return nil
		})
		require.NoErrorf(t, err, "walking the manifest corpus %q", root)
	}

	sort.Strings(manifests)
	require.NotEmpty(t, manifests)

	return manifests
}

// trackedFiles lists the files known to git, relative to this package. It returns nil
// when git is unavailable, in which case every manifest found is checked.
func trackedFiles(t *testing.T) map[string]bool {
	t.Helper()

	repository := filepath.Join("..", "..", "..")

	output, err := exec.Command("git", "-C", repository, "ls-files", "-z").Output()
	if err != nil {
		t.Logf("git is unavailable, checking every manifest found: %s", err)
		return nil
	}

	tracked := map[string]bool{}
	for _, name := range strings.Split(string(output), "\x00") {
		if name == "" {
			continue
		}
		tracked[filepath.Join(repository, name)] = true
	}

	return tracked
}

// TestCorpusValidatesWithoutError ensures every manifest shipped with Updatecli passes
// schema validation. It is the guard against a schema that rejects manifests Updatecli
// happily runs.
func TestCorpusValidatesWithoutError(t *testing.T) {

	options := DefaultSchemaValidationOptions()

	failures := []string{}

	for _, manifest := range corpusManifests(t) {
		content, err := os.ReadFile(manifest)
		require.NoError(t, err)

		report, err := ValidateSchema(manifest, content, options)
		if err != nil {
			// A manifest is templated before being parsed, so one using templating is
			// not necessarily valid YAML on its own. Anything else is a genuine failure.
			if bytes.Contains(content, []byte("{{")) {
				t.Logf("skipping the templated manifest %s: %s", manifest, err)
				continue
			}

			t.Errorf("%s cannot be parsed: %s", manifest, err)
			continue
		}

		for _, problem := range report.Errors() {
			failures = append(failures, problem.String())
		}
	}

	if len(failures) > 0 {
		t.Errorf("%d manifest(s) rejected by the schema:\n%s",
			len(failures), strings.Join(failures, "\n"))
	}
}

// TestCorpusReportsDeprecationsAsWarnings pins that a manifest using a deprecated but
// still accepted keyword is reported without being rejected.
func TestCorpusReportsDeprecationsAsWarnings(t *testing.T) {

	deprecated := filepath.Join("..", "..", "..", "e2e", "updatecli.d", "deprecated.d")

	content, err := os.ReadFile(filepath.Join(deprecated, "githubPullrequest.yaml"))
	require.NoError(t, err)

	report, err := ValidateSchema("githubPullrequest.yaml", content, DefaultSchemaValidationOptions())
	require.NoError(t, err)

	require.Emptyf(t, report.Errors(), "a deprecated manifest must not be rejected:\n%s", report.Error())

	warnings := []string{}
	for _, problem := range report.Warnings() {
		warnings = append(warnings, problem.String())
	}

	require.NotEmpty(t, warnings, "the deprecated 'pullrequests' keyword must be reported")
	fmt.Println(strings.Join(warnings, "\n"))
}
