package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	applyCommit bool
	applyClean  bool
	applyPush   bool

	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply checks if an update is needed then apply the changes",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Config.ManifestFile = cfgFile
			e.Options.Config.ValuesFiles = valuesFiles
			e.Options.Config.SecretsFiles = secretsFiles

			e.Options.Pipeline.Commit = applyCommit
			e.Options.Pipeline.Push = applyPush
			e.Options.Pipeline.Clean = applyClean
			e.Options.Pipeline.DryRun = false

			err := run("apply")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updatecli.yaml", "Sets config file or directory. (default: './updatecli.yaml')")
	applyCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	applyCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")

	applyCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Record changes to the repository, '--commit=false' (default: true)")
	applyCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Update remote refs '--push=false' (default: true)")
	applyCmd.Flags().BoolVarP(&applyClean, "clean", "", true, "Remove updatecli working directory like '--clean=false '(default: true)")
}
