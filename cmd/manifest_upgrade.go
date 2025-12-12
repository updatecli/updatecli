package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	manifestUpgradeInPlace bool

	manifestUpgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade executes manifest upgrade task",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Manifests = append(e.Options.Manifests, manifest.Manifest{
				Manifests: manifestFiles,
			})

			e.Options.Config.DisableTemplating = true

			err := run("manifest/upgrade")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestUpgradeCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	manifestUpgradeCmd.Flags().BoolVarP(&manifestUpgradeInPlace, "in-place", "i", false, "Write updated Updatecli manifest back to the same file instead of stdout")
	manifestUpgradeCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their IDs, accepted a comma separated list")

	manifestCmd.AddCommand(manifestUpgradeCmd)
}
