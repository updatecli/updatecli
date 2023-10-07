package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/compose"
)

var (
	composeApplyCommit bool
	composeApplyClean  bool
	composeApplyPush   bool

	composeApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply checks and apply changes defined by the compose file",
		Run: func(cmd *cobra.Command, args []string) {

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

			e.Options.Pipeline.Target.Commit = composeApplyCommit
			e.Options.Pipeline.Target.Push = composeApplyPush
			e.Options.Pipeline.Target.Clean = composeApplyClean
			e.Options.Pipeline.Target.DryRun = false

			err = run("compose/apply")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	composeApplyCmd.Flags().StringVarP(&composeCmdFile, "file", "f", "update-compose.yaml", "Define the update-compose file")

	composeApplyCmd.Flags().BoolVarP(&composeApplyCommit, "commit", "", true, "Record changes to the repository, '--commit=false'")
	composeApplyCmd.Flags().BoolVarP(&composeApplyPush, "push", "", true, "Update remote refs '--push=false'")
	composeApplyCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	composeApplyCmd.Flags().BoolVar(&composeApplyClean, "clean", false, "Remove updatecli working directory like '--clean=true'")

	composeCmd.AddCommand(composeApplyCmd)
}
