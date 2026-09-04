// Package containment holds the cross-resource regression tests for path handling.
//
// It lives in its own package on purpose: it exercises resources through
// resource.New, which is the code path a pipeline actually takes, and which
// cannot be imported from inside any resource package without a cycle.
package containment

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// escapingPaths returns the two ways a resource path can leave its working directory.
func escapingPaths(t *testing.T) map[string]string {
	t.Helper()

	return map[string]string{
		"absolute path":     filepath.Join(t.TempDir(), "pwned.txt"),
		"dot dot traversal": filepath.Join("..", "..", "pwned.txt"),
	}
}

// containedKinds lists the file based resources and how to point them at a file, so a
// kind added later is one table entry away from being covered.
var containedKinds = []struct {
	kind string
	spec func(filePath string) interface{}
}{
	{kind: "json", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": ".version"} }},
	{kind: "toml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "version"} }},
	{kind: "yaml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "$.version"} }},
	{kind: "xml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "path": "/version"} }},
	{kind: "file", spec: func(p string) interface{} { return map[string]interface{}{"file": p} }},
	{kind: "toolversions", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "golang"} }},
	{kind: "hcl", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "path": "resource.version"} }},
	{kind: "bazelmod", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "module": "rules_go"} }},
	{kind: "systemd", spec: func(p string) interface{} {
		return map[string]interface{}{"file": p, "section": "Service", "option": "ExecStart"}
	}},
	{kind: "golang/gomod", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "module": "example.com/x"} }},
	{kind: "terraform/lock", spec: func(p string) interface{} {
		return map[string]interface{}{
			"file":      p,
			"provider":  "registry.terraform.io/hashicorp/random",
			"platforms": []string{"linux_amd64"},
		}
	}},
	{kind: "terraform/provider", spec: func(p string) interface{} {
		return map[string]interface{}{"file": p, "provider": "random", "value": "1.0.0"}
	}},
}

// TestConditionPathContainment is the regression test for GHSA-hj4x-hm4v-7wpw across every
// file based kind, not just the handful that had a test of their own.
//
// The check runs through Condition because every kind here implements it, while a few
// (terraform/lock, terraform/provider) have no source stage.
//
// When Updatecli works from an SCM checkout, a path that leaves it — through an absolute
// path or through ".." — must fail the pipeline rather than read an arbitrary file. The
// path is often (attacker) controlled, e.g. templated from a source output.
func TestConditionPathContainment(t *testing.T) {
	for _, tt := range containedKinds {
		t.Run(tt.kind, func(t *testing.T) {
			for name, escapingPath := range escapingPaths(t) {
				t.Run(name, func(t *testing.T) {
					// Nested two levels down so that "../.." still lands inside the
					// temporary directory rather than in a shared location.
					workingDir := filepath.Join(t.TempDir(), "checkout", "nested")
					require.NoError(t, os.MkdirAll(workingDir, 0o700))

					sut, err := resource.New(resource.ResourceConfig{
						Kind: tt.kind,
						Spec: tt.spec(escapingPath),
					})
					require.NoError(t, err)

					mockSCM := &scm.MockScm{WorkingDir: workingDir}

					// A non empty source: a few kinds validate it before touching
					// the filesystem, and the point here is to reach the path.
					_, _, gotErr := sut.Condition(
						context.Background(),
						"1.0.0",
						mockSCM,
						utils.NewResolver(mockSCM, ""))

					require.Error(t, gotErr, "a path escaping the working directory must be rejected")
					assert.True(t,
						strings.Contains(gotErr.Error(), "is not allowed") ||
							strings.Contains(gotErr.Error(), "escapes the working directory"),
						"expected a containment error, got: %s", gotErr)
				})
			}
		})
	}
}

// TestConditionLocalRunAcceptsAbsolutePath is the counterpart: without an SCM there is no
// boundary, so an absolute spec.file is an ordinary local path.
//
// Every kind here used to accept one; json, toml, csv and toolversions started rejecting
// it when containment began treating the process working directory as a boundary.
func TestConditionLocalRunAcceptsAbsolutePath(t *testing.T) {
	for _, tt := range containedKinds {
		t.Run(tt.kind, func(t *testing.T) {
			absentFilePath := filepath.Join(t.TempDir(), "does-not-exist.txt")

			sut, err := resource.New(resource.ResourceConfig{
				Kind: tt.kind,
				Spec: tt.spec(absentFilePath),
			})
			require.NoError(t, err)

			// The file does not exist, so an error is expected either way. What matters
			// is that it is a "missing file" error and never a containment refusal.
			_, _, gotErr := sut.Condition(context.Background(), "1.0.0", nil, utils.NewResolver(nil, ""))
			if gotErr == nil {
				return
			}

			assert.NotContains(t, gotErr.Error(), "is not allowed")
			assert.NotContains(t, gotErr.Error(), "escapes the working directory")
		})
	}
}
