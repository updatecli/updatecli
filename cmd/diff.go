package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	diffClean bool

	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "diff shows changes",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.ValuesFiles = valuesFiles
			e.Options.Config.SecretsFiles = secretsFiles

			e.Options.Pipeline.AutoDiscovery.Disabled = autoDiscoveryDisabled

			e.Options.Pipeline.Target.Commit = false
			e.Options.Pipeline.Target.Push = false
			e.Options.Pipeline.Target.Clean = diffClean
			e.Options.Pipeline.Target.DryRun = true

			err := run("diff")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	diffCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Sets config file or directory.")
	diffCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	diffCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	diffCmd.Flags().BoolVarP(&diffClean, "clean", "", true, "Remove updatecli working directory like '--clean=false '(default: true)")
	diffCmd.Flags().BoolVar(&autoDiscoveryDisabled, "autodiscovery-disable", false, "Discovery automatically available Updatecli manifest")
}
