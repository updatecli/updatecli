package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	validateJsonSchema string
	validateConfig     string

	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "validate checks updatecli configuration syntax",
		Run: func(cmd *cobra.Command, args []string) {

			err := run("validate")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	validateCmd.Flags().StringVarP(&validateConfig, "config", "c", "./updatecli.yaml", "Sets config file or directory. (default: './updatecli.yaml')")
	validateCmd.Flags().StringVarP(&validateJsonSchema, "schema", "s", "./config.json", "Sets Json schema (default: 'config.json')")
}
