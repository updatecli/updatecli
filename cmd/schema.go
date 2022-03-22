package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	schemaCmd = &cobra.Command{
		Use:   "shema",
		Short: "Export Json Schema to file",
		Run: func(cmd *cobra.Command, args []string) {

			err := run("schema")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
	schemaDirectory string
	schemaBaseID    string
)

func init() {
	schemaCmd.Flags().StringVarP(&schemaDirectory, "directory", "d", "./schema", "Export schema to directory (default: './schema')")
	schemaCmd.Flags().StringVarP(&schemaBaseID, "baseID", "b", "https://www.updatecli.io/schema", "Define schema baseID")
}
