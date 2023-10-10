package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/compose"
)

var (
	composeDiffCmd = &cobra.Command{
		Use:   "diff",
		Short: "diff show changes defined by the compose file",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO: To be removed once not experimental anymore
			if !experimental {
				logrus.Warningf("The 'compose' subcommand requires the flag experimental to work, such as:\n\t`updatecli compose diff --experimental`")
				os.Exit(1)
			}

			c, err := compose.New(composeCmdFile)
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			policies, err := c.GetPolicies(disableTLS)
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			e.Options.Manifests = append(e.Options.Manifests, policies...)

			e.Options.Pipeline.Target.Commit = false
			e.Options.Pipeline.Target.Push = false
			e.Options.Pipeline.Target.Clean = composeCmdClean
			e.Options.Pipeline.Target.DryRun = true

			err = run("compose/diff")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	composeDiffCmd.Flags().StringVarP(&composeCmdFile, "file", "f", "update-compose.yaml", "Define the update-compose file")
	composeDiffCmd.Flags().BoolVar(&composeCmdClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	composeDiffCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")

	composeCmd.AddCommand(composeDiffCmd)
}
