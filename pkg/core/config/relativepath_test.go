package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// TestConfig_BaseDir covers where the relative paths of a manifest resolve from, which is
// the whole point of the "options.relativepaths" key.
func TestConfig_BaseDir(t *testing.T) {
	testdata := []struct {
		name            string
		filename        string
		generatedFrom   string
		relativePaths   RelativePathBase
		expectedBaseDir string
	}{
		{
			name:            "an undeclared base keeps the historical working directory behavior",
			filename:        filepath.Join("updatecli.d", "nodejs.yaml"),
			expectedBaseDir: "",
		},
		{
			name:            "an explicit working directory base resolves from the process directory",
			filename:        filepath.Join("updatecli.d", "nodejs.yaml"),
			relativePaths:   RelativePathBaseWorkingDirectory,
			expectedBaseDir: "",
		},
		{
			name:            "a manifest base resolves from the directory holding the manifest",
			filename:        filepath.Join("updatecli.d", "nodejs.yaml"),
			relativePaths:   RelativePathBaseManifest,
			expectedBaseDir: "updatecli.d",
		},
		{
			name:            "a manifest at the root of the working directory resolves from it",
			filename:        "updatecli.yaml",
			relativePaths:   RelativePathBaseManifest,
			expectedBaseDir: ".",
		},
		{
			// An autodiscovery generated manifest has no file on disk, so there is
			// nothing to be relative to: the crawler pins the directory instead.
			name:            "a generated manifest resolves from the directory it was crawled from",
			generatedFrom:   filepath.Join("some", "checkout"),
			relativePaths:   RelativePathBaseManifest,
			expectedBaseDir: filepath.Join("some", "checkout"),
		},
		{
			name:            "a generated manifest without a crawl directory resolves from the process directory",
			relativePaths:   RelativePathBaseManifest,
			expectedBaseDir: "",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				filename: tt.filename,
				Spec:     Spec{Options: ManifestOptions{RelativePaths: tt.relativePaths}},
			}
			if tt.generatedFrom != "" {
				config.SetBaseDir(tt.generatedFrom)
			}

			assert.Equal(t, tt.expectedBaseDir, config.BaseDir())
		})
	}
}

// TestManifestOptions_Merge pins the rule that a manifest overrides the command line
// setting by setting, rather than the whole section winning as a block.
//
// It matters as soon as a second setting exists: a manifest naming only "relativepaths"
// must still inherit the others, and a setting meant to prevent something would have to
// keep the safer of the two values instead.
func TestManifestOptions_Merge(t *testing.T) {
	testdata := []struct {
		name           string
		manifest       ManifestOptions
		defaults       ManifestOptions
		expectedResult ManifestOptions
	}{
		{
			name:           "an undefined setting takes the default",
			manifest:       ManifestOptions{},
			defaults:       ManifestOptions{RelativePaths: RelativePathBaseManifest},
			expectedResult: ManifestOptions{RelativePaths: RelativePathBaseManifest},
		},
		{
			name:           "the manifest overrides the default",
			manifest:       ManifestOptions{RelativePaths: RelativePathBaseWorkingDirectory},
			defaults:       ManifestOptions{RelativePaths: RelativePathBaseManifest},
			expectedResult: ManifestOptions{RelativePaths: RelativePathBaseWorkingDirectory},
		},
		{
			name:           "no default leaves the manifest untouched",
			manifest:       ManifestOptions{RelativePaths: RelativePathBaseManifest},
			defaults:       ManifestOptions{},
			expectedResult: ManifestOptions{RelativePaths: RelativePathBaseManifest},
		},
		{
			name:           "neither side defining anything stays undefined",
			manifest:       ManifestOptions{},
			defaults:       ManifestOptions{},
			expectedResult: ManifestOptions{},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.manifest
			gotResult.Merge(tt.defaults)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

// TestSpec_MarshalOmitsEmptyOptions guards the manifests already on disk.
//
// "manifest upgrade" and SaveOnDisk write a Spec back out, so an "options" section that
// does not omit its zero value would be injected into every manifest Updatecli rewrites.
func TestSpec_MarshalOmitsEmptyOptions(t *testing.T) {
	gotResult, err := yaml.Marshal(Spec{Name: "a pipeline"})
	require.NoError(t, err)

	assert.NotContains(t, string(gotResult), "options")

	// The counterpart: a section that is set must survive the round trip.
	gotResult, err = yaml.Marshal(Spec{
		Name:    "a pipeline",
		Options: ManifestOptions{RelativePaths: RelativePathBaseManifest},
	})
	require.NoError(t, err)

	assert.Contains(t, string(gotResult), "options:")
	assert.Contains(t, string(gotResult), "relativepaths: manifest")
}

// TestManifestOptions_Validate ensures a misspelled setting is reported rather than
// silently falling back to the default.
func TestManifestOptions_Validate(t *testing.T) {
	testdata := []struct {
		relativePaths RelativePathBase
		wantErr       bool
	}{
		{relativePaths: RelativePathBaseUndefined},
		{relativePaths: RelativePathBaseWorkingDirectory},
		{relativePaths: RelativePathBaseManifest},
		{relativePaths: "manifests", wantErr: true},
		{relativePaths: "Manifest", wantErr: true},
	}

	for _, tt := range testdata {
		t.Run(string(tt.relativePaths), func(t *testing.T) {
			gotErr := ManifestOptions{RelativePaths: tt.relativePaths}.Validate()

			if tt.wantErr {
				require.Error(t, gotErr)
				assert.Contains(t, gotErr.Error(), "relativepaths")
				return
			}

			require.NoError(t, gotErr)
		})
	}
}

// TestConfig_ValidateOptions checks that a bad setting surfaces through the manifest
// validation, pointing at the section it came from.
func TestConfig_ValidateOptions(t *testing.T) {
	config := Config{Spec: Spec{
		Name:    "test",
		Options: ManifestOptions{RelativePaths: "manifests"},
	}}

	gotErr := config.Validate()

	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "options validation error")
}

// TestNew_ManifestOptionsPrecedence checks that a manifest declaring its own settings wins
// over the ones configured on the command line.
func TestNew_ManifestOptionsPrecedence(t *testing.T) {
	testdata := []struct {
		name           string
		manifest       string
		globalDefault  ManifestOptions
		expectedResult RelativePathBase
	}{
		{
			name:           "the global default applies when the manifest is silent",
			manifest:       "testdata/updatecli.d/jenkins.yaml",
			globalDefault:  ManifestOptions{RelativePaths: RelativePathBaseManifest},
			expectedResult: RelativePathBaseManifest,
		},
		{
			name:           "no global default keeps the manifest silent",
			manifest:       "testdata/updatecli.d/jenkins.yaml",
			expectedResult: RelativePathBaseUndefined,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := New(Option{ManifestFile: tt.manifest}, tt.globalDefault, nil, nil)
			require.NoError(t, err)
			require.Len(t, configs, 1)

			assert.Equal(t, tt.expectedResult, configs[0].Spec.Options.RelativePaths)
		})
	}
}
