package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

func TestDisableVersionCheckFlagRegistration(t *testing.T) {
	t.Setenv(DisableVersionCheckEnvVar, "false")

	flag := rootCmd.PersistentFlags().Lookup("disable-version-check")
	if flag == nil {
		t.Fatal("flag not registered")
	}

	if flag.Usage == "" {
		t.Error("flag help text is empty")
	}

	if !strings.Contains(flag.Usage, DisableVersionCheckEnvVar) {
		t.Errorf(
			"flag help text does not mention env var %q: help text is %q",
			DisableVersionCheckEnvVar,
			flag.Usage,
		)
	}
}

func TestDisableVersionCheckFlagUsesEnvDefault(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedDef string
	}{
		{
			name:        "env_var_true",
			envValue:    "true",
			expectedDef: "true",
		},
		{
			name:        "env_var_false",
			envValue:    "false",
			expectedDef: "false",
		},
		{
			name:        "env_var_1",
			envValue:    "1",
			expectedDef: "true",
		},
		{
			name:        "env_var_0",
			envValue:    "0",
			expectedDef: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(DisableVersionCheckEnvVar, tt.envValue)

			cmd := &cobra.Command{Use: "test"}
			var val bool
			cmd.PersistentFlags().BoolVar(
				&val,
				"disable-version-check",
				getEnvBoolOrDefault(DisableVersionCheckEnvVar, false),
				"Disable version check",
			)

			flag := cmd.PersistentFlags().Lookup("disable-version-check")
			if flag == nil {
				t.Fatal("flag not registered")
			}

			if flag.DefValue != tt.expectedDef {
				t.Errorf(
					"flag default value: got %q, expected %q",
					flag.DefValue,
					tt.expectedDef,
				)
			}
		})
	}
}

func TestDisableVersionCheckFlagOverridesEnv(t *testing.T) {
	t.Setenv(DisableVersionCheckEnvVar, "true")

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	var val bool
	cmd.PersistentFlags().BoolVar(
		&val,
		"disable-version-check",
		getEnvBoolOrDefault(DisableVersionCheckEnvVar, false),
		"Disable version check",
	)

	cmd.SetArgs([]string{"--disable-version-check=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val {
		t.Error("expected flag to override env var")
	}
}

func TestSkipVersionCheckCommandsContainsExpectedCommands(t *testing.T) {
	expected := []string{"completion", "__complete", "__completeNoDesc", "docs", "man", "jsonschema"}
	for _, name := range expected {
		found := false
		for _, s := range skipVersionCheckCommands {
			if s == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("skipVersionCheckCommands missing %q", name)
		}
	}
}

func TestRunPrepareReturnsError(t *testing.T) {
	invalidManifest := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(invalidManifest, []byte("sources: [\n"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "deprecated_prepare",
			command: "prepare",
		},
		{
			name:    "pipeline_prepare",
			command: "pipeline/prepare",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := e
			t.Cleanup(func() { e = saved })

			e.Options.Manifests = []manifest.Manifest{{Manifests: []string{invalidManifest}}}

			if err := run(tt.command); err == nil {
				t.Fatal("expected run to return the prepare error, got nil")
			}
		})
	}
}
