package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Print the configuration that will be executed",
		Run: func(cmd *cobra.Command, args []string) {

			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.ValuesFiles = valuesFiles
			e.Options.Config.SecretsFiles = secretsFiles
			e.Options.Pipeline.AutoDiscovery.Disabled = autoDiscoveryDisabled

			err := run("show")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	showCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updatecli.yaml", "Sets config file or directory. (default: './updatecli.yaml')")
	showCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	showCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	diffCmd.Flags().BoolVar(&autoDiscoveryDisabled, "autodiscovery-disable", false, "Discovery automatically available Updatecli manifest")
}
