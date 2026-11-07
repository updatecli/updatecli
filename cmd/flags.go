package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

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

// udashAPIURL selects which of the credentials stored by `updatecli udash login`
// to publish with, by API URL. It is empty when the default one should be used.
var udashAPIURL string

// addUdashAPIURLFlag registers the shared --udash-api-url flag on the provided command.
func addUdashAPIURLFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&udashAPIURL,
		"udash-api-url",
		"",
		"Udash API URL selecting which service stored by udash login receives pipeline reports",
	)
}

// addRelativePathsFlag registers the shared --relative-paths flag on the provided command,
// using the value from UPDATECLI_RELATIVE_PATHS as the default when the flag is not
// explicitly passed. A manifest declaring "relativepaths" overrides it.
func addRelativePathsFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(
		dest,
		"relative-paths",
		getEnvStringOrDefault(RelativePathsEnvVar, string(config.RelativePathBaseWorkingDirectory)),
		"What the relative paths of a manifest resolve against: \"workingdirectory\" or \"manifest\" (env: "+RelativePathsEnvVar+")",
	)
}

// devModePolicy tells the user what dev mode leaves out, as enforced by
// engine.enforceLocalScm, the pipeline scm, target and action setup and the gitbranch and
// gittag targets.
const devModePolicy = "only targets changing local files are applied: those of other repositories, which are cloned read only, " +
	"and those that may change a remote, such as git branches, tags or releases, run in dry run. " +
	"Nothing is committed or pushed, actions are skipped and no report is published to Udash"

// devModeUsage describes --dev in the command help.
const devModeUsage = "[Experimental, requires --experimental] Dev mode, for writing and testing a manifest: " +
	"every scm describing the git repository of the current directory, as one of its git remotes tells, runs against it instead of a fresh clone, " +
	"so apply changes your local files in place. Run diff first. " +
	"Not meant for CI. In dev mode, " + devModePolicy + ". " +
	"Its behavior may change in any release"

// addDevFlag registers the shared --dev flag on the provided command.
func addDevFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&devMode, "dev", false, devModeUsage)
}

// errDevModeNeedsExperimental is returned when --dev is used without --experimental.
var errDevModeNeedsExperimental = errors.New("--dev is experimental: add --experimental to use it")

// applyDevFlag points every scm at the git repository of the current directory when --dev
// is set. The pipeline options then enforce what a local checkout allows.
func applyDevFlag() error {
	if !devMode {
		return nil
	}

	if !cmdoptions.Experimental {
		return errDevModeNeedsExperimental
	}

	directory, err := scm.LocalDirectory(".")
	if err != nil {
		return fmt.Errorf("--dev: %w", err)
	}

	log.Infof("Dev mode: running scms against the local checkout %q: %s", directory, devModePolicy)

	e.Options.Pipeline.LocalScmDirectory = directory

	return nil
}
