package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	jsonschemaCmd = &cobra.Command{
		Use:   "jsonschema",
		Short: "Export Json Schema to file",
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
	jsonschemaCmd.Flags().StringVarP(&jsonschemaDirectory, "directory", "d", "./schema", "Export schema to directory (default: './schema')")
	jsonschemaCmd.Flags().StringVarP(&jsonschemaBaseID, "baseID", "b", "https://www.updatecli.io/schema", "Define schema baseID")
}
