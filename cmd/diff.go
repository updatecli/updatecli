package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "diff shows if targets update are needed",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Apply"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			e.Options.Target.Commit = false
			e.Options.Target.Push = false
			e.Options.Target.Clean = false

			run(
				"apply",
			)
		},
	}
)

func init() {
	diffCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "config file (default is ./updateCli.yaml)")
	diffCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "values file use for templating (required {.tpl,.tmpl} config)")
}
