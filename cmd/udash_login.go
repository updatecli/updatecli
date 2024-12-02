package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	udashOAuthAccessToken string
	udashOAuthClientID    string
	udashOAuthIssuer      string
	udashOAuthAudience    string
	udashEndpointURL      string
	udashEndpointAPIURL   string

	udashLoginCmd = &cobra.Command{
		Use:     "login url",
		Short:   "[Experimental] login authenticates with the Udash.",
		Example: "updatecli udash login app.updatecli.io",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO: To be removed once not experimental anymore
			if !experimental {
				logrus.Warningf("The 'login' feature requires the flag experimental to work, such as:\n\t`updatecli udash login --experimental https://app.updatecli.io`")
				os.Exit(1)
			}

			switch len(args) {
			case 0:
				logrus.Errorf("missing URL to login to")
				os.Exit(1)
			case 1:
				udashEndpointURL = args[0]
			default:
				logrus.Errorf("can only login to one URL at a time")
				os.Exit(1)
			}

			err := run("udash/login")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	udashLoginCmd.Flags().StringVar(&udashOAuthClientID, "oauth-clientId", "", "oauth-clientId defines the Oauth client ID")
	udashLoginCmd.Flags().StringVar(&udashOAuthIssuer, "oauth-issuer", "", "oauth-issuer defines the Oauth authentication URL")
	udashLoginCmd.Flags().StringVar(&udashOAuthAudience, "oauth-audience", "", "oauth-audience defines the Oauth audience URL")
	udashLoginCmd.Flags().StringVar(&udashOAuthAccessToken, "oauth-access-token", "", "oauth-access-token defines the Oauth access token")
	udashLoginCmd.Flags().StringVar(&udashEndpointAPIURL, "api-url", "", "api-url defines the udash API URL")

	udashCmd.AddCommand(udashLoginCmd)
}
