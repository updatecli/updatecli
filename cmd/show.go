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
			run(cfgFile, "show")
		},
	}
)

func init() {
	showCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "config file (default is ./updateCli.yaml)")
	showCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "values file uses for templating (required {.tpl,.tmpl} as config)")
}
