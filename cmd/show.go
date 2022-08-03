package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	showClean          bool
	showDisablePrepare bool

	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Print the configuration that will be executed",
		Run: func(cmd *cobra.Command, args []string) {

			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.ValuesFiles = valuesFiles
			e.Options.Config.SecretsFiles = secretsFiles
			e.Options.Pipeline.AutoDiscovery.Disabled = autoDiscoveryDisabled
			e.Options.Pipeline.Target.Clean = showClean

			err := run("show")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	showCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Sets config file or directory.")
	showCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	showCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	showCmd.Flags().BoolVar(&autoDiscoveryDisabled, "disable-local-autodiscovery", false, "Discovery automatically available Updatecli manifest")
	showCmd.Flags().BoolVar(&showClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	showCmd.Flags().BoolVar(&showDisablePrepare, "disable-prepare", false, "--disable-prepare skip the Updatecli 'prepare' stage'--disable-prepare=true'")
}
