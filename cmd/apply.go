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

			e.Options.Pipeline.Target.Commit = applyCommit
			e.Options.Pipeline.Target.Push = applyPush
			e.Options.Pipeline.Target.Clean = applyClean
			e.Options.Pipeline.Target.DryRun = false

			err := run("apply")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	applyCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	applyCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")

	applyCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Record changes to the repository, '--commit=false' (default: true)")
	applyCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Update remote refs '--push=false' (default: true)")

	applyCmd.Flags().BoolVar(&applyClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
}
