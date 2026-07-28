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
