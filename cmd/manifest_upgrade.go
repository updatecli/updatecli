package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	manifestCmd = &cobra.Command{
		Use:   "manifest",
		Short: "manifest executes specific manifest task such as upgrade",
	}

	manifestUpgradeInPlace bool

	manifestUpgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade executes manifest upgrade task",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.DisableTemplating = true

			err := run("manifest/upgrade")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestUpgradeCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	manifestUpgradeCmd.Flags().BoolVarP(&manifestUpgradeInPlace, "in-place", "i", false, "Write updated Updatecli manifest back to the same file instead of stdout")

	manifestCmd.AddCommand(manifestUpgradeCmd)
}
