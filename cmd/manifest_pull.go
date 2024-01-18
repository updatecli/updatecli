package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	manifestPullPolicyReference string

	manifestPullCmd = &cobra.Command{
		Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		Use:   "pull NAME[:TAG|@DIGEST]",
		Short: "pull manifest(s) from an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			e.Options.Pipeline.Target.Clean = manifestShowClean
			manifestPullPolicyReference = args[0]

			err := run("manifest/pull")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestPullCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	manifestCmd.AddCommand(manifestPullCmd)
}
