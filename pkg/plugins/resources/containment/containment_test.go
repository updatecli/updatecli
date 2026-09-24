// Package containment holds the cross-resource regression tests for path handling.
//
// It is a separate package because it calls resources through resource.New,
// the code path a pipeline takes, which no resource package can import without
// a cycle.
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
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// escapingPaths returns the two ways a resource path can leave its working directory.
func escapingPaths(t *testing.T) map[string]string {
	t.Helper()

	return map[string]string{
		"absolute path":     filepath.Join(t.TempDir(), "pwned.txt"),
		"dot dot traversal": filepath.Join("..", "..", "pwned.txt"),
	}
}

// resourceKind is a file based resource and how to point it at a file.
type resourceKind struct {
	kind string
	spec func(filePath string) interface{}
}

// containedKinds lists the file based resources whose paths must stay within the SCM
// checkout, so a kind added later is one table entry away from being covered.
var containedKinds = []resourceKind{
	{kind: "json", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": ".version"} }},
	{kind: "toml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "version"} }},
	{kind: "yaml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "$.version"} }},
	{kind: "xml", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "path": "/version"} }},
	{kind: "file", spec: func(p string) interface{} { return map[string]interface{}{"file": p} }},
	{kind: "toolversions", spec: func(p string) interface{} { return map[string]interface{}{"file": p, "key": "golang"} }},
}

// unconfinedKinds lists the file based resources that always accepted absolute and parent
// directory paths with an SCM. Confining them would break existing manifests, which only a
// major release may do.
var unconfinedKinds = []resourceKind{
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
// file based kind.
//
// The check runs through Condition because every kind here implements it.
//
// When Updatecli works from an SCM checkout, a path that leaves it, whether absolute or
// through "..", must fail the pipeline instead of reading an arbitrary file. An attacker
// can often control that path, for example when it is templated from a source output.
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

					// A non empty source, because a few kinds validate it before
					// touching the filesystem and this test must reach the path check.
					_, _, gotErr := sut.Condition(
						context.Background(),
						"1.0.0",
						mockSCM,
						pathresolver.New(mockSCM, ""))

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

// TestConditionLocalRunAcceptsAbsolutePath checks the opposite case: without an SCM there
// is no boundary, so an absolute spec.file is an ordinary local path. The process working
// directory must not act as a boundary.
func TestConditionLocalRunAcceptsAbsolutePath(t *testing.T) {
	for _, tt := range append(append([]resourceKind{}, containedKinds...), unconfinedKinds...) {
		t.Run(tt.kind, func(t *testing.T) {
			absentFilePath := filepath.Join(t.TempDir(), "does-not-exist.txt")

			sut, err := resource.New(resource.ResourceConfig{
				Kind: tt.kind,
				Spec: tt.spec(absentFilePath),
			})
			require.NoError(t, err)

			// The file does not exist, so an error is expected. It must be a "missing
			// file" error, never a containment refusal.
			_, _, gotErr := sut.Condition(context.Background(), "1.0.0", nil, pathresolver.New(nil, ""))
			if gotErr == nil {
				return
			}

			assert.NotContains(t, gotErr.Error(), "is not allowed")
			assert.NotContains(t, gotErr.Error(), "escapes the working directory")
		})
	}
}

// TestConditionUnconfinedKindsAcceptEscapingPath checks that the kinds which never enforced
// the SCM boundary still accept absolute and parent directory paths with an SCM.
func TestConditionUnconfinedKindsAcceptEscapingPath(t *testing.T) {
	for _, tt := range unconfinedKinds {
		t.Run(tt.kind, func(t *testing.T) {
			for name, escapingPath := range escapingPaths(t) {
				t.Run(name, func(t *testing.T) {
					workingDir := filepath.Join(t.TempDir(), "checkout", "nested")
					require.NoError(t, os.MkdirAll(workingDir, 0o700))

					sut, err := resource.New(resource.ResourceConfig{
						Kind: tt.kind,
						Spec: tt.spec(escapingPath),
					})
					require.NoError(t, err)

					mockSCM := &scm.MockScm{WorkingDir: workingDir}

					// The file does not exist, so an error is expected. It must be a
					// "missing file" error, never a containment refusal.
					_, _, gotErr := sut.Condition(
						context.Background(),
						"1.0.0",
						mockSCM,
						pathresolver.New(mockSCM, ""))
					if gotErr == nil {
						return
					}

					assert.NotContains(t, gotErr.Error(), "is not allowed")
					assert.NotContains(t, gotErr.Error(), "escapes the working directory")
				})
			}
		})
	}
}
