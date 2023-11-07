package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	udashConfigCmd = &cobra.Command{
		Use:     "config",
		Short:   "[Experimental] config shows the config file path",
		Example: "updatecli udash config",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO: To be removed once not experimental anymore
			if !experimental {
				logrus.Warningf("The 'config' feature requires the flag experimental to work, such as:\n\t`updatecli udash config --experimental`")
				os.Exit(1)
			}

			err := run("udash/config")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	udashCmd.AddCommand(udashConfigCmd)
}
