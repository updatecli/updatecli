package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	jsonschemaCmd = &cobra.Command{
		Use:   "jsonschema",
		Short: "**Experimental** Export JsonSchema to file",
		Run: func(cmd *cobra.Command, args []string) {

			err := run("jsonschema")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
	jsonschemaDirectory string
	jsonschemaBaseID    string
)

func init() {
	jsonschemaCmd.Flags().StringVarP(&jsonschemaDirectory, "directory", "d", "./", "Export schema to directory")
	jsonschemaCmd.Flags().StringVarP(&jsonschemaBaseID, "baseID", "b", "https://www.updatecli.io/latest/schema", "Define schema baseID")
}
