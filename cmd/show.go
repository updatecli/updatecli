package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Print the configuration that will be executed",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Show"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			run("show")
		},
	}
)

func init() {
	showCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	showCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "Sets values file uses for templating (required {.tpl,.tmpl} config)")
}
