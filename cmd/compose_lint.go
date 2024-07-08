package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	composeLintPolicyRootDir string
	composeLintCmd           = &cobra.Command{
		Use:   "lint <path>",
		Short: "lint compose policy",
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {

			composeLintPolicyRootDir = "updatecli/policies"
			if len(args) == 1 {
				composeLintPolicyRootDir = args[0]
			}

			err := run("compose/lint")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

		},
	}
)

func init() {
	composeLintCmd.Flags().StringVarP(&policyFolder, "policy", "p", "updatecli/policies", "Define the policy folder")

	composeCmd.AddCommand(composeLintCmd)
}
