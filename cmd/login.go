package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	oAuthClientID string
	oAuthIssuer   string
	oAuthAudience string
	endpointURL   string

	loginCmd = &cobra.Command{
		Use:     "login url",
		Short:   "[Experimental] login authenticates with the Updatecli service.",
		Example: "updatecli login app.updatecli.io",
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
				endpointURL = args[0]
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
	loginCmd.Flags().StringVar(&oAuthIssuer, "oauth-issuer", "", "oauth-issuer defines the Oauth authentication URL")
	loginCmd.Flags().StringVar(&oAuthAudience, "oauth-audience", "", "oauth-audience defines the Oauth audience URL")
}
