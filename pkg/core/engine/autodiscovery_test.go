package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/plugins/resources/updateclihttp"
	githubscm "github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

func TestAutodiscoveryManifestFingerprintIgnoresSecrets(t *testing.T) {
	t.Parallel()

	manifestWithFirstSecrets := config.Spec{
		Name: "discovered manifest",
		Sources: map[string]source.Config{
			"version": {
				ResourceConfig: resource.ResourceConfig{
					Name: "version",
					Kind: "http",
					Spec: updateclihttp.Spec{
						Url:                  "https://user:first-secret@example.com/releases/latest",
						ReturnResponseHeader: "etag",
					},
				},
			},
		},
		SCMs: map[string]scm.Config{
			"default": {
				Kind: "github",
				Spec: githubscm.Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Token:      "first-token",
					Branch:     "main",
				},
			},
		},
	}

	manifestWithSecondSecrets := config.Spec{
		Name: "discovered manifest",
		Sources: map[string]source.Config{
			"version": {
				ResourceConfig: resource.ResourceConfig{
					Name: "version",
					Kind: "http",
					Spec: updateclihttp.Spec{
						Url:                  "https://user:second-secret@example.com/releases/latest",
						ReturnResponseHeader: "etag",
					},
				},
			},
		},
		SCMs: map[string]scm.Config{
			"default": {
				Kind: "github",
				Spec: githubscm.Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Token:      "second-token",
					Branch:     "main",
				},
			},
		},
	}

	firstFingerprint, err := autodiscoveryManifestFingerprint(manifestWithFirstSecrets, "parent-pipeline")
	require.NoError(t, err)

	secondFingerprint, err := autodiscoveryManifestFingerprint(manifestWithSecondSecrets, "parent-pipeline")
	require.NoError(t, err)

	require.Equal(t, firstFingerprint, secondFingerprint)
}

func TestAutodiscoveryManifestFingerprintTracksSanitizedIdentity(t *testing.T) {
	t.Parallel()

	manifestA := config.Spec{
		Name: "discovered manifest",
		Sources: map[string]source.Config{
			"version": {
				ResourceConfig: resource.ResourceConfig{
					Name: "version",
					Kind: "http",
					Spec: updateclihttp.Spec{
						Url:                  "https://example.com/releases/latest",
						ReturnResponseHeader: "etag",
					},
				},
			},
		},
		SCMs: map[string]scm.Config{
			"default": {
				Kind: "github",
				Spec: githubscm.Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
		},
	}

	manifestB := config.Spec{
		Name: "discovered manifest",
		Sources: map[string]source.Config{
			"version": {
				ResourceConfig: resource.ResourceConfig{
					Name: "version",
					Kind: "http",
					Spec: updateclihttp.Spec{
						Url:                  "https://example.com/releases/stable",
						ReturnResponseHeader: "etag",
					},
				},
			},
		},
		SCMs: map[string]scm.Config{
			"default": {
				Kind: "github",
				Spec: githubscm.Spec{
					Owner:      "updatecli",
					Repository: "website",
					Branch:     "main",
				},
			},
		},
	}

	fingerprintA, err := autodiscoveryManifestFingerprint(manifestA, "parent-pipeline")
	require.NoError(t, err)

	fingerprintB, err := autodiscoveryManifestFingerprint(manifestB, "parent-pipeline")
	require.NoError(t, err)

	require.NotEqual(t, fingerprintA, fingerprintB)
}
