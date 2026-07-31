package cmd

import "github.com/spf13/cobra"

// addDisableChangelogFlag registers the shared --disable-changelog flag on the
// provided command, using the value from UPDATECLI_DISABLE_CHANGELOG as the
// default when the flag is not explicitly passed.
func addDisableChangelogFlag(cmd *cobra.Command, dest *bool) {
	cmd.Flags().BoolVar(
		dest,
		"disable-changelog",
		getEnvBoolOrDefault(DisableChangelogEnvVar, false),
		"Disable changelog retrieval to avoid unnecessary requests (env: "+DisableChangelogEnvVar+")",
	)
}

// addValidateSchemaFlag registers the shared --validate-schema flag on the provided
// command, using the value from UPDATECLI_VALIDATE_SCHEMA as the default when the flag is
// not explicitly passed.
func addValidateSchemaFlag(cmd *cobra.Command, dest *bool) {
	cmd.Flags().BoolVar(
		dest,
		"validate-schema",
		getEnvBoolOrDefault(ValidateSchemaEnvVar, false),
		"Report manifest keys not matching the Updatecli schema as warnings (env: "+ValidateSchemaEnvVar+")",
	)
}

// addExportReportToYAMLFlag registers the shared --export-report-to-yaml flag on
// the provided command. Exporting is opt-in as it writes files to disk.
func addExportReportToYAMLFlag(cmd *cobra.Command, dest *bool) {
	cmd.Flags().BoolVar(
		dest,
		"export-report-to-yaml",
		false,
		"Export pipeline reports to YAML files",
	)
}

// addDisableUdashReportFlag registers the shared --disable-udash-report flag on
// the provided command. Publishing is opt-out as it only happens when a Udash
// endpoint is already configured.
func addDisableUdashReportFlag(cmd *cobra.Command, dest *bool) {
	cmd.Flags().BoolVar(
		dest,
		"disable-udash-report",
		false,
		"Disable publishing pipeline reports to Udash",
	)
}
