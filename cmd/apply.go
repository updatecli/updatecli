package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	applyCommit bool
	applyClean  bool
	applyPush   bool

	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply checks if an updated is needed then apply the changes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Apply"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			e.Options.Target.Commit = applyCommit
			e.Options.Target.Push = applyPush
			e.Options.Target.Clean = applyClean

			run(
				"apply",
			)
		},
	}
)

func init() {
	applyCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "config file (default is ./updateCli.yaml)")
	applyCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "values file use for templating (required {.tpl,.tmpl} config)")

	showCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Commit")
	showCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Push changes")
	showCmd.Flags().BoolVarP(&applyClean, "clean", "", true, "clean working directory")
}
