package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	oAuthClientID   string
	oAuthAuthDomain string
	reportAPI       string

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "[Experimental] login authenticates with the Updatecli service.",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO: To be removed once not experimental anymore
			if !experimental {
				logrus.Warningf("The 'login' feature requires the flag experimental to work, such as:\n\t`updatecli login --experimental`")
				os.Exit(1)
			}

			switch len(args) {
			case 0:
				logrus.Errorf("missing URL to login to")
				os.Exit(1)
			case 1:
				reportAPI = args[0]
			default:
				logrus.Errorf("can only login to one URL at a time")
				os.Exit(1)
			}

			err := run("login")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	loginCmd.Flags().StringVar(&oAuthClientID, "oauth-clientId", "", "oauth-clientId defines the Oauth client ID")
	loginCmd.Flags().StringVar(&oAuthAuthDomain, "oauth-authDomain", "", "oauth-authDomain defines the Oauth authentication URL")
}
