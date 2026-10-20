package cmd

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/spf13/cobra"
)

var (
	udashLogoutCmd = &cobra.Command{
		Use:     "logout url",
		Short:   "[Experimental] logout from an Udash service.",
		Example: "updatecli udash logout app.updatecli.io",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO: To be removed once not experimental anymore
			if !experimental {
				log.Warningf("The 'logout' feature requires the flag experimental to work, such as:\n\t`updatecli udash logout --experimental https://app.updatecli.io`")
				os.Exit(1)
			}

			switch len(args) {
			case 0:
				log.Errorf("missing URL to logout from")
				os.Exit(1)
			case 1:
				udashEndpointURL = args[0]
			default:
				log.Errorf("can only logout from one URL at a time")
				os.Exit(1)
			}

			err := run("udash/logout")
			if err != nil {
				log.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	udashCmd.AddCommand(udashLogoutCmd)
}
