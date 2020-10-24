package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	diffClean bool

	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "diff shows changes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Diff"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			e.Options.Target.Commit = false
			e.Options.Target.Push = false
			e.Options.Target.Clean = diffClean
			e.Options.Target.DryRun = true

			run(
				"diff",
			)
		},
	}
)

func init() {
	diffCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	diffCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "Sets values file uses for templating (required {.tpl,.tmpl} config)")
	diffCmd.Flags().BoolVarP(&diffClean, "clean", "", true, "Remove updatecli working directory like '--clean=false '(default: true)")
}
