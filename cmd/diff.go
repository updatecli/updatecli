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

			e.Options.Pipeline.Commit = false
			e.Options.Pipeline.Push = false
			e.Options.Pipeline.Clean = diffClean
			e.Options.Pipeline.DryRun = true

			err := run("diff")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	diffCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updatecli.yaml", "Sets config file or directory. (default: './updatecli.yaml')")
	diffCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	diffCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	diffCmd.Flags().BoolVarP(&diffClean, "clean", "", true, "Remove updatecli working directory like '--clean=false '(default: true)")
}
