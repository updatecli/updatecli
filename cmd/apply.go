package cmd

import (
	"os"
	"strings"

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
			logrus.Infof("\n%s\n", strings.ToTitle("Apply"))

			e.Options.File = cfgFile
			e.Options.ValuesFiles = valuesFiles
			e.Options.SecretsFiles = secretsFiles

			e.Options.Target.Commit = applyCommit
			e.Options.Target.Push = applyPush
			e.Options.Target.Clean = applyClean
			e.Options.Target.DryRun = false

			err := run("apply")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	applyCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	applyCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")

	applyCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Record changes to the repository, '--commit=false' (default: true)")
	applyCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Update remote refs '--push=false' (default: true)")
	applyCmd.Flags().BoolVarP(&applyClean, "clean", "", true, "Remove updatecli working directory like '--clean=false '(default: true)")
}
