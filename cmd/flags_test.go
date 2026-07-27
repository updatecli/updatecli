package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddDisableChangelogFlagRegistration(t *testing.T) {
	t.Setenv(DisableChangelogEnvVar, "false")

	cmd := &cobra.Command{
		Use: "test",
	}

	var disableChangelog bool
	addDisableChangelogFlag(cmd, &disableChangelog)

	// Check that flag exists
	flag := cmd.Flags().Lookup("disable-changelog")
	if flag == nil {
		t.Fatal("flag not registered")
	}

	// Check flag has help text
	if flag.Usage == "" {
		t.Error("flag help text is empty")
	}

	// Check that env var name appears in help text
	if !strings.Contains(flag.Usage, DisableChangelogEnvVar) {
		t.Errorf(
			"flag help text does not mention env var %q: help text is %q",
			DisableChangelogEnvVar,
			flag.Usage,
		)
	}

	// Check default value is "false" when env var not set
	if flag.DefValue != "false" {
		t.Errorf(
			"flag default value when no env var: got %q, expected %q",
			flag.DefValue,
			"false",
		)
	}
}

func TestAddExportReportToYAMLFlagRegistration(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	var exportReportToYAML bool
	addExportReportToYAMLFlag(cmd, &exportReportToYAML)

	// Check that flag exists
	flag := cmd.Flags().Lookup("export-report-to-yaml")
	if flag == nil {
		t.Fatal("flag not registered")
	}

	// Check flag has help text
	if flag.Usage == "" {
		t.Error("flag help text is empty")
	}

	// Exporting writes files to disk, so it must be opt-in
	if flag.DefValue != "false" {
		t.Errorf(
			"flag default value: got %q, expected %q",
			flag.DefValue,
			"false",
		)
	}
}

func TestAddDisableUdashReportFlagRegistration(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	var disableUdashReport bool
	addDisableUdashReportFlag(cmd, &disableUdashReport)

	// Check that flag exists
	flag := cmd.Flags().Lookup("disable-udash-report")
	if flag == nil {
		t.Fatal("flag not registered")
	}

	// Check flag has help text
	if flag.Usage == "" {
		t.Error("flag help text is empty")
	}

	// Publishing only happens when a Udash endpoint is configured, so it is opt-out
	if flag.DefValue != "false" {
		t.Errorf(
			"flag default value: got %q, expected %q",
			flag.DefValue,
			"false",
		)
	}
}

func TestAddDisableChangelogFlagUsesEnvDefault(t *testing.T) {
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
			t.Setenv(DisableChangelogEnvVar, tt.envValue)

			cmd := &cobra.Command{
				Use: "test",
			}

			var disableChangelog bool
			addDisableChangelogFlag(cmd, &disableChangelog)

			flag := cmd.Flags().Lookup("disable-changelog")
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

func TestDisableChangelogFlagOverridesEnv(t *testing.T) {
	t.Setenv(DisableChangelogEnvVar, "true")

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	var disableChangelog bool
	addDisableChangelogFlag(cmd, &disableChangelog)

	cmd.SetArgs([]string{"--disable-changelog=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if disableChangelog {
		t.Error("expected flag to override env var")
	}
}
