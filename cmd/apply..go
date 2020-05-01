package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply will execute updateCli and update target if needed",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Apply"))
			run(cfgFile, "apply")
		},
	}
)

func init() {
	applyCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "config file (default is ./updateCli.yaml)")
	applyCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "values file use for templating (required {.tpl,.tmpl} config)")
}
