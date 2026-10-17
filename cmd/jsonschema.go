package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/log"
)

const (
	// jsonschemaCmdName is the name of the jsonschema command, also used to skip
	// the version check so that generated content stays clean.
	jsonschemaCmdName = "jsonschema"
)

var (
	jsonschemaCmd = &cobra.Command{
		Use:   jsonschemaCmdName,
		Short: "**Experimental** Export JsonSchema to file",
		Run: func(cmd *cobra.Command, args []string) {
			err := run(jsonschemaCmdName)
			if err != nil {
				log.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
	jsonschemaDirectory string
	jsonschemaBaseID    string
)

func init() {
	jsonschemaCmd.Flags().StringVarP(&jsonschemaDirectory, "directory", "d", "./", "Export schema to directory")
	jsonschemaCmd.Flags().StringVarP(&jsonschemaBaseID, "baseid", "b", "https://www.updatecli.io/schema/latest", "Define schema baseid")
}
