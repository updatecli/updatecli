package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	prepareClean bool

	prepareCmd = &cobra.Command{
		Use:   "prepare",
		Short: "prepare run tasks needed for a run like `git clone`",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Prepare"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			e.Options.Target.Clean = prepareClean

			run(
				"prepare",
			)
		},
	}
)

func init() {
	prepareCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	prepareCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "Sets values file uses for templating (required {.tpl,.tmpl} config)")
	prepareCmd.Flags().BoolVarP(&prepareClean, "clean", "", false, "Remove updatecli working directory like '--clean=true '(default: false)")
}
