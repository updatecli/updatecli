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
		"Disable changelog retrieval to avoid unnecessary requests (env: UPDATECLI_DISABLE_CHANGELOG)",
	)
}
