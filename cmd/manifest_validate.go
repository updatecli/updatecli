package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

var (
	manifestValidateDisableTemplating bool
	manifestValidateStrict            bool

	manifestValidateCmd = &cobra.Command{
		Use:   "validate",
		Short: "**Experimental** validate manifest(s) against the Updatecli schema",
		Long: `**Experimental** validate reports the manifest keys that do not match the
Updatecli schema, such as a misspelled one, which Updatecli would otherwise silently
ignore.

A deprecated keyword, or a key Updatecli cannot check reliably, is reported as a warning
and does not fail the command unless '--strict' is specified.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: To be removed once not experimental anymore
			if !cmdoptions.Experimental {
				logrus.Warningf("The 'manifest validate' command is experimental, please use the '--experimental' flag to enable it")
				os.Exit(1)
			}

			e.Options.Manifests = append(e.Options.Manifests, manifest.Manifest{
				Manifests:    manifestFiles,
				Values:       valuesFiles,
				ValuesInline: valuesInline,
				Secrets:      secretsFiles,
			})

			e.Options.Config.DisableTemplating = manifestValidateDisableTemplating

			err := run("manifest/validate")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestValidateCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	manifestValidateCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	manifestValidateCmd.Flags().StringArrayVarP(&valuesInline, "values-inline", "i", []string{}, "Sets inline values uses for templating, accepted valid json/yaml string")
	manifestValidateCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	manifestValidateCmd.Flags().BoolVar(&manifestValidateDisableTemplating, "disable-templating", false, "Disable manifest templating")
	manifestValidateCmd.Flags().BoolVar(&manifestValidateStrict, "strict", false, "Report warnings as errors")

	manifestCmd.AddCommand(manifestValidateCmd)
}
