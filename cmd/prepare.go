package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	prepareClean bool

	prepareCmd = &cobra.Command{
		Use:   "prepare",
		Short: "prepare run tasks needed for a run like `git clone`",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.File = cfgFile
			e.Options.ValuesFiles = valuesFiles
			e.Options.SecretsFiles = secretsFiles

			e.Options.Pipeline.Target.Clean = prepareClean

			err := run("prepare")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	prepareCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	prepareCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	prepareCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	prepareCmd.Flags().BoolVarP(&prepareClean, "clean", "", false, "Remove updatecli working directory like '--clean=true '(default: false)")
}
