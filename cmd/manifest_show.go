package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	manifestShowClean             bool
	manifestShowDisablePrepare    bool
	manifestShowDisableTemplating bool

	manifestShowCmd = &cobra.Command{
		Use:   "show",
		Short: "show manifest(s) which will be executed",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.ValuesFiles = valuesFiles
			e.Options.Config.SecretsFiles = secretsFiles
			e.Options.Pipeline.AutoDiscovery.Disabled = autoDiscoveryDisabled
			e.Options.Pipeline.Target.Clean = manifestShowClean

			e.Options.Config.DisableTemplating = manifestShowDisableTemplating

			err := run("manifest/show")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestShowCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Sets config file or directory.")
	manifestShowCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	manifestShowCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	manifestShowCmd.Flags().BoolVar(&autoDiscoveryDisabled, "disable-local-autodiscovery", false, "Discovery automatically available Updatecli manifest")
	manifestShowCmd.Flags().BoolVar(&manifestShowClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	manifestShowCmd.Flags().BoolVar(&manifestShowDisablePrepare, "disable-prepare", false, "--disable-prepare skip the Updatecli 'prepare' stage")
	manifestShowCmd.Flags().BoolVar(&manifestShowDisableTemplating, "disable-templating", false, "Disable manifest templating")

	manifestCmd.AddCommand(manifestShowCmd)
}
